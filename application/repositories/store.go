package repositories

import (
	"context"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type StoreRepository interface {
	Connect(cfg *config.StoreConfig) error
	Close() error
	GetDB() any
}

type MeterStoreRepository interface {
	CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error)
	GetMeterByIDorSlug(ctx context.Context, idOrSlug string) (*models.Meter, error)
	ListMeters(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Meter], error)
	ListMetersByEventTypes(ctx context.Context, eventTypes []string) ([]*models.Meter, error)
	DeleteMeterByIDorSlug(ctx context.Context, idOrSlug string) error
	UpdateMeterByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdateMeterInput) (*models.Meter, error)
}
