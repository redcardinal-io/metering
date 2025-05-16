package models

// Feature represents a feature in the system
type Feature struct {
	Base
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	TenantSlug  string `json:"tenant_slug" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=static metered"`
	Config      string `json:"config,omitempty"`
}

// CreateFeatureInput represents the input for creating a new feature
type CreateFeatureInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	TenantSlug  string `json:"tenant_slug" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=static metered"`
	Config      string `json:"config,omitempty"`
	CreatedBy   string `json:"created_by" validate:"required"`
}

// UpdateFeatureInput represents the input for updating an existing feature
type UpdateFeatureInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Config      string `json:"config,omitempty"`
	UpdatedBy   string `json:"updated_by" validate:"required"`
}
