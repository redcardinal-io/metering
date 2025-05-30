package features

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"go.uber.org/zap"
)

// @Summary List features
// @Description Get a list of features for the tenant
// @Tags features
// @Accept json
// @Produce json
// @Param X-Tenant-Slug header string true "Tenant Slug"
// @Param page query integer false "Page number"
// @Param limit query integer false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort order (asc/desc)"
// @Param type query string false "Filter by feature type (static/metered)"
// @Success 200 {object} models.HttpResponse[[]models.Feature] "Features retrieved successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/features [get]
func (h *httpHandler) list(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)

	// Create pagination input
	paginationInput := pagination.ExtractPaginationFromContext(ctx)
	// validate the pagination input
	if paginationInput.Queries["type"] != "" {
		if paginationInput.Queries["type"] != string(models.FeatureTypeStatic) &&
			paginationInput.Queries["type"] != string(models.FeatureTypeMetered) {
			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "invalid feature type")
			h.logger.Error("invalid feature type", zap.Reflect("error", errResp))
			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
		}
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	features, err := h.featureSvc.ListFeatures(c, paginationInput)
	if err != nil {
		h.logger.Error("failed to list features", zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(features, "features retrieved successfully", fiber.StatusOK))
}
