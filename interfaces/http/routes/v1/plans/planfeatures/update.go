package planfeatures

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type updatePlanFeatureRequest struct {
	Config    map[string]interface{} `json:"config,omitempty"`
	UpdatedBy string                 `json:"updated_by" validate:"required"`
}

// @Description Updates configuration of a feature associated with a plan
// @Tags Plan Features
// @Accept json
// @Produce json
// @Param X-Tenant header string true "Tenant slug"
// @Param planID path string true "Plan ID" format(uuid)
// @Param featureID path string true "Feature ID" format(uuid)
// @Param body body updatePlanFeatureRequest true "Updated plan feature details"
// @Success 200 {object} models.HttpResponse[models.PlanFeature] "Plan feature updated successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 401 {object} domainerrors.ErrorResponse "Unauthorized"
// @Failure 404 {object} domainerrors.ErrorResponse "Plan feature not found"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{planID}/features/{featureID} [put]
func (h *httpHandler) update(ctx *fiber.Ctx) error {
	// Get tenant slug from context
	tenantSlug := ctx.Get(constants.TenantHeader)

	// Get plan ID from URL parameter
	planIDStr := ctx.Params("planID")
	planID := uuid.MustParse(planIDStr)

	// Get feature ID from URL parameter
	featureIDStr := ctx.Params("featureID")
	featureID := uuid.MustParse(featureIDStr)

	var req updatePlanFeatureRequest
	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"failed to parse request body",
		)
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Validate the request body
	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid request body",
		)
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Create context with tenant slug
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	// Call service to update plan feature
	planFeature, err := h.planSvc.UpdatePlanFeature(c, planID, featureID, models.UpdatePlanFeatureInput{
		Config:    req.Config,
		UpdatedBy: req.UpdatedBy,
	})
	if err != nil {
		h.logger.Error("failed to update plan feature",
			zap.String("plan_id", planID.String()),
			zap.String("feature_id", featureID.String()),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.Status(fiber.StatusOK).JSON(models.NewHttpResponse(
		planFeature,
		"plan feature updated successfully",
		fiber.StatusOK,
	))
}
