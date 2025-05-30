package features

import (
	"context"

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

	return toFeatureModel(m), nil
}
