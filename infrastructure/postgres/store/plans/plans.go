package plans

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
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

func toPlanModel(m gen.Plan) *models.Plan {
	return &models.Plan{
		Name:        m.Name,
		Slug:        m.Slug,
		Type:        models.PlanTypeEnum(m.Type),
		Description: m.Description.String,
		ArchivedAt:  m.ArchivedAt,
		TenantSlug:  m.TenantSlug,
		Base: models.Base{
			ID:        uuid.UUID(m.ID.Bytes),
			CreatedAt: m.CreatedAt.Time,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt.Time,
		},
	}
}
