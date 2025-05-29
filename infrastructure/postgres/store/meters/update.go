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
			Name:        pgtype.Text{String: arg.Name, Valid: arg.Name != ""},
			Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
			TenantSlug:  tenantSlug,
			ID:          pgtype.UUID{Bytes: id, Valid: true},
			UpdatedBy:   arg.UpdatedBy,
		})
	} else {
		// Not a UUID, update by slug
		m, updateErr = s.q.UpdateMeterBySlug(ctx, gen.UpdateMeterBySlugParams{
			Name:        pgtype.Text{String: arg.Name, Valid: arg.Name != ""},
			Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
			TenantSlug:  tenantSlug,
			Slug:        idOrSlug,
			UpdatedBy:   arg.UpdatedBy,
		})
	}

	if updateErr != nil {
		return nil, postgres.MapError(updateErr, "Postgres.UpdateMeterByIDorSlug")
	}

	// Valid UUID, delete by ID
	return toMeterModel(m), nil
}
