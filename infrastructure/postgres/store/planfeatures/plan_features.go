package planfeatures

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgPlanFeatureStoreRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

func NewPgPlanFeatureStoreRepository(db any, logger *logger.Logger) repositories.PlanFeatureStoreRepository {
	return &PgPlanFeatureStoreRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}

func UnMarshalPlanFeatureConfig(configBytes []byte) map[string]any {
	if configBytes == nil {
		return nil
	}

	var config map[string]any
	// INFO: As configBytes is returned from database, it is expected to be in JSON format. So we can unmarshal it directly.
	_ = json.Unmarshal(configBytes, &config)
	return config
}
