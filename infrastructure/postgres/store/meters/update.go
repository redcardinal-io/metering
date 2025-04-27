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

func (s *PgMeterStoreRepository) UpdateMeterByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdateMeterInput) (*models.Meter, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	id, err := uuid.Parse(idOrSlug)
	var updateErr error
	var m gen.Meter
	if err == nil {
		m, updateErr = s.q.UpdateMeterByID(ctx, gen.UpdateMeterByIDParams{
			Column1:     arg.Name,
			Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
			TenantSlug:  tenantSlug,
			ID:          pgtype.UUID{Bytes: id, Valid: true},
		})
	} else {
		// Not a UUID, update by slug
		m, updateErr = s.q.UpdateMeterBySlug(ctx, gen.UpdateMeterBySlugParams{
			Column1:     arg.Name,
			Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
			TenantSlug:  tenantSlug,
			Slug:        idOrSlug,
		})
	}

	if updateErr != nil {
		return nil, postgres.MapError(updateErr, "Postgres.UpdateMeterByIDorSlug")
	}

	uuid, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	// Valid UUID, delete by ID
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
		TenantSlug:    m.TenantSlug,
	}, nil
}
