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
	TenantSlug    string          `json:"tenant_slug"`
}

// CreateMeterInput represents the input for creating a new meter
type CreateMeterInput struct {
	Name          string
	MeterSlug     string
	EventType     string
	Description   string
	ValueProperty string
	Properties    []string
	Aggregation   AggregationEnum
	Populate      bool
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

type QueryMeterInput struct {
	MeterSlug      string
	FilterGroupBy  map[string][]string
	From           *time.Time
	To             *time.Time
	GroupBy        []string
	WindowSize     *WindowSize
	WindowTimeZone *string
}

type QueryMeterOutput struct {
	WindowStart *time.Time      `json:"window_start"`
	WindowEnd   *time.Time      `json:"window_end"`
	WindowSize  *WindowSize     `json:"window_size,omitempty"`
	Data        []QueryMeterRow `json:"data"`
}

type QueryMeterRow struct {
	WindowStart time.Time         `json:"window_start"`
	WindowEnd   time.Time         `json:"window_end"`
	Value       float64           `json:"value"`
	GroupBy     map[string]string `json:"group_by,omitempty"`
}

type UpdateMeterInput struct {
	Name        string
	Description string
}
