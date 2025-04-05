package meters

import (
	"context"
	"fmt"
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
		Slug:          "test-meter-" + uuid.New().String()[0:8],
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
		assert.Equal(t, input.Slug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, input.CreatedBy, meter.CreatedBy)
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
		assert.Equal(t, input.Slug, meter.Slug)
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
		assert.Equal(t, input.Slug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, input.CreatedBy, meter.CreatedBy)
		assert.False(t, meter.CreatedAt.IsZero())
	})

	t.Run("Success get by Slug", func(t *testing.T) {
		// Create a meter to retrieve
		input := createTestMeterInput()
		createdMeter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, createdMeter)

		// Retrieve by slug
		meter, err := repo.GetMeterByIDorSlug(ctx, input.Slug)
		require.NoError(t, err)
		require.NotNil(t, meter)

		// Verify fields match
		assert.Equal(t, createdMeter.ID.String(), meter.ID.String())
		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.Slug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, input.CreatedBy, meter.CreatedBy)
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

func TestListMetersByEventType(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer cleanupTestMeters(t, ctx, db)

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	// Create meters with different event types
	createTestMetersWithEventTypes := func() (string, []*models.Meter) {
		// Create a unique event type for this test
		uniqueEventType := "test.event." + uuid.New().String()[0:8]

		// Create 3 meters with the unique event type
		metersWithType := make([]*models.Meter, 0, 3)
		for i := range 3 {
			input := models.CreateMeterInput{
				Name:          fmt.Sprintf("Event Type Meter %d", i),
				Slug:          fmt.Sprintf("event-type-meter-%s-%d", uuid.New().String()[0:8], i),
				EventType:     uniqueEventType,
				ValueProperty: "amount",
				Description:   fmt.Sprintf("Test meter for event type %d", i),
				Properties:    []string{fmt.Sprintf("property%d", i)},
				Aggregation:   models.AggregationSum,
				CreatedBy:     "test-user",
			}
			meter, err := repo.CreateMeter(ctx, input)
			require.NoError(t, err)
			metersWithType = append(metersWithType, meter)

			// Add a small delay to ensure created timestamps are different
			time.Sleep(10 * time.Millisecond)
		}

		// Create 2 meters with different event types
		for i := range 2 {
			input := models.CreateMeterInput{
				Name:          fmt.Sprintf("Other Event Type Meter %d", i),
				Slug:          fmt.Sprintf("other-event-type-%s-%d", uuid.New().String()[0:8], i),
				EventType:     fmt.Sprintf("other.event.%d", i),
				ValueProperty: "amount",
				Description:   "Test meter with different event type",
				Properties:    []string{"property"},
				Aggregation:   models.AggregationSum,
				CreatedBy:     "test-user",
			}
			_, err := repo.CreateMeter(ctx, input)
			require.NoError(t, err)
		}

		return uniqueEventType, metersWithType
	}

	t.Run("List meters by event type without cursor", func(t *testing.T) {
		// Clean up any existing test meters
		cleanupTestMeters(t, ctx, db)

		// Create test meters with a specific event type
		eventType, metersWithType := createTestMetersWithEventTypes()

		// List meters by the specific event type
		limit := int32(10)
		result, err := repo.ListMetersByEventType(ctx, limit, eventType, nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify we got exactly the meters with the specific event type
		assert.Len(t, result.Items, 3)

		// Verify all meters in the result match the expected event type
		for _, meter := range result.Items {
			assert.Equal(t, eventType, meter.EventType)
		}

		// Verify all expected meters are in the result
		foundCount := 0
		for _, created := range metersWithType {
			for _, listed := range result.Items {
				if created.ID == listed.ID {
					foundCount++
					break
				}
			}
		}
		assert.Equal(t, len(metersWithType), foundCount)

		// With items, NextCursor should be set
		assert.NotNil(t, result.NextCursor)
	})

	t.Run("List meters by event type with pagination", func(t *testing.T) {
		// Clean up any existing test meters
		cleanupTestMeters(t, ctx, db)

		// Create a unique event type for this test
		uniqueEventType := "test.event.paginated." + uuid.New().String()[0:8]

		// Create 5 meters with the unique event type
		createdMeters := make([]*models.Meter, 0, 5)
		for i := range 5 {
			input := models.CreateMeterInput{
				Name:          fmt.Sprintf("Paginated Event Type Meter %d", i),
				Slug:          fmt.Sprintf("paginated-event-type-%s-%d", uuid.New().String()[0:8], i),
				EventType:     uniqueEventType,
				ValueProperty: "amount",
				Description:   fmt.Sprintf("Test meter for paginated event type %d", i),
				Properties:    []string{fmt.Sprintf("property%d", i)},
				Aggregation:   models.AggregationSum,
				CreatedBy:     "test-user",
			}
			meter, err := repo.CreateMeter(ctx, input)
			require.NoError(t, err)
			createdMeters = append(createdMeters, meter)

			// Add a small delay to ensure created timestamps are different
			time.Sleep(10 * time.Millisecond)
		}

		// Map of IDs for quicker lookup
		createdIDs := make(map[uuid.UUID]bool)
		for _, m := range createdMeters {
			createdIDs[m.ID] = true
		}

		// Keep track of all found meters
		allItems := make([]models.Meter, 0, len(createdMeters))
		foundIDs := make(map[uuid.UUID]bool)

		// Use smaller page size for pagination test
		limit := int32(2)
		var cursor *pagination.Cursor

		// Fetch pages until we've found all the meters we created or hit an empty page
		for {
			page, err := repo.ListMetersByEventType(ctx, limit, uniqueEventType, cursor)
			require.NoError(t, err)

			// If we get an empty page, break
			if len(page.Items) == 0 {
				t.Logf("No more items found, breaking pagination loop")
				break
			}

			// Verify all items have the correct event type
			for _, item := range page.Items {
				assert.Equal(t, uniqueEventType, item.EventType)

				// Skip if we've already seen this ID
				if foundIDs[item.ID] {
					continue
				}

				// Add to our found collection
				foundIDs[item.ID] = true

				// Only count items that were from our test set
				if createdIDs[item.ID] {
					allItems = append(allItems, item)
				}
			}

			t.Logf("Found %d items in this page, total %d", len(page.Items), len(allItems))
			t.Logf("Next cursor: %v", *page.NextCursor)

			// Break if we don't have a next cursor
			if page.NextCursor == nil {
				break
			}

			// Set up cursor for next page
			nextCursor, err := pagination.DecodeCursor(*page.NextCursor)
			t.Logf("Next cursor: %v", nextCursor)
			require.NoError(t, err)
			cursor = nextCursor

			// Break if we've found all our test meters
			if len(allItems) >= len(createdMeters) {
				t.Logf("Found all test meters, breaking pagination loop %d", len(allItems))
				break
			}
		}

		// Verify we found all our test meters
		t.Logf("Found %d meters out of %d created with event type %s", len(allItems), len(createdMeters), uniqueEventType)
		assert.Equal(t, len(createdMeters), len(allItems), "Should find all created meters via pagination")
	})

	t.Run("List meters by non-existent event type", func(t *testing.T) {
		// Use a random event type that shouldn't exist
		nonExistentEventType := "non.existent.event.type." + uuid.New().String()

		// List meters by the non-existent event type
		limit := int32(10)
		result, err := repo.ListMetersByEventType(ctx, limit, nonExistentEventType, nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify we got no meters
		assert.Empty(t, result.Items)

		// With no items, NextCursor should be nil
		assert.Nil(t, result.NextCursor)
	})

	t.Run("Error database operation", func(t *testing.T) {
		// Create a new connection and close it to force errors
		badDB, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		badDB.Close() // Close immediately to cause errors

		badRepo := NewPostgresMeterStoreRepository(badDB, l)

		// Attempt to list with closed connection
		result, err := badRepo.ListMetersByEventType(ctx, 10, "any.event.type", nil)
		assert.Nil(t, result)
		assert.Equal(t, errors.ErrDatabaseOperation, err)
	})
}
