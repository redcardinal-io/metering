package features

import (
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PgFeatureRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}
