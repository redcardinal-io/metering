package plans

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type updatePlanRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description string `json:"description,omitempty" validate:"omitempty,min=10,max=255"`
	UpdatedBy   string `json:"updated_by" validate:"required,min=3,max=255"`
}

func (h *httpHandler) update(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req updatePlanRequest
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
	if req.Name == "" && req.Description == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "at least one field (name or description) is required")
		h.logger.Error("at least one field (name or description) is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	plan, err := h.planSvc.UpdatePlanByIDorSlug(c, idOrSlug, models.UpdatePlanInput{
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   req.UpdatedBy,
	})
	if err != nil {
		h.logger.Error("failed to update plan", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(plan, "plan updated successfully", fiber.StatusOK))
}

type archivePlanRequest struct {
	UpdatedBy string `json:"updated_by" validate:"required,min=3,max=255"`
	Archive   bool   `json:"archive"`
}

func (h *httpHandler) archive(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "Plan Id or Slug is required")
		h.logger.Error("Plan Id or Slug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req archivePlanRequest
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

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	err := h.planSvc.ArchivePlanByIDorSlug(c, idOrSlug, models.ArchivePlanInput{
		UpdatedBy: req.UpdatedBy,
		Archive:   req.Archive,
	})
	if err != nil {
		h.logger.Error("failed to toggle archive plan", zap.String("id", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		SendStatus(fiber.StatusNoContent)
}
