package planfeatures

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanFeatureStoreRepository) ListPlanFeaturesByPlan(ctx context.Context, planID uuid.UUID, filter models.PlanFeatureListFilter) ([]models.PlanFeature, error) {
	params := gen.ListPlanFeaturesByPlanParams{
		PlanID: pgtype.UUID{Bytes: planID, Valid: true},
	}
	if filter.FeatureType != "" {
		params.FeatureType = gen.NullFeatureEnum{
			FeatureEnum: gen.FeatureEnum(filter.FeatureType),
			Valid:       true,
		}
	}

	rows, err := p.q.ListPlanFeaturesByPlan(ctx, params)
	if err != nil {
		p.logger.Error("Error listing plan features: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListPlanFeatures")
	}

	planFeatures := make([]models.PlanFeature, 0, len(rows))
	for _, row := range rows {
		id, err := uuid.FromBytes(row.PlanFeatureID.Bytes[:])
		if err != nil {
			p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
			return nil, postgres.MapError(err, "Postgres.ParseUUID")
		}

		planIDResult, err := uuid.FromBytes(row.PlanID.Bytes[:])
		if err != nil {
			p.logger.Error("failed to parse plan UUID from bytes", zap.Error(err))
			return nil, postgres.MapError(err, "Postgres.ParseUUID")
		}

		featureIDResult, err := uuid.FromBytes(row.FeatureID.Bytes[:])
		if err != nil {
			p.logger.Error("failed to parse feature UUID from bytes", zap.Error(err))
			return nil, postgres.MapError(err, "Postgres.ParseUUID")
		}

		planFeatures = append(planFeatures, models.PlanFeature{
			PlanID:      planIDResult,
			FeatureID:   featureIDResult,
			Config:      row.Config,
			FeatureName: row.FeatureName,
			FeatureSlug: row.FeatureSlug,
			Type:        models.FeatureTypeEnum(row.FeatureType),
			Base: models.Base{
				ID:        id,
				CreatedAt: row.CreatedAt,
				CreatedBy: row.CreatedBy,
				UpdatedAt: row.UpdatedAt,
				UpdatedBy: row.UpdatedBy,
			},
		})
	}

	return planFeatures, nil
}
