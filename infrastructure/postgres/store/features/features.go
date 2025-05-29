package features

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
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

func toFeatureModel(m gen.Feature) *models.Feature {
	config := make(map[string]any)
	_ = json.Unmarshal(m.Config, &config)
	return &models.Feature{
		Name:        m.Name,
		Description: m.Description.String,
		Slug:        m.Slug,
		TenantSlug:  m.TenantSlug,
		Type:        models.FeatureTypeEnum(m.Type),
		Config:      config,
		Base: models.Base{
			ID:        uuid.UUID(m.ID.Bytes),
			CreatedAt: m.CreatedAt.Time,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt.Time,
		},
	}
}
