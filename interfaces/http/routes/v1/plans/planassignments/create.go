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

type createAssignmentRequest struct {
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	ValidFrom      time.Time `json:"valid_from" validate:"required"`
	ValidUntil     time.Time `json:"valid_until" validate:"required"`
	CreatedBy      string    `json:"created_by" validate:"required"`
}

func (h *httpHandler) create(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	var req createAssignmentRequest

	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan idOrSlug is required")
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

	if req.ValidFrom.After(req.ValidUntil) {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_until must be after valid_from")
		h.logger.Error("valid_until must be after valid_from", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ValidFrom.IsZero() {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_from is required")
		h.logger.Error("valid_from is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ValidUntil.IsZero() {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "valid_until is required")
		h.logger.Error("valid_until is required", zap.Reflect("error", errResp))
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

	planID, err := getPlanIDFromIdentifier(c, idOrSlug, h.planSvc)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid plan id or slug")
		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	planAssignment, err := h.planSvc.CreateAssignment(c, models.CreateAssignmentInput{
		PlanID:         planID,
		OrganizationID: req.OrganizationID,
		UserID:         req.UserID,
		ValidFrom:      req.ValidFrom.UTC(),
		ValidUntil:     req.ValidUntil.UTC(),
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
