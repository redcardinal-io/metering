package meters

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
)

const testTenantSlug = "test-tenant-123"

func setupTestDB(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true environment variable to run.")
	}

	// Get database connection string
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		t.Fatal("DATABASE_URL environment variable not set, skipping integration test") // Use Fatalf for setup errors
	}

	ctx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
	db, err := pgxpool.New(ctx, connString)
	require.NoError(t, err, "Failed to connect to database")

	cleanupTestMeters(t, ctx, db)

	return db, ctx
}

// createTestLogger creates a logger for testing
func createTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	l, err := logger.NewLogger(&config.LoggerConfig{
		Level:   "debug", // Use debug for more info during tests if needed
		LogFile: "",      // Don't write to file during tests
		Mode:    "development",
	})
	require.NoError(t, err, "Failed to create logger")
	return l
}

// createTestMeterInput creates a meter input with unique slug for testing
func createTestMeterInput(baseSuffix string) models.CreateMeterInput {
	uniquePart := uuid.New().String()
	finalSuffix := baseSuffix
	if finalSuffix != "" {
		finalSuffix += "-" + uniquePart
	} else {
		finalSuffix = uniquePart
	}

	return models.CreateMeterInput{
		Name:          "Test Meter " + finalSuffix,
		MeterSlug:     "test-meter-" + finalSuffix,
		EventType:     "test.event." + finalSuffix,
		ValueProperty: "amount",
		Description:   "Test meter description " + finalSuffix,
		Properties:    []string{"prop1", "prop2"},
		Aggregation:   models.AggregationSum,
		CreatedBy:     "test-user-" + finalSuffix,
	}
}

func cleanupTestMeters(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()
	tenant, ok := ctx.Value(constants.TenantSlugKey).(string)
	if !ok || tenant == "" {
		t.Logf("Skipping cleanup: Tenant slug missing or empty in context.") // Log instead of Fatal
		return
	}
	_, err := db.Exec(ctx, "DELETE FROM meter WHERE tenant_slug = $1 AND name LIKE 'Test Meter %'", tenant)
	if err != nil {
		t.Errorf("Failed to clean up test meters for tenant %s: %v", tenant, err)
	}
}

