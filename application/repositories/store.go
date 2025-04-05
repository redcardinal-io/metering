package repositories

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type StoreRepository interface {
	Connect(cfg *config.PostgresConfig) error
	Close() error
	GetDB() any
}

type MeterStoreRepository interface {
	CreateMeter(ctx context.Context, arg models.CreateMeterInput) (models.Meter, error)
	GetMeterByID(ctx context.Context, id uuid.UUID) (models.Meter, error)
	GetMeterBySlug(ctx context.Context, slug string) (models.Meter, error)
	ListMeters(ctx context.Context, limit int, page int) (pagination.Result[models.Meter], error)
	ListMetersByEventType(ctx context.Context, eventType string) ([]models.Meter, error)
	DeleteMeterByID(ctx context.Context, id uuid.UUID) error
	DeleteMeterBySlug(ctx context.Context, slug string) error
}
