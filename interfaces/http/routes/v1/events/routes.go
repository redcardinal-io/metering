package events

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger       *logger.Logger
	publishTopic string
	producer     *services.ProducerService
	validator    *validator.Validate
}

type HttpHandlerParams struct {
	PublishTopic string
	Producer     *services.ProducerService
	Logger       *logger.Logger
}

func NewHTTPHandler(params HttpHandlerParams) *httpHandler {
	validator := validator.New()
	return &httpHandler{
		logger:       params.Logger,
		publishTopic: params.PublishTopic,
		producer:     params.Producer,
		validator:    validator,
	}
}

func (httph *httpHandler) RegisterRoutes(r fiber.Router) {
	r.Post("/events", httph.publishEvent)
}
