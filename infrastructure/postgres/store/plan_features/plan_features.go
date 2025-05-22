package plan_features

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgPlanFeatureStoreRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

// NewPostgresPlanFeatureStoreRepository creates a new PostgreSQL-backed plan feature store repository using the provided database connection and logger.
func NewPostgresPlanFeatureStoreRepository(db any, logger *logger.Logger) repositories.PlanFeatureStoreRepository {
	return &PgPlanFeatureStoreRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}
