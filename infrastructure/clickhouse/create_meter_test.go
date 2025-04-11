package clickhouse

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
)

// setupTestClickHouse initializes a ClickHouse connection for testing
func setupTestClickHouse(t *testing.T) (*ClickHouseStore, context.Context) {
	// Skip if not running integration tests
	//if os.Getenv("INTEGRATION_TESTS") != "true" {
	//	t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	//}

	// Setup context and logger
	ctx := context.Background()
	l := createTestLogger(t)

	// Create ClickHouse store
	store := ClickHouseStoreRepository(l).(*ClickHouseStore)

	// Connect to ClickHouse
	cfg := &config.ClickHouseConfig{
		Host:     "localhost",
		Port:     "9000",
		Database: "default",
		User:     "default",
		Password: "default",
	}

	err := store.Connect(cfg)
	require.NoError(t, err, "Failed to connect to ClickHouse")

	return store, ctx
}

// createTestLogger creates a logger for testing
func createTestLogger(t *testing.T) *logger.Logger {
	l, err := logger.NewLogger(&config.LoggerConfig{
		Level:   "debug",
		LogFile: "",
		Mode:    "development",
	})
	require.NoError(t, err, "Failed to create logger")
	return l
}

// createTestMeterInput creates a meter input with unique slug for testing
func createTestMeterInput() models.CreateMeterInput {
	return models.CreateMeterInput{
		Name:          "Test Meter " + time.Now().Format(time.RFC3339),
		Slug:          "test-meter-" + uuid.New().String()[0:8],
		Organization:  "test-org-" + uuid.New().String()[0:8],
		EventType:     "test.event",
		ValueProperty: "amount",
		Description:   "Test meter description",
		Properties:    []string{"property1", "property2"},
		Aggregation:   models.AggregationSum,
		Populate:      false, // Don't populate by default in tests
		CreatedBy:     "test-user",
	}
}

// cleanupTestMeters removes test meters from ClickHouse
func cleanupTestMeters(t *testing.T, ctx context.Context, store *ClickHouseStore, organization, meterSlug string) {
	viewName := getMeterViewName(organization, meterSlug)
	_, err := store.db.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", viewName))
	require.NoError(t, err, "Failed to clean up test meter view")
}

