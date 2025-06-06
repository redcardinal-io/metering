package assignments

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	assignments := r.Group("/plans/assignments")

	assignments.Get("/", h.list)
	assignments.Get("/history", h.listhistory)
	assignments.Post("/", h.create)
	assignments.Put("/", h.update)
	assignments.Delete("/", h.delete)
}

// getPlanIDFromIdentifier retrieves the UUID of a plan given its ID or slug identifier.
// Returns a pointer to the plan's UUID if found, or an error if the plan does not exist.
func getPlanIDFromIdentifier(ctx context.Context, idOrSlug string, planSvc *services.PlanManagementService) (*uuid.UUID, error) {
	plan, err := planSvc.GetPlanByIDorSlug(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}

	return &plan.ID, nil
}
