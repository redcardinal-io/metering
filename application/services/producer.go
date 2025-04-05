package services

import (
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"go.uber.org/zap"
)

type ProducerService struct {
	logger   *logger.Logger
	producer repositories.ProducerRepository
}

func NewProducerService(producer repositories.ProducerRepository, logger *logger.Logger) *ProducerService {
	return &ProducerService{
		logger:   logger,
		producer: producer,
	}
}

func (p *ProducerService) PublishEvents(topic string, event *models.EventBatch) error {
	p.logger.Debug("Publishing event", zap.String("topic", topic), zap.Int("events-count", len(event.Events)))
	return p.producer.PublishEvents(topic, event)
}
