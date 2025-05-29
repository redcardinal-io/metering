package planassignments

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
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

func toPlanAssignmentModel(m gen.PlanAssignment) *models.PlanAssignment {
	return &models.PlanAssignment{
		Base: models.Base{
			ID:        uuid.UUID(m.ID.Bytes),
			CreatedAt: m.CreatedAt.Time,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt.Time,
		},
		PlanID:         m.PlanID.String(),
		OrganizationID: m.OrganizationID.String,
		UserID:         m.UserID.String,
		ValidFrom:      m.ValidFrom.Time,
		ValidUntil:     m.ValidUntil.Time,
	}
}