func TestCreateMeter_Integration(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer func() {
		cleanupCtx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
		cleanupTestMeters(t, cleanupCtx, db)
	}()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	t.Run("Success with all fields", func(t *testing.T) {
		input := createTestMeterInput("all-fields")
		input.Properties = []string{"propA", "propB"}

		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)

		assert.NotEqual(t, uuid.Nil, meter.ID)
		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.MeterSlug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, input.ValueProperty, meter.ValueProperty)
		assert.Equal(t, input.Description, meter.Description)
		assert.Equal(t, input.Properties, meter.Properties)
		assert.Equal(t, input.Aggregation, meter.Aggregation)
		assert.Equal(t, testTenantSlug, meter.TenantSlug)
		assert.Equal(t, input.CreatedBy, meter.CreatedBy)
		assert.Equal(t, input.CreatedBy, meter.UpdatedBy)
		assert.NotEmpty(t, meter.CreatedAt)
		assert.NotEmpty(t, meter.UpdatedAt)
		assert.WithinDuration(t, time.Now(), meter.CreatedAt, 10*time.Second)
		assert.WithinDuration(t, time.Now(), meter.UpdatedAt, 10*time.Second)
	})

	t.Run("Success with minimum fields (empty properties array)", func(t *testing.T) {
		input := createTestMeterInput("minimal-empty-props")
		input.Description = ""
		input.ValueProperty = ""
		input.Properties = []string{}

		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)

		assert.Equal(t, input.Name, meter.Name)
		assert.Equal(t, input.MeterSlug, meter.Slug)
		assert.Equal(t, input.EventType, meter.EventType)
		assert.Equal(t, "", meter.Description)
		assert.Equal(t, "", meter.ValueProperty)

		assert.NotNil(t, meter.Properties, "Properties should be a non-nil empty slice, not nil")
		assert.Len(t, meter.Properties, 0, "Properties should be an empty slice")

		assert.Equal(t, testTenantSlug, meter.TenantSlug)
	})

	t.Run("Error null properties for NOT NULL column", func(t *testing.T) {
		input := createTestMeterInput("null-props-fail")
		input.Properties = nil

		_, err := repo.CreateMeter(ctx, input)
		require.Error(t, err, "Creating meter with nil properties should fail due to NOT NULL constraint")

		assert.NotNil(t, err)

		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr, "Expected error to wrap a PgError for NOT NULL violation") {
			assert.Equal(t, "23502", pgErr.Code, "Expected PostgreSQL NOT NULL violation code 23502")
		}
	})

	t.Run("Error duplicate slug", func(t *testing.T) {
		inputSfx := "duplicate-slug-base" // Base part of the suffix
		input := createTestMeterInput(inputSfx)

		_, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err, "First creation should succeed")

		fixedUniqueSlug := "test-meter-fixed-duplicate-" + uuid.NewString()[:8]
		fixedInput1 := models.CreateMeterInput{
			Name:        "Test Meter Fixed Duplicate 1",
			MeterSlug:   fixedUniqueSlug,
			EventType:   "test.event.fixeddup1",
			Properties:  []string{"prop"},
			Aggregation: models.AggregationSum,
			CreatedBy:   "user1",
		}
		fixedInput2 := models.CreateMeterInput{
			Name:        "Test Meter Fixed Duplicate 2",
			MeterSlug:   fixedUniqueSlug,
			EventType:   "test.event.fixeddup2",
			Properties:  []string{"prop"},
			Aggregation: models.AggregationSum,
			CreatedBy:   "user2",
		}

		_, err = repo.CreateMeter(ctx, fixedInput1)
		require.NoError(t, err, "Creation of first meter with fixed slug should succeed")

		_, err = repo.CreateMeter(ctx, fixedInput2)
		require.Error(t, err, "Second creation with the exact same slug should fail")

		assert.NotNil(t, err)
		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr, "Expected error to wrap a PgError") {
			assert.Equal(t, "23505", pgErr.Code, "Expected PostgreSQL unique violation code 23505")
		}
	})

	t.Run("Error invalid aggregation type", func(t *testing.T) {
		input := createTestMeterInput("invalid-agg")
		input.Aggregation = "INVALID_AGG_TYPE"

		_, err := repo.CreateMeter(ctx, input)
		require.Error(t, err, "Creation with invalid aggregation type should fail")

		assert.NotNil(t, err)
		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr) {
			assert.Contains(t, []string{"22P02", "23514"}, pgErr.Code, "Expected invalid text representation (22P02) or check constraint violation (23514)")
		}
	})
}

