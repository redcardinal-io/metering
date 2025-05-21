package plans

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/interfaces/http/routes/v1/plans/planassignments"
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
	// Group all plans routes
	plans := r.Group("/plans")

	// Plan collection routes
	plans.Post("/", h.create)
	plans.Get("/", h.list)

	// Single Plan routes with id or slug parameter
	plan := plans.Group("/:idOrSlug")
	plan.Get("/", h.details)
	plan.Put("/", h.update)
	plan.Delete("/", h.delete_h)
	plan.Put("/archive", h.archive)

	assignments := planassignments.NewHTTPHandler(h.logger, h.planSvc)
	assignments.RegisterRoutes(plan)
}
