package plans

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgPlanStoreRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

// NewPostgresPlanStoreRepository creates a new PostgreSQL-backed plan store repository using the provided database connection and logger.
func NewPostgresPlanStoreRepository(db any, logger *logger.Logger) repositories.PlanStoreRepository {
	return &PgPlanStoreRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}
