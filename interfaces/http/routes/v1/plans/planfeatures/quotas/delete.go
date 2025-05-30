package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

func (h *httpHandler) delete(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)

	planID := ctx.Params("planID")
	featureID := ctx.Params("featureID")

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	if err := h.planSvc.DeletePlanFeatureQuota(c, planID, featureID); err != nil {
		errResp := domainerrors.NewErrorResponse(err)
		h.logger.Error("failed to delete plan feature quota", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
