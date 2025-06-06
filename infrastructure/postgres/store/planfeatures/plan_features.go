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

// NewPgPlanFeatureStoreRepository creates a new PostgreSQL-backed plan feature store repository using the provided database connection and logger.
func NewPgPlanFeatureStoreRepository(db any, logger *logger.Logger) repositories.PlanFeatureStoreRepository {
	return &PgPlanFeatureStoreRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}

// UnMarshalPlanFeatureConfig parses JSON-encoded configuration data into a map.
// Returns nil if the input is nil or if unmarshaling fails.
func UnMarshalPlanFeatureConfig(configBytes []byte) map[string]any {
	if configBytes == nil {
		return nil
	}

	var config map[string]any
	// INFO: As configBytes is returned from database, it is expected to be in JSON format. So we can unmarshal it directly.
	_ = json.Unmarshal(configBytes, &config)
	return config
}
