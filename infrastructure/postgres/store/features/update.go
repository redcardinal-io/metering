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
)

func (p *PgFeatureRepository) UpdateFeatureByIDorSlug(ctx context.Context, idOrSlug string, input models.UpdateFeatureInput) (*models.Feature, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	parsedId, parseErr := uuid.Parse(idOrSlug)
	var updateErr error
	var m gen.Feature

	configJson, err := json.Marshal(input.Config)
	if err != nil {
		return nil, postgres.MapError(parseErr, "Postgres.MarshalConfig")
	}

	if parseErr == nil {
		m, updateErr = p.q.UpdateFeatureByID(ctx, gen.UpdateFeatureByIDParams{
			Name:        pgtype.Text{String: input.Name, Valid: input.Name != ""},
			Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
			TenantSlug:  tenantSlug,
			ID:          pgtype.UUID{Bytes: parsedId, Valid: true},
			Config:      configJson,
			UpdatedBy:   input.UpdatedBy,
		})
	} else {
		// Not a UUID, update by slug
		m, updateErr = p.q.UpdateFeatureBySlug(ctx, gen.UpdateFeatureBySlugParams{
			Name:        pgtype.Text{String: input.Name, Valid: input.Name != ""},
			Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
			TenantSlug:  tenantSlug,
			Slug:        idOrSlug,
			UpdatedBy:   input.UpdatedBy,
		})
	}

	if updateErr != nil {
		return nil, postgres.MapError(updateErr, "Postgres.UpdateFeatureByID")
	}

	uuid, parseErr := uuid.FromBytes(m.ID.Bytes[:])
	if parseErr != nil {
		return nil, postgres.MapError(parseErr, "Postgres.ParseUUID")
	}

	config := make(map[string]any)
	_ = json.Unmarshal(m.Config, &config)

	return &models.Feature{
		Name:        m.Name,
		Description: m.Description.String,
		Slug:        m.Slug,
		TenantSlug:  m.TenantSlug,
		Type:        models.FeatureTypeEnum(m.Type),
		Config:      config,
		Base: models.Base{
			ID:        uuid,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
		},
	}, nil
}
