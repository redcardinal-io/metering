package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/domain/models"
)

type PlanFeatureStoreRepository interface {
	CreatePlanFeature(ctx context.Context, arg models.CreatePlanFeatureInput) (*models.PlanFeature, error)
	UpdatePlanFeature(ctx context.Context, planID, featureID uuid.UUID, arg models.UpdatePlanFeatureInput) (*models.PlanFeature, error)
	DeletePlanFeature(ctx context.Context, arg models.DeletePlanFeatureInput) error
	ListPlanFeaturesByPlan(ctx context.Context, planID uuid.UUID, filter models.PlanFeatureListFilter) (*[]models.PlanFeature, error)
}
