package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redcardinal-io/metering/application/repositories"
	domainerrors "github.com/redcardinal-io/metering/domain/errors" // Assuming AppError is defined here or accessible
	"github.com/redcardinal-io/metering/domain/models"
)

const (
	MaxBatchSize = 100
)

type ProducerService struct {
	producer repositories.ProducerRepository
	store    repositories.MeterStoreRepository
}

func NewProducerService(producer repositories.ProducerRepository, store repositories.MeterStoreRepository) *ProducerService {
	return &ProducerService{
		producer: producer,
		store:    store,
	}
}

type meterConfig struct {
	exists             map[string]bool
	requiredProperties map[string][]string
}

type validationResult struct {
	validEvents  []*models.Event
	failedEvents []*models.FailedEvent
}

func (p *ProducerService) PublishEvents(ctx context.Context, topic string, events *models.EventBatch, allowPartialSuccess bool) (*models.PublishEventsResult, error) {
	if err := validateBatchSize(events); err != nil {
		return nil, err
	}

	config, err := p.fetchAndPrepareMeterConfig(ctx, events)
	if err != nil {
		return nil, err
	}

	valResult, err := p.validateEventsAgainstConfig(events, config, allowPartialSuccess)
	if err != nil {
		return nil, err
	}

	result := &models.PublishEventsResult{
		SuccessCount: 0,
		FailedEvents: valResult.failedEvents,
	}

	if len(valResult.validEvents) == 0 {
		if len(result.FailedEvents) > 0 && allowPartialSuccess {
			return result, domainerrors.New(
				fmt.Errorf("all %d events failed validation", len(events.Events)),
				domainerrors.EINVALID,
				"all events failed validation",
			)
		}
		return result, nil
	}

	validBatch := &models.EventBatch{Events: valResult.validEvents}
	err = p.producer.PublishEvents(topic, validBatch)
	if err != nil {
		return result, domainerrors.New(
			fmt.Errorf("failed to publish valid events: %w", err),
			domainerrors.EINTERNAL,
			"event publishing failed",
		)
	}

	result.SuccessCount = len(valResult.validEvents)
	return result, nil
}

func validateBatchSize(events *models.EventBatch) error {
	if len(events.Events) > MaxBatchSize {
		return domainerrors.New(
			fmt.Errorf("batch size (%d) exceeds maximum limit of %d events", len(events.Events), MaxBatchSize),
			domainerrors.EINVALID,
			"batch size too large",
		)
	}
	return nil
}

func (p *ProducerService) fetchAndPrepareMeterConfig(ctx context.Context, events *models.EventBatch) (*meterConfig, error) {
	eventTypeSet := make(map[string]struct{})
	for _, event := range events.Events {
		if event != nil && event.Type != "" {
			eventTypeSet[event.Type] = struct{}{}
		}
	}

	if len(eventTypeSet) == 0 {
		return nil, domainerrors.New(
			fmt.Errorf("no valid event types found in the batch"),
			domainerrors.EINVALID,
			"no valid event types",
		)
	}

	eventTypes := make([]string, 0, len(eventTypeSet))
	for eventType := range eventTypeSet {
		eventTypes = append(eventTypes, eventType)
	}

	meters, err := p.store.ListMetersByEventTypes(ctx, eventTypes)
	if err != nil {
		return nil, domainerrors.New(
			fmt.Errorf("failed to fetch meters for event types %v: %w", eventTypes, err),
			domainerrors.EINTERNAL,
			"meter fetching error",
		)
	}

	config := &meterConfig{
		exists:             make(map[string]bool),
		requiredProperties: make(map[string][]string),
	}
	tempProperties := make(map[string]map[string]struct{})

	for _, meter := range meters {
		if meter == nil {
			continue
		}
		eventType := meter.EventType
		config.exists[eventType] = true

		if _, ok := tempProperties[eventType]; !ok {
			tempProperties[eventType] = make(map[string]struct{})
		}

		if meter.Properties != nil {
			for _, prop := range meter.Properties {
				if prop != "" {
					tempProperties[eventType][prop] = struct{}{}
				}
			}
		}
		if meter.ValueProperty != "" {
			tempProperties[eventType][meter.ValueProperty] = struct{}{}
		}
	}

	for eventType, propsSet := range tempProperties {
		propsList := make([]string, 0, len(propsSet))
		for prop := range propsSet {
			propsList = append(propsList, prop)
		}
		config.requiredProperties[eventType] = propsList
	}

	return config, nil
}

func (p *ProducerService) validateEventsAgainstConfig(events *models.EventBatch, config *meterConfig, allowPartialSuccess bool) (*validationResult, error) {
	result := &validationResult{
		validEvents:  make([]*models.Event, 0, len(events.Events)),
		failedEvents: make([]*models.FailedEvent, 0),
	}

	for _, event := range events.Events {
		var validationErr error

		if event == nil {
			validationErr = domainerrors.New(fmt.Errorf("event is nil"), domainerrors.EINVALID, "nil event in batch")
		} else if event.Type == "" {
			validationErr = domainerrors.New(fmt.Errorf("event type is empty"), domainerrors.EINVALID, "missing event type")
		} else if !config.exists[event.Type] {
			validationErr = domainerrors.New(
				fmt.Errorf("no meter configured for event type: %s", event.Type),
				domainerrors.EINVALID,
				"missing meter configuration",
			)
		} else {
			requiredProps := config.requiredProperties[event.Type]
			validationErr = p.validateEventProperties(event, requiredProps)
		}

		if validationErr != nil {
			if !allowPartialSuccess {
				return nil, validationErr
			}
			result.failedEvents = append(result.failedEvents, &models.FailedEvent{
				Event: event,
				Error: validationErr,
			})
		} else {
			result.validEvents = append(result.validEvents, event)
		}
	}

	return result, nil
}

func (p *ProducerService) validateEventProperties(event *models.Event, requiredProps []string) error {
	if len(requiredProps) == 0 {
		return nil
	}

	var eventProperties map[string]any
	if event.Properties == "" || event.Properties == "{}" {
		eventProperties = make(map[string]any)
	} else {
		if err := json.Unmarshal([]byte(event.Properties), &eventProperties); err != nil {
			return domainerrors.New(
				fmt.Errorf("failed to unmarshal event properties for event ID %s (type %s): %w", event.ID, event.Type, err),
				domainerrors.EINVALID,
				"invalid event properties format",
			)
		}
	}

	missingProps := make([]string, 0)
	for _, reqProp := range requiredProps {
		value, exists := eventProperties[reqProp]
		if !exists || isEmptyValue(value) {
			missingProps = append(missingProps, reqProp)
		}
	}

	if len(missingProps) > 0 {
		return domainerrors.New(
			fmt.Errorf("event ID %s (type %s) missing or empty required properties: %v", event.ID, event.Type, missingProps),
			domainerrors.EINVALID,
			"missing required event properties",
		)
	}

	return nil
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return false
	case bool:
		return false
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}
