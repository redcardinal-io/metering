package meters

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger   *logger.Logger
	meterSvc *services.MeterService
}

func NewHTTPHandler(logger *logger.Logger, meterSvc *services.MeterService) *httpHandler {
	return &httpHandler{
		logger:   logger,
		meterSvc: meterSvc,
	}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	r.Post("/meters", h.create)
}
