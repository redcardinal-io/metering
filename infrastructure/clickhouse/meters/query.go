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
	var adjustedFrom, adjustedTo time.Time

	tz := "UTC" // Default timezone
	// TODO: Handle time zone conversion in ClickHouse
	if q.WindowTimeZone != nil {
		tz = *q.WindowTimeZone
	}

	// Handle window size grouping
	groupByWindowSize := q.WindowSize != nil
	if groupByWindowSize {
		if q.From == nil || q.To == nil {
			return "", nil, fmt.Errorf("From/To must be provided when WindowSize is set")
		}
		switch *q.WindowSize {
		case models.WindowSizeMinute:
			// Truncate 'from' to the start of the minute
			adjustedFrom = q.From.Truncate(time.Minute)
			// Extend 'to' to the end of the minute if it's not already at the start
			truncatedTo := q.To.Truncate(time.Minute)
			if !truncatedTo.Equal(*q.To) {
				adjustedTo = truncatedTo.Add(time.Minute)
			} else {
				adjustedTo = *q.To
			}
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleStart(windowstart, toIntervalMinute(1), '%s') AS windowstart", tz))
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleEnd(windowend, toIntervalMinute(1), '%s') AS windowend", tz))
		case models.WindowSizeHour:
			adjustedFrom = q.From.Truncate(time.Hour)
			truncatedTo := q.To.Truncate(time.Hour)
			if !truncatedTo.Equal(*q.To) {
				adjustedTo = truncatedTo.Add(time.Hour)
			} else {
				adjustedTo = *q.To
			}
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleStart(windowstart, toIntervalHour(1), '%s') AS windowstart", tz))
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleEnd(windowend, toIntervalHour(1), '%s') AS windowend", tz))
		case models.WindowSizeDay:
			from := *q.From
			to := *q.To
			adjustedFrom = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
			if to.Hour() != 0 || to.Minute() != 0 || to.Second() != 0 || to.Nanosecond() != 0 {
				adjustedTo = time.Date(to.Year(), to.Month(), to.Day()+1, 0, 0, 0, 0, to.Location())
			} else {
				adjustedTo = to
			}
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleStart(windowstart, toIntervalDay(1), '%s') AS windowstart", tz))
			selectColumns = append(selectColumns, fmt.Sprintf("tumbleEnd(windowend, toIntervalDay(1), '%s') AS windowend", tz))
		default:
			adjustedFrom = *q.From
			adjustedTo = *q.To
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
		if len(values) > 1 {
			filterArgs := make([]interface{}, len(values))
			for i, v := range values {
				filterArgs[i] = v
			}
			builder.Where(builder.In(safeCol, filterArgs...))
		} else {
			builder.Where(builder.Equal(safeCol, values[0]))
		}
	}

	// Add time range filters
	if !adjustedFrom.IsZero() {
		builder.Where(builder.GE("windowstart", adjustedFrom.Unix()))
	}
	if !adjustedTo.IsZero() {
		builder.Where(builder.LE("windowend", adjustedTo.Unix()))
	}

	// Add GROUP BY clause
	if len(groupByColumns) > 0 {
		builder.GroupBy(groupByColumns...)
	}
	builder.Select(selectColumns...)

	sql, args := builder.Build()

	return sql, args, nil
}
