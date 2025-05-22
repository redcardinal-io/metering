package plan_features

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanFeatureStoreRepository) DeletePlanFeature(ctx context.Context, arg models.DeletePlanFeatureInput) error {
	err := p.q.DeletePlanFeature(ctx, gen.DeletePlanFeatureParams{
		PlanID:    pgtype.UUID{Bytes: arg.PlanID, Valid: true},
		FeatureID: pgtype.UUID{Bytes: arg.FeatureID, Valid: true},
	})
	if err != nil {
		p.logger.Error("failed to delete plan feature", zap.Error(err))
		return postgres.MapError(err, "Postgres.DeletePlanFeature")
	}

	return nil
}
