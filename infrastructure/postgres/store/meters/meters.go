package meters

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgMeterStoreRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

func NewPostgresMeterStoreRepository(db any, logger *logger.Logger) repositories.MeterStoreRepository {
	return &PgMeterStoreRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}
