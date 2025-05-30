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

type createPlanFeatureRequest struct {
	FeatureID string                 `json:"feature_id" validate:"required,uuid"`
	Config    map[string]interface{} `json:"config,omitempty"`
	CreatedBy string                 `json:"created_by" validate:"required"`
}

// @Summary Create plan feature
// @Description Adds a new feature to a plan
// @Tags Plan Features
// @Accept json
// @Produce json
// @Param X-Tenant header string true "Tenant slug"
// @Param planID path string true "Plan ID" format(uuid)
// @Param body body createPlanFeatureRequest true "Plan feature details"
// @Success 201 {object} models.HttpResponse[models.PlanFeature] "Plan feature created successfully"
// @Failure 400 {object} domainerrors.ErrorResponse "Invalid request"
// @Failure 401 {object} domainerrors.ErrorResponse "Unauthorized"
// @Failure 409 {object} domainerrors.ErrorResponse "Plan feature already exists"
// @Failure 500 {object} domainerrors.ErrorResponse "Internal server error"
// @Router /v1/plans/{planID}/features [post]
func (h *httpHandler) create(ctx *fiber.Ctx) error {
	// Get tenant slug from context
	tenantSlug := ctx.Get(constants.TenantHeader)

	planIDStr := ctx.Params("planID")
	planID := uuid.MustParse(planIDStr)

	var req createPlanFeatureRequest
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

	// Parse feature ID
	featureID, err := uuid.Parse(req.FeatureID)
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

	belongs, err := h.planSvc.CheckPlanAndFeatureForTenant(c, planID, featureID)
	if err != nil {
		h.logger.Error("failed to check plan and feature for tenant",
			zap.String("tenant", tenantSlug),
			zap.String("plan_id", planID.String()),
			zap.String("feature_id", featureID.String()),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if !belongs {
		errResp := domainerrors.NewErrorResponseWithOpts(
			nil,
			domainerrors.EUNAUTHORIZED,
			"plan feature not found or does not belong to tenant",
		)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	// Call service to create plan feature
	planFeature, err := h.planSvc.CreatePlanFeature(c, models.CreatePlanFeatureInput{
		PlanID:    planID,
		FeatureID: featureID,
		Config:    req.Config,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		h.logger.Error("failed to create plan feature",
			zap.String("plan_id", planID.String()),
			zap.String("feature_id", featureID.String()),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.Status(fiber.StatusCreated).JSON(models.NewHttpResponse(
		planFeature,
		"plan feature created successfully",
		fiber.StatusCreated,
	))
}
