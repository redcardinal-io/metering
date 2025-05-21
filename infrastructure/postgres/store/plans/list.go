package plans

import (
	"context"

	"github.com/google/uuid"
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
		id, err := uuid.FromBytes(plan.ID.Bytes[:])
		if err != nil {
			p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
			return nil, postgres.MapError(err, "Postgres.ParseUUID")
		}
		plans = append(plans, models.Plan{
			Name:        plan.Name,
			Slug:        plan.Slug,
			Type:        models.PlanTypeEnum(plan.Type),
			ArchivedAt:  plan.ArchivedAt,
			Description: plan.Description.String,
			TenantSlug:  plan.TenantSlug,
			Base: models.Base{
				ID:        id,
				CreatedAt: plan.CreatedAt,
				CreatedBy: plan.CreatedBy,
				UpdatedBy: plan.UpdatedBy,
				UpdatedAt: plan.UpdatedAt,
			},
		})
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
