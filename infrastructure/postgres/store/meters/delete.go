package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/errors"
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
		if pgErr, ok := deleteErr.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return errors.ErrMeterNotFound
		}
		return errors.ErrDatabaseOperation
	}

	return nil
}