func TestUpdateMeterByIDorSlug_Integration(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer func() {
		cleanupCtx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
		cleanupTestMeters(t, cleanupCtx, db)
	}()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	initialInput := createTestMeterInput("update-target")
	createdMeter, err := repo.CreateMeter(ctx, initialInput)
	require.NoError(t, err, "Setup: Failed to create initial meter for update tests. Slug: %s", initialInput.MeterSlug)
	require.NotNil(t, createdMeter)
	originalUpdatedAt := createdMeter.UpdatedAt

	time.Sleep(50 * time.Millisecond)

	t.Run("Success update by ID", func(t *testing.T) {
		updateInput := models.UpdateMeterInput{
			Name:        "Updated Name by ID",
			Description: "Updated Description by ID",
			UpdatedBy:   "updater-id",
		}

		time.Sleep(50 * time.Millisecond)
		updatedMeter, err := repo.UpdateMeterByIDorSlug(ctx, createdMeter.ID.String(), updateInput)
		require.NoError(t, err)
		require.NotNil(t, updatedMeter)

		assert.Equal(t, createdMeter.ID, updatedMeter.ID)
		assert.Equal(t, updateInput.Name, updatedMeter.Name)
		assert.Equal(t, updateInput.Description, updatedMeter.Description)
		assert.Equal(t, updateInput.UpdatedBy, updatedMeter.UpdatedBy)
		assert.Equal(t, testTenantSlug, updatedMeter.TenantSlug)

		assert.Equal(t, createdMeter.Slug, updatedMeter.Slug)
		assert.Equal(t, createdMeter.EventType, updatedMeter.EventType)
		assert.Equal(t, createdMeter.ValueProperty, updatedMeter.ValueProperty)
		assert.Equal(t, createdMeter.Properties, updatedMeter.Properties)
		assert.Equal(t, createdMeter.Aggregation, updatedMeter.Aggregation)
		assert.Equal(t, createdMeter.CreatedBy, updatedMeter.CreatedBy)
		assert.Equal(t, createdMeter.CreatedAt, updatedMeter.CreatedAt)
		assert.True(t, updatedMeter.UpdatedAt.After(originalUpdatedAt), "UpdatedAt (%v) should be newer than original (%v)", updatedMeter.UpdatedAt, originalUpdatedAt)
		originalUpdatedAt = updatedMeter.UpdatedAt
	})

	t.Run("Success update by Slug", func(t *testing.T) {
		updateInput := models.UpdateMeterInput{
			Name:        "Updated Name by Slug",
			Description: "Updated Description by Slug",
			UpdatedBy:   "updater-slug",
		}

		time.Sleep(50 * time.Millisecond)
		updatedMeter, err := repo.UpdateMeterByIDorSlug(ctx, createdMeter.Slug, updateInput)
		require.NoError(t, err)
		require.NotNil(t, updatedMeter)

		assert.Equal(t, createdMeter.ID, updatedMeter.ID)
		assert.Equal(t, updateInput.Name, updatedMeter.Name)
		assert.Equal(t, updateInput.Description, updatedMeter.Description)
		assert.Equal(t, updateInput.UpdatedBy, updatedMeter.UpdatedBy)
		assert.Equal(t, testTenantSlug, updatedMeter.TenantSlug)

		assert.Equal(t, createdMeter.Slug, updatedMeter.Slug)
		assert.Equal(t, createdMeter.EventType, updatedMeter.EventType)
		assert.True(t, updatedMeter.UpdatedAt.After(originalUpdatedAt), "UpdatedAt (%v) should be newer than previous (%v)", updatedMeter.UpdatedAt, originalUpdatedAt)
		originalUpdatedAt = updatedMeter.UpdatedAt
	})

	t.Run("Success update only Name by ID (empty description ignored)", func(t *testing.T) {
		updateInput := models.UpdateMeterInput{
			Name:      "Only Name Updated",
			UpdatedBy: "updater-name-only",
		}
		time.Sleep(50 * time.Millisecond)

		currentDescription := "Updated Description by Slug"

		updatedMeter, err := repo.UpdateMeterByIDorSlug(ctx, createdMeter.ID.String(), updateInput)
		require.NoError(t, err)
		require.NotNil(t, updatedMeter)

		assert.Equal(t, updateInput.Name, updatedMeter.Name)
		assert.Equal(t, updateInput.UpdatedBy, updatedMeter.UpdatedBy)
		assert.Equal(t, currentDescription, updatedMeter.Description)
		assert.True(t, updatedMeter.UpdatedAt.After(originalUpdatedAt), "UpdatedAt (%v) should be newer than previous (%v)", updatedMeter.UpdatedAt, originalUpdatedAt)
		originalUpdatedAt = updatedMeter.UpdatedAt
	})

	t.Run("Success update only Description by Slug (empty name ignored)", func(t *testing.T) {
		updateInput := models.UpdateMeterInput{
			Description: "Only Description Updated",
			UpdatedBy:   "updater-desc-only",
		}
		time.Sleep(50 * time.Millisecond)
		currentName := "Only Name Updated"

		updatedMeter, err := repo.UpdateMeterByIDorSlug(ctx, createdMeter.Slug, updateInput)
		require.NoError(t, err)
		require.NotNil(t, updatedMeter)

		assert.Equal(t, updateInput.Description, updatedMeter.Description)
		assert.Equal(t, updateInput.UpdatedBy, updatedMeter.UpdatedBy)
		assert.Equal(t, currentName, updatedMeter.Name)
		assert.True(t, updatedMeter.UpdatedAt.After(originalUpdatedAt), "UpdatedAt (%v) should be newer than previous (%v)", updatedMeter.UpdatedAt, originalUpdatedAt)
	})

	t.Run("Error update non-existent ID", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		updateInput := models.UpdateMeterInput{Name: "Update Fail NonExistent", UpdatedBy: "updater-fail"}

		_, err := repo.UpdateMeterByIDorSlug(ctx, nonExistentID, updateInput)
		require.Error(t, err)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected pgx.ErrNoRows to be wrapped")
	})

	t.Run("Error update non-existent Slug", func(t *testing.T) {
		nonExistentSlug := "non-existent-slug-" + uuid.NewString()
		updateInput := models.UpdateMeterInput{Name: "Update Fail NonExistent", UpdatedBy: "updater-fail"}

		_, err := repo.UpdateMeterByIDorSlug(ctx, nonExistentSlug, updateInput)
		require.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected pgx.ErrNoRows to be wrapped")
	})

	t.Run("Error update with invalid ID format (falls back to slug)", func(t *testing.T) {
		invalidIDFormat := "this-is-definitely-not-a-uuid"
		updateInput := models.UpdateMeterInput{Name: "Update Fail Invalid ID", UpdatedBy: "updater-fail"}

		_, err := repo.UpdateMeterByIDorSlug(ctx, invalidIDFormat, updateInput)
		require.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected pgx.ErrNoRows to be wrapped")
	})

	t.Run("Attempt update meter from different tenant (by ID)", func(t *testing.T) {
		ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "different-tenant-456")
		updateInput := models.UpdateMeterInput{Name: "Tenant Mismatch Update ID", UpdatedBy: "updater-tenant2"}

		_, err := repo.UpdateMeterByIDorSlug(ctxTenant2, createdMeter.ID.String(), updateInput)
		require.Error(t, err, "Should get an error when updating across tenants by ID")
		assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected pgx.ErrNoRows to be wrapped")
	})

	t.Run("Attempt update meter from different tenant (by Slug)", func(t *testing.T) {
		ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "different-tenant-789")
		updateInput := models.UpdateMeterInput{Name: "Tenant Mismatch Update Slug", UpdatedBy: "updater-tenant2-slug"}

		_, err := repo.UpdateMeterByIDorSlug(ctxTenant2, createdMeter.Slug, updateInput)
		require.Error(t, err, "Should get an error when updating across tenants by slug")
		assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected pgx.ErrNoRows to be wrapped")
	})
}

