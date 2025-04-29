package meters

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

func (p *PgMeterStoreRepository) ListMeters(ctx context.Context, page pagination.Pagination) (*pagination.PaginationView[models.Meter], error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	m, err := p.q.ListMetersPaginated(ctx, gen.ListMetersPaginatedParams{
		Limit:      int32(page.Limit),
		Offset:     int32(page.GetOffset()),
		TenantSlug: ctx.Value(constants.TenantSlugKey).(string),
	})

	if err != nil {
		p.logger.Error("Error listing meters: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListMeters")
	}

	meters := make([]models.Meter, 0, len(m))
	for _, meter := range m {
		id, _ := uuid.FromBytes(meter.ID.Bytes[:])
		meters = append(meters, models.Meter{
			ID:            id,
			Name:          meter.Name,
			Slug:          meter.Slug,
			ValueProperty: meter.ValueProperty.String,
			EventType:     meter.EventType.String,
			Description:   meter.Description.String,
			Properties:    meter.Properties,
			Aggregation:   models.AggregationEnum(meter.Aggregation),
			CreatedAt:     meter.CreatedAt.Time,
			TenantSlug:    meter.TenantSlug,
		})
	}

	count, err := p.q.CountMeters(ctx, tenantSlug)
	if err != nil {
		p.logger.Error("Error counting meters: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountMeters")
	}

	result := pagination.FormatWith(page, int(count), meters)

	return &result, nil
}

func (p *PgMeterStoreRepository) ListMetersByEventTypes(
	ctx context.Context,
	eventTypes []string,
) ([]*models.Meter, error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)

	metesrs, err := p.q.ListMetersByEventTypes(ctx, gen.ListMetersByEventTypesParams{
		Column1:    eventTypes,
		TenantSlug: tenantSlug,
	})
	if err != nil {
		p.logger.Error("Error listing meters by event type: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListMetersByEventType")
	}

	meters := make([]*models.Meter, 0, len(metesrs))
	for _, meter := range metesrs {
		id, _ := uuid.FromBytes(meter.ID.Bytes[:])
		meters = append(meters, &models.Meter{
			ID:            id,
			Name:          meter.Name,
			Slug:          meter.Slug,
			ValueProperty: meter.ValueProperty.String,
			EventType:     meter.EventType.String,
			Description:   meter.Description.String,
			Properties:    meter.Properties,
			Aggregation:   models.AggregationEnum(meter.Aggregation),
			CreatedAt:     meter.CreatedAt.Time,
			TenantSlug:    meter.TenantSlug,
		})
	}

	return meters, nil
}
