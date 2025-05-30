package models

import (
	"time"

	"github.com/google/uuid"
)

// PlanTypeEnum represents the possible types of a plan
type PlanTypeEnum string

const (
	Standard PlanTypeEnum = "standard"
	Custom   PlanTypeEnum = "custom"
)

// PlanAssignmentHistoryActionEnum represents the possible actions in plan_assignment_history
type HistoryActionEnum string

const (
	Insert HistoryActionEnum = "CREATE"
	Update HistoryActionEnum = "UPDATE"
	Delete HistoryActionEnum = "DELETE"
)

// ValidateHistoryAction checks whether the given string matches a defined HistoryActionEnum value.
func ValidateHistoryAction(value string) bool {
	switch HistoryActionEnum(value) {
	case Insert, Update, Delete:
		return true
	default:
		return false
	}
}

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
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Slug        string       `json:"slug"`
	Type        PlanTypeEnum `json:"type"`
	ArchivedAt  time.Time    `json:"archived_at"`
	TenantSlug  string       `json:"tenant_slug"`
}

// PlanAssignment represents a plan_assignment entity from the database
type PlanAssignment struct {
	Base
	PlanID         string    `json:"plan_id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
}

// PlanAssignmentHistory represents a plan_assignment_history entity from the database
type PlanAssignmentHistory struct {
	Base
	PlanID         string    `json:"plan_id"`
	Action         string    `json:"action"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
}

type QueryPlanAssignmentInput struct {
	PlanID         *uuid.UUID
	OrganizationID string
	UserID         string
	ValidFrom      time.Time
	ValidUntil     time.Time
}

type QueryPlanAssignmentHistoryInput struct {
	PlanID           *uuid.UUID
	Action           string
	OrganizationID   string
	UserID           string
	ValidFromBefore  time.Time
	ValidFromAfter   time.Time
	ValidUntilBefore time.Time
	ValidUntilAfter  time.Time
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

type CreateAssignmentInput struct {
	UserID         string
	OrganizationID string
	PlanID         *uuid.UUID
	ValidFrom      time.Time
	ValidUntil     time.Time
	CreatedBy      string
}

type UpdateAssignmentInput struct {
	PlanID              *uuid.UUID
	UserID              string
	OrganizationID      string
	ValidFrom           time.Time
	ValidUntil          time.Time
	UpdatedBy           string
	SetValidUntilToZero bool
}

type TerminateAssignmentInput struct {
	PlanID         *uuid.UUID
	UserID         string
	OrganizationID string
}
