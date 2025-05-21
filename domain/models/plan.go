package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// PlanTypeEnum represents the possible types of a plan
type PlanTypeEnum string

const (
	Standard PlanTypeEnum = "standard"
	Custom   PlanTypeEnum = "custom"
)

// ValidatePlanType checks whether the given string matches a defined PlanTypeEnum value.
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

// PlanAssignment represents a plan_assignment entity from the database
type PlanAssignment struct {
	Base
	PlanId         string    `json:"plan_id"`
	OrganizationId string    `json:"organization_id"`
	UserId         string    `json:"user_id"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
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

type AssignPlanInput struct {
	UserID         string
	OrganizationID string
	PlanID         string
	ValidFrom      time.Time
	ValidUntil     time.Time
	CreatedBy      string
}

type UpdateAssignedPlanInput struct {
	PlanID         string
	UserID         string
	OrganizationID string
	ValidFrom      time.Time
	ValidUntil     time.Time
	UpdatedBy      string
}

type TerminateAssignedPlanInput struct {
	PlanId         string
	UserId         string
	OrganizationId string
}
