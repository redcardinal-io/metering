package quotas

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

type httpHandler struct {
	logger    *logger.Logger
	planSvc   *services.PlanManagementService
	validator *validator.Validate
}

func NewHTTPHandler(logger *logger.Logger, planSvc *services.PlanManagementService) *httpHandler {
	validator := validator.New()
	return &httpHandler{
		logger:    logger,
		planSvc:   planSvc,
		validator: validator,
	}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	r.Get("/", h.details)
	r.Post("/", h.create)
	r.Put("/", h.update)
	r.Delete("/", h.delete)
}
