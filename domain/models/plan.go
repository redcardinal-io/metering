package models

// Plan represents a plan entity from the database
type Plan struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	TenantSlug  string `json:"tenant_slug"`
}

// CreatePlanInput represents the input for creating a new plan
type CreatePlanInput struct {
	Name        string
	Description string
	CreatedBy   string
}

type UpdatePlanInput struct {
	Name        string
	Description string
	UpdatedBy   string
}
