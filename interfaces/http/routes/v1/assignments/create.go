package assignments

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type createAssignmentRequest struct {
	PlanIDOrSlug   string    `json:"plan_id_or_slug" validate:"required"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	ValidFrom      time.Time `json:"valid_from" validate:"required"`
	ValidUntil     time.Time `json:"valid_until" validate:"omitempty"`
	CreatedBy      string    `json:"created_by" validate:"required"`
}

// @Summary Create a new plan assignment
// @Description Assign a plan to either an organization or a user
// @Tags plan-assignments
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param assignment body createAssignmentRequest true "Assignment information"
// @Success 201 {object} models.HttpResponse[models.PlanAssignment] "Plan assignment created successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/assignments [post]
func (h *httpHandler) create(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	var req createAssignmentRequest

	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// validate the request body
	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ValidFrom.IsZero() {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_from is required")
		h.logger.Error("valid_from is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if !req.ValidUntil.IsZero() && req.ValidFrom.After(req.ValidUntil) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_until must be after valid_from")
		h.logger.Error("valid_until must be after valid_from", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.OrganizationID != "" && req.UserID != "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "organization_id and user_id are mutually exclusive, provide any one")
		h.logger.Error("organization_id and user_id are mutually exclusive, provide any one", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.OrganizationID == "" && req.UserID == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "organization_id or user_id is required")
		h.logger.Error("organization_id or user_id is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	planID, err := getPlanIDFromIdentifier(c, req.PlanIDOrSlug, h.planSvc)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	planAssignment, err := h.planSvc.CreateAssignment(c, models.CreateAssignmentInput{
		PlanID:         planID,
		OrganizationID: req.OrganizationID,
		UserID:         req.UserID,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		CreatedBy:      req.CreatedBy,
	})
	if err != nil {
		h.logger.Error("failed to create plan assignment", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(planAssignment, "plan assignment created successfully", fiber.StatusCreated))
}
