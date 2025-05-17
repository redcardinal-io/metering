package services

import (
	"context"

	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type PlanManagementService struct {
	planStore    repositories.PlanStoreRepository
	featureStore repositories.FeatureStoreRepository
}

// NewPlanService returns a new PlanService that uses the given PlanStoreRepository for plan operations.
func NewPlanService(planStore repositories.PlanStoreRepository, featureStore repositories.FeatureStoreRepository) *PlanManagementService {
	return &PlanManagementService{
		planStore:    planStore,
		featureStore: featureStore,
	}
}

func (s *PlanManagementService) CreatePlan(ctx context.Context, arg models.CreatePlanInput) (*models.Plan, error) {
	// Store the plan in the database
	m, err := s.planStore.CreatePlan(ctx, arg)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanManagementService) GetPlanByIDorSlug(ctx context.Context, IDorSlug string) (*models.Plan, error) {
	// Call the repository to get the plan
	m, err := s.planStore.GetPlanByIDorSlug(ctx, IDorSlug)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *PlanManagementService) DeletePlanIDOrSlug(ctx context.Context, IDorSlug string) error {
	err := s.planStore.DeletePlanByIDorSlug(ctx, IDorSlug)
	if err != nil {
		return err
	}

	return nil
}

func (s *PlanManagementService) ArchivePlanByIDorSlug(ctx context.Context, IDorSlug string, arg models.ArchivePlanInput) error {
	err := s.planStore.ArchivePlanByIDorSlug(ctx, IDorSlug, arg)
	if err != nil {
		return err
	}

	return nil
}

func (s *PlanManagementService) UpdatePlanByIDorSlug(ctx context.Context, IDorSlug string, arg models.UpdatePlanInput) (*models.Plan, error) {
	// Call the store repository to update the plan
	m, err := s.planStore.UpdatePlanByIDorSlug(ctx, IDorSlug, arg)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanManagementService) ListPlans(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Plan], error) {
	// Call the store repository to list the plans
	m, err := s.planStore.ListPlans(ctx, pagination)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanManagementService) CreateFeature(ctx context.Context, arg models.CreateFeatureInput) (*models.Feature, error) {
	return s.featureStore.CreateFeature(ctx, arg)
}

func (s *PlanManagementService) GetFeatureByIDorSlug(ctx context.Context, idOrSlug string) (*models.Feature, error) {
	return s.featureStore.GetFeatureByIDorSlug(ctx, idOrSlug)
}

func (s *PlanManagementService) DeleteFeatureByIDorSlug(ctx context.Context, idOrSlug string) error {
	return s.featureStore.DeleteFeatureByIDorSlug(ctx, idOrSlug)
}

func (s *PlanManagementService) UpdateFeatureByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdateFeatureInput) (*models.Feature, error) {
	return s.featureStore.UpdateFeatureByIDorSlug(ctx, idOrSlug, arg)
}

func (s *PlanManagementService) ListFeatures(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Feature], error) {
	return s.featureStore.ListFeatures(ctx, pagination)
}
