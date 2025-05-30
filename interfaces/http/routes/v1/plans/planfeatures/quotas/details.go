package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

// @Summary Get plan feature quota details
// @Description Retrieves quota configuration for a specific plan feature
// @Tags Plan Feature Quotas
// @Accept json
// @Produce json
// @Param X-Tenant header string true "Tenant slug"
// @Param planID path string true "Plan ID" format(uuid)
// @Param featureID path string true "Feature ID" format(uuid)
// @Success 200 {object} models.HttpResponse[models.PlanFeatureQuota] "Quota details retrieved successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 401 {object} domainerrors.ErrorResponse "Unauthorized"
// @Failure 404 {object} domainerrors.ErrorResponse "Quota not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{planID}/features/{featureID}/quotas [get]
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
