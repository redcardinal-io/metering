package plans

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

func (p *PgPlanStoreRepository) GetPlanByIDorSlug(ctx context.Context, idOrSlug string) (*models.Plan, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	parsedId, err := uuid.Parse(idOrSlug)
	var detailsErr error
	var m gen.Plan
	if err == nil {
		// Valid UUID, get details by ID
		m, detailsErr = p.q.GetPlanByID(ctx, gen.GetPlanByIDParams{
			ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
			TenantSlug: tenantSlug,
		})
	} else {
		// Not a UUID, get details by slug
		m, detailsErr = p.q.GetPlanBySlug(ctx, gen.GetPlanBySlugParams{
			Slug:       idOrSlug,
			TenantSlug: tenantSlug,
		})
	}

	// Handle errors from either get operation
	if detailsErr != nil {
		return nil, postgres.MapError(detailsErr, "Postgres.GetPlanByIDorSlug")
	}

	return toPlanModel(m), nil
}
