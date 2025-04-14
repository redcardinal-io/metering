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

// Map aggregation types to their corresponding ClickHouse functions
var aggregationToFunctions = map[models.AggregationEnum]struct {
	aggFunc      string
	aggMergeFunc string
	aggStateFunc string
	dataType     string
}{
	"sum":          {"sum", "sumMerge", "sumState", "Float64"},
	"avg":          {"avg", "avgMerge", "avgState", "Float64"},
	"min":          {"min", "minMerge", "minState", "Float64"},
	"max":          {"max", "maxMerge", "maxState", "Float64"},
	"count":        {"count", "countMerge", "countState", "Float64"},
	"unique_count": {"uniq", "uniqMerge", "uniqState", "String"},
}

// TablePrefix defines the prefix used for all tables and views
const TablePrefix = "rc_"
const EventsTableName = "events"

// CreateMeter creates a materialized view for meter data in ClickHouse
func (store *ClickHouseStore) CreateMeter(ctx context.Context, arg models.CreateMeterInput) error {
	// Build the SQL for creating the materialized view
	sql, params, err := buildCreateMeterSQL(arg)
	if err != nil {
		store.logger.Error("Failed to build meter view SQL", zap.Error(err))
		return fmt.Errorf("failed to build meter view SQL: %w", err)
	}

	// Execute the SQL with the provided parameters
	_, err = store.db.ExecContext(ctx, sql, params...)
	if err != nil {
		store.logger.Error("Failed to create meter view", zap.Error(err), zap.String("sql", sql))
		return fmt.Errorf("failed to create meter view: %w", err)
	}

	store.logger.Info("Successfully created meter view",
		zap.String("organization", arg.Organization),
		zap.String("meter_slug", arg.Slug))
	return nil
}

// buildCreateMeterSQL builds the SQL for creating the materialized view
func buildCreateMeterSQL(arg models.CreateMeterInput) (string, []any, error) {
	// Get the view name
	viewName := getMeterViewName(arg.Organization, arg.Slug)
	eventsTable := getEventsTableName()

	// Validate the aggregation type
	aggInfo, ok := aggregationToFunctions[arg.Aggregation]
	if !ok {
		return "", nil, fmt.Errorf("invalid aggregation type: %s", arg.Aggregation)
	}

	// Build the column definitions for the materialized view
	var columns []string
	columns = append(columns,
		"organization String",       // Organization identifier
		"user String",
		"windowstart DateTime",
		"windowend DateTime",
		fmt.Sprintf("value AggregateFunction(%s, %s)", aggInfo.aggFunc, aggInfo.dataType),
	)

	// Define the ORDER BY columns
	orderByColumns := []string{
		"windowstart",
		"windowend",
		"organization",
		"user",
	}

	// Sort group_by keys for consistency
	groupByKeys := make([]string, 0, len(arg.Properties))
	for k := range arg.Properties {
		groupByKeys = append(groupByKeys, k)
	}
	sort.Strings(groupByKeys)

	// Add each group_by property as a column
	for _, k := range groupByKeys {
		columnName := escape(k)
		columns = append(columns, fmt.Sprintf("%s String", columnName))
		orderByColumns = append(orderByColumns, columnName)
	}

	// Build the SELECT query for the materialized view
	selectSQL, selectParams, err := buildMeterSelectSQL(arg, eventsTable, aggInfo.aggStateFunc, aggInfo.dataType)
	if err != nil {
		return "", nil, err
	}

	// Handle the POPULATE option
	populateClause := ""
	if arg.Populate {
		populateClause = "POPULATE"
	}

	// Construct the complete CREATE MATERIALIZED VIEW statement
	createSQL := fmt.Sprintf(
		`CREATE MATERIALIZED VIEW IF NOT EXISTS %s (
			%s
		) ENGINE = AggregatingMergeTree()
		PARTITION BY toYYYYMM(windowstart)
		ORDER BY (%s)
		%s
		AS
		%s`,
		viewName,
		strings.Join(columns, ",\n\t\t\t"),
		strings.Join(orderByColumns, ", "),
		populateClause,
		selectSQL,
	)

	return createSQL, selectParams, nil
}

// buildMeterSelectSQL builds the SELECT statement for populating the materialized view
func buildMeterSelectSQL(arg models.CreateMeterInput, eventsTable string, aggStateFunc string, dataType string) (string, []any, error) {
	// Initialize the list of columns to select
	selects := []string{
		"organization", // Organization identifier
		"user",         // User identifier
		"tumbleStart(time, toIntervalMinute(1)) AS windowstart",
		"tumbleEnd(time, toIntervalMinute(1)) AS windowend",
	}

	var params []any

	// Handle the value aggregation based on the aggregation type
	valueSelect := ""
	if arg.ValueProperty == "" && arg.Aggregation == "count" {
		valueSelect = fmt.Sprintf("%s(*) AS value", aggStateFunc)
	} else if arg.Aggregation == "unique_count" {
		valueSelect = fmt.Sprintf("%s(JSONExtractString(data, '%s')) AS value",
			aggStateFunc, escape(arg.ValueProperty))
	} else {
		valueSelect = fmt.Sprintf("%s(cast(JSONExtractString(data, '%s'), '%s')) AS value",
			aggStateFunc, escape(arg.ValueProperty), dataType)
	}
	selects = append(selects, valueSelect)

	// Define the ORDER BY columns
	orderBy := []string{
		"windowstart",
		"windowend",
		"subject",
	}

	propertyPaths := make(map[string]string)
	for _, prop := range arg.Properties {
		propertyPaths[prop] = prop // Using property name as the JSON path
	}

	// Sort group_by keys for consistency
	groupByKeys := make([]string, 0, len(arg.propertyPaths))
	for k := range arg.propertyPaths {
		groupByKeys = append(groupByKeys, k)
	}
	sort.Strings(groupByKeys)

	// Add each group_by property as a column
	for _, k := range groupByKeys {
		v := arg.propertyPaths[k]
		columnName := escape(k)
		orderBy = append(orderBy, columnName)
		selects = append(selects, fmt.Sprintf("JSONExtractString(data, '%s') as %s",
			escape(v), columnName))
	}

	// Construct the WHERE clause to filter events
	whereClauses := []string{
		"empty(validation_error) = 1",
		"type = ?",
	}
	params = append(params, arg.EventType)

	// Complete SELECT query
	sql := fmt.Sprintf(
		"SELECT %s\nFROM %s\nWHERE %s\nGROUP BY %s",
		strings.Join(selects, ", "),
		eventsTable,
		strings.Join(whereClauses, " AND "),
		strings.Join(orderBy, ", "),
	)

	return sql, params, nil
}

// getEventsTableName returns the fully qualified name of the events table
func getEventsTableName() string {
	return fmt.Sprintf("%s%s", TablePrefix, EventsTableName)
}

func getMeterViewName(organization string, meterSlug string) string {
	// Replace invalid characters with underscores for ClickHouse identifiers
	sanitizeIdentifier := func(s string) string {
		// Replace hyphens and other invalid characters with underscores
		re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
		return re.ReplaceAllString(s, "_")
	}

	return fmt.Sprintf("meter_%s_%s_mv",
		sanitizeIdentifier(organization),
		sanitizeIdentifier(meterSlug))
}

// escape escapes a string for use in SQL queries
func escape(s string) string {
	return strings.Replace(s, "$", "$$", -1)
}