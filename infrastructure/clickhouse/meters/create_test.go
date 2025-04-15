package meters

import (
	"testing"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateMeter(t *testing.T) {

	tests := []struct {
		name     string
		meter    createMeter
		wantSQL  string
		wantArgs []any
		wantErr  bool
	}{
		{
			name: "Simple meter with sum aggregation",
			meter: createMeter{
				Name:          "Page Views",
				Slug:          "page_views",
				EventType:     "page_view",
				Description:   "Count of page views",
				ValueProperty: "count",
				Properties:    []string{"path", "referrer"},
				Aggregation:   models.AggregationSum,
				CreatedBy:     "test_user",
				Populate:      false,
				TenantSlug:    "test_tenant",
			},
			wantSQL: `CREATE MATERIALIZED VIEW IF NOT EXISTS rc_test_tenant_page_views_mv ( ? ) ENGINE = AggregatingMergeTree()
			ORDER BY (?) ?
			 AS SELECT
				organization,
				user,
				tumbleStart(time, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(time, toIntervalMinute(1)) AS windowend,
				sumState(cast(JSONExtractString(properties, 'count'), 'Float64')) AS value,
				JSONExtractString(properties, 'path') as path,
				JSONExtractString(properties, 'referrer') as referrer
			FROM rc_events
			WHERE empty(rc_events.validation_error) = 1
			AND rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, path, referrer`,
			wantArgs: []any{"organization String, user String, windowstart DateTime, windowend DateTime, value AggregateFunction(sum, Float64), path String, referrer String", "windowstart, windowend, organization, user, path, referrer", "", "page_view"},
			wantErr:  false,
		},
		{
			name: "Meter with unique count aggregation",
			meter: createMeter{
				Name:          "Unique Users",
				Slug:          "unique_users",
				EventType:     "user_login",
				Description:   "Count of unique users",
				ValueProperty: "user_id",
				Properties:    []string{"country", "device"},
				Aggregation:   models.AggregationUniqueCount,
				CreatedBy:     "test_user",
				Populate:      true,
				TenantSlug:    "test_tenant",
			},
			wantSQL: `CREATE MATERIALIZED VIEW IF NOT EXISTS rc_test_tenant_unique_users_mv ( ? ) ENGINE = AggregatingMergeTree()
			ORDER BY (?) ?
			AS SELECT
				organization,
				user,
				tumbleStart(time, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(time, toIntervalMinute(1)) AS windowend,
				uniqState(JSONExtractString(properties, 'user_id')) AS value,
				JSONExtractString(properties, 'country') as country,
				JSONExtractString(properties, 'device') as device
			FROM rc_events
			WHERE empty(rc_events.validation_error) = 1
			AND rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, country, device`,
			wantArgs: []any{"organization String, user String, windowstart DateTime, windowend DateTime, value AggregateFunction(uniq, String), country String, device String", "windowstart, windowend, organization, user, country, device", "POPULATE", "user_login"},
			wantErr:  false,
		},
		{
			name: "Meter with count aggregation without value property",
			meter: createMeter{
				Name:          "API Requests",
				Slug:          "api_requests",
				EventType:     "api_request",
				Description:   "Count of API requests",
				ValueProperty: "",
				Properties:    []string{"endpoint", "method"},
				Aggregation:   models.AggregationCount,
				CreatedBy:     "test_user",
				Populate:      false,
				TenantSlug:    "test_tenant",
			},
			wantSQL: `CREATE MATERIALIZED VIEW IF NOT EXISTS rc_test_tenant_api_requests_mv ( ? ) ENGINE = AggregatingMergeTree()
			ORDER BY (?) ?
			 AS SELECT
				organization,
				user,
				tumbleStart(time, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(time, toIntervalMinute(1)) AS windowend,
				countState(*) AS value,
				JSONExtractString(properties, 'endpoint') as endpoint,
				JSONExtractString(properties, 'method') as method
			FROM rc_events
			WHERE empty(rc_events.validation_error) = 1
			AND rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, endpoint, method`,
			wantArgs: []any{"organization String, user String, windowstart DateTime, windowend DateTime, value AggregateFunction(count, Float64), endpoint String, method String", "windowstart, windowend, organization, user, endpoint, method", "", "api_request"},
			wantErr:  false,
		},
		{
			name: "Invalid aggregation type",
			meter: createMeter{
				Name:          "Invalid Meter",
				Slug:          "invalid_meter",
				EventType:     "event",
				Description:   "This has an invalid aggregation",
				ValueProperty: "value",
				Properties:    []string{"property"},
				Aggregation:   "invalid_aggregation", // Invalid aggregation
				CreatedBy:     "test_user",
				Populate:      false,
				TenantSlug:    "test_tenant",
			},
			wantSQL:  "",
			wantArgs: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := tt.meter.toCreateSQL()

			// Check error expectation
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Normalize and compare SQL
			assert.Equal(t, normalizeSQL(tt.wantSQL), normalizeSQL(gotSQL))

			// Compare args
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

// Also test the toSeleteSQL method separately
func TestCreateMeterSelectSQL(t *testing.T) {
	tests := []struct {
		name      string
		meter     createMeter
		stateFunc string
		dataType  string
		wantSQL   string
		wantArgs  []any
	}{
		{
			name: "Select SQL for sum aggregation",
			meter: createMeter{
				EventType:     "page_view",
				ValueProperty: "count",
				Properties:    []string{"path", "referrer"},
				TenantSlug:    "test_tenant",
			},
			stateFunc: "sumState",
			dataType:  "Float64",
			wantSQL: `SELECT
				organization,
				user,
				tumbleStart(time, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(time, toIntervalMinute(1)) AS windowend,
				sumState(cast(JSONExtractString(properties, 'count'), 'Float64')) AS value,
				JSONExtractString(properties, 'path') as path,
				JSONExtractString(properties, 'referrer') as referrer
			FROM rc_events
			WHERE empty(rc_events.validation_error) = 1
			AND rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, path, referrer`,
			wantArgs: []any{"page_view"},
		},
		{
			name: "Select SQL for count aggregation",
			meter: createMeter{
				EventType:     "api_request",
				ValueProperty: "",
				Properties:    []string{"endpoint", "method"},
				Aggregation:   models.AggregationCount,
				TenantSlug:    "test_tenant",
			},
			stateFunc: "countState",
			dataType:  "Float64",
			wantSQL: `SELECT
				organization,
				user,
				tumbleStart(time, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(time, toIntervalMinute(1)) AS windowend,
				countState(*) AS value,
				JSONExtractString(properties, 'endpoint') as endpoint,
				JSONExtractString(properties, 'method') as method
			FROM rc_events
			WHERE empty(rc_events.validation_error) = 1
			AND rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, endpoint, method`,
			wantArgs: []any{"api_request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := tt.meter.toSeleteSQL(tt.stateFunc, tt.dataType)

			// Normalize and compare SQL
			assert.Equal(t, normalizeSQL(tt.wantSQL), normalizeSQL(gotSQL))

			// Compare args
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
