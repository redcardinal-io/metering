package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

// AggregationEnum represents the possible aggregation types for a meter
type AggregationEnum string

const (
	AggregationCount       AggregationEnum = "count"
	AggregationSum         AggregationEnum = "sum"
	AggregationAvg         AggregationEnum = "avg"
	AggregationUniqueCount AggregationEnum = "unique_count"
	AggregationMin         AggregationEnum = "min"
	AggregationMax         AggregationEnum = "max"
)

// Meter represents a meter entity from the database
type Meter struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Slug          string          `json:"slug" db:"slug"`
	EventType     string          `json:"event_type" db:"event_type"`
	Description   string          `json:"description,omitempty" db:"description"`
	ValueProperty string          `json:"value_property,omitempty" db:"value_property"`
	Properties    []string        `json:"properties" db:"properties"`
	Aggregation   AggregationEnum `json:"aggregation" db:"aggregation"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	CreatedBy     string          `json:"created_by" db:"created_by"`
}

// CreateMeterInput represents the input for creating a new meter
type CreateMeterInput struct {
	Name          string          `json:"name" validate:"required"`
	Slug          string          `json:"slug" validate:"required"`
	EventType     string          `json:"event_type" validate:"required"`
	Description   string          `json:"description,omitempty"`
	ValueProperty string          `json:"value_property,omitempty"`
	Properties    []string        `json:"properties" validate:"required,min=1"`
	Aggregation   AggregationEnum `json:"aggregation" validate:"required,oneof=count sum avg unique_count min max"`
	CreatedBy     string          `json:"created_by" validate:"required"`
}

// ValidateAggregation checks if a string is a valid aggregation enum value
func ValidateAggregation(value string) bool {
	switch AggregationEnum(value) {
	case AggregationCount, AggregationSum, AggregationAvg,
		AggregationUniqueCount, AggregationMin, AggregationMax:
		return true
	default:
		return false
	}
}

func (m Meter) Cursor() pagination.Cursor {
	return pagination.NewCursor(m.CreatedAt, m.ID.String())
}
