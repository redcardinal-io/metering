package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/application/repositories"
	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type PlanManagementService struct {
	planStore            repositories.PlanStoreRepository
	planAssignmentsStore repositories.PlanAssignmentsStoreRepository
	featureStore         repositories.FeatureStoreRepository
	planFeatureStore     repositories.PlanFeatureStoreRepository
	planFeatureQuotaRepo repositories.PlanFeatureQuotaStoreRepository
}

func NewPlanService(
	planStore repositories.PlanStoreRepository,
	featureStore repositories.FeatureStoreRepository,
	planFeatureStore repositories.PlanFeatureStoreRepository,
	planAssignmentsStore repositories.PlanAssignmentsStoreRepository,
	planFeatureQuotaRepo repositories.PlanFeatureQuotaStoreRepository,
) *PlanManagementService {
	return &PlanManagementService{
		planStore:            planStore,
		featureStore:         featureStore,
		planFeatureStore:     planFeatureStore,
		planAssignmentsStore: planAssignmentsStore,
		planFeatureQuotaRepo: planFeatureQuotaRepo,
	}
}

func (s *PlanManagementService) CreateAssignment(ctx context.Context, arg models.CreateAssignmentInput) (*models.PlanAssignment, error) {
	// Assign the plan based on isOrg parameter in the database
	return s.planAssignmentsStore.CreateAssignment(ctx, arg)
}

func (s *PlanManagementService) TerminateAssignment(ctx context.Context, arg models.TerminateAssignmentInput) error {
	return s.planAssignmentsStore.TerminateAssignment(ctx, arg)
}

func (s *PlanManagementService) UpdateAssignment(ctx context.Context, arg models.UpdateAssignmentInput) (*models.PlanAssignment, error) {
	return s.planAssignmentsStore.UpdateAssignment(ctx, arg)
}

func (s *PlanManagementService) ListAssignments(ctx context.Context, arg models.QueryPlanAssignmentInput, pagination pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {
	// Call the store repository to list the assignments
	m, err := s.planAssignmentsStore.ListAssignments(ctx, arg, pagination)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanManagementService) ListAssignmentsHistory(ctx context.Context, arg models.QueryPlanAssignmentHistoryInput, pagination pagination.Pagination) (*pagination.PaginationView[models.PlanAssignmentHistory], error) {
	// Call the store repository to list the assignments history
	m, err := s.planAssignmentsStore.ListAssignmentsHistory(ctx, arg, pagination)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *PlanManagementService) ListAllAssignments(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {
	// Call the store repository to list all assignments
	m, err := s.planAssignmentsStore.ListAllAssignments(ctx, pagination)
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

// PlanFeature methods
func (s *PlanManagementService) CreatePlanFeature(ctx context.Context, arg models.CreatePlanFeatureInput) (*models.PlanFeature, error) {
	return s.planFeatureStore.CreatePlanFeature(ctx, arg)
}

func (s *PlanManagementService) UpdatePlanFeature(ctx context.Context, planID, featureID uuid.UUID, arg models.UpdatePlanFeatureInput) (*models.PlanFeature, error) {
	return s.planFeatureStore.UpdatePlanFeature(ctx, planID, featureID, arg)
}

func (s *PlanManagementService) DeletePlanFeature(ctx context.Context, arg models.DeletePlanFeatureInput) error {
	return s.planFeatureStore.DeletePlanFeature(ctx, arg)
}

func (s *PlanManagementService) ListPlanFeaturesByPlan(ctx context.Context, planID uuid.UUID, filter models.PlanFeatureListFilter) ([]models.PlanFeature, error) {
	return s.planFeatureStore.ListPlanFeaturesByPlan(ctx, planID, filter)
}

func (s *PlanManagementService) CheckPlanAndFeatureForTenant(ctx context.Context, planID, featureID uuid.UUID) (bool, error) {
	return s.planFeatureStore.CheckPlanAndFeatureForTenant(ctx, planID, featureID)
}

// CreatePlanFeatureQuota creates a new quota for a plan feature
func (s *PlanManagementService) CreatePlanFeatureQuota(ctx context.Context, arg models.CreatePlanFeatureQuotaInput, planID, featureID string) (*models.PlanFeatureQuota, error) {
	planFeatureID, err := s.planFeatureStore.GetPlanFeatureIDByPlanAndFeature(ctx, uuid.MustParse(planID), uuid.MustParse(featureID))
	if err != nil {
		return nil, err
	}
	arg.PlanFeatureID = planFeatureID.String()
	return s.planFeatureQuotaRepo.CreatePlanFeatureQuota(ctx, arg)
}

// GetPlanFeatureQuota retrieves a quota by plan feature ID
func (s *PlanManagementService) GetPlanFeatureQuota(ctx context.Context, planID, featureID string) (*models.PlanFeatureQuota, error) {
	// Validate input
	planFeatureID, err := s.planFeatureStore.GetPlanFeatureIDByPlanAndFeature(ctx, uuid.MustParse(planID), uuid.MustParse(featureID))
	if err != nil {
		return nil, err
	}

	return s.planFeatureQuotaRepo.GetPlanFeatureQuota(ctx, planFeatureID)
}

// UpdatePlanFeatureQuota updates an existing quota
func (s *PlanManagementService) UpdatePlanFeatureQuota(ctx context.Context, input models.UpdatePlanFeatureQuotaInput, planID, featureID string) (*models.PlanFeatureQuota, error) {
	// Check if the quota exists
	planFeatureID, err := s.planFeatureStore.GetPlanFeatureIDByPlanAndFeature(ctx, uuid.MustParse(planID), uuid.MustParse(featureID))
	if err != nil {
		return nil, err
	}

	input.PlanFeatureID = planFeatureID.String()
	existing, err := s.planFeatureQuotaRepo.GetPlanFeatureQuota(ctx, planFeatureID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domainerrors.New(errors.New("plan feature quota not found"), domainerrors.ENOTFOUND, "The specified plan feature quota does not exist")
	}

	// Update the quota
	return s.planFeatureQuotaRepo.UpdatePlanFeatureQuota(ctx, input)
}

// DeletePlanFeatureQuota deletes a quota by plan feature ID
func (s *PlanManagementService) DeletePlanFeatureQuota(ctx context.Context, planID, featureID string) error {
	planFeatureID, err := s.planFeatureStore.GetPlanFeatureIDByPlanAndFeature(ctx, uuid.MustParse(planID), uuid.MustParse(featureID))
	if err != nil {
		return err
	}

	return s.planFeatureQuotaRepo.DeletePlanFeatureQuota(ctx, planFeatureID)
}