func TestPgErrorHandling(t *testing.T) {
	t.Run("Handle duplicate key error (23505)", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"meter_slug_tenant_slug_key\""}
		mappedErr := postgres.MapError(pgErr, "TestContext.DuplicateKey")

		require.Error(t, mappedErr)
	})

	t.Run("Handle foreign key violation (23503)", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23503", Message: "insert or update on table \"meter\" violates foreign key constraint \"meter_tenant_slug_fkey\""}
		mappedErr := postgres.MapError(pgErr, "TestContext.ForeignKey")

		require.Error(t, mappedErr)
	})

	t.Run("Handle not null violation (23502)", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23502", Message: "null value in column \"name\" violates not-null constraint"}
		mappedErr := postgres.MapError(pgErr, "TestContext.NotNullViolation")
		require.Error(t, mappedErr)
	})

	t.Run("Handle no rows found (pgx.ErrNoRows)", func(t *testing.T) {
		originalErr := pgx.ErrNoRows
		mappedErr := postgres.MapError(originalErr, "TestContext.NoRowsPgx")

		require.Error(t, mappedErr)
	})

	t.Run("Handle no rows found (sql.ErrNoRows)", func(t *testing.T) {
		originalErr := sql.ErrNoRows
		mappedErr := postgres.MapError(originalErr, "TestContext.NoRowsSql")

		require.Error(t, mappedErr)
	})

	t.Run("Handle check constraint violation (23514)", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23514", Message: "new row for relation \"meter\" violates check constraint \"meter_aggregation_check\""}
		mappedErr := postgres.MapError(pgErr, "TestContext.CheckConstraint")

		require.Error(t, mappedErr)
	})

	t.Run("Handle invalid text representation (22P02)", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "22P02", Message: "invalid input syntax for type aggregation_enum: \"INVALID_AGG_TYPE\""}
		mappedErr := postgres.MapError(pgErr, "TestContext.InvalidEnum")

		require.Error(t, mappedErr)
	})

	t.Run("Handle other database errors (e.g., connection error)", func(t *testing.T) {
		originalErr := fmt.Errorf("some underlying network error")
		mappedErr := postgres.MapError(originalErr, "TestContext.Generic")

		require.Error(t, mappedErr)
	})
}

