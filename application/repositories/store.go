package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

type StoreRepository interface {
	Connect(cfg *config.StoreConfig) error
	Close() error
	GetDB() any
}

type MeterStoreRepository interface {
	CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error)
	GetMeterByIDorSlug(ctx context.Context, idOrSlug string) (*models.Meter, error)
	ListMeters(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Meter], error)
	ListMetersByEventTypes(ctx context.Context, eventTypes []string) ([]*models.Meter, error)
	DeleteMeterByIDorSlug(ctx context.Context, idOrSlug string) error
	UpdateMeterByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdateMeterInput) (*models.Meter, error)
}

type PlanStoreRepository interface {
	CreatePlan(ctx context.Context, arg models.CreatePlanInput) (*models.Plan, error)
	GetPlanByIDorSlug(ctx context.Context, idOrSlug string) (*models.Plan, error)
	ListPlans(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Plan], error)
	DeletePlanByIDorSlug(ctx context.Context, idOrSlug string) error
	UpdatePlanByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdatePlanInput) (*models.Plan, error)
	ArchivePlanByIDorSlug(ctx context.Context, idOrSlug string, arg models.ArchivePlanInput) error
}

type PlanAssignmentsStoreRepository interface {
	CreateAssignment(ctx context.Context, arg models.CreateAssignmentInput) (*models.PlanAssignment, error)
	TerminateAssignment(ctx context.Context, arg models.TerminateAssignmentInput) error
	UpdateAssignment(ctx context.Context, arg models.UpdateAssignmentInput) (*models.PlanAssignment, error)
	ListAssignments(ctx context.Context, arg models.QueryPlanAssignmentInput, pagination pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error)
	ListAssignmentsHistory(ctx context.Context, arg models.QueryPlanAssignmentHistoryInput, pagination pagination.Pagination) (*pagination.PaginationView[models.PlanAssignmentHistory], error)
	ListAllAssignments(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error)
}

type FeatureStoreRepository interface {
	CreateFeature(ctx context.Context, arg models.CreateFeatureInput) (*models.Feature, error)
	GetFeatureByIDorSlug(ctx context.Context, idOrSlug string) (*models.Feature, error)
	ListFeatures(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Feature], error)
	DeleteFeatureByIDorSlug(ctx context.Context, idOrSlug string) error
	UpdateFeatureByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdateFeatureInput) (*models.Feature, error)
}

type PlanFeatureStoreRepository interface {
	CreatePlanFeature(ctx context.Context, arg models.CreatePlanFeatureInput) (*models.PlanFeature, error)
	UpdatePlanFeature(ctx context.Context, planID, featureID uuid.UUID, arg models.UpdatePlanFeatureInput) (*models.PlanFeature, error)
	DeletePlanFeature(ctx context.Context, arg models.DeletePlanFeatureInput) error
	ListPlanFeaturesByPlan(ctx context.Context, planID uuid.UUID, filter models.PlanFeatureListFilter) ([]models.PlanFeature, error)
	CheckPlanAndFeatureForTenant(ctx context.Context, planID, featureID uuid.UUID) (bool, error)
	GetPlanFeatureIDByPlanAndFeature(ctx context.Context, planID, featureID uuid.UUID) (uuid.UUID, error)
}

type PlanFeatureQuotaStoreRepository interface {
	CreatePlanFeatureQuota(ctx context.Context, arg models.CreatePlanFeatureQuotaInput) (*models.PlanFeatureQuota, error)
	GetPlanFeatureQuota(ctx context.Context, planFeatureID uuid.UUID) (*models.PlanFeatureQuota, error)
	UpdatePlanFeatureQuota(ctx context.Context, arg models.UpdatePlanFeatureQuotaInput) (*models.PlanFeatureQuota, error)
	DeletePlanFeatureQuota(ctx context.Context, planFeatureID uuid.UUID) error
	CheckMeteredFeature(ctx context.Context, planFeatureID uuid.UUID) (bool, error)
}
