package meters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgMeterStoreRepository) ListMeters(ctx context.Context, limit int32, cursor *pagination.Cursor) (*pagination.Result[models.Meter], error) {
	var params gen.ListMetersCursorPaginatedParams

	if cursor != nil {
		id := uuid.MustParse(cursor.ID)
		params.CursorID = pgtype.UUID{Bytes: id, Valid: true}
		params.Limit = limit
		params.UseCursor = true
		params.CursorTime = pgtype.Timestamptz{Time: cursor.Time, Valid: true}
	} else {
		params.Limit = limit
		params.UseCursor = false
	}

	m, err := p.q.ListMetersCursorPaginated(ctx, params)
	if err != nil {
		p.logger.Error("Error listing meters: ", zap.Error(err))
		return nil, errors.ErrDatabaseOperation
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

	result := pagination.NewResult(meters)

	return &result, nil
}

func (p *PgMeterStoreRepository) ListMetersByEventType(
	ctx context.Context,
	limit int32,
	eventType string,
	cursor *pagination.Cursor) (*pagination.Result[models.Meter], error) {
	var params gen.ListMetersCursorPaginatedByEventTypeParams

	params.EventType = pgtype.Text{String: eventType, Valid: true}
	if cursor != nil {
		id := uuid.MustParse(cursor.ID)
		params.CursorID = pgtype.UUID{Bytes: id, Valid: true}
		params.Limit = limit
		params.UseCursor = true
		params.CursorTime = pgtype.Timestamptz{Time: cursor.Time, Valid: true}
	} else {
		params.Limit = limit
		params.UseCursor = false
	}

	m, err := p.q.ListMetersCursorPaginatedByEventType(ctx, params)
	if err != nil {
		p.logger.Error("Error listing meters by event type: ", zap.Error(err))
		return nil, errors.ErrDatabaseOperation
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

	result := pagination.NewResult(meters)
	return &result, nil
}
