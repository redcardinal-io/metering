package planfeaturequota

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (r *PlanFeatureQuotaRepository) CreatePlanFeatureQuota(ctx context.Context, arg models.CreatePlanFeatureQuotaInput) (*models.PlanFeatureQuota, error) {
	// First check if the feature is metered
	isMetered, err := r.CheckMeteredFeature(ctx, arg.PlanFeatureID)
	if err != nil {
		r.logger.Error("failed to check if feature is metered", zap.Error(err))
		return nil, postgres.MapError(err, "failed to check if feature is metered")
	}
	if !isMetered {
		r.logger.Error("quota can only be set for metered features", zap.String("planFeatureID", arg.PlanFeatureID))
		return nil, errors.New("quota can only be set for metered features")
	}

	planFeatureID, err := uuid.Parse(arg.PlanFeatureID)
	if err != nil {
		return nil, err
	}

	// Convert reset period to the generated enum type
	resetPeriod := gen.MeteredResetPeriodEnum(arg.ResetPeriod)

	// Convert action at limit to the generated enum type
	actionAtLimit := gen.MeteredActionAtLimitEnum(arg.ActionAtLimit)

	// Handle custom period minutes
	var customPeriodMinutes pgtype.Int8
	if arg.CustomPeriodMinutes != nil {
		customPeriodMinutes = pgtype.Int8{Int64: *arg.CustomPeriodMinutes, Valid: true}
	}

	// Create the quota
	quota, err := r.q.CreatePlanFeatureQuota(ctx, gen.CreatePlanFeatureQuotaParams{
		PlanFeatureID:       pgtype.UUID{Bytes: planFeatureID, Valid: true},
		LimitValue:          arg.LimitValue,
		ResetPeriod:         resetPeriod,
		CustomPeriodMinutes: customPeriodMinutes,
		ActionAtLimit:       actionAtLimit,
	})
	if err != nil {
		r.logger.Error("failed to create plan feature quota", zap.Error(err), zap.String("planFeatureID", arg.PlanFeatureID))
		return nil, postgres.MapError(err, "failed to create plan feature quota")
	}

	return toPlanFeatureQuotaModel(quota), nil
}