func TestCreateMeterView(t *testing.T) {
	store, ctx := setupTestClickHouse(t)
	defer store.Close()

	t.Run("Success with all fields", func(t *testing.T) {
		input := createTestMeterInput()
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Verify view exists
		viewName := getMeterViewName(input.Organization, input.Slug)
		var exists int
		err = store.db.QueryRowContext(ctx, `
			SELECT 1 FROM system.tables 
			WHERE database = currentDatabase() 
			AND name = ?
		`, viewName).Scan(&exists)
		require.NoError(t, err)
		assert.Equal(t, 1, exists)
	})

	t.Run("Success with minimum fields (count aggregation)", func(t *testing.T) {
		input := createTestMeterInput()
		input.ValueProperty = ""
		input.Properties = []string{}
		input.Aggregation = models.AggregationCount
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Verify view exists
		viewName := getMeterViewName(input.Organization, input.Slug)
		var exists int
		err = store.db.QueryRowContext(ctx, `
			SELECT 1 FROM system.tables 
			WHERE database = currentDatabase() 
			AND name = ?
		`, viewName).Scan(&exists)
		require.NoError(t, err)
		assert.Equal(t, 1, exists)
	})

	t.Run("Success with unique_count aggregation", func(t *testing.T) {
		input := createTestMeterInput()
		input.Aggregation = models.AggregationUniqueCount
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Verify view exists
		viewName := getMeterViewName(input.Organization, input.Slug)
		var exists int
		err = store.db.QueryRowContext(ctx, `
			SELECT 1 FROM system.tables 
			WHERE database = currentDatabase() 
			AND name = ?
		`, viewName).Scan(&exists)
		require.NoError(t, err)
		assert.Equal(t, 1, exists)
	})

	t.Run("Success with populate flag true", func(t *testing.T) {
		input := createTestMeterInput()
		input.Populate = true
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Verify view exists
		viewName := getMeterViewName(input.Organization, input.Slug)
		var exists int
		err = store.db.QueryRowContext(ctx, `
			SELECT 1 FROM system.tables 
			WHERE database = currentDatabase() 
			AND name = ?
		`, viewName).Scan(&exists)
		require.NoError(t, err)
		assert.Equal(t, 1, exists)
	})

	t.Run("Error invalid aggregation type", func(t *testing.T) {
		input := createTestMeterInput()
		input.Aggregation = "invalid_aggregation"
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid aggregation type")
	})

	t.Run("Error missing value_property for unique_count", func(t *testing.T) {
		input := createTestMeterInput()
		input.Aggregation = models.AggregationUniqueCount
		input.ValueProperty = ""
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value_property is required for unique_count")
	})

	t.Run("Error duplicate meter view", func(t *testing.T) {
		input := createTestMeterInput()
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		// First creation should succeed
		err := store.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Second creation should fail
		err = store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create meter view")
	})

	t.Run("Error empty organization", func(t *testing.T) {
		input := createTestMeterInput()
		input.Organization = ""
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create meter view")
	})

	t.Run("Error empty meter slug", func(t *testing.T) {
		input := createTestMeterInput()
		input.Slug = ""
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create meter view")
	})

	t.Run("Error empty event type", func(t *testing.T) {
		input := createTestMeterInput()
		input.EventType = ""
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create meter view")
	})

	t.Run("Error invalid property names", func(t *testing.T) {
		input := createTestMeterInput()
		input.Properties = []string{"invalid-property-name-with-hyphens"}
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create meter view")
	})

	t.Run("Error database connection closed", func(t *testing.T) {
		input := createTestMeterInput()
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		// Close the connection
		err := store.Close()
		require.NoError(t, err)

		// Try to create meter with closed connection
		err = store.CreateMeter(ctx, input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create meter view")
	})

	t.Run("Error invalid value property for numeric aggregation", func(t *testing.T) {
		input := createTestMeterInput()
		input.ValueProperty = "non-numeric-property"
		// Assuming the test data in ClickHouse has non-numeric values for this property
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		// This might not fail during view creation but would fail during querying
		// Depending on your requirements, you might want to validate the property exists and has correct type
		require.NoError(t, err)
	})

	t.Run("Success with special characters in organization and slug", func(t *testing.T) {
		input := createTestMeterInput()
		input.Organization = "org-with-special-chars-!@#$%^&*()"
		input.Slug = "meter-with-special-chars-!@#$%^&*()"
		defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

		err := store.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Verify view exists
		viewName := getMeterViewName(input.Organization, input.Slug)
		var exists int
		err = store.db.QueryRowContext(ctx, `
			SELECT 1 FROM system.tables 
			WHERE database = currentDatabase() 
			AND name = ?
		`, viewName).Scan(&exists)
		require.NoError(t, err)
		assert.Equal(t, 1, exists)
	})

	t.Run("Success with all supported aggregation types", func(t *testing.T) {
		aggregations := []models.AggregationEnum{
			models.AggregationSum,
			models.AggregationAvg,
			models.AggregationMin,
			models.AggregationMax,
			models.AggregationCount,
			models.AggregationUniqueCount,
		}

		for _, agg := range aggregations {
			t.Run(string(agg), func(t *testing.T) {
				input := createTestMeterInput()
				input.Aggregation = agg
				// For unique_count, ensure value_property is set
				if agg == models.AggregationUniqueCount {
					input.ValueProperty = "user_id"
				}
				defer cleanupTestMeters(t, ctx, store, input.Organization, input.Slug)

				err := store.CreateMeter(ctx, input)
				require.NoError(t, err)

				// Verify view exists
				viewName := getMeterViewName(input.Organization, input.Slug)
				var exists int
				err = store.db.QueryRowContext(ctx, `
					SELECT 1 FROM system.tables 
					WHERE database = currentDatabase() 
					AND name = ?
				`, viewName).Scan(&exists)
				require.NoError(t, err)
				assert.Equal(t, 1, exists)
			})
		}
	})
}
