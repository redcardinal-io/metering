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
	meters, err := p.store.ListMetersByEventType(context.Background(), events.Events[0].Type)
	if err != nil {
		return err
	}
	for _, event := range events.Events {
		if err := p.ValidateEvent(event, meters); err != nil {
			return err
		}
	}

	return p.producer.PublishEvents(topic, events)
}

func (p *ProducerService) ValidateEvent(event *models.Event, meters []*models.Meter) error {
	if event == nil {
		return domainerrors.New(fmt.Errorf("event cannot be nil"), domainerrors.EINVALID, "invalid event")
	}

	if event.Type == "" {
		return domainerrors.New(fmt.Errorf("event type cannot be empty"), domainerrors.EINVALID, "invalid event type")
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
