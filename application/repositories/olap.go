package repositories

import (
	"context"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
)

type OlapRepository interface {
	Connect(cfg *config.ClickHouseConfig) error
	CreateMeter(ctx context.Context, arg models.CreateMeterInput) error
	DeleteMeter(ctx context.Context, organization string, meterSlug string) error
	Close() error
	GetDB() any
}
