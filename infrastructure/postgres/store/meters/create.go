package meters

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

func (p *PgMeterStoreRepository) CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := p.q.CreateMeter(ctx, gen.CreateMeterParams{
		Slug:          arg.MeterSlug,
		Name:          arg.Name,
		EventType:     pgtype.Text{String: arg.EventType, Valid: true},
		Description:   pgtype.Text{String: arg.Description, Valid: arg.Description != ""},
		ValueProperty: pgtype.Text{String: arg.ValueProperty, Valid: arg.ValueProperty != ""},
		Properties:    arg.Properties,
		Aggregation:   gen.AggregationEnum(arg.Aggregation),
		TenantSlug:    tenantSlug,
		CreatedBy:     arg.CreatedBy,
	})

	if err != nil {
		p.logger.Error("failed to create meter", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreateMeter")
	}

	id, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	meter := &models.Meter{
		Name:          m.Name,
		Slug:          m.Slug,
		ValueProperty: m.ValueProperty.String,
		EventType:     m.EventType.String,
		Description:   m.Description.String,
		Properties:    m.Properties,
		Aggregation:   models.AggregationEnum(m.Aggregation),
		TenantSlug:    m.TenantSlug,
		Base: models.Base{
			ID:        id,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
		},
	}

	return meter, nil
}
