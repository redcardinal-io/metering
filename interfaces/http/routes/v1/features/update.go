package features

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type updateFeatureRequest struct {
	Name        string         `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description string         `json:"description" validate:"omitempty,min=10,max=255"`
	Config      map[string]any `json:"config" validate:"omitempty"`
	UpdatedBy   string         `json:"updated_by" validate:"required,min=3,max=100"`
}

// @Summary Update a feature
// @Description Update a feature by ID or slug
// @Tags features
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param idOrSlug path string true "Feature ID or Slug"
// @Param feature body updateFeatureRequest true "Feature update data"
// @Success 200 {object} models.HttpResponse[models.Feature] "Feature updated successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Feature not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/features/{idOrSlug} [put]
func (h *httpHandler) updateByIDorSlug(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "feature idOrSlug is required")
		h.logger.Error("feature idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req updateFeatureRequest
	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}
	if req.Name == "" && req.Description == "" && req.Config == nil {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "at least one field (name or description or config) is required")
		h.logger.Error("at least one field (name or description or config) is required ", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	feature, err := h.featureSvc.UpdateFeatureByIDorSlug(c, idOrSlug, models.UpdateFeatureInput{
		Name:        req.Name,
		UpdatedBy:   req.UpdatedBy,
		Config:      req.Config,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to update feature", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(feature, "feature updated successfully", fiber.StatusOK))
}
