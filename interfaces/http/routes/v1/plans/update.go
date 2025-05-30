package plans

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// updatePlanRequest defines the request body for updating a plan
type updatePlanRequest struct {
	Name string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description string `json:"description,omitempty" validate:"omitempty,min=10,max=255"`
	UpdatedBy string `json:"updated_by" validate:"required,min=3,max=255"`
}

// @Summary Update a plan
// @Description Update a plan's details by ID or slug
// @Tags plans
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param idOrSlug path string true "Plan ID or slug"
// @Param plan body updatePlanRequest true "Updated plan information"
// @Success 200 {object} models.HttpResponse[models.Plan] "Plan updated successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{idOrSlug} [put]
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

// archivePlanRequest defines the request body for archiving or unarchiving a plan
type archivePlanRequest struct {
	UpdatedBy string `json:"updated_by" validate:"required,min=3,max=255"`
	Archive bool `json:"archive"`
}

// @Summary Archive or unarchive a plan
// @Description Archive or unarchive a plan by ID or slug
// @Tags plans
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param idOrSlug path string true "Plan ID or slug"
// @Param archiveRequest body archivePlanRequest true "Archive request information"
// @Success 204 "Plan archived/unarchived successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{idOrSlug}/archive [put]
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
