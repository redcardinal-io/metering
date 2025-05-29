package plans

import (
	"context"

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
		Slug:        arg.Slug,
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

	return toPlanModel(m), nil
}
