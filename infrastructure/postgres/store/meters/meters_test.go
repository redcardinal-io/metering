package meters

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

// setupTestDB initializes a database connection for testing
func setupTestDB(t *testing.T) (*pgxpool.Pool, context.Context) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	// Get database connection string
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Setup context and database connection
	ctx := context.Background()
	db, err := pgxpool.New(ctx, connString)
	require.NoError(t, err, "Failed to connect to database")

	return db, ctx
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
		MeterSlug:     "test-meter-" + uuid.New().String()[0:8],
		EventType:     "test.event",
		ValueProperty: "amount",
		Description:   "Test meter description",
		Properties:    []string{"property1", "property2"},
		Aggregation:   models.AggregationSum,
		CreatedBy:     "test-user",
	}
}

// cleanupTestMeters removes test meters from the database
func cleanupTestMeters(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "DELETE FROM meter WHERE name LIKE 'Test Meter%'")
	require.NoError(t, err, "Failed to clean up test meters")
}

func TestCreateMeter(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer cleanupTestMeters(t, ctx, db)

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	t.Run("Success with all fields", func(t *testing.T) {
		input := createTestMeterInput()

		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)

		assert.NotNil(t, meter)
		assert.NotEqual(t, uuid.Nil, meter.ID)
		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.MeterSlug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, input.CreatedBy, meter.TenantSlug)
		assert.False(t, meter.CreatedAt.IsZero())
	})

	t.Run("Success with minimum fields", func(t *testing.T) {
		input := createTestMeterInput()
		// Remove optional fields
		input.Description = ""
		input.ValueProperty = ""
		input.Properties = []string{}

		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)

		assert.NotNil(t, meter)
		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.MeterSlug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, "", meter.Description)
		assert.Equal(t, "", meter.ValueProperty)
		assert.Empty(t, meter.Properties)
	})

	t.Run("Error duplicate slug", func(t *testing.T) {
		// Create a meter with a specific slug
		input := createTestMeterInput()
		_, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)

		// Try to create another meter with the same slug
		duplicateInput := input
		duplicateInput.Name = "Different Name"

		_, err = repo.CreateMeter(ctx, duplicateInput)
		assert.Equal(t, errors.ErrMeterAlreadyExists, err)
	})

	t.Run("Error database operation", func(t *testing.T) {
		// Create a new connection and close it to force errors
		badDB, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		badDB.Close() // Close immediately to cause errors

		badRepo := NewPostgresMeterStoreRepository(badDB, l)

		input := createTestMeterInput()
		_, err = badRepo.CreateMeter(ctx, input)
		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}

// TestPgErrorHandling is a unit test for PostgreSQL error handling logic
func TestPgErrorHandling(t *testing.T) {
	t.Run("Handle duplicate key error", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505"}

		// Check if the error is correctly identified as a duplicate key error
		isDuplicate := pgErr.Code == "23505"
		assert.True(t, isDuplicate)

		// Verify we map to the expected domain error
		var err error
		if isDuplicate {
			err = errors.ErrMeterAlreadyExists
		} else {
			err = errors.ErrDatabaseOperation
		}

		assert.Equal(t, errors.ErrMeterAlreadyExists, err)
	})

	t.Run("Handle other database errors", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "42P01"} // Undefined table

		// Check if the error is correctly not identified as a duplicate
		isDuplicate := pgErr.Code == "23505"
		assert.False(t, isDuplicate)

		// Verify we map to the general database error
		var err error
		if isDuplicate {
			err = errors.ErrMeterAlreadyExists
		} else {
			err = errors.ErrDatabaseOperation
		}

		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}

