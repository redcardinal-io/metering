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
