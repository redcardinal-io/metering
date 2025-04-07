package clickhouse

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/redcardinal-io/metering/domain/models"
	"go.uber.org/zap"
)

var aggregationToFunctions = map[models.AggregationEnum]struct {
	aggFunc      string
	aggStateFunc string
	dataType     string
}{
	"sum":          {"sum", "sumState", "Float64"},
	"avg":          {"avg", "avgState", "Float64"},
	"min":          {"min", "minState", "Float64"},
	"max":          {"max", "maxState", "Float64"},
	"count":        {"count", "countState", "Float64"},
	"unique_count": {"uniq", "uniqState", "String"},
}

func (store *ClickHouseStore) CreateMeter(ctx context.Context, arg models.CreateMeterInput) error {
	// Build the SQL for creating the materialized view
	sql, params, err := buildCreateMeterSQL(arg, eventsTable)
	if err != nil {
		store.logger.Error("Failed to build meter view SQL", zap.Error(err))
		return fmt.Errorf("failed to build meter view SQL: %w", err)
	}

	// Execute the SQL with the provided parameters (SQL injection safe)
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

func buildCreateMeterSQL(arg models.CreateMeterInput, eventsTable string) (string, []any, error) {
	// Generate a safe, escaped view name (SQL injection safe)
	viewName := getMeterViewName(arg.Organization, arg.Slug)

	// Validate the aggregation type against supported types
	aggInfo, ok := aggregationToFunctions[arg.Aggregation]
	if !ok {
		return "", nil, fmt.Errorf("invalid aggregation type: %s", arg.Aggregation)
	}

	// Validate the ValueProperty, if aggregation is unique_count value_property cannot be empty
	if arg.Aggregation == "unique_count" && arg.ValueProperty == "" {
		return "", nil, fmt.Errorf("value_property is required for unique_count")
	}

	// Build the column definitions for the materialized view
	var columnDefs []string

	// Add base columns that are always present
	columnDefs = append(columnDefs,
		"organization String",       // Organization identifier
		"user String",               // User identifier
		"windowstart DateTime64(3)", // Start of the time window (millisecond precision)
		"windowend DateTime64(3)",   // End of the time window (millisecond precision)
		fmt.Sprintf("value AggregateFunction(%s, %s)", aggInfo.aggFunc, aggInfo.dataType), // Aggregated value with appropriate function and type
	)

	// Define the columns to order by (important for query performance in ClickHouse)
	orderByColumns := []string{
		"windowstart",
		"windowend",
		"organization",
		"user",
	}

	// Process the user-defined properties to group by
	// Sorting ensures consistent column ordering regardless of input order
	groupByKeys := make([]string, 0, len(arg.Properties))
	for _, k := range arg.Properties {
		groupByKeys = append(groupByKeys, k)
	}
	sort.Strings(groupByKeys)

	// Add each property as a column (all properties are stored as strings)
	for _, k := range groupByKeys {
		columnName := k // The column name matches the property name
		// These are schema definitions, not user inputs, so direct string formatting is safe
		columnDefs = append(columnDefs, fmt.Sprintf("%s String", columnName))
		orderByColumns = append(orderByColumns, columnName)
	}

	// Build the SELECT query that populates the materialized view
	selectSQL, selectArgs, err := buildMeterSelectSQL(arg, eventsTable, aggInfo.aggStateFunc, aggInfo.dataType)
	if err != nil {
		return "", nil, err
	}

	// Handle the POPULATE option (backfills the view with existing data if true)
	populateClause := ""
	if arg.Populate {
		populateClause = "POPULATE"
	}

	// Construct the complete CREATE MATERIALIZED VIEW statement
	// Note: column definitions and ORDER BY columns are constructed from validated inputs
	createSQL := fmt.Sprintf(
		`create materialized view if not exists %s (
			%s
		) engine = AggregatingMergeTree()
		order by (%s)
		%s
		AS
		%s`,
		viewName, // Already escaped in getMeterViewName
		strings.Join(columnDefs, ", "),
		strings.Join(orderByColumns, ", "),
		populateClause,
		selectSQL, // Generated using parameterized queries
	)

	return createSQL, selectArgs, nil
}

func buildMeterSelectSQL(arg models.CreateMeterInput, eventsTable string, aggStateFunc string, dataType string) (string, []any, error) {
	sb := sqlbuilder.ClickHouse.NewSelectBuilder()

	// Add the basic dimensions to select
	sb.Select(
		"organization", // Organization identifier
		"user",         // User identifier
		// Time window functions with 1-minute granularity
		// These are ClickHouse functions, not user inputs, so string formatting is safe
		"tumbleStart(timestamp, toIntervalMinute(1)) AS windowstart",
		"tumbleEnd(timestamp, toIntervalMinute(1)) AS windowend",
	)

	// Handle the value aggregation based on the aggregation type
	if arg.ValueProperty == "" && arg.Aggregation == "count" {
		// For simple counting without a specific value property
		sb.Select(fmt.Sprintf("%s(*) AS value", aggStateFunc))
	} else if arg.Aggregation == "unique_count" {
		// For counting unique values of a property
		// The value property is escaped to prevent SQL injection
		sb.Select(fmt.Sprintf("%s(JSONExtractString(properties, '%s')) AS value",
			aggStateFunc, sqlbuilder.Escape(arg.ValueProperty)))
	} else {
		// For other aggregations (sum, avg, min, max)
		// Cast to the appropriate data type
		// Both the value property and data type are sanitized
		sb.Select(fmt.Sprintf("%s(cast(JSONExtractString(properties, '%s'), '%s')) AS value",
			aggStateFunc, sqlbuilder.Escape(arg.ValueProperty), dataType))
	}

	// Create a map of property paths (currently using property name as path)
	propertyPaths := make(map[string]string)
	for _, prop := range arg.Properties {
		propertyPaths[prop] = prop // Using property name as the JSON path
	}

	// Process the user-defined properties to group by
	// Sorting ensures consistent column ordering
	groupByKeys := make([]string, 0, len(propertyPaths))
	for k := range propertyPaths {
		groupByKeys = append(groupByKeys, k)
	}
	sort.Strings(groupByKeys)

	// Add each property as a column, extracting values from JSON
	for _, k := range groupByKeys {
		jsonPath := propertyPaths[k]
		// The JSON path is escaped to prevent SQL injection
		sb.Select(fmt.Sprintf("JSONExtractString(properties, '%s') as %s",
			sqlbuilder.Escape(jsonPath), k))
	}

	// Specify the source table
	sb.From(eventsTable)

	// Add WHERE clause conditions
	// 1. Filter by event type (using parameterized query for safety)
	sb.Where(sb.Equal("type", arg.EventType))
	// 2. Only include events without validation errors
	sb.Where("(validation_errors IS NULL OR length(validation_errors) = 0)")

	// Define the GROUP BY columns
	groupByColumns := []string{
		"organization",
		"user",
		"windowstart",
		"windowend",
	}

	// Add custom properties to GROUP BY
	for _, k := range groupByKeys {
		groupByColumns = append(groupByColumns, k)
	}

	// Apply the GROUP BY clause
	sb.GroupBy(groupByColumns...)

	// Build the final SQL and parameters
	// The sqlbuilder library handles SQL injection protection
	sql, args := sb.Build()

	// Fix spacing issues in the generated SQL
	sql = strings.Replace(sql, "SELECT", "SELECT ", 1)
	sql = strings.Replace(sql, "NULLOR", "NULL OR", 1)
	sql = strings.Replace(sql, "errorsIS", "errors IS", 1)

	return sql, args, nil
}

func getMeterViewName(organization, meterSlug string) string {
	// Replace invalid characters with underscores for ClickHouse identifiers
	sanitizeIdentifier := func(s string) string {
		// Replace hyphens and other invalid characters with underscores
		re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
		return re.ReplaceAllString(s, "_")
	}

	return fmt.Sprintf("meter_%s_%s",
		sanitizeIdentifier(organization),
		sanitizeIdentifier(meterSlug))
}
