package planfeatures

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

// NewHTTPHandler creates and returns a new httpHandler configured with the given logger and plan management service.
func NewHTTPHandler(logger *logger.Logger, planSvc *services.PlanManagementService) *httpHandler {
	validator := validator.New()
	return &httpHandler{
		logger:    logger,
		planSvc:   planSvc,
		validator: validator,
	}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	// Register plan feature routes
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Put("/:featureID", h.TenantPlanFeatureMiddleware(), h.update)
	r.Delete("/:featureID", h.TenantPlanFeatureMiddleware(), h.delete)
}
