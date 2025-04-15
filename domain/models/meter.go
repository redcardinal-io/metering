package models

import (
	"time"

	"github.com/google/uuid"
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
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Slug          string          `json:"slug"`
	EventType     string          `json:"event_type"`
	Description   string          `json:"description,omitempty"`
	ValueProperty string          `json:"value_property,omitempty"`
	Properties    []string        `json:"properties"`
	Aggregation   AggregationEnum `json:"aggregation"`
	CreatedAt     time.Time       `json:"created_at"`
	CreatedBy     string          `json:"created_by"`
}

// CreateMeterInput represents the input for creating a new meter
type CreateMeterInput struct {
	Name          string
	Slug          string
	EventType     string
	Description   string
	ValueProperty string
	Properties    []string
	Aggregation   AggregationEnum
	CreatedBy     string
	Populate      bool
	TenantSlug    string
}

type WindowSize string

const (
	WindowSizeMinute WindowSize = "minute"
	WindowSizeHour   WindowSize = "hour"
	WindowSizeDay    WindowSize = "day"
)

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
