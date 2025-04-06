package repositories

import (
	"context"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
)

type OlapRepository interface {
	Connect(cfg *config.ClickHouseConfig) error
	Close() error
	GetDB() any
}

type OlapMeterRepository interface {
	CreateMaterializedView(ctx context.Context, arg models.MaterializedView) error
}
