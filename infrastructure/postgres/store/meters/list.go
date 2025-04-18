package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgMeterStoreRepository) ListMeters(ctx context.Context, page pagination.Pagination) (*pagination.PaginationView[models.Meter], error) {

	m, err := p.q.ListMetersPaginated(ctx, gen.ListMetersPaginatedParams{
		Limit:  int32(page.Limit),
		Offset: int32(page.GetOffset()),
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
			CreatedBy:     meter.CreatedBy,
		})
	}

	count, err := p.q.CountMeters(ctx)
	if err != nil {
		p.logger.Error("Error counting meters: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountMeters")
	}

	result := pagination.FormatWith(page, int(count), meters)

	return &result, nil
}

func (p *PgMeterStoreRepository) ListMetersByEventType(
	ctx context.Context,
	eventType string,
	page pagination.Pagination) (*pagination.PaginationView[models.Meter], error) {

	m, err := p.q.ListMetersPaginatedByEventType(ctx, gen.ListMetersPaginatedByEventTypeParams{
		EventType: pgtype.Text{String: eventType, Valid: true},
		Limit:     int32(page.Limit),
		Offset:    int32(page.GetOffset()),
	})
	if err != nil {
		p.logger.Error("Error listing meters by event type: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListMetersByEventType")
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
			CreatedBy:     meter.CreatedBy,
		})
	}

	count, err := p.q.CountMetersByEventType(ctx, pgtype.Text{String: eventType, Valid: true})
	if err != nil {
		p.logger.Error("Error counting meters by event type: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountMetersByEventType")
	}

	result := pagination.FormatWith(page, int(count), meters)
	return &result, nil
}
