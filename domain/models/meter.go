package models

import (
	"time"
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
	Base
	Name          string          `json:"name"`
	Slug          string          `json:"slug"`
	EventType     string          `json:"event_type"`
	Description   string          `json:"description,omitempty"`
	ValueProperty string          `json:"value_property,omitempty"`
	Properties    []string        `json:"properties"`
	Aggregation   AggregationEnum `json:"aggregation"`
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
	CreatedBy     string
}

type WindowSize string

const (
	WindowSizeMinute WindowSize = "minute"
	WindowSizeHour   WindowSize = "hour"
	WindowSizeDay    WindowSize = "day"
)

func IsValidWindowSize(ws *WindowSize) bool {
	switch *ws {
	case WindowSizeMinute, WindowSizeHour, WindowSizeDay:
		return true
	default:
		return false
	}
}

// ValidateAggregation returns true if the provided string is a valid aggregation type.
func ValidateAggregation(value string) bool {
	switch AggregationEnum(value) {
	case AggregationCount, AggregationSum, AggregationAvg,
		AggregationUniqueCount, AggregationMin, AggregationMax:
		return true
	default:
		return false
	}
}

type QueryMeterParams struct {
	MeterSlug      string
	FilterGroupBy  map[string][]string
	From           *time.Time
	To             *time.Time
	GroupBy        []string
	WindowSize     *WindowSize
	WindowTimeZone *string
}

type QueryMeterResult struct {
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
	UpdatedBy   string
}
