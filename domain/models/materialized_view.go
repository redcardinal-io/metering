package models

import (
	"time"
)

// Meter represents a meter entity from the database
type MaterializedView struct {
	Subject                  string          `json:"subject"`
	WindowStart              time.Time       `json:"windowstart"`
	WindowEnd                time.Time       `json:"windowend"`
	AggregationFunctionValue AggregationEnum `json:"aggregationfunctionvalue"`
	DynamicGroupByColumns    []string        `json:"dynamicgroupbycolumns"`
	OrderBy                  []string        `json:"orderby"`
}
