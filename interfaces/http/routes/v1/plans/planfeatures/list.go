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

type listPlanFeaturesQuery struct {
	FeatureType *string `query:"feature_type"`
}

func (h *httpHandler) list(ctx *fiber.Ctx) error {
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

	// Parse query parameters
	var query listPlanFeaturesQuery
	if err := ctx.QueryParser(&query); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"failed to parse query parameters",
		)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Create context with tenant slug
	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	// Call service to list plan features
	filter := models.PlanFeatureListFilter{
		FeatureType: (*models.FeatureTypeEnum)(query.FeatureType),
	}

	planFeatures, err := h.planSvc.ListPlanFeaturesByPlan(c, planID, filter)
	if err != nil {
		h.logger.Error("failed to list plan features",
			zap.String("plan_id", planID.String()),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.Status(fiber.StatusOK).JSON(models.NewHttpResponse(
		planFeatures,
		"plan features retrieved successfully",
		fiber.StatusOK,
	))
}
