package repositories

import "github.com/redcardinal-io/metering/domain/models"

type ProducerRepository interface {
	PublishEvents(topic string, event *models.EventBatch) error
	Close() error
}
