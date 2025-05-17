package features

import (
	"context"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgFeatureRepository) CreateFeature(ctx context.Context, arg models.CreateFeatureInput) (*models.Feature, error) {
	// no need validate tenant slug, it is already validated
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := p.q.CreateFeature(ctx, gen.CreateFeatureParams{
		Name:        arg.Name,
		Description: arg.Description,
		TenantSlug:  tenantSlug,
		CreatedBy:   arg.CreatedBy,
		UpdatedBy:   arg.CreatedBy,
	})
	if err != nil {
		p.logger.Error("failed to create feature", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreateFeature")
	}

	id, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}
}