func TestDeleteMeterByIDorSlug_Integration(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer func() {
		cleanupCtx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
		cleanupTestMeters(t, cleanupCtx, db)
	}()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	t.Run("Success delete by ID", func(t *testing.T) {
		input := createTestMeterInput("delete-id")
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)

		err = repo.DeleteMeterByIDorSlug(ctx, meter.ID.String())
		require.NoError(t, err)
	})

	t.Run("Success delete by Slug", func(t *testing.T) {
		input := createTestMeterInput("delete-slug")
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)

		err = repo.DeleteMeterByIDorSlug(ctx, meter.Slug)
		require.NoError(t, err)
	})

	t.Run("Success delete non-existent ID (no error expected)", func(t *testing.T) {
		err := repo.DeleteMeterByIDorSlug(ctx, uuid.New().String())
		assert.NoError(t, err)
	})

	t.Run("Success delete non-existent Slug (no error expected)", func(t *testing.T) {
		err := repo.DeleteMeterByIDorSlug(ctx, "non-existent-slug-"+uuid.NewString())
		assert.NoError(t, err)
	})

	t.Run("Success delete with invalid ID format (falls back to non-existent slug)", func(t *testing.T) {
		err := repo.DeleteMeterByIDorSlug(ctx, "not-a-uuid-at-all")
		assert.NoError(t, err)
	})

	t.Run("Attempt delete meter from different tenant (by ID)", func(t *testing.T) {
		inputTenant1 := createTestMeterInput("tenant1-delete-id")
		meterTenant1, err := repo.CreateMeter(ctx, inputTenant1)
		require.NoError(t, err)
		ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "different-tenant-delete-id")

		err = repo.DeleteMeterByIDorSlug(ctxTenant2, meterTenant1.ID.String())
		assert.NoError(t, err) // Should be no error as 0 rows affected is mapped to nil

		_, err = repo.GetMeterByIDorSlug(ctx, meterTenant1.ID.String()) // Check original
		assert.NoError(t, err, "Meter should still exist under the original tenant")
	})

	t.Run("Attempt delete meter from different tenant (by Slug)", func(t *testing.T) {
		inputTenant1 := createTestMeterInput("tenant1-delete-slug")
		meterTenant1, err := repo.CreateMeter(ctx, inputTenant1)
		require.NoError(t, err)
		ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "different-tenant-delete-slug")

		err = repo.DeleteMeterByIDorSlug(ctxTenant2, meterTenant1.Slug)
		assert.NoError(t, err) // Should be no error as 0 rows affected is mapped to nil

		_, err = repo.GetMeterByIDorSlug(ctx, meterTenant1.Slug) // Check original
		assert.NoError(t, err, "Meter should still exist under the original tenant")
	})
}

