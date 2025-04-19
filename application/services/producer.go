package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redcardinal-io/metering/application/repositories"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
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

func (p *ProducerService) PublishEvents(topic string, events *models.EventBatch) error {
	for _, event := range events.Events {
		if err := p.ValidateEvent(event); err != nil {
			return err
		}
	}

	return p.producer.PublishEvents(topic, events)
}

func (p *ProducerService) ValidateEvent(event *models.Event) error {
	if event == nil {
		return domainerrors.New(fmt.Errorf("event cannot be nil"), domainerrors.EINVALID, "invalid event")
	}

	if event.Type == "" {
		return domainerrors.New(fmt.Errorf("event type cannot be empty"), domainerrors.EINVALID, "invalid event type")
	}

	ctx := context.Background()
	meters, err := p.store.ListMetersByEventType(ctx, event.Type)
	if err != nil {
		return domainerrors.New(err, domainerrors.EINTERNAL, "failed to retrieve meters")
	}

	// If no meters are found for the event type, return nil
	if len(meters) == 0 {
		return nil
	}

	// Collect all required properties from meters
	requiredProperties := make(map[string]int)
	for _, meter := range meters {
		for _, prop := range meter.Properties {
			requiredProperties[prop] = requiredProperties[prop] + 1
		}
	}

	// Parse event properties
	var eventProperties map[string]any
	if err := json.Unmarshal([]byte(event.Properties), &eventProperties); err != nil {
		return domainerrors.New(err, domainerrors.EINVALID, "invalid event properties format")
	}

	// Validate event properties against required properties
	missingProps := make([]string, 0)
	for reqProp := range requiredProperties {
		value, exists := eventProperties[reqProp]
		if !exists || value == "" {
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
