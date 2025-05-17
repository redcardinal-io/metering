package features

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgFeatureRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

// NewPgFeatureStoreRepository creates a new PostgreSQL-backed feature repository using the provided database connection and logger.
func NewPgFeatureStoreRepository(db any, logger *logger.Logger) repositories.FeatureStoreRepository {
	return &PgFeatureRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}
