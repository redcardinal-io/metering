package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"go.uber.org/zap"
)

func (p *PgMeterStoreRepository) DeleteMeterByIDorSlug(ctx context.Context, idOrSlug string) error {
	// Try to parse as UUID first
	id, err := uuid.Parse(idOrSlug)
	var deleteErr error
	if err == nil {
		// Valid UUID, delete by ID
		deleteErr = p.q.DeleteMeterByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	} else {
		// Not a UUID, delete by slug
		deleteErr = p.q.DeleteMeterBySlug(ctx, idOrSlug)
	}

	// Handle errors from either delete operation
	if deleteErr != nil {
		p.logger.Error("failed to delete meter", zap.Error(deleteErr))
		return postgres.MapError(deleteErr, "Postgres.DeleteMeter")
	}

	return nil
}
