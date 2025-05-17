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

func (p *PgFeatureRepository) GetFeatureByIDorSlug(ctx context.Context, idOrSlug string) (*models.Feature, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	parsedId, err := uuid.Parse(idOrSlug)
	var detailsErr error
	var m gen.Feature
	if err == nil {
		// Valid UUID, get details by ID
		m, detailsErr = p.q.GetFeatureByID(ctx, gen.GetFeatureByIDParams{
			ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
			TenantSlug: tenantSlug,
		})
	} else {
		// Not a UUID, get details by slug
		m, detailsErr = p.q.GetFeatureBySlug(ctx, gen.GetFeatureBySlugParams{
			Slug:       idOrSlug,
			TenantSlug: tenantSlug,
		})
	}

	if detailsErr != nil {
		return nil, postgres.MapError(detailsErr, "Postgres.GetFeatureByIDorSlug")
	}

	uuid, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
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
