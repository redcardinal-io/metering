// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CountMeters(ctx context.Context) (int64, error)
	CountMetersByEventType(ctx context.Context, eventType pgtype.Text) (int64, error)
	CreateMeter(ctx context.Context, arg CreateMeterParams) (Meter, error)
	DeleteMeterByID(ctx context.Context, id pgtype.UUID) error
	DeleteMeterBySlug(ctx context.Context, slug string) error
	GetEventTypes(ctx context.Context) ([]pgtype.Text, error)
	GetMeterByID(ctx context.Context, id pgtype.UUID) (Meter, error)
	GetMeterBySlug(ctx context.Context, slug string) (Meter, error)
	GetPropertiesByEventType(ctx context.Context, eventType pgtype.Text) ([]interface{}, error)
	GetValuePropertiesByEventType(ctx context.Context, eventType pgtype.Text) ([]pgtype.Text, error)
	ListMetersPaginated(ctx context.Context, arg ListMetersPaginatedParams) ([]Meter, error)
	ListMetersPaginatedByEventType(ctx context.Context, arg ListMetersPaginatedByEventTypeParams) ([]Meter, error)
}

var _ Querier = (*Queries)(nil)
