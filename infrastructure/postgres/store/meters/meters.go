package meters

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
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

func toMeterModel(m gen.Meter) *models.Meter {
	return &models.Meter{
		Name:          m.Name,
		Slug:          m.Slug,
		ValueProperty: m.ValueProperty.String,
		EventType:     m.EventType,
		Description:   m.Description.String,
		Properties:    m.Properties,
		Aggregation:   models.AggregationEnum(m.Aggregation),
		TenantSlug:    m.TenantSlug,
		Base: models.Base{
			ID:        uuid.UUID(m.ID.Bytes),
			CreatedAt: m.CreatedAt.Time,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt.Time,
		},
	}
}
