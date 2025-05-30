package quotas

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"go.uber.org/zap"
)

func (r *PlanFeatureQuotaRepository) DeletePlanFeatureQuota(ctx context.Context, planFeatureID uuid.UUID) error {
	err := r.q.DeletePlanFeatureQuota(ctx, pgtype.UUID{Bytes: planFeatureID, Valid: true})
	if err != nil {
		r.logger.Error("failed to delete plan feature quota", zap.Error(err), zap.String("planFeatureID", planFeatureID.String()))
		return postgres.MapError(err, "failed to delete plan feature quota")
	}

	return nil
}
