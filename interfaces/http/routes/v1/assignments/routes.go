package assignments

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/application/services"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
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

func getPlanIDFromIdentifier(ctx context.Context, idOrSlug string, planSvc *services.PlanManagementService) (*uuid.UUID, error) {
	plan, err := planSvc.GetPlanByIDorSlug(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}

	return &plan.ID, nil
}

func getPlanTimeRange(ctx context.Context, planId *uuid.UUID, orgId string, userId string, planSvc *services.PlanManagementService) (*time.Time, *time.Time, error) {
	assignmentQueryInput := models.QueryPlanAssignmentInput{
		PlanID:         planId,
		OrganizationID: orgId,
		UserID:         userId,
	}
	plan, err := planSvc.ListAssignments(ctx, assignmentQueryInput, pagination.Pagination{Limit: 1, Page: 0})
	if err != nil {
		return nil, nil, err
	}

	validFrom := plan.Results[0].ValidFrom
	validUntil := plan.Results[0].ValidUntil

	return &validFrom, &validUntil, nil
}
