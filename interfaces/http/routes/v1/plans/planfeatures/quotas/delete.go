package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// @Summary Delete plan feature quota
// @Description Removes the quota configuration from a plan feature
// @Tags Plan Feature Quotas
// @Accept json
// @Produce json
// @Param X-Tenant header string true "Tenant slug"
// @Param planID path string true "Plan ID" format(uuid)
// @Param featureID path string true "Feature ID" format(uuid)
// @Success 204 "Quota deleted successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 401 {object} domainerrors.ErrorResponse "Unauthorized"
// @Failure 404 {object} domainerrors.ErrorResponse "Quota not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{planID}/features/{featureID}/quotas [delete]
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
