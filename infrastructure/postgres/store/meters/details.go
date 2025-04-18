package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

func (p *PgMeterStoreRepository) GetMeterByIDorSlug(ctx context.Context, idOrSlug string) (*models.Meter, error) {
	// Try to parse as UUID first
	id, err := uuid.Parse(idOrSlug)
	var detailsErr error
	var m gen.Meter
	if err == nil {
		// Valid UUID, get details by ID
		m, detailsErr = p.q.GetMeterByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	} else {
		// Not a UUID, get details by slug
		m, detailsErr = p.q.GetMeterBySlug(ctx, idOrSlug)
	}

	// Handle errors from either get operation
	if detailsErr != nil {
		return nil, postgres.MapError(detailsErr, "Postgres.GetMeterByIDorSlug")
	}

	uuid, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	return &models.Meter{
		ID:            uuid,
		Name:          m.Name,
		Slug:          m.Slug,
		ValueProperty: m.ValueProperty.String,
		EventType:     m.EventType.String,
		Description:   m.Description.String,
		Properties:    m.Properties,
		Aggregation:   models.AggregationEnum(m.Aggregation),
		CreatedAt:     m.CreatedAt.Time,
		CreatedBy:     m.CreatedBy,
	}, nil
}
