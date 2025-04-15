package meters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/redcardinal-io/metering/domain/models"
)

// createMeter creates a materialized view for meter data in ClickHouse
type createMeter struct {
	Name          string
	Slug          string
	EventType     string
	Description   string
	ValueProperty string
	Properties    []string
	Aggregation   models.AggregationEnum
	CreatedBy     string
	Populate      bool
	TenantSlug    string
}

func (c *createMeter) toCreateSQL() (string, []any, error) {
	agg, ok := aggregationMap[c.Aggregation]
	if !ok {
		return "", nil, fmt.Errorf("invalid aggregation type: %s", c.Aggregation)
	}

	// Get view name
	viewName := getMeterViewName(c.TenantSlug, c.Slug)

	// Build columns for the materialized view
	var columns []string
	columns = append(columns,
		"organization String",
		"user String",
		"windowstart DateTime",
		"windowend DateTime",
		fmt.Sprintf("value AggregateFunction(%s, %s)", agg.mergeFunc, agg.dataType),
	)

	// Define order by columns
	orderByColumns := []string{
		"windowstart",
		"windowend",
		"organization",
		"user",
	}

	propertyNames := make([]string, len(c.Properties))
	copy(propertyNames, c.Properties)
	sort.Strings(propertyNames)
	// Add each property as a column
	for _, name := range propertyNames {
		columnName := sqlbuilder.Escape(name)
		columns = append(columns, fmt.Sprintf("%s String", columnName))
		orderByColumns = append(orderByColumns, columnName)
	}

	// Build the SELECT query using the helper method
	selectSQL, selectArgs := c.toSeleteSQL(agg.stateFunc, agg.dataType)

	// Handle POPULATE option
	populateClause := ""
	if c.Populate {
		populateClause = "POPULATE"
	}

	// Construct the complete CREATE MATERIALIZED VIEW statement
	builder := sqlbuilder.Buildf(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS %v (
		 %s
		) ENGINE = AggregatingMergeTree()
		ORDER BY (%v)
		%v AS %v
		`, sqlbuilder.Raw(viewName),
		strings.Join(columns, ", "),
		strings.Join(orderByColumns, ", "),
		populateClause,
		sqlbuilder.Raw(selectSQL),
	)

	createSQL, createArgs := builder.Build()

	return createSQL, append(createArgs, selectArgs...), nil
}

func (c *createMeter) toSeleteSQL(aggStateFunc, dataType string) (string, []any) {
	// Create the select builder
	query := sqlbuilder.ClickHouse.NewSelectBuilder()

	// Add basic columns

	var valueColumn string
	// Add value column based on aggregation type
	if c.ValueProperty == "" && c.Aggregation == models.AggregationCount {
		valueColumn = fmt.Sprintf("%s(*) AS value", aggStateFunc)
	} else if c.Aggregation == models.AggregationUniqueCount {
		valueColumn = fmt.Sprintf("%s(JSONExtractString(properties, '%s')) AS value",
			aggStateFunc, sqlbuilder.Escape(c.ValueProperty))
	} else {
		valueColumn = fmt.Sprintf("%s(cast(JSONExtractString(properties, '%s'), '%s')) AS value",
			aggStateFunc, sqlbuilder.Escape(c.ValueProperty), dataType)
	}

	columnNames := []string{
		"organization",
		"user",
		"tumbleStart(time, toIntervalMinute(1)) AS windowstart",
		"tumbleEnd(time, toIntervalMinute(1)) AS windowend",
		valueColumn,
	}

	// Get property names and add them to SELECT
	propertyNames := make([]string, len(c.Properties))
	copy(propertyNames, c.Properties)
	sort.Strings(propertyNames)
	// Add property columns to SELECT
	for _, name := range propertyNames {
		path := name
		columnName := sqlbuilder.Escape(name)
		columnNames = append(columnNames, fmt.Sprintf("JSONExtractString(properties, '%s') as %s",
			sqlbuilder.Escape(path), columnName))
	}

	query.Select(columnNames...)
	query.From(eventsTable)

	// Set WHERE clause with validation check
	query.Where(fmt.Sprintf("empty(%s.validation_error) = 1", eventsTable))
	query.Where(query.Equal(fmt.Sprintf("%s.type", eventsTable), c.EventType))

	// Set GROUP BY clause
	groupByColumns := []string{"windowstart", "windowend", "organization", "user"}
	for _, name := range propertyNames {
		groupByColumns = append(groupByColumns, sqlbuilder.Escape(name))
	}
	query.GroupBy(groupByColumns...)

	sql, args := query.Build()
	return sql, args
}
