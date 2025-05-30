package features

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// @Summary Delete a feature
// @Description Delete a feature by ID or slug
// @Tags features
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param idOrSlug path string true "Feature ID or Slug"
// @Success 204 "Feature deleted successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 404 {object} domainerrors.ErrorResponse "Feature not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/features/{idOrSlug} [delete]
func (h *httpHandler) deleteByIDorSlug(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "feature ID or Slug is required")
		h.logger.Error("feature ID or Slug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)
	err := h.featureSvc.DeleteFeatureByIDorSlug(c, idOrSlug)
	if err != nil {
		h.logger.Error("failed to delete feature", zap.String("id", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		SendStatus(fiber.StatusNoContent)
}
