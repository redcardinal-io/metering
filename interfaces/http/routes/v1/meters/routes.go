package meters

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger    *logger.Logger
	meterSvc  *services.MeterService
	validator *validator.Validate
}

func NewHTTPHandler(logger *logger.Logger, meterSvc *services.MeterService) *httpHandler {
	validator := validator.New()
	return &httpHandler{
		logger:    logger,
		meterSvc:  meterSvc,
		validator: validator,
	}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	r.Post("/meters", h.create)
	r.Post("/meters/query", h.query)
}
