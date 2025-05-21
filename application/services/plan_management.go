package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type PlanManagementService struct {
	planStore            repositories.PlanStoreRepository
	planAssignmentsStore repositories.PlanAssignmentsStoreRepository
	featureStore         repositories.FeatureStoreRepository
}

// NewPlanService creates a new PlanManagementService with the provided plan and feature repositories.
func NewPlanService(planStore repositories.PlanStoreRepository, featureStore repositories.FeatureStoreRepository, planAssignmentsStore repositories.PlanAssignmentsStoreRepository) *PlanManagementService {
	return &PlanManagementService{
		planStore:            planStore,
		featureStore:         featureStore,
		planAssignmentsStore: planAssignmentsStore,
	}
}

func (s *PlanManagementService) AssignPlan(ctx context.Context, planId uuid.UUID, arg models.AssignOrUpdateAssignedPlanInput, isOrg bool) (*models.PlanAssignment, error) {
	// Assign the plan based on isOrg parameter in the database

	var m *models.PlanAssignment
	var err error

	if isOrg {
		m, err = s.planAssignmentsStore.AssignPlanToOrg(ctx, planId, arg)
	} else {
		m, err = s.planAssignmentsStore.AssignPlanToUser(ctx, planId, arg)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *PlanManagementService) UnAssignPlan(ctx context.Context, planId uuid.UUID, orgOrUserId uuid.UUID, isOrg bool) error {
	var err error

	if isOrg {
		err = s.planAssignmentsStore.UnAssignPlanToOrg(ctx, planId, orgOrUserId)
	} else {
		err = s.planAssignmentsStore.UnAssignPlanToUser(ctx, planId, orgOrUserId)
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *PlanManagementService) UpdateAssignedPlan(ctx context.Context, planId uuid.UUID, arg models.AssignOrUpdateAssignedPlanInput, isOrg bool) (*models.PlanAssignment, error) {
	// Call the store repository to update the assigned plan
	var m *models.PlanAssignment
	var err error

	if isOrg {
		m, err = s.planAssignmentsStore.UpdateAssignedPlanToOrg(ctx, planId, arg)
	} else {
		m, err = s.planAssignmentsStore.UpdateAssignedPlanToUser(ctx, planId, arg)
	}
	if err != nil {
		return nil, err
	}

	return m, nil
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
