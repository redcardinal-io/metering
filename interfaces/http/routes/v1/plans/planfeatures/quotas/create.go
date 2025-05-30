package quotas

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"go.uber.org/zap"
)

type createQuotaRequest struct {
	LimitValue          int64  `json:"limit_value" validate:"required"`
	ResetPeriod         string `json:"reset_period" validate:"required,oneof=day week month year custom rolling never"`
	CustomPeriodMinutes *int64 `json:"custom_period_minutes,omitempty" validate:"required_if=ResetPeriod custom,omitempty,gt=0"`
	ActionAtLimit       string `json:"action_at_limit" validate:"required,oneof=none block throttle"`
	CreatedBy           string `json:"created_by" validate:"required"`
}

func (h *httpHandler) create(ctx *fiber.Ctx) error {
	tenantSlug := ctx.Get(constants.TenantHeader)

	planID := ctx.Params("planID")
	featureID := ctx.Params("featureID")

	_, err := uuid.Parse(planID)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid plan ID",
		)
		h.logger.Error("invalid plan ID", zap.String("planID", planID), zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	_, err = uuid.Parse(featureID)
	if err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid feature ID",
		)
		h.logger.Error("invalid feature ID", zap.String("featureID", featureID), zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	var req createQuotaRequest

	if err := ctx.BodyParser(&req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"failed to parse request body",
		)
		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if err := h.validator.Struct(req); err != nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			err,
			domainerrors.EINVALID,
			"invalid request body",
		)
		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ResetPeriod == "custom" && req.CustomPeriodMinutes == nil {
		errResp := domainerrors.NewErrorResponseWithOpts(
			nil,
			domainerrors.EINVALID,
			"custom period minutes is required when reset period is custom",
		)
		h.logger.Error("custom period minutes required", zap.Reflect("error", errResp))
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	if req.ResetPeriod != "custom" && req.CustomPeriodMinutes != nil {
		req.CustomPeriodMinutes = nil // Clear custom period if not using custom reset period
	}

	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenantSlug)
	quota, err := h.planSvc.CreatePlanFeatureQuota(c, models.CreatePlanFeatureQuotaInput{
		LimitValue:          req.LimitValue,
		ResetPeriod:         models.MeteredResetPeriod(req.ResetPeriod),
		CustomPeriodMinutes: req.CustomPeriodMinutes,
		ActionAtLimit:       models.MeteredActionAtLimit(req.ActionAtLimit),
		CreatedBy:           req.CreatedBy,
	}, planID, featureID)
	if err != nil {
		h.logger.Error("failed to create plan feature quota",
			zap.String("plan_id", planID),
			zap.String("feature_id", featureID),
			zap.Error(err))
		errResp := domainerrors.NewErrorResponse(err)
		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
	}

	h.logger.Info("created plan feature quota")

	return ctx.Status(fiber.StatusCreated).JSON(models.NewHttpResponse(quota, "plan feature quota created successfully", fiber.StatusCreated))
}
