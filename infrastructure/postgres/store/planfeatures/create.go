package planfeatures

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanFeatureStoreRepository) CreatePlanFeature(ctx context.Context, arg models.CreatePlanFeatureInput) (*models.PlanFeature, error) {
	var configBytes []byte
	var err error

	if arg.Config != nil {
		configBytes, err = json.Marshal(arg.Config)
		if err != nil {
			p.logger.Error("failed to marshal plan feature config", zap.Error(err))
			return nil, postgres.MapError(err, "Postgres.MarshalPlanFeatureConfig")
		}
	}

	m, err := p.q.CreatePlanFeature(ctx, gen.CreatePlanFeatureParams{
		PlanID:    pgtype.UUID{Bytes: arg.PlanID, Valid: true},
		FeatureID: pgtype.UUID{Bytes: arg.FeatureID, Valid: true},
		Config:    configBytes,
		CreatedBy: arg.CreatedBy,
		UpdatedBy: arg.CreatedBy,
	})
	if err != nil {
		p.logger.Error("failed to create plan feature", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreatePlanFeature")
	}

	id, err := uuid.FromBytes(m.PlanFeatureID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	planID, err := uuid.FromBytes(m.PlanID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse plan UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	featureID, err := uuid.FromBytes(m.FeatureID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse feature UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	planFeature := &models.PlanFeature{
		PlanID:    planID,
		FeatureID: featureID,
		Config:    m.Config,
		Base: models.Base{
			ID:        id,
			CreatedAt: m.CreatedAt.Time,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt.Time,
		},
	}

	return planFeature, nil
}
