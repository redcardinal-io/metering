package meters

import (
	"strings"
	"testing"
	"time"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestQueryMeterToSQL(t *testing.T) {
	// Helper function to create a consistent time for testing
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	fromTime := &baseTime
	to := baseTime.Add(24 * time.Hour) // One day later
	toTime := &to

	// Define UTC timezone for tests
	utc := "UTC"

	// Define window sizes
	minute := models.WindowSizeMinute
	hour := models.WindowSizeHour
	day := models.WindowSizeDay
	invalidWindow := models.WindowSize("invalid")

	tests := []struct {
		name        string
		query       QueryMeter
		wantErr     bool
		checkResult func(t *testing.T, sql string, args []any)
	}{
		{
			name: "Basic query with sum aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "SELECT min(windowstart) AS windowstart, max(windowend) AS windowend, sumMerge(value) AS value")
				assert.Contains(t, sql, "FROM rc_test_tenant_page_views_mv")
				assert.Contains(t, sql, "WHERE windowstart >= ? AND windowend <= ?")
				assert.Len(t, args, 2)
				assert.Equal(t, fromTime.Unix(), args[0])
				assert.Equal(t, toTime.Unix(), args[1])
			},
		},
		{
			name: "Query with minute window size",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				From:           fromTime,
				To:             toTime,
				WindowSize:     &minute,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "tumbleStart(windowstart, toIntervalMinute(1), 'UTC') AS windowstart")
				assert.Contains(t, sql, "tumbleEnd(windowend, toIntervalMinute(1), 'UTC') AS windowend")
				assert.Contains(t, sql, "GROUP BY windowstart, windowend")
			},
		},
		{
			name: "Query with hour window size",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				From:           fromTime,
				To:             toTime,
				WindowSize:     &hour,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "tumbleStart(windowstart, toIntervalHour(1), 'UTC') AS windowstart")
				assert.Contains(t, sql, "tumbleEnd(windowend, toIntervalHour(1), 'UTC') AS windowend")
				assert.Contains(t, sql, "GROUP BY windowstart, windowend")
			},
		},
		{
			name: "Query with day window size",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				From:           fromTime,
				To:             toTime,
				WindowSize:     &day,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "tumbleStart(windowstart, toIntervalDay(1), 'UTC') AS windowstart")
				assert.Contains(t, sql, "tumbleEnd(windowend, toIntervalDay(1), 'UTC') AS windowend")
				assert.Contains(t, sql, "GROUP BY windowstart, windowend")
			},
		},
		{
			name: "Query with avg aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationAvg,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "avgMerge(value) AS value")
			},
		},
		{
			name: "Query with min aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationMin,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "minMerge(value) AS value")
			},
		},
		{
			name: "Query with max aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationMax,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "maxMerge(value) AS value")
			},
		},
		{
			name: "Query with unique count aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "unique_users",
				Aggregation:    models.AggregationUniqueCount,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "toFloat64(uniqMerge(value)) AS value")
			},
		},
		{
			name: "Query with count aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "api_requests",
				Aggregation:    models.AggregationCount,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				assert.Contains(t, sql, "toFloat64(countMerge(value)) AS value")
			},
		},
		{
			name: "Query with custom dimension filter",
			query: QueryMeter{
				TenantSlug:  "test_tenant",
				MeterSlug:   "page_views",
				Aggregation: models.AggregationSum,
				FilterGroupBy: map[string][]string{
					"path":     {"home", "about"},
					"referrer": {"google", "facebook"},
				},
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				// Check for the filter expressions - should contain path and referrer filters
				assert.True(t, strings.Contains(sql, "path IN (?, ?)") || strings.Contains(sql, "\"path\" IN (?, ?)"),
					"SQL should contain path filter condition")
				assert.True(t, strings.Contains(sql, "referrer IN (?, ?)") || strings.Contains(sql, "\"referrer\" IN (?, ?)"),
					"SQL should contain referrer filter condition")

				// Create a copy of args to work with
				argsCopy := make([]any, len(args))
				copy(argsCopy, args)

				// Check for filter values
				foundValues := make(map[string]bool)
				for _, arg := range argsCopy {
					// Skip timestamp values
					if _, ok := arg.(int64); ok {
						continue
					}

					// Check string values
					if s, ok := arg.(string); ok {
						foundValues[s] = true
					}
				}

				assert.True(t, foundValues["home"], "Args should contain 'home'")
				assert.True(t, foundValues["about"], "Args should contain 'about'")
				assert.True(t, foundValues["google"], "Args should contain 'google'")
				assert.True(t, foundValues["facebook"], "Args should contain 'facebook'")
			},
		},
		{
			name: "Query with group by dimensions",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				GroupBy:        []string{"path", "referrer"},
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				// Check for GROUP BY clause
				assert.True(t, strings.Contains(sql, "GROUP BY"), "SQL should contain GROUP BY clause")

				// Check for dimension columns (handle potential escaping)
				assert.True(t, strings.Contains(sql, "path") || strings.Contains(sql, "\"path\""),
					"SQL should include path column")
				assert.True(t, strings.Contains(sql, "referrer") || strings.Contains(sql, "\"referrer\""),
					"SQL should include referrer column")
			},
		},
		{
			name: "Query with window size and group by",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				GroupBy:        []string{"path", "referrer"},
				From:           fromTime,
				To:             toTime,
				WindowSize:     &hour,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)
				// Check for window functions
				assert.Contains(t, sql, "tumbleStart")
				assert.Contains(t, sql, "tumbleEnd")

				// Check for GROUP BY with windows and dimensions
				assert.Contains(t, sql, "GROUP BY windowstart, windowend")

				// Check for dimension columns
				assert.True(t, strings.Contains(sql, "path") || strings.Contains(sql, "\"path\""),
					"SQL should include path column")
				assert.True(t, strings.Contains(sql, "referrer") || strings.Contains(sql, "\"referrer\""),
					"SQL should include referrer column")
			},
		},
		{
			name: "Query with combined filters and grouping",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				FilterGroupBy:  map[string][]string{"path": {"home"}},
				GroupBy:        []string{"referrer"},
				From:           fromTime,
				To:             toTime,
				WindowSize:     &day,
				WindowTimeZone: &utc,
			},
			wantErr: false,
			checkResult: func(t *testing.T, sql string, args []any) {
				sql = normalizeSQL(sql)

				// The SQL will contain organization and user filters at the beginning
				// and the WHERE clause for time range at the end
				// When using the ClickHouse SQL builder, these might appear in different ways

				// Check that path filter exists
				pathFilter := strings.Contains(sql, "path = ?") || strings.Contains(sql, "\"path\" = ?")
				assert.True(t, pathFilter, "SQL should contain path filter")

				// Create a copy of args to work with
				argsCopy := make([]any, len(args))
				copy(argsCopy, args)

				// Check for filter values
				foundValues := make(map[string]bool)
				for _, arg := range argsCopy {
					// Skip timestamp values
					if _, ok := arg.(int64); ok {
						continue
					}

					// Check string values
					if s, ok := arg.(string); ok {
						foundValues[s] = true
					}
				}

				assert.True(t, foundValues["home"], "Args should contain 'home'")

				// Check for group by with window
				assert.Contains(t, sql, "GROUP BY windowstart, windowend")

				// Check for referrer column in group by
				assert.True(t, strings.Contains(sql, "referrer"), "SQL should contain referrer column")
			},
		},
		{
			name: "Error case - unsupported timezone",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: func() *string { s := "EST"; return &s }(),
			},
			wantErr: true,
			checkResult: func(t *testing.T, sql string, args []any) {
				// Should not reach here due to error
			},
		},
		{
			name: "Error case - invalid aggregation",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    "invalid_aggregation",
				From:           fromTime,
				To:             toTime,
				WindowTimeZone: &utc,
			},
			wantErr: true,
			checkResult: func(t *testing.T, sql string, args []any) {
				// Should not reach here due to error
			},
		},
		{
			name: "Error case - invalid window size",
			query: QueryMeter{
				TenantSlug:     "test_tenant",
				MeterSlug:      "page_views",
				Aggregation:    models.AggregationSum,
				From:           fromTime,
				To:             toTime,
				WindowSize:     &invalidWindow,
				WindowTimeZone: &utc,
			},
			wantErr: true,
			checkResult: func(t *testing.T, sql string, args []any) {
				// Should not reach here due to error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := tt.query.ToSQL()

			// Check error expectation
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Use the custom check function for validating results
			tt.checkResult(t, gotSQL, gotArgs)
		})
	}
}
