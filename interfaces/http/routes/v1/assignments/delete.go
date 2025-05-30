package assignments

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type terminatePlanRequest struct {
	PlanIDOrSlug   string `json:"plan_id_or_slug" validate:"required"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
}

// @Summary Delete a plan assignment
// @Description Terminate a plan assignment for an organization or user
// @Tags plan-assignments
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param assignment body terminatePlanRequest true "Assignment termination information"
// @Success 204 "No Content"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan assignment not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/assignments [delete]
func (h *httpHandler) delete(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req terminatePlanRequest

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

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenant_slug)

	planID, getErr := getPlanIDFromIdentifier(c, req.PlanIDOrSlug, h.planSvc)
	if getErr != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(getErr, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	err := h.planSvc.TerminateAssignment(c, models.TerminateAssignmentInput{
		PlanID:         planID,
		OrganizationID: req.OrganizationID,
		UserID:         req.UserID,
	})
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to terminate plan assignment")
		h.logger.Error("failed to terminate plan assignment", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
