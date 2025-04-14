package clickhouse

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/redcardinal-io/metering/domain/models"
	"go.uber.org/zap"
)

// WindowSize represents different time window sizes for querying meter data
type WindowSize int

const (
	WindowSizeMinute WindowSize = iota
	WindowSizeHour
	WindowSizeDay
)

// NewWindowSize creates a WindowSize from a string representation
func NewWindowSize(value string) WindowSize {
	switch strings.ToLower(value) {
	case "minute", "1m", "min":
		return WindowSizeMinute
	case "hour", "1h":
		return WindowSizeHour
	case "day", "1d":
		return WindowSizeDay
	default:
		return WindowSizeMinute
	}
}

// QueryMeterView represents parameters for querying a materialized view
type QueryMeterView struct {
	Namespace       string
	MeterSlug       string
	Aggregation     models.AggregationEnum
	Subjects        []string
	FilterGroupBy   map[string][]string
	From            time.Time
	To              time.Time
	GroupBy         []string
	WindowSize      *WindowSize
	WindowTimeZone  *string
}

// QueryMeter executes a query against a meter view
func (store *ClickHouseStore) QueryMeter(ctx context.Context, query QueryMeterView) ([]models.QueryMeterRow, error) {
	sql, params, err := buildQueryMeterSQL(query)
	if err != nil {
		store.logger.Error("Failed to build query meter SQL", zap.Error(err))
		return nil, fmt.Errorf("failed to build query meter SQL: %w", err)
	}

	// Execute the query and process results
	// This would depend on your specific ClickHouse client and models
	// Implementation omitted for brevity
	
	return nil, fmt.Errorf("not implemented")
}

// buildQueryMeterSQL builds the SQL for querying a meter view
func buildQueryMeterSQL(query QueryMeterView) (string, []any, error) {
	viewName := getMeterViewName(query.Namespace, query.MeterSlug)
	selectColumns := []string{}
	whereClauses := []string{}
	var params []any
	
	// Get the aggregation functions
	aggInfo, ok := aggregationToFunctions[query.Aggregation]
	if !ok {
		return "", nil, fmt.Errorf("invalid aggregation type: %s", query.Aggregation)
	}
	
	// Handle window size grouping
	groupByWindowSize := query.WindowSize != nil
	tz := "UTC"
	if query.WindowTimeZone != nil {
		tz = *query.WindowTimeZone
	}
	
	if groupByWindowSize {
		switch *query.WindowSize {
		case WindowSizeMinute:
			selectColumns = append(selectColumns, 
				"tumbleStart(windowstart, toIntervalMinute(1), ?) AS start",
				"tumbleEnd(windowend, toIntervalMinute(1), ?) AS end")
			params = append(params, tz, tz)
		case WindowSizeHour:
			selectColumns = append(selectColumns, 
				"tumbleStart(windowstart, toIntervalHour(1), ?) AS start",
				"tumbleEnd(windowend, toIntervalHour(1), ?) AS end")
			params = append(params, tz, tz)
		case WindowSizeDay:
			selectColumns = append(selectColumns, 
				"tumbleStart(windowstart, toIntervalDay(1), ?) AS start",
				"tumbleEnd(windowend, toIntervalDay(1), ?) AS end")
			params = append(params, tz, tz)
		}
	} else {
		selectColumns = append(selectColumns, "min(windowstart)", "max(windowend)")
	}
	
	// Add value column based on aggregation type
	valueColumn := ""
	switch query.Aggregation {
	case "sum":
		valueColumn = "sumMerge(value) AS value"
	case "avg":
		valueColumn = "avgMerge(value) AS value"
	case "min":
		valueColumn = "minMerge(value) AS value"
	case "max":
		valueColumn = "maxMerge(value) AS value"
	case "unique_count":
		valueColumn = "toFloat64(uniqMerge(value)) AS value"
	case "count":
		valueColumn = "toFloat64(countMerge(value)) AS value"
	}
	selectColumns = append(selectColumns, valueColumn)
	
	// Handle subject and group_by concatenation for GroupBy field
	groupByColumns := []string{"subject"}
	if len(query.GroupBy) > 0 {
		concatParts := []string{"concat(subject, '-subject::')"}
		columnNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		
		for i, c := range query.GroupBy {
			escapedC := escape(c)
			
			// Validate column name
			if !columnNameRegex.MatchString(escapedC) {
				return "", nil, fmt.Errorf("invalid group_by column name: %s. Only alphanumeric characters and underscores are allowed", escapedC)
			}
			
			params = append(params, escapedC)
			concatExpr := ""
			if i == len(query.GroupBy)-1 {
				concatExpr = fmt.Sprintf("concat(%s, concat('-', ?))", escapedC)
			} else {
				concatExpr = fmt.Sprintf("concat(%s, concat('-', concat(?, '::')))", escapedC)
			}
			concatParts = append(concatParts, concatExpr)
		}
		concatExpr := fmt.Sprintf("concat(%s)", strings.Join(concatParts, ", "))
		selectColumns = append(selectColumns, fmt.Sprintf("%s AS group_by", concatExpr))
		groupByColumns = append(groupByColumns, query.GroupBy...)
	} else {
		selectColumns = append(selectColumns, "concat(subject, '-subject') AS group_by")
	}
	
	// Handle subject filtering
	if len(query.Subjects) > 0 {
		subjects := make([]string, len(query.Subjects))
		for i := range query.Subjects {
			subjects[i] = "subject = ?"
			params = append(params, escape(query.Subjects[i]))
		}
		whereClauses = append(whereClauses, fmt.Sprintf("(%s)", strings.Join(subjects, " OR ")))
	}
	
	// Handle additional group by filtering
	if len(query.FilterGroupBy) > 0 {
		columns := make([]string, 0, len(query.FilterGroupBy))
		for column := range query.FilterGroupBy {
			columns = append(columns, column)
		}
		sort.Strings(columns)
		
		for _, column := range columns {
			values := query.FilterGroupBy[column]
			if len(values) == 0 {
				return "", nil, fmt.Errorf("empty filter for group by: %s", column)
			}
			
			filters := make([]string, len(values))
			for i := range values {
				filters[i] = fmt.Sprintf("%s = ?", escape(column))
				params = append(params, escape(values[i]))
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", strings.Join(filters, " OR ")))
		}
	}
	
	// Handle time range filtering
	whereClauses = append(whereClauses, "windowstart >= ?")
	params = append(params, query.From.Unix())
	
	whereClauses = append(whereClauses, "windowend <= ?")
	params = append(params, query.To.Unix())
	
	// Construct the WHERE clause
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = fmt.Sprintf("WHERE %s", strings.Join(whereClauses, " AND "))
	}
	
	// Construct the GROUP BY clause
	groupByClause := ""
	if groupByWindowSize {
		groupByClause = fmt.Sprintf("GROUP BY start, end, %s", strings.Join(groupByColumns, ", "))
	} else {
		groupByClause = fmt.Sprintf("GROUP BY %s", strings.Join(groupByColumns, ", "))
	}
	
	// Construct the ORDER BY clause
	orderByClause := ""
	if groupByWindowSize {
		orderByClause = "ORDER BY start"
	}
	
	// Construct the full query
	sql := fmt.Sprintf(
		"SELECT %s\nFROM %s\n%s\n%s\n%s",
		strings.Join(selectColumns, ", "),
		viewName,
		whereClause,
		groupByClause,
		orderByClause,
	)
	
	return sql, params, nil
}