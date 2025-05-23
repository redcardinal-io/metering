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

func (s *PgPlanStoreRepository) UpdatePlanByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdatePlanInput) (*models.Plan, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// Try to parse as UUID first
	parsedId, err := uuid.Parse(idOrSlug)
	var updateErr error
	var m gen.Plan
	if err == nil {
		m, updateErr = s.q.UpdatePlanByID(ctx, gen.UpdatePlanByIDParams{
			Name:        pgtype.Text{String: arg.Name, Valid: arg.Name != ""},
			Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
			TenantSlug:  tenantSlug,
			ID:          pgtype.UUID{Bytes: parsedId, Valid: true},
			UpdatedBy:   arg.UpdatedBy,
		})
	} else {
		// Not a UUID, update by slug
		m, updateErr = s.q.UpdatePlanBySlug(ctx, gen.UpdatePlanBySlugParams{
			Name:        pgtype.Text{String: arg.Name, Valid: arg.Name != ""},
			Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
			TenantSlug:  tenantSlug,
			Slug:        idOrSlug,
			UpdatedBy:   arg.UpdatedBy,
		})
	}

	if updateErr != nil {
		return nil, postgres.MapError(updateErr, "Postgres.UpdatePlanByID")
	}

	uuid, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	// Valid UUID, delete by ID
	return &models.Plan{
		Name:        m.Name,
		Description: m.Description.String,
		Slug:        m.Slug,
		Type:        models.PlanTypeEnum(m.Type),
		ArchivedAt:  m.ArchivedAt,
		TenantSlug:  m.TenantSlug,
		Base: models.Base{
			ID:        uuid,
			CreatedAt: m.CreatedAt,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
		},
	}, nil
}
