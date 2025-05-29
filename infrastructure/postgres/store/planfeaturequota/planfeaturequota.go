package planfeaturequota

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func NewPlanFeatureQuotaRepository(db any, logger *logger.Logger) repositories.PlanFeatureQuotaStoreRepository {
	return &PlanFeatureQuotaRepository{
		q:      gen.New(db.(*pgxpool.Pool)),
		logger: logger,
	}
}

func toPlanFeatureQuotaModel(quota gen.PlanFeatureQuotum) *models.PlanFeatureQuota {
	// Convert UUID to string
	var planFeatureID string
	if quota.PlanFeatureID.Valid {
		if id, err := uuid.FromBytes(quota.PlanFeatureID.Bytes[:]); err == nil {
			planFeatureID = id.String()
		}
	}

	// Initialize the result with the database values
	result := &models.PlanFeatureQuota{
		Base: models.Base{
			ID: quota.ID.Bytes,
			CreatedAt: pgtype.Timestamptz{
				Time:  quota.CreatedAt.Time,
				Valid: quota.CreatedAt.Valid,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  quota.UpdatedAt.Time,
				Valid: quota.UpdatedAt.Valid,
			},
		},
		PlanFeatureID: planFeatureID,
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
