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
	FeatureType string `query:"feature_type"`
}

// @Summary List plan features
// @Description Retrieves all features associated with a plan
// @Tags Plan Features
// @Accept json
// @Produce json
// @Param X-Tenant header string true "Tenant slug"
// @Param planID path string true "Plan ID" format(uuid)
// @Param feature_type query string false "Filter by feature type"
// @Success 200 {object} models.HttpResponse[[]models.PlanFeature]  "List of plan features"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 401 {object} domainerrors.ErrorResponse "Unauthorized"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{planID}/features [get]
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
	featureType := models.FeatureTypeEnum(query.FeatureType)

	// Call service to list plan features
	filter := models.PlanFeatureListFilter{
		FeatureType: featureType,
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
