package quotas

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"go.uber.org/zap"
)

func (r *PlanFeatureQuotaRepository) GetPlanFeatureQuota(ctx context.Context, planFeatureID uuid.UUID) (*models.PlanFeatureQuota, error) {
	quota, err := r.q.GetPlanFeatureQuotaByPlanFeatureID(ctx, pgtype.UUID{Bytes: planFeatureID, Valid: true})
	if err != nil {
		r.logger.Error("failed to get plan feature quota", zap.Error(err), zap.String("planFeatureID", planFeatureID.String()))
		return nil, postgres.MapError(err, "failed to get plan feature quota")
	}

	return toPlanFeatureQuotaModel(quota), nil
}

func (r *PlanFeatureQuotaRepository) CheckMeteredFeature(ctx context.Context, planFeatureID uuid.UUID) (bool, error) {
	isMetered, err := r.q.CheckMeteredFeature(ctx, pgtype.UUID{Bytes: planFeatureID, Valid: true})
	if err != nil {
		r.logger.Error("failed to check if feature is metered", zap.Error(err), zap.String("planFeatureID", planFeatureID.String()))
		return false, postgres.MapError(err, "failed to check if feature is metered")
	}

	return isMetered, nil
}
