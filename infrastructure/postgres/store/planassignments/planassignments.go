package planassignments

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgPlanAssignmentsStoreRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

// NewPostgresPlanAssignmentsStoreRepository creates a new PostgreSQL-backed plan assignments store repository using the provided database connection and logger.
func NewPostgresPlanAssignmentsStoreRepository(db any, logger *logger.Logger) repositories.PlanAssignmentsStoreRepository {
	return &PgPlanAssignmentsStoreRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}
