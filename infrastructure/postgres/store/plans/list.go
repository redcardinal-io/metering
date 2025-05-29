package plans

import (
	"context"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanStoreRepository) ListPlans(ctx context.Context, page pagination.Pagination) (*pagination.PaginationView[models.Plan], error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := p.q.ListPlansPaginated(ctx, gen.ListPlansPaginatedParams{
		Limit:      int32(page.Limit),
		Offset:     int32(page.GetOffset()),
		TenantSlug: ctx.Value(constants.TenantSlugKey).(string),
		Type:       createPlanTypeEnum(page.Queries["type"]),
	})
	if err != nil {
		p.logger.Error("Error listing plans: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListPlans")
	}

	plans := make([]models.Plan, 0, len(m))
	for _, plan := range m {
		plans = append(plans, *toPlanModel(plan))
	}

	count, err := p.q.CountPlans(ctx, gen.CountPlansParams{
		TenantSlug: tenantSlug,
		Type:       createPlanTypeEnum(page.Queries["type"]),
	})
	if err != nil {
		p.logger.Error("Error counting plans: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountMeters")
	}

	result := pagination.FormatWith(page, int(count), plans)

	return &result, nil
}

func createPlanTypeEnum(planType string) gen.NullPlanTypeEnum {
	return gen.NullPlanTypeEnum{
		PlanTypeEnum: gen.PlanTypeEnum(planType),
		Valid:        planType != "",
	}
}
