package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

func (h *httpHandler) details(ctx *fiber.Ctx) error {
	// Get tenant slug from context
	tenantSlug := ctx.Get(constants.TenantHeader)

	// Get plan ID from URL parameter
	planID := ctx.Params("planID")
	_, err := uuid.Parse(planID)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid plan ID",
		)
		h.logger.Error("invalid plan ID", zap.String("planID", planID), zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Get feature ID from URL parameter
	featureID := ctx.Params("featureID")
	_, err = uuid.Parse(featureID)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid feature ID",
		)
		h.logger.Error("invalid feature ID", zap.String("featureID", featureID), zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	quota, err := h.planSvc.GetPlanFeatureQuota(c, planID, featureID)
	if err != nil {
		errResp := domainerrors.NewErrorResponse(err)
		h.logger.Error("failed to get plan feature quota details", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.JSON(models.NewHttpResponse(quota, "plan feature quota details retrieved successfully", fiber.StatusOK))
}