func TestGetMeterByIDorSlug_Integration(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer func() {
		cleanupCtx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
		cleanupTestMeters(t, cleanupCtx, db)
	}()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	input := createTestMeterInput("get-target")
	createdMeter, err := repo.CreateMeter(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, createdMeter)

	t.Run("Success get by ID", func(t *testing.T) {
		meter, err := repo.GetMeterByIDorSlug(ctx, createdMeter.ID.String())
		require.NoError(t, err)
		require.NotNil(t, meter)
		assert.Equal(t, createdMeter.ID, meter.ID)
		assert.Equal(t, createdMeter.Name, meter.Name)
	})

	t.Run("Success get by Slug", func(t *testing.T) {
		meter, err := repo.GetMeterByIDorSlug(ctx, createdMeter.Slug)
		require.NoError(t, err)
		require.NotNil(t, meter)
		assert.Equal(t, createdMeter.ID, meter.ID)
		assert.Equal(t, createdMeter.Slug, meter.Slug)
	})

	t.Run("Error meter not found by ID", func(t *testing.T) {
		_, err := repo.GetMeterByIDorSlug(ctx, uuid.New().String())
		require.Error(t, err)
	})

	t.Run("Error meter not found by Slug", func(t *testing.T) {
		_, err := repo.GetMeterByIDorSlug(ctx, "non-existent-slug-"+uuid.NewString())
		require.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("Error get with invalid ID format (falls back to slug)", func(t *testing.T) {
		_, err := repo.GetMeterByIDorSlug(ctx, "this-is-not-a-uuid")
		require.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("Attempt get meter from different tenant (by ID)", func(t *testing.T) {
		ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "different-tenant-get-id")
		_, err := repo.GetMeterByIDorSlug(ctxTenant2, createdMeter.ID.String())
		require.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("Attempt get meter from different tenant (by Slug)", func(t *testing.T) {
		ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "different-tenant-get-slug")
		_, err := repo.GetMeterByIDorSlug(ctxTenant2, createdMeter.Slug)
		require.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestListMeters_Integration(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer func() {
		cleanupCtx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
		cleanupTestMeters(t, cleanupCtx, db)
	}()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	defer func() {
		_, err := db.Exec(context.Background(), "DELETE FROM meter")
		if err != nil {
			t.Logf("Failed to cleanup other tenant meter: %v", err)
		}
	}()

	numMeters := 5
	createdMeterIDs := make(map[uuid.UUID]bool)
	for i := range make([]struct{}, numMeters) {
		input := createTestMeterInput(fmt.Sprintf("list-%d", i))
		meter, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, meter)
		createdMeterIDs[meter.ID] = true
	}

	t.Run("Success list meters with pagination", func(t *testing.T) {
		page := pagination.Pagination{Page: 1, Limit: 3}
		result, err := repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 3, result.Limit)
		assert.Len(t, result.Results, 3)
		for _, meter := range result.Results {
			assert.Equal(t, testTenantSlug, meter.TenantSlug)
			assert.True(t, createdMeterIDs[meter.ID])
		}

		page.Page = 2
		result, err = repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, result.Page)
		assert.Equal(t, numMeters, result.Total-1)
		assert.Len(t, result.Results, numMeters-2)
		for _, meter := range result.Results {
			assert.Equal(t, testTenantSlug, meter.TenantSlug)
		}
	})

	t.Run("Success empty result with high page number", func(t *testing.T) {
		page := pagination.Pagination{Page: 100, Limit: 10}
		result, err := repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, numMeters, result.Total-1)
		assert.Empty(t, result.Results)
	})

	t.Run("Success list with default limit if limit is zero or negative", func(t *testing.T) {
		page := pagination.Pagination{Page: 1, Limit: 0}
		result, err := repo.ListMeters(ctx, page)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, numMeters, result.Total-1)
	})
}

func TestListMetersByEventTypes_Integration(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()
	defer func() {
		cleanupCtx := context.WithValue(context.Background(), constants.TenantSlugKey, testTenantSlug)
		cleanupTestMeters(t, cleanupCtx, db)
	}()

	l := createTestLogger(t)
	repo := NewPostgresMeterStoreRepository(db, l)

	eventType1 := "list.event.type1." + uuid.NewString()[:8]
	eventType2 := "list.event.type2." + uuid.NewString()[:8]
	eventType3 := "list.event.type3." + uuid.NewString()[:8]

	for i := range make([]struct{}, 3) {
		input := createTestMeterInput(fmt.Sprintf("list-et1-%d", i))
		input.EventType = eventType1
		_, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
	}
	for i := range make([]struct{}, 2) {
		input := createTestMeterInput(fmt.Sprintf("list-et2-%d", i))
		input.EventType = eventType2
		_, err := repo.CreateMeter(ctx, input)
		require.NoError(t, err)
	}

	ctxTenant2 := context.WithValue(context.Background(), constants.TenantSlugKey, "other-list-et-tenant")
	otherInput := createTestMeterInput("other-tenant-list-et")
	otherInput.EventType = eventType1
	_, err := repo.CreateMeter(ctxTenant2, otherInput)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec(context.Background(), "DELETE FROM meter WHERE tenant_slug = $1 AND event_type = $2", "other-list-et-tenant", eventType1)
	}()

	t.Run("Success list meters by single event type", func(t *testing.T) {
		result, err := repo.ListMetersByEventTypes(ctx, []string{eventType1})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result, 3)
		for _, meter := range result {
			assert.Equal(t, eventType1, meter.EventType)
			assert.Equal(t, testTenantSlug, meter.TenantSlug)
		}
	})

	t.Run("Success list meters by multiple event types", func(t *testing.T) {
		result, err := repo.ListMetersByEventTypes(ctx, []string{eventType1, eventType2})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result, 5)
	})

	t.Run("Success empty result for non-existent event type", func(t *testing.T) {
		result, err := repo.ListMetersByEventTypes(ctx, []string{eventType3})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Success empty result for empty event type list", func(t *testing.T) {
		result, err := repo.ListMetersByEventTypes(ctx, []string{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Success list includes existing and non-existing types", func(t *testing.T) {
		result, err := repo.ListMetersByEventTypes(ctx, []string{eventType2, eventType3})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result, 2)
		for _, meter := range result {
			assert.Equal(t, eventType2, meter.EventType)
		}
	})
}
