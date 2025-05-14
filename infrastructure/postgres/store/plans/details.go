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

func (p *PgPlanStoreRepository) GetPlanByID(ctx context.Context, id string) (*models.Plan, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	parsedId, err := uuid.Parse(id)
	var detailsErr error
	var m gen.Plan
	if err == nil {
		// Valid UUID, get details by ID
		m, detailsErr = p.q.GetPlanByID(ctx, gen.GetPlanByIDParams{
			ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
			TenantSlug: tenantSlug,
		})
	}

	// Handle errors from either get operation
	if detailsErr != nil {
		return nil, postgres.MapError(detailsErr, "Postgres.GetPlanByID")
	}

	uuid, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	return &models.Plan{
		Name:        m.Name,
		Description: m.Description.String,
		TenantSlug:  m.TenantSlug,
		Base: models.Base{
			ID:        uuid,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
		},
	}, nil
}
