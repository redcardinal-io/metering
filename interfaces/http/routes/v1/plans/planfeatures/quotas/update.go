package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type updatePlanFeatureQuotaRequest struct {
	LimitValue          int64  `json:"limit_value,omitempty" validate:"omitempty,gt=0"`
	ResetPeriod         string `json:"reset_period,omitempty" validate:"omitempty,oneof=day week month year custom rolling never"`
	CustomPeriodMinutes int64  `json:"custom_period_minutes,omitempty" validate:"omitempty,gt=0"`
	ActionAtLimit       string `json:"action_at_limit,omitempty" validate:"omitempty,oneof=none block throttle"`
}

func (h *httpHandler) update(ctx *fiber.Ctx) error {
	// Get tenant slug from context
	tenantSlug := ctx.Get(constants.TenantHeader)

	// Get plan ID from URL parameter
	planID := ctx.Params("planID")
	// Get feature ID from URL parameter
	featureID := ctx.Params("featureID")

	var req updatePlanFeatureQuotaRequest
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

	if req.LimitValue == 0 && req.ResetPeriod == "" && req.CustomPeriodMinutes == 0 && req.ActionAtLimit == "" {
		errResp := domainerrors.NewErrorResponseWithOpts(
			nil,
			domainerrors.EINVALID,
			"at least one field must be provided to update",
		)
		h.logger.Error("no fields provided for update", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ResetPeriod == "custom" && req.CustomPeriodMinutes == 0 {
		errResp := domainerrors.NewErrorResponseWithOpts(
			nil,
			domainerrors.EINVALID,
			"custom period minutes is required when reset period is custom",
		)
		h.logger.Error("custom period minutes required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ResetPeriod != "custom" && req.CustomPeriodMinutes != 0 {
		req.CustomPeriodMinutes = 0 // Clear custom period if reset period is not custom
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)

	resetPeriod := models.MeteredResetPeriod(req.ResetPeriod)
	actionAtLimit := models.MeteredActionAtLimit(req.ActionAtLimit)
	quota, err := h.planSvc.UpdatePlanFeatureQuota(c, models.UpdatePlanFeatureQuotaInput{
		LimitValue:          req.LimitValue,
		ResetPeriod:         resetPeriod,
		CustomPeriodMinutes: req.CustomPeriodMinutes,
		ActionAtLimit:       actionAtLimit,
	}, planID, featureID)
	if err != nil {
		h.logger.Error("failed to update plan feature quota",
			zap.String("plan_id", planID),
			zap.String("feature_id", featureID),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	return ctx.JSON(models.NewHttpResponse(quota, "plan feature quota updated successfully", fiber.StatusOK))
}
