package plans

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

// NewHTTPHandler initializes a new httpHandler with the provided logger and plan service for handling plan-related HTTP endpoints.
func NewHTTPHandler(logger *logger.Logger, planSvc *services.PlanManagementService) *httpHandler {
	validator := validator.New()
	return &httpHandler{
		logger:    logger,
		planSvc:   planSvc,
		validator: validator,
	}
}

func (h *httpHandler) RegisterRoutes(r fiber.Router) {
	// Group all plans routes
	plans := r.Group("/plans")

	// Plan collection routes
	plans.Post("/", h.create)
	plans.Get("/", h.list)

	// Single Plan routes with id parameter
	plans.Get("/:idOrSlug", h.getByIDorSlug)
	plans.Put("/:idOrSlug", h.updateByIDorSlug)
	plans.Delete("/:idOrSlug", h.deleteByIDorSlug)

	// Toggle Plan Archive
	plans.Put("/:idOrSlug/archive", h.archive)
}
