package models

type FeatureTypeEnum string

const (
	// FeatureTypeStandard represents a standard feature
	FeatureTypeStandard FeatureTypeEnum = "standard"
	// FeatureTypeMetered represents a custom feature
	FeatureTypeMetered FeatureTypeEnum = "metered"
)

// Feature represents a feature in the system
type Feature struct {
	Base
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Description string          `json:"description,omitempty"`
	TenantSlug  string          `json:"tenant_slug"`
	Type        FeatureTypeEnum `json:"type"`
	Config      map[string]any  `json:"config,omitempty"`
}

// CreateFeatureInput represents the input for creating a new feature
type CreateFeatureInput struct {
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Description string          `json:"description,omitempty"`
	TenantSlug  string          `json:"tenant_slug"`
	Type        FeatureTypeEnum `json:"type"`
	Config      map[string]any  `json:"config,omitempty"`
	CreatedBy   string          `json:"created_by"`
}

// UpdateFeatureInput represents the input for updating an existing feature
type UpdateFeatureInput struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Config      map[string]any `json:"config,omitempty"`
	UpdatedBy   string         `json:"updated_by"`
}
