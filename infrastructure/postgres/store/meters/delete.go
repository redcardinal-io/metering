package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgMeterStoreRepository) DeleteMeterByIDorSlug(ctx context.Context, idOrSlug string) error {
	// Try to parse as UUID first
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	id, err := uuid.Parse(idOrSlug)
	var deleteErr error
	if err == nil {
		// Valid UUID, delete by ID
		deleteErr = p.q.DeleteMeterByID(ctx, gen.DeleteMeterByIDParams{
			ID:         pgtype.UUID{Bytes: id, Valid: true},
			TenantSlug: tenantSlug,
		})
	} else {
		// Not a UUID, delete by slug
		deleteErr = p.q.DeleteMeterBySlug(ctx, gen.DeleteMeterBySlugParams{
			Slug:       idOrSlug,
			TenantSlug: tenantSlug,
		})
	}

	// Handle errors from either delete operation
	if deleteErr != nil {
		p.logger.Error("failed to delete meter", zap.Error(deleteErr))
		return postgres.MapError(deleteErr, "Postgres.DeleteMeter")
	}

	return nil
}
