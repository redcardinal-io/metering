package planfeatures

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

func (s *PgPlanFeatureStoreRepository) CheckPlanAndFeatureForTenant(ctx context.Context, planID, featureID uuid.UUID) (bool, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := s.q.CheckPlanAndFeatureForTenant(ctx, gen.CheckPlanAndFeatureForTenantParams{
		ID:         pgtype.UUID{Bytes: planID, Valid: true},
		ID_2:       pgtype.UUID{Bytes: featureID, Valid: true},
		TenantSlug: tenantSlug,
	})
	if err != nil {
		return false, postgres.MapError(err, "Postgres.CheckPlanAndFeatureForTenant")
	}

	return m, nil
}

func (s *PgPlanFeatureStoreRepository) GetPlanFeatureIDByPlanAndFeature(ctx context.Context, planID, featureID uuid.UUID) (uuid.UUID, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	id, err := s.q.GetPlanFeatureIDByPlanAndFeature(ctx, gen.GetPlanFeatureIDByPlanAndFeatureParams{
		PlanID:     pgtype.UUID{Bytes: planID, Valid: true},
		FeatureID:  pgtype.UUID{Bytes: featureID, Valid: true},
		TenantSlug: tenantSlug,
	})
	if err != nil {
		return uuid.Nil, postgres.MapError(err, "Postgres.GetPlanFeatureIDByPlanAndFeature")
	}
	return id.Bytes, nil
}
