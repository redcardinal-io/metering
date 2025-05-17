package plans

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanStoreRepository) DeletePlanByIDorSlug(ctx context.Context, idOrSlug string) error {
	// Try to parse as UUID first
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	parsedId, err := uuid.Parse(idOrSlug)
	var deleteErr error
	if err == nil {
		// Valid UUID, delete by ID
		deleteErr = p.q.DeletePlanByID(ctx, gen.DeletePlanByIDParams{
			ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
			TenantSlug: tenantSlug,
		})
	} else {
		// Not a UUID, delete by slug
		deleteErr = p.q.DeletePlanBySlug(ctx, gen.DeletePlanBySlugParams{
			Slug:       idOrSlug,
			TenantSlug: tenantSlug,
		})
	}

	// Handle errors from either delete operation
	if deleteErr != nil {
		p.logger.Error("failed to delete plan", zap.Error(deleteErr))
		return postgres.MapError(deleteErr, "Postgres.DeletePlan")
	}

	return nil
}
