package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// PlanTypeEnum represents the possible types of a plan
type PlanTypeEnum string

const (
	Standard PlanTypeEnum = "standard"
	Custom   PlanTypeEnum = "custom"
)

// ValidatePlanType returns true if the provided string is a valid plan type.
func ValidatePlanType(value string) bool {
	switch PlanTypeEnum(value) {
	case Standard, Custom:
		return true
	default:
		return false
	}
}

// Plan represents a plan entity from the database
type Plan struct {
	Base
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Slug        string             `json:"slug"`
	Type        PlanTypeEnum       `json:"type"`
	ArchivedAt  pgtype.Timestamptz `json:"archived_at"`
	TenantSlug  string             `json:"tenant_slug"`
}

// CreatePlanInput represents the input for creating a new plan
type CreatePlanInput struct {
	Name        string
	Slug        string
	Type        PlanTypeEnum
	Description string
	CreatedBy   string
}

type UpdatePlanInput struct {
	Name        string
	Description string
	UpdatedBy   string
}

type ArchivePlanInput struct {
	UpdatedBy string
	Archive   bool
}
