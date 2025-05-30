package quotas

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
)

type PlanFeatureQuotaRepository struct {
	q      *gen.Queries
	logger *logger.Logger
}

// NewPlanFeatureQuotaRepository creates a new PlanFeatureQuotaRepository using the provided database connection and logger.
func NewPlanFeatureQuotaRepository(db any, logger *logger.Logger) repositories.PlanFeatureQuotaStoreRepository {
	return &PlanFeatureQuotaRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}

// toPlanFeatureQuotaModel converts a PlanFeatureQuotum database record into a PlanFeatureQuota domain model.
// It maps all relevant fields and sets CustomPeriodMinutes if present in the database record.
func toPlanFeatureQuotaModel(quota gen.PlanFeatureQuotum) *models.PlanFeatureQuota {
	// Initialize the result with the database values
	result := &models.PlanFeatureQuota{
		Base: models.Base{
			ID:        uuid.UUID(quota.ID.Bytes),
			CreatedAt: quota.CreatedAt.Time,
			UpdatedAt: quota.UpdatedAt.Time,
		},
		PlanFeatureID: quota.PlanFeatureID.String(),
		LimitValue:    quota.LimitValue,
		ResetPeriod:   models.MeteredResetPeriod(quota.ResetPeriod),
		ActionAtLimit: models.MeteredActionAtLimit(quota.ActionAtLimit),
	}

	// Set custom period minutes if valid
	if quota.CustomPeriodMinutes.Valid {
		customPeriod := quota.CustomPeriodMinutes.Int64
		result.CustomPeriodMinutes = &customPeriod
	}

	return result
}
