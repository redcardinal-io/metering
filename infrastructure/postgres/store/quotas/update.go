package quotas

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (r *PlanFeatureQuotaRepository) UpdatePlanFeatureQuota(ctx context.Context, arg models.UpdatePlanFeatureQuotaInput) (*models.PlanFeatureQuota, error) {
	// Parse the plan feature ID
	planFeatureIDStr := arg.PlanFeatureID
	planFeatureID := uuid.MustParse(planFeatureIDStr)

	// Prepare the update parameters
	params := gen.UpdatePlanFeatureQuotaParams{
		PlanFeatureID: pgtype.UUID{Bytes: planFeatureID, Valid: true},
	}

	// Only update the fields that are provided
	if arg.LimitValue != 0 {
		params.LimitValue = pgtype.Int8{Int64: arg.LimitValue, Valid: true}
	}
	if arg.ResetPeriod != "" {
		params.ResetPeriod = gen.NullMeteredResetPeriodEnum{
			Valid:                  true,
			MeteredResetPeriodEnum: gen.MeteredResetPeriodEnum(arg.ResetPeriod),
		}
	}
	if arg.CustomPeriodMinutes != 0 {
		params.CustomPeriodMinutes = pgtype.Int8{Int64: arg.CustomPeriodMinutes, Valid: true}
	}
	if arg.ActionAtLimit != "" {
		params.ActionAtLimit = gen.NullMeteredActionAtLimitEnum{
			Valid:                    true,
			MeteredActionAtLimitEnum: gen.MeteredActionAtLimitEnum(arg.ActionAtLimit),
		}
	}

	// Perform the update
	updatedQuota, err := r.q.UpdatePlanFeatureQuota(ctx, params)
	if err != nil {
		r.logger.Error("failed to update plan feature quota", zap.Error(err), zap.String("planFeatureID", planFeatureIDStr))
		return nil, postgres.MapError(err, "failed to update plan feature quota")
	}

	return toPlanFeatureQuotaModel(updatedQuota), nil
}
