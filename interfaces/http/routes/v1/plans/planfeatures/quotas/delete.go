package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

func (h *httpHandler) delete(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)

	planID := ctx.Params("planID")
	featureID := ctx.Params("featureID")

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

	if err := h.planSvc.DeletePlanFeatureQuota(c, planID, featureID); err != nil {
		errResp := domainerrors.NewErrorResponse(err)
		h.logger.Error("failed to delete plan feature quota", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
