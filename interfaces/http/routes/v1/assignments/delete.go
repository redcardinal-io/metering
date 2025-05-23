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
	PlanIDOrSlug   string `json:"plan_idorslug" validate:"required"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
}

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
