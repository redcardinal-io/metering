package meters

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgMeterStoreRepository) CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := p.q.CreateMeter(ctx, gen.CreateMeterParams{
		Slug:          arg.MeterSlug,
		Name:          arg.Name,
		EventType:     arg.EventType,
		Description:   pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
		ValueProperty: pgtype.Text{String: arg.ValueProperty, Valid: arg.ValueProperty != ""},
		Properties:    arg.Properties,
		Aggregation:   gen.AggregationEnum(arg.Aggregation),
		TenantSlug:    tenantSlug,
		CreatedBy:     arg.CreatedBy,
		UpdatedBy:     arg.CreatedBy,
	})
	if err != nil {
		p.logger.Error("failed to create meter", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreateMeter")
	}

	return toMeterModel(m), nil
}
