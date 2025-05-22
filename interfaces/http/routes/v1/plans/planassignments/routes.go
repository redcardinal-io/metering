package planassignments

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

	assignments.Post("/", h.create)
	assignments.Put("/", h.update)
	assignments.Delete("/", h.delete)
}

func getPlanIDFromIdentifier(ctx context.Context, idOrSlug string, planSvc *services.PlanManagementService) (*uuid.UUID, error) {
	planId, err := uuid.Parse(idOrSlug)
	if err != nil {
		plan, err := planSvc.GetPlanByIDorSlug(ctx, idOrSlug)
		if err != nil {
			return nil, err
		}
		return &plan.ID, nil
	}
	return &planId, nil
}
