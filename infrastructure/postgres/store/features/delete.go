package features

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgFeatureRepository) DeleteFeatureByIDorSlug(ctx context.Context, idOrSlug string) error {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	parsedId, err := uuid.Parse(idOrSlug)
	var deleteErr error
	if err == nil {
		// Valid UUID, delete by ID
		deleteErr = p.q.DeleteFeatureByID(ctx, gen.DeleteFeatureByIDParams{
			ID:         pgtype.UUID{Bytes: parsedId, Valid: true},
			TenantSlug: tenantSlug,
		})
	} else {
		// Not a UUID, delete by slug
		deleteErr = p.q.DeleteFeatureBySlug(ctx, gen.DeleteFeatureBySlugParams{
			Slug:       idOrSlug,
			TenantSlug: tenantSlug,
		})
	}

	if deleteErr != nil {
		p.logger.Error("failed to delete feature", zap.Error(deleteErr))
		return postgres.MapError(deleteErr, "Postgres.DeleteFeature")
	}

	return nil
}
