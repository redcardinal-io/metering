package planassignments

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type upadateAssignedPlanRequest struct {
	OrganizationID string     `json:"organization_id"`
	UserID         string     `json:"user_id"`
	ValidFrom      *time.Time `json:"valid_from" validate:"omitempty"`
	ValidUntil     *time.Time `json:"valid_until" validate:"omitempty"`
	UpdatedBy      string     `json:"updated_by" validate:"required"`
}

func (h *httpHandler) update(ctx *fiber.Ctx) error {
	tenant_slug := ctx.Get(constants.TenantHeader)
	var req upadateAssignedPlanRequest

	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

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

	planId, getErr := getPlanIDFromIdentifier(c, idOrSlug, h.planSvc)
	if getErr != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(getErr, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var utcValidFrom, utcValidUntil time.Time
	if req.ValidFrom != nil {
		utcValidFrom = req.ValidFrom.UTC()
	}
	if req.ValidUntil != nil {
		utcValidUntil = req.ValidUntil.UTC()
	}

	updatedAssignment, err := h.planSvc.UpdateAssignment(c, models.UpdateAssignmentInput{
		PlanID:         planId,
		OrganizationID: req.OrganizationID,
		UserID:         req.UserID,
		ValidFrom:      &utcValidFrom,
		ValidUntil:     &utcValidUntil,
		UpdatedBy:      req.UpdatedBy,
	})
	if err != nil {
		h.logger.Error("failed to update plan assignment", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(updatedAssignment, "updated assignment successfully", fiber.StatusCreated))
}
