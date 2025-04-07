package clickhouse

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"regexp"
	"testing"
)

func getTestStore(t *testing.T) (*ClickHouseStore, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	zapLogger, _ := zap.NewDevelopment()
	testLogger := &logger.Logger{Logger: zapLogger}
	store := &ClickHouseStore{
		db:     sqlx.NewDb(db, "clickhouse"),
		logger: testLogger,
	}
	return store, mock
}

func TestCreateMeter_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		input    models.CreateMeterInput
		mockErr  error
		wantErr  bool
		errMatch string
	}{
		{
			name: "count aggregation",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "count_slug", EventType: "evt",
				Aggregation: "count", Properties: []string{"country"},
			},
		},
		{
			name: "sum aggregation",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "sum_slug", EventType: "evt",
				Aggregation: "sum", ValueProperty: "amount", Properties: []string{"user"},
			},
		},
		{
			name: "avg aggregation",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "avg_slug", EventType: "evt",
				Aggregation: "avg", ValueProperty: "duration", Properties: []string{"page"},
			},
		},
		{
			name: "min aggregation",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "min_slug", EventType: "evt",
				Aggregation: "min", ValueProperty: "latency", Properties: []string{"server"},
			},
		},
		{
			name: "max aggregation",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "max_slug", EventType: "evt",
				Aggregation: "max", ValueProperty: "latency", Properties: []string{"server"},
			},
		},
		{
			name: "unique_count with value property",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "unique_slug", EventType: "evt",
				Aggregation: "unique_count", ValueProperty: "session_id", Properties: []string{"browser"},
			},
		},
		{
			name: "unique_count missing value property",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "bad_unique", EventType: "evt",
				Aggregation: "unique_count", ValueProperty: "", Properties: []string{"country"},
			},
			wantErr:  true,
			errMatch: "value_property is required for unique_count",
		},
		{
			name: "invalid aggregation",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "bad_agg", EventType: "evt",
				Aggregation: "median", ValueProperty: "duration",
			},
			wantErr:  true,
			errMatch: "invalid aggregation type",
		},
		{
			name: "empty properties",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "no_props", EventType: "evt",
				Aggregation: "count", Properties: []string{},
			},
		},
		{
			name: "populate = true",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "populate_true", EventType: "evt",
				Aggregation: "sum", ValueProperty: "duration", Properties: []string{"url"}, Populate: true,
			},
		},
		{
			name: "populate = false",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "populate_false", EventType: "evt",
				Aggregation: "sum", ValueProperty: "duration", Properties: []string{"url"}, Populate: false,
			},
		},
		{
			name: "sql failure",
			input: models.CreateMeterInput{
				Organization: "org", Slug: "sql_fail", EventType: "evt",
				Aggregation: "sum", ValueProperty: "value", Properties: []string{"user"},
			},
			mockErr:  errors.New("clickhouse syntax error"),
			wantErr:  true,
			errMatch: "failed to create meter view",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store, mock := getTestStore(t)
			defer store.db.Close()

			if tc.mockErr != nil {
				mock.ExpectExec(regexp.QuoteMeta("create materialized view")).WillReturnError(tc.mockErr)
			} else if !tc.wantErr {
				mock.ExpectExec(regexp.QuoteMeta("create materialized view")).WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := store.CreateMeter(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMatch != "" {
					assert.Contains(t, err.Error(), tc.errMatch)
				}
			} else {
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
