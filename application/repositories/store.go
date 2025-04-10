package repositories

import (
	"context"

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
	CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error)
	GetMeterByIDorSlug(ctx context.Context, idOrSlug string) (*models.Meter, error)
	ListMeters(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Meter], error)
	ListMetersByEventType(ctx context.Context, eventType string, pagination pagination.Pagination) (*pagination.PaginationView[models.Meter], error)
	DeleteMeterByIDorSlug(ctx context.Context, idOrSlug string) error
}
