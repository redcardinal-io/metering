package planfeaturequota

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"go.uber.org/zap"
)

func (r *PlanFeatureQuotaRepository) GetPlanFeatureQuota(ctx context.Context, planFeatureID string) (*models.PlanFeatureQuota, error) {
	pfID := uuid.MustParse(planFeatureID)

	quota, err := r.q.GetPlanFeatureQuotaByPlanFeatureID(ctx, pgtype.UUID{Bytes: pfID, Valid: true})
	if err != nil {
		r.logger.Error("failed to get plan feature quota", zap.Error(err), zap.String("planFeatureID", planFeatureID))
		return nil, postgres.MapError(err, "failed to get plan feature quota")
	}

	return toPlanFeatureQuotaModel(quota), nil
}

func (r *PlanFeatureQuotaRepository) CheckMeteredFeature(ctx context.Context, planFeatureID string) (bool, error) {
	pfID := uuid.MustParse(planFeatureID)

	isMetered, err := r.q.CheckMeteredFeature(ctx, pgtype.UUID{Bytes: pfID, Valid: true})
	if err != nil {
		r.logger.Error("failed to check if feature is metered", zap.Error(err), zap.String("planFeatureID", planFeatureID))
		return false, postgres.MapError(err, "failed to check if feature is metered")
	}

	return isMetered, nil
}
