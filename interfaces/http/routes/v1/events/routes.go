package events

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger       *logger.Logger
	publishTopic string
	producer     *services.ProducerService
}

type HttpHandlerParams struct {
	PublishTopic string
	Producer     *services.ProducerService
	Logger       *logger.Logger
}

func NewHTTPHandler(params HttpHandlerParams) *httpHandler {
	return &httpHandler{
		logger:       params.Logger,
		publishTopic: params.PublishTopic,
		producer:     params.Producer,
	}
}

func (httph *httpHandler) RegisterRoutes(r fiber.Router) {
	r.Post("/events", httph.publishEvent)
}
