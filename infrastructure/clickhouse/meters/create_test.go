package meters

import (
	"fmt"
	"testing"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateMeter(t *testing.T) {

	tests := []struct {
		name     string
		meter    CreateMeter
		wantSQL  string
		wantArgs []any
		wantErr  bool
	}{
		{
			name: "Simple meter with sum aggregation",
			meter: CreateMeter{
				Slug:          "page_views",
				EventType:     "page_view",
				ValueProperty: "count",
				Properties:    []string{"path", "referrer"},
				Aggregation:   models.AggregationSum,
				Populate:      false,
				TenantSlug:    "test_tenant",
			},
			wantSQL: `CREATE MATERIALIZED VIEW IF NOT EXISTS rc_test_tenant_page_views_mv ( %s ) ENGINE = AggregatingMergeTree()
			ORDER BY (%s)
			 AS SELECT
				organization,
				user,
				tumbleStart(timestamp, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(timestamp, toIntervalMinute(1)) AS windowend,
				sumState(cast(JSONExtractString(properties, 'count'), 'Float64')) AS value,
				JSONExtractString(properties, 'path') as path,
				JSONExtractString(properties, 'referrer') as referrer
			FROM rc_events
			WHERE rc_events.type = ? 
      GROUP BY windowstart, windowend, organization, user, path, referrer`,
			wantArgs: []any{"organization String, user String, windowstart DateTime, windowend DateTime, value AggregateFunction(sum, Float64), path String, referrer String", "windowstart, windowend, organization, user, path, referrer", "", "page_view"},
			wantErr:  false,
		},
		{
			name: "Meter with unique count aggregation",
			meter: CreateMeter{
				Slug:          "unique_users",
				EventType:     "user_login",
				ValueProperty: "user_id",
				Properties:    []string{"country", "device"},
				Aggregation:   models.AggregationUniqueCount,
				Populate:      true,
				TenantSlug:    "test_tenant",
			},
			wantSQL: `CREATE MATERIALIZED VIEW IF NOT EXISTS rc_test_tenant_unique_users_mv ( %s ) ENGINE = AggregatingMergeTree()
			ORDER BY (%s)
      POPULATE
			AS SELECT
				organization,
				user,
				tumbleStart(timestamp, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(timestamp, toIntervalMinute(1)) AS windowend,
				uniqState(JSONExtractString(properties, 'user_id')) AS value,
				JSONExtractString(properties, 'country') as country,
				JSONExtractString(properties, 'device') as device
			FROM rc_events
			WHERE rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, country, device`,
			wantArgs: []any{"organization String, user String, windowstart DateTime, windowend DateTime, value AggregateFunction(uniq, String), country String, device String", "windowstart, windowend, organization, user, country, device", "POPULATE", "user_login"},
			wantErr:  false,
		},
		{
			name: "Meter with count aggregation without value property",
			meter: CreateMeter{
				Slug:          "api_requests",
				EventType:     "api_request",
				ValueProperty: "",
				Properties:    []string{"endpoint", "method"},
				Aggregation:   models.AggregationCount,
				Populate:      false,
				TenantSlug:    "test_tenant",
			},
			wantSQL: `CREATE MATERIALIZED VIEW IF NOT EXISTS rc_test_tenant_api_requests_mv ( %s ) ENGINE = AggregatingMergeTree()
			ORDER BY (%s)
			 AS SELECT
				organization,
				user,
				tumbleStart(timestamp, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(timestamp, toIntervalMinute(1)) AS windowend,
				countState(*) AS value,
				JSONExtractString(properties, 'endpoint') as endpoint,
				JSONExtractString(properties, 'method') as method
			FROM rc_events
			WHERE rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, endpoint, method`,
			wantArgs: []any{"organization String, user String, windowstart DateTime, windowend DateTime, value AggregateFunction(count, Float64), endpoint String, method String", "windowstart, windowend, organization, user, endpoint, method", "", "api_request"},
			wantErr:  false,
		},
		{
			name: "Invalid aggregation type",
			meter: CreateMeter{
				Slug:          "invalid_meter",
				EventType:     "event",
				ValueProperty: "value",
				Properties:    []string{"property"},
				Aggregation:   "invalid_aggregation", // Invalid aggregation
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
			gotSQL, _, err := tt.meter.ToCreateSQL()

			// Check error expectation
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.wantSQL = fmt.Sprintf(tt.wantSQL, tt.wantArgs[0], tt.wantArgs[1])
			// Normalize and compare SQL
			assert.Equal(t, normalizeSQL(tt.wantSQL), normalizeSQL(gotSQL))
		})
	}
}

// Also test the toSeleteSQL method separately
func TestCreateMeterSelectSQL(t *testing.T) {
	tests := []struct {
		name      string
		meter     CreateMeter
		stateFunc string
		dataType  string
		wantSQL   string
		wantArgs  []any
	}{
		{
			name: "Select SQL for sum aggregation",
			meter: CreateMeter{
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
				tumbleStart(timestamp, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(timestamp, toIntervalMinute(1)) AS windowend,
				sumState(cast(JSONExtractString(properties, 'count'), 'Float64')) AS value,
				JSONExtractString(properties, 'path') as path,
				JSONExtractString(properties, 'referrer') as referrer
			FROM rc_events
			WHERE rc_events.type = ?
			GROUP BY windowstart, windowend, organization, user, path, referrer`,
			wantArgs: []any{"page_view"},
		},
		{
			name: "Select SQL for count aggregation",
			meter: CreateMeter{
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
				tumbleStart(timestamp, toIntervalMinute(1)) AS windowstart,
				tumbleEnd(timestamp, toIntervalMinute(1)) AS windowend,
				countState(*) AS value,
				JSONExtractString(properties, 'endpoint') as endpoint,
				JSONExtractString(properties, 'method') as method
			FROM rc_events
			WHERE rc_events.type = ?
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
