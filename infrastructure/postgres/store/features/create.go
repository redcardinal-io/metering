package features

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgFeatureRepository) CreateFeature(ctx context.Context, arg models.CreateFeatureInput) (*models.Feature, error) {
	// no need validate tenant slug, it is already validated
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	configBytes, err := json.Marshal(arg.Config)
	if err != nil {
		p.logger.Error("failed to marshal config", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreateFeature.MarshalConfig")
	}

	m, err := p.q.CreateFeature(ctx, gen.CreateFeatureParams{
		Name:        arg.Name,
		Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
		Slug:        arg.Slug,
		Type:        gen.FeatureEnum(arg.Type),
		Config:      configBytes,
		TenantSlug:  tenantSlug,
		CreatedBy:   arg.CreatedBy,
		UpdatedBy:   arg.CreatedBy,
	})
	if err != nil {
		p.logger.Error("failed to create feature", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreateFeature")
	}

	id, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	config := make(map[string]any)
	_ = json.Unmarshal(m.Config, config)
	return &models.Feature{
		Name:        m.Name,
		Description: m.Description.String,
		Slug:        m.Slug,
		TenantSlug:  m.TenantSlug,
		Type:        models.FeatureTypeEnum(m.Type),
		Config:      config,
		Base: models.Base{
			ID:        id,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
		},
	}, nil
}