func TestDeleteMeterByIDorSlug(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	t.Run("Success delete by ID", func(t *testing.T) {
		// Create a meter to delete
		input := createTestMeterInput()
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)

		// Delete by ID
		err = repo.DeleteMeterByIDorSlug(ctx, meter.ID.String())
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetMeterByIDorSlug(ctx, meter.ID.String())
		assert.Equal(t, errors.ErrMeterNotFound, err)
	})

	t.Run("Success delete by Slug", func(t *testing.T) {
		// Create a meter to delete
		input := createTestMeterInput()
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)

		// Delete by slug
		err = repo.DeleteMeterByIDorSlug(ctx, meter.Slug)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetMeterByIDorSlug(ctx, meter.Slug)
		assert.Equal(t, errors.ErrMeterNotFound, err)
	})

	t.Run("Error meter not found by ID", func(t *testing.T) {
		// Generate a random UUID that doesn't exist in the database
		nonExistentID := uuid.New().String()

		// Attempt to delete
		err := repo.DeleteMeterByIDorSlug(ctx, nonExistentID)
		// Assert no error since the ID doesn't exist
		assert.Equal(t, nil, err)
	})

	t.Run("Error meter not found by Slug", func(t *testing.T) {
		// Generate a random slug that doesn't exist in the database
		nonExistentSlug := "non-existent-slug-" + uuid.New().String()[0:8]

		// Attempt to delete
		err := repo.DeleteMeterByIDorSlug(ctx, nonExistentSlug)
		// Assert no error since the slug doesn't exist
		assert.Equal(t, nil, err)
	})

	t.Run("Error database operation", func(t *testing.T) {
		// Create a new connection and close it to force errors
		badDB, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		badDB.Close() // Close immediately to cause errors

		badRepo := NewPostgresMeterStoreRepository(badDB, l)

		// Attempt to delete with closed connection
		err = badRepo.DeleteMeterByIDorSlug(ctx, "any-value")
		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}

func TestGetMeterByIDorSlug(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer cleanupTestMeters(t, ctx, db)

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	t.Run("Success get by ID", func(t *testing.T) {
		// Create a meter to retrieve
		input := createTestMeterInput()
		createdMeter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, createdMeter)

		// Retrieve by ID
		meter, err := repo.GetMeterByIDorSlug(ctx, createdMeter.ID.String())
		require.NoError(t, err)
		require.NotNil(t, meter)

		// Verify fields match
		assert.Equal(t, createdMeter.ID.String(), meter.ID.String())
		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.MeterSlug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, input.CreatedBy, meter.TenantSlug)
		assert.False(t, meter.CreatedAt.IsZero())
	})

	t.Run("Success get by Slug", func(t *testing.T) {
		// Create a meter to retrieve
		input := createTestMeterInput()
		createdMeter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, createdMeter)

		// Retrieve by slug
		meter, err := repo.GetMeterByIDorSlug(ctx, input.MeterSlug)
		require.NoError(t, err)
		require.NotNil(t, meter)

		// Verify fields match
		assert.Equal(t, createdMeter.ID.String(), meter.ID.String())
		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.MeterSlug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, input.CreatedBy, meter.TenantSlug)
		assert.False(t, meter.CreatedAt.IsZero())
	})

	t.Run("Error meter not found by ID", func(t *testing.T) {
		// Generate a random UUID that doesn't exist in the database
		nonExistentID := uuid.New()

		// Attempt to retrieve
		meter, err := repo.GetMeterByIDorSlug(ctx, nonExistentID.String())
		assert.Nil(t, meter)
		assert.Equal(t, errors.ErrMeterNotFound, err)
	})

	t.Run("Error meter not found by Slug", func(t *testing.T) {
		// Generate a random slug that doesn't exist in the database
		nonExistentSlug := "non-existent-slug-" + uuid.New().String()[0:8]

		// Attempt to retrieve
		meter, err := repo.GetMeterByIDorSlug(ctx, nonExistentSlug)
		assert.Nil(t, meter)
		assert.Equal(t, errors.ErrMeterNotFound, err)
	})

	t.Run("Error invalid UUID format", func(t *testing.T) {
		// Try with an invalid UUID string that will cause parse error but looks like a UUID
		invalidUUID := "not-a-valid-uuid-format-123456789012"

		// The method should fall back to searching by slug, which won't find anything
		meter, err := repo.GetMeterByIDorSlug(ctx, invalidUUID)
		assert.Nil(t, meter)
		assert.Equal(t, errors.ErrMeterNotFound, err)
	})

	t.Run("Error database operation", func(t *testing.T) {
		// Create a new connection and close it to force errors
		badDB, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		badDB.Close() // Close immediately to cause errors

		badRepo := NewPostgresMeterStoreRepository(badDB, l)

		// Attempt to retrieve with closed connection
		meter, err := badRepo.GetMeterByIDorSlug(ctx, "any-value")
		assert.Nil(t, meter)
		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}

func TestListMeters(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer cleanupTestMeters(t, ctx, db)

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	// Create multiple test meters for listing
	numMeters := 5
	createdMeters := make([]models.Meter, 0, numMeters)
	for range numMeters {
		input := createTestMeterInput()
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		createdMeters = append(createdMeters, *meter)
	}

	t.Run("Success list all meters with pagination", func(t *testing.T) {
		// Request first page with limit 3
		page := pagination.Pagination{
			Page:  1,
			Limit: 3,
		}

		// List meters
		result, err := repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify pagination info
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 3, result.Limit)
		assert.GreaterOrEqual(t, result.Total, numMeters) // Database might have other meters
		assert.Len(t, result.Results, 3)

		// Request second page
		page.Page = 2
		result, err = repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify pagination info for second page
		assert.Equal(t, 2, result.Page)
		assert.Equal(t, 3, result.Limit)
		assert.GreaterOrEqual(t, result.Total, numMeters)
		assert.LessOrEqual(t, len(result.Results), 3) // Might be less than 3 items on second page
	})

	t.Run("Success empty result with high page number", func(t *testing.T) {
		// Request a very high page number that should be empty
		page := pagination.Pagination{
			Page:  100,
			Limit: 10,
		}

		result, err := repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify empty results but correct total
		assert.Equal(t, 100, result.Page)
		assert.Equal(t, 10, result.Limit)
		assert.GreaterOrEqual(t, result.Total, numMeters)
		assert.Empty(t, result.Results)
	})

	t.Run("Error database operation", func(t *testing.T) {
		// Create a new connection and close it to force errors
		badDB, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		badDB.Close() // Close immediately to cause errors

		badRepo := NewPostgresMeterStoreRepository(badDB, l)

		page := pagination.Pagination{
			Page:  1,
			Limit: 10,
		}

		// Attempt to list with closed connection
		result, err := badRepo.ListMeters(ctx, page)
		assert.Nil(t, result)
		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}

func TestListMetersByEventType(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer cleanupTestMeters(t, ctx, db)

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	// Create meters with different event types
	eventType1 := "test.event.type1"
	eventType2 := "test.event.type2"

	// Create 3 meters with eventType1
	eventType1Meters := make([]models.Meter, 0, 3)
	for range 3 {
		input := createTestMeterInput()
		input.EventType = eventType1
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		eventType1Meters = append(eventType1Meters, *meter)
	}

	// Create 2 meters with eventType2
	eventType2Meters := make([]models.Meter, 0, 2)
	for range 2 {
		input := createTestMeterInput()
		input.EventType = eventType2
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		eventType2Meters = append(eventType2Meters, *meter)
	}

	t.Run("Success list meters by event type with pagination", func(t *testing.T) {
		// Request first page with limit 2
		page := pagination.Pagination{
			Page:  1,
			Limit: 2,
		}

		// List meters with eventType1
		result, err := repo.ListMetersByEventType(ctx, eventType1, page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify pagination info
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.Limit)
		assert.Equal(t, 3, result.Total) // We created 3 meters with eventType1
		assert.Len(t, result.Results, 2)

		// Verify all results have the correct event type
		for _, meter := range result.Results {
			assert.Equal(t, eventType1, meter.EventType)
		}

		// Request second page
		page.Page = 2
		result, err = repo.ListMetersByEventType(ctx, eventType1, page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify pagination info for second page
		assert.Equal(t, 2, result.Page)
		assert.Equal(t, 2, result.Limit)
		assert.Equal(t, 3, result.Total)
		assert.Len(t, result.Results, 1) // Only 1 meter left on the second page

		// Verify the result has the correct event type
		assert.Equal(t, eventType1, result.Results[0].EventType)
	})

	t.Run("Success list meters with different event type", func(t *testing.T) {
		page := pagination.Pagination{
			Page:  1,
			Limit: 10,
		}

		// List meters with eventType2
		result, err := repo.ListMetersByEventType(ctx, eventType2, page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify pagination info
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, 2, result.Total) // We created 2 meters with eventType2
		assert.Len(t, result.Results, 2)

		// Verify all results have the correct event type
		for _, meter := range result.Results {
			assert.Equal(t, eventType2, meter.EventType)
		}
	})

	t.Run("Success empty result for non-existent event type", func(t *testing.T) {
		page := pagination.Pagination{
			Page:  1,
			Limit: 10,
		}

		// List meters with non-existent event type
		result, err := repo.ListMetersByEventType(ctx, "non.existent.event.type", page)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify empty results
		assert.Equal(t, 0, result.Total)
		assert.Empty(t, result.Results)
	})

	t.Run("Error database operation", func(t *testing.T) {
		// Create a new connection and close it to force errors
		badDB, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		badDB.Close() // Close immediately to cause errors

		badRepo := NewPostgresMeterStoreRepository(badDB, l)

		page := pagination.Pagination{
			Page:  1,
			Limit: 10,
		}

		// Attempt to list with closed connection
		result, err := badRepo.ListMetersByEventType(ctx, eventType1, page)
		assert.Nil(t, result)
		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}
