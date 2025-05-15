package services

import (
	"context"

	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type PlanService struct {
	store repositories.PlanStoreRepository
}

// NewPlanService returns a new PlanService that uses the given PlanStoreRepository for plan operations.
func NewPlanService(store repositories.PlanStoreRepository) *PlanService {
	return &PlanService{
		store: store,
	}
}

func (s *PlanService) CreatePlan(ctx context.Context, arg models.CreatePlanInput) (*models.Plan, error) {
	// Store the plan in the database
	m, err := s.store.CreatePlan(ctx, arg)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanService) GetPlanByID(ctx context.Context, ID string) (*models.Plan, error) {
	// Call the repository to get the plan
	m, err := s.store.GetPlanByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *PlanService) DeletePlan(ctx context.Context, ID string) error {

	err := s.store.DeletePlanByID(ctx, ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PlanService) UpdatePlan(ctx context.Context, ID string, arg models.UpdatePlanInput) (*models.Plan, error) {
	// Call the store repository to update the plan
	m, err := s.store.UpdatePlanByID(ctx, ID, arg)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanService) ListPlans(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Plan], error) {
	// Call the store repository to list the plans
	m, err := s.store.ListPlans(ctx, pagination)
	if err != nil {
		return nil, err
	}

	return m, nil
}
