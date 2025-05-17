package plans

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanStoreRepository) ArchivePlanByIDorSlug(ctx context.Context, idOrSlug string, arg models.ArchivePlanInput) error {
	// Try to parse as UUID first
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	parsedId, err := uuid.Parse(idOrSlug)
	var archiveErr error
	if err == nil {
		// Valid UUID, get details by ID
		if arg.Archive {
			_, archiveErr = p.q.ArchivePlanByID(ctx, gen.ArchivePlanByIDParams{
				ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
				UpdatedBy:  arg.UpdatedBy,
				TenantSlug: tenantSlug,
			})
		} else {
			_, archiveErr = p.q.UnArchivePlanByID(ctx, gen.UnArchivePlanByIDParams{
				ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
				UpdatedBy:  arg.UpdatedBy,
				TenantSlug: tenantSlug,
			})
		}

	} else {
		// Not a UUID, get details by slug
		if arg.Archive {
			_, archiveErr = p.q.ArchivePlanBySlug(ctx, gen.ArchivePlanBySlugParams{
				Slug:       idOrSlug,
				TenantSlug: tenantSlug,
				UpdatedBy:  arg.UpdatedBy,
			})
		} else {
			_, archiveErr = p.q.UnArchivePlanBySlug(ctx, gen.UnArchivePlanBySlugParams{
				Slug:       idOrSlug,
				TenantSlug: tenantSlug,
				UpdatedBy:  arg.UpdatedBy,
			})
		}
	}

	// Handle errors from either delete operation
	if archiveErr != nil {
		p.logger.Error("failed to archive plan", zap.Error(archiveErr))
		return postgres.MapError(archiveErr, "Postgres.ArchivePlan")
	}

	return nil
}
