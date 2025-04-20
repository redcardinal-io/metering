package meters

import (
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/redcardinal-io/metering/domain/models"
)

// QueryMeter represents the parameters used for querying meter data.
type QueryMeter struct {
	TenantSlug     string                 // Unique identifier for the tenant
	MeterSlug      string                 // Unique identifier for the meter
	Aggregation    models.AggregationEnum // Type of aggregation to apply (sum, count, etc.)
	Organizations  []string               // List of organization IDs to filter by
	Users          []string               // List of user IDs to filter by
	FilterGroupBy  map[string][]string    // Custom dimensions to filter and group by
	From           *time.Time             // Start time of the query range
	To             *time.Time             // End time of the query range
	GroupBy        []string               // Dimensions to group results by
	WindowSize     *models.WindowSize     // Time window size for time-based aggregations
	WindowTimeZone *string                // Timezone to use for time-based windows (default is UTC)

}

func (q *QueryMeter) ToSQL() (string, []any, error) {
	if q.WindowTimeZone != nil && *q.WindowTimeZone != "UTC" {
		return "", nil, fmt.Errorf("Currently, only UTC is supported for WindowTimeZone")
	}

	viewName := GetMeterViewName(q.TenantSlug, q.MeterSlug)
	var selectColumns []string
	var groupByColumns []string

	tz := "UTC" // Default timezone
	// TODO: Handle time zone conversion in ClickHouse
	if q.WindowTimeZone != nil {
		tz = *q.WindowTimeZone
	}

	// Handle window size grouping
	groupByWindowSize := q.WindowSize != nil
	if groupByWindowSize {
		switch *q.WindowSize {
		case models.WindowSizeMinute:
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleStart(windowstart, toIntervalMinute(1), '%s') AS windowstart", tz))
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleEnd(windowend, toIntervalMinute(1), '%s') AS windowend", tz))
		case models.WindowSizeHour:
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleStart(windowstart, toIntervalHour(1), '%s') AS windowstart", tz))
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleEnd(windowend, toIntervalHour(1), '%s') AS windowend", tz))
		case models.WindowSizeDay:
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleStart(windowstart, toIntervalDay(1), '%s') AS windowstart", tz))
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleEnd(windowend, toIntervalDay(1), '%s') AS windowend", tz))
		default:
			return "", nil, fmt.Errorf("unsupported window size")
		}
		groupByColumns = append(groupByColumns, "windowstart", "windowend")
	} else {
		selectColumns = append(selectColumns, "min(windowstart) AS windowstart", "max(windowend) AS windowend")
	}

	// Add value aggregation
	switch q.Aggregation {
	case models.AggregationSum:
		selectColumns = append(selectColumns, "sumMerge(value) AS value")
	case models.AggregationAvg:
		selectColumns = append(selectColumns, "avgMerge(value) AS value")
	case models.AggregationMin:
		selectColumns = append(selectColumns, "minMerge(value) AS value")
	case models.AggregationMax:
		selectColumns = append(selectColumns, "maxMerge(value) AS value")
	case models.AggregationUniqueCount:
		selectColumns = append(selectColumns, "toFloat64(uniqMerge(value)) AS value")
	case models.AggregationCount:
		selectColumns = append(selectColumns, "toFloat64(countMerge(value)) AS value")
	default:
		return "", nil, fmt.Errorf("invalid aggregation type: %s", q.Aggregation)
	}

	// Build the query using sqlbuilder
	builder := sqlbuilder.ClickHouse.NewSelectBuilder()
	builder.From(viewName)

	// Add organization filter if any
	if len(q.Organizations) > 0 {
		for _, org := range q.Organizations {
			builder.Where(builder.Equal("organization", org))
		}
	}

	// Add user filter if any
	if len(q.Users) > 0 {
		for _, user := range q.Users {
			builder.Where(builder.Equal("user", user))
		}
	}

	// Add group by columns
	for _, column := range q.GroupBy {
		safeCol := sqlbuilder.Escape(column)
		selectColumns = append(selectColumns, safeCol)
		groupByColumns = append(groupByColumns, safeCol)
	}

	// Add group by filter conditions
	for column, values := range q.FilterGroupBy {
		if len(values) == 0 {
			continue // Skip empty filters
		}

		safeCol := sqlbuilder.Escape(column)
		for _, value := range values {
			builder.Where(builder.Equal(safeCol, value))
		}
	}

	// Add time range filters
	if q.From != nil {
		builder.Where(builder.GE("windowstart", q.From.Unix()))
	}
	if q.To != nil {
		builder.Where(builder.LE("windowend", q.To.Unix()))
	}

	// Add GROUP BY clause
	if len(groupByColumns) > 0 {
		builder.GroupBy(groupByColumns...)
	}
	builder.Select(selectColumns...)

	sql, args := builder.Build()

	return sql, args, nil
}
