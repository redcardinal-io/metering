package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

func (p *PgMeterStoreRepository) GetMeterByIDorSlug(ctx context.Context, idOrSlug string) (*models.Meter, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	id, err := uuid.Parse(idOrSlug)
	var detailsErr error
	var m gen.Meter
	if err == nil {
		// Valid UUID, get details by ID
		m, detailsErr = p.q.GetMeterByID(ctx, gen.GetMeterByIDParams{
			ID:         pgtype.UUID{Bytes: id, Valid: true},
			TenantSlug: tenantSlug,
		})
	} else {
		// Not a UUID, get details by slug
		m, detailsErr = p.q.GetMeterBySlug(ctx, gen.GetMeterBySlugParams{
			Slug:       idOrSlug,
			TenantSlug: tenantSlug,
		})
	}

	// Handle errors from either get operation
	if detailsErr != nil {
		return nil, postgres.MapError(detailsErr, "Postgres.GetMeterByIDorSlug")
	}

	return toMeterModel(m), nil
}
