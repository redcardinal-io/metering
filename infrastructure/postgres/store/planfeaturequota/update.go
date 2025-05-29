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

func (r *PlanFeatureQuotaRepository) UpdatePlanFeatureQuota(ctx context.Context, arg models.UpdatePlanFeatureQuotaInput) (*models.PlanFeatureQuota, error) {
	// First get the existing quota to ensure it exists
	existingQuota, err := r.GetPlanFeatureQuota(ctx, arg.PlanFeatureID)
	if err != nil {
		return nil, err
	}
	if existingQuota == nil {
		r.logger.Error("plan feature quota not found", zap.String("planFeatureID", arg.PlanFeatureID))
		return nil, errors.New("plan feature quota not found")
	}

	// Parse the plan feature ID
	planFeatureIDStr := arg.PlanFeatureID
	planFeatureID := uuid.MustParse(planFeatureIDStr)

	// Prepare the update parameters
	params := gen.UpdatePlanFeatureQuotaParams{
		PlanFeatureID: pgtype.UUID{Bytes: planFeatureID, Valid: true},
	}

	// Only update the fields that are provided
	if arg.LimitValue != nil {
		params.LimitValue = *arg.LimitValue
	} else {
		params.LimitValue = existingQuota.LimitValue
	}

	if arg.ResetPeriod != nil {
		params.ResetPeriod = gen.MeteredResetPeriodEnum(*arg.ResetPeriod)
	} else {
		params.ResetPeriod = gen.MeteredResetPeriodEnum(existingQuota.ResetPeriod)
	}

	if arg.CustomPeriodMinutes != nil {
		params.CustomPeriodMinutes = pgtype.Int8{Int64: *arg.CustomPeriodMinutes, Valid: true}
	} else if existingQuota.CustomPeriodMinutes != nil {
		params.CustomPeriodMinutes = pgtype.Int8{Int64: *existingQuota.CustomPeriodMinutes, Valid: true}
	}

	if arg.ActionAtLimit != nil {
		params.ActionAtLimit = gen.MeteredActionAtLimitEnum(*arg.ActionAtLimit)
	} else {
		params.ActionAtLimit = gen.MeteredActionAtLimitEnum(existingQuota.ActionAtLimit)
	}

	// Perform the update
	updatedQuota, err := r.q.UpdatePlanFeatureQuota(ctx, params)
	if err != nil {
		r.logger.Error("failed to update plan feature quota", zap.Error(err), zap.String("planFeatureID", planFeatureIDStr))
		return nil, postgres.MapError(err, "failed to update plan feature quota")
	}

	return toPlanFeatureQuotaModel(updatedQuota), nil
}
