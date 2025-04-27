package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redcardinal-io/metering/application/repositories"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
)

const (
	// Maximum number of events allowed in a single batch
	MaxBatchSize = 1000
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

// PublishEvents publishes a batch of events to the specified topic, with options for validation
func (p *ProducerService) PublishEvents(ctx context.Context, topic string, events *models.EventBatch, allowPartialSuccess bool) (*models.PublishEventsResult, error) {
	if events == nil || len(events.Events) == 0 {
		return &models.PublishEventsResult{SuccessCount: 0}, nil
	}

	// Check batch size limit
	if len(events.Events) > MaxBatchSize {
		return nil, domainerrors.New(
			fmt.Errorf("batch size exceeds maximum limit of %d events", MaxBatchSize),
			domainerrors.EINVALID,
			"batch size too large",
		)
	}

	// Extract unique event types
	eventTypeMap := make(map[string]bool)
	for _, event := range events.Events {
		if event != nil {
			eventTypeMap[event.Type] = true
		}
	}

	// Convert to slice
	eventTypes := make([]string, 0, len(eventTypeMap))
	for eventType := range eventTypeMap {
		eventTypes = append(eventTypes, eventType)
	}

	// Fetch meters for each event type
	meters := make([]*models.Meter, 0)
	meters, err := p.store.ListMetersByEventTypes(ctx, eventTypes)

	// Process required properties by event type
	properties := listPropertiesForEventType(meters)

	// Validate and prepare events
	validEvents := make([]*models.Event, 0, len(events.Events))
	result := &models.PublishEventsResult{
		FailedEvents: make([]*models.FailedEvent, 0),
	}

	for _, event := range events.Events {
		if event == nil {
			if !allowPartialSuccess {
				return nil, domainerrors.New(
					fmt.Errorf("event cannot be nil"),
					domainerrors.EINVALID,
					"invalid event",
				)
			}
			result.FailedEvents = append(result.FailedEvents, &models.FailedEvent{
				Event: nil,
				Error: domainerrors.New(
					fmt.Errorf("event cannot be nil"),
					domainerrors.EINVALID,
					"invalid event",
				),
			})
			continue
		}

		// Validate the event
		err := p.ValidateEvent(event, properties[event.Type])
		if err != nil {
			if !allowPartialSuccess {
				return nil, err
			}
			result.FailedEvents = append(result.FailedEvents, &models.FailedEvent{
				Event: event,
				Error: err,
			})
			continue
		}

		validEvents = append(validEvents, event)
	}

	if len(validEvents) == 0 {
		if len(result.FailedEvents) > 0 {
			return result, domainerrors.New(
				fmt.Errorf("all events failed validation"),
				domainerrors.EINVALID,
				"validation errors",
			)
		}
		return result, nil
	}

	// Create a new batch with only valid events
	validBatch := &models.EventBatch{
		Events: validEvents,
	}

	// Publish the valid events
	err = p.producer.PublishEvents(topic, validBatch)
	if err != nil {
		return result, domainerrors.New(
			fmt.Errorf("failed to publish events: %w", err),
			domainerrors.EINTERNAL,
			"publishing error",
		)
	}

	result.SuccessCount = len(validEvents)
	return result, nil
}

// ValidateEvent validates a single event against required properties
func (p *ProducerService) ValidateEvent(event *models.Event, properties []string) error {
	if event == nil {
		return domainerrors.New(fmt.Errorf("event cannot be nil"), domainerrors.EINVALID, "invalid event")
	}

	if event.Type == "" {
		return domainerrors.New(fmt.Errorf("event type cannot be empty"), domainerrors.EINVALID, "invalid event type")
	}

	// Skip property validation if no properties are required
	if len(properties) == 0 {
		return nil
	}

	// Parse event properties once
	var eventProperties map[string]any
	if event.Properties == "" {
		eventProperties = make(map[string]any)
	} else {
		if err := json.Unmarshal([]byte(event.Properties), &eventProperties); err != nil {
			return domainerrors.New(err, domainerrors.EINVALID, "invalid event properties format")
		}
	}

	// Validate event properties against required properties
	missingProps := make([]string, 0)
	for _, reqProp := range properties {
		value, exists := eventProperties[reqProp]
		if !exists || isEmptyValue(value) {
			missingProps = append(missingProps, reqProp)
		}
	}

	if len(missingProps) > 0 {
		return domainerrors.New(
			fmt.Errorf("missing or empty required properties: %v", missingProps),
			domainerrors.EINVALID,
			"invalid event properties",
		)
	}

	return nil
}

// isEmptyValue checks if a value should be considered empty
func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return false // Numbers are considered non-empty regardless of value
	case bool:
		return false // Boolean values are considered non-empty
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		// For complex types, consider them non-empty
		return false
	}
}

// listPropertiesForEventType collects required properties for each event type
func listPropertiesForEventType(meters []*models.Meter) map[string][]string {
	properties := make(map[string]map[string]struct{})

	// First pass: collect unique properties by event type using maps
	for _, meter := range meters {
		if _, exists := properties[meter.EventType]; !exists {
			properties[meter.EventType] = make(map[string]struct{})
		}

		// Check if Properties is nil before iterating
		if meter.Properties != nil {
			for _, prop := range meter.Properties {
				properties[meter.EventType][prop] = struct{}{}
			}
		}
	}

	// Second pass: convert maps to slices
	result := make(map[string][]string)
	for eventType, uniqueProps := range properties {
		propsList := make([]string, 0, len(uniqueProps))
		for prop := range uniqueProps {
			propsList = append(propsList, prop)
		}
		result[eventType] = propsList
	}

	return result
}
