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

type updateAssignedPlanRequest struct {
	PlanIDOrSlug        string    `json:"plan_id_or_slug" validate:"required"`
	OrganizationID      string    `json:"organization_id"`
	UserID              string    `json:"user_id"`
	ValidFrom           time.Time `json:"valid_from"`
	ValidUntil          time.Time `json:"valid_until"`
	UpdatedBy           string    `json:"updated_by" validate:"required"`
	SetValidUntilToZero bool      `json:"set_valid_until_to_zero" validate:"omitempty"`
}

// @Summary Update a plan assignment
// @Description Update the validity period for a plan assignment
// @Tags plan-assignments
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param assignment body updateAssignedPlanRequest true "Assignment update information"
// @Success 200 {object} models.HttpResponse[models.PlanAssignment] "Assignment updated successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan assignment not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/assignments [put]
func (h *httpHandler) update(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req updateAssignedPlanRequest

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

	if !req.ValidUntil.IsZero() && req.SetValidUntilToZero {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "cannot set value to valid_unitl if set_valid_until_to_zero is true")
		h.logger.Error("cannot set value to valid_unitl if set_valid_until_to_zero is true", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if !req.ValidFrom.IsZero() && !req.ValidUntil.IsZero() && req.ValidFrom.After(req.ValidUntil) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_until must be after valid_from")
		h.logger.Error("valid_until must be after valid_from", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenant_slug)

	planId, getErr := getPlanIDFromIdentifier(c, req.PlanIDOrSlug, h.planSvc)
	if getErr != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(getErr, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// set valid_until to zero because request might contain valid_until as non zero
	// breaking  validation logic in the service layer
	if req.SetValidUntilToZero {
		req.ValidUntil = time.Time{} // set valid_until to zero
	}

	updatedAssignment, err := h.planSvc.UpdateAssignment(c, models.UpdateAssignmentInput{
		PlanID:              planId,
		OrganizationID:      req.OrganizationID,
		UserID:              req.UserID,
		ValidFrom:           req.ValidFrom,
		ValidUntil:          req.ValidUntil,
		UpdatedBy:           req.UpdatedBy,
		SetValidUntilToZero: req.SetValidUntilToZero,
	})
	if err != nil {
		h.logger.Error("failed to update plan assignment", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(updatedAssignment, "updated assignment successfully", fiber.StatusCreated))
}
