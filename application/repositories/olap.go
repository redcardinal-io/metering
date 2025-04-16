package repositories

import (
	"context"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
)

type OlapRepository interface {
	Connect(cfg *config.OlapConfig) error
	Close() error
	GetDB() any

	// meter methods
	CreateMeter(ctx context.Context, arg models.CreateMeterInput) error
	QueryMeter(ctx context.Context, arg models.QueryMeterInput, agg *models.AggregationEnum) (*models.QueryMeterOutput, error)
	DeleteMeter(ctx context.Context, arg models.DeleteMeterInput) error
}
