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

func (p *PgPlanStoreRepository) CreatePlan(ctx context.Context, arg models.CreatePlanInput) (*models.Plan, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := p.q.CreatePlan(ctx, gen.CreatePlanParams{
		Name:        arg.Name,
		Slug:        arg.PlanSlug,
		Type:        gen.PlanTypeEnum(arg.Type),
		Description: pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
		TenantSlug:  tenantSlug,
		CreatedBy:   arg.CreatedBy,
		UpdatedBy:   arg.CreatedBy,
	})

	if err != nil {
		p.logger.Error("failed to create plan", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreatePlan")
	}

	id, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	plan := &models.Plan{
		Name:        m.Name,
		Slug:        m.Slug,
		Type:        models.PlanTypeEnum(m.Type),
		Description: m.Description.String,
		ArchivedAt:  m.ArchivedAt,
		TenantSlug:  m.TenantSlug,
		Base: models.Base{
			ID:        id,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
		},
	}

	return plan, nil
}
