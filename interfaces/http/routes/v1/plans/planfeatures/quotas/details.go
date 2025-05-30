package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
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
	// Get feature ID from URL parameter
	featureID := ctx.Params("featureID")

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	quota, err := h.planSvc.GetPlanFeatureQuota(c, planID, featureID)
	if err != nil {
		errResp := domainerrors.NewErrorResponse(err)
		h.logger.Error("failed to get plan feature quota details", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.JSON(models.NewHttpResponse(quota, "plan feature quota details retrieved successfully", fiber.StatusOK))
}
