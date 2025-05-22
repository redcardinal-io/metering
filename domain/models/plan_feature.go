package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// PlanFeature represents a feature associated with a plan
type PlanFeature struct {
	Base
	PlanID      uuid.UUID       `json:"plan_id"`
	FeatureID   uuid.UUID       `json:"feature_id"`
	Config      json.RawMessage `json:"config,omitempty"`
	FeatureName string          `json:"feature_name,omitempty"`
	FeatureSlug string          `json:"feature_slug,omitempty"`
	Type        FeatureTypeEnum `json:"feature_type,omitempty"`
}

// CreatePlanFeatureInput represents the input for creating a new plan feature association
type CreatePlanFeatureInput struct {
	PlanID    uuid.UUID
	FeatureID uuid.UUID
	Config    map[string]any
	CreatedBy string
}

// UpdatePlanFeatureInput represents the input for updating an existing plan feature association
type UpdatePlanFeatureInput struct {
	Config    map[string]any
	UpdatedBy string
}

// DeletePlanFeatureInput represents the input for deleting a plan feature association
type DeletePlanFeatureInput struct {
	PlanID    uuid.UUID
	FeatureID uuid.UUID
}

// PlanFeatureListFilter represents filters for listing plan features
type PlanFeatureListFilter struct {
	FeatureType *FeatureTypeEnum
}
