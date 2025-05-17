package features

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

func (h *httpHandler) getByIDorSlug(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)
	idOrSlug := ctx.Params("idOrSlug")

	if idOrSlug == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "feature ID is required")
		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	feature, err := h.featureSvc.GetFeatureByIDorSlug(c, idOrSlug)
	if err != nil {
		h.logger.Error("failed to get feature", zap.String("idOrSlug", idOrSlug), zap.Reflect("error", err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.
		Status(fiber.StatusOK).JSON(models.NewHttpResponse(feature, "plan retrieved successfully", fiber.StatusOK))
}
