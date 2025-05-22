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

func (h *httpHandler) delete(ctx *fiber.Ctx) error {
	// Get tenant slug from context
	tenantSlug := ctx.Get(constants.TenantHeader)

	// Get plan ID from URL parameter
	planIDStr := ctx.Params("planID")
	if planIDStr == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(
			nil,
			domainerrors.EINVALID,
			"plan ID is required",
		)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid plan ID format",
		)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Get feature ID from URL parameter
	featureIDStr := ctx.Params("featureID")
	if featureIDStr == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(
			nil,
			domainerrors.EINVALID,
			"feature ID is required",
		)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	featureID, err := uuid.Parse(featureIDStr)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid feature ID format",
		)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Create context with tenant slug
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	// Call service to delete plan feature
	err = h.planSvc.DeletePlanFeature(c, models.DeletePlanFeatureInput{
		PlanID:    planID,
		FeatureID: featureID,
	})
	if err != nil {
		h.logger.Error("failed to delete plan feature",
			zap.String("plan_id", planID.String()),
			zap.String("feature_id", featureID.String()),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.Status(fiber.StatusOK).JSON(models.NewHttpResponse[*string](
		nil,
		"plan feature deleted successfully",
		fiber.StatusOK,
	))
}
