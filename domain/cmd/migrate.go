package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/redcardinal-io/metering/domain/pkg/config"
	"github.com/redcardinal-io/metering/domain/pkg/logger"
	_ "github.com/redcardinal-io/metering/migrations/clickhouse"
	_ "github.com/redcardinal-io/metering/migrations/postgres"
)

var (
	pgDbString string
	chDbString string

	pgMigrationsDir = "migrations/postgres"
	chMigrationsDir = "migrations/clickhouse"
	lg              *logger.Logger
)

func init() {
	lg, _ = logger.NewLogger(&config.LoggerConfig{
		Level: "debug",
		Mode:  "dev",
	})
	// Add connection string flags to the migrate parent command
	migrateCmd.Flags().StringVarP(
		&pgDbString,
		"postgres-db-string",
		"p",
		os.Getenv("RCMETERING_POSTGRES_URL"),
		"PostgreSQL database connection string (or set RCMETERING_POSTGRES_URL env var)",
	)

	migrateCmd.Flags().StringVarP(
		&chDbString,
		"clickhouse-db-string",
		"c",
		os.Getenv("RCMETERING_CLICKHOUSE_URL"),
		"ClickHouse database connection string (or set RCMETERING_CLICKHOUSE_URL env var)",
	)

	// Add subcommands to the migrate command
	migrateCmd.AddCommand(migratePgCmd)
	migrateCmd.AddCommand(migrateChCmd)
	migrateCmd.AddCommand(migrateAllCmd)

	// Add down migration subcommands
	migrateCmd.AddCommand(migratePgDownCmd)
	migrateCmd.AddCommand(migrateChDownCmd)
	migrateCmd.AddCommand(migrateAllDownCmd)

	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  "Run database migrations for PostgreSQL, ClickHouse, or both",
}

var migratePgCmd = &cobra.Command{
	Use:   "pg",
	Short: "Run PostgreSQL database migrations up",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPostgresMigrations(false)
	},
}

var migrateChCmd = &cobra.Command{
	Use:   "ch",
	Short: "Run ClickHouse database migrations up",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClickHouseMigrations(false)
	},
}

var migrateAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run both PostgreSQL and ClickHouse migrations up",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPostgresMigrations(false); err != nil {
			return err
		}
		return runClickHouseMigrations(false)
	},
}

var migratePgDownCmd = &cobra.Command{
	Use:   "pg-down",
	Short: "Roll back the most recent PostgreSQL database migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPostgresMigrations(true)
	},
}

var migrateChDownCmd = &cobra.Command{
	Use:   "ch-down",
	Short: "Roll back the most recent ClickHouse database migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClickHouseMigrations(true)
	},
}

var migrateAllDownCmd = &cobra.Command{
	Use:   "all-down",
	Short: "Roll back the most recent PostgreSQL and ClickHouse migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPostgresMigrations(true); err != nil {
			return err
		}
		return runClickHouseMigrations(true)
	},
}

func runPostgresMigrations(down bool) error {
	if pgDbString == "" {
		return fmt.Errorf("PostgreSQL database connection string is required")
	}

	action := "up"
	if down {
		action = "down"
	}

	lg.Info(fmt.Sprintf("Running PostgreSQL migrations (%s)...", action))
	db, err := sql.Open("pgx", pgDbString)
	if err != nil {
		lg.Error("failed to connect to PostgreSQL database: %v", zap.Error(err))
		return fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		lg.Error("failed to ping PostgreSQL database: %v", zap.Error(err))
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		lg.Error("failed to set PostgreSQL dialect: %v", zap.Error(err))
		return fmt.Errorf("failed to set PostgreSQL dialect: %w", err)
	}

	var migrateErr error
	if down {
		migrateErr = goose.Down(db, pgMigrationsDir)
	} else {
		migrateErr = goose.Up(db, pgMigrationsDir)
	}

	if migrateErr != nil {
		lg.Error(fmt.Sprintf("failed to run PostgreSQL migrations (%s): %%v", action), zap.Error(migrateErr))
		return fmt.Errorf("failed to run PostgreSQL migrations (%s): %%w", action, migrateErr)
	}

	lg.Info(fmt.Sprintf("PostgreSQL migrations (%s) completed successfully", action))
	return nil
}

func runClickHouseMigrations(down bool) error {
	if chDbString == "" {
		return fmt.Errorf("ClickHouse database connection string is required")
	}

	action := "up"
	if down {
		action = "down"
	}

	lg.Info(fmt.Sprintf("Running ClickHouse migrations (%s)...", action))
	db, err := sql.Open("clickhouse", chDbString)
	if err != nil {
		lg.Error("failed to connect to ClickHouse database: %v", zap.Error(err))
		return fmt.Errorf("failed to connect to ClickHouse database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		lg.Error("failed to ping ClickHouse database: %v", zap.Error(err))
		return fmt.Errorf("failed to ping ClickHouse database: %w", err)
	}

	if err := goose.SetDialect("clickhouse"); err != nil {
		lg.Error("failed to set ClickHouse dialect: %v", zap.Error(err))
		return fmt.Errorf("failed to set ClickHouse dialect: %w", err)
	}

	var migrateErr error
	if down {
		migrateErr = goose.Down(db, chMigrationsDir)
	} else {
		migrateErr = goose.Up(db, chMigrationsDir)
	}

	if migrateErr != nil {
		lg.Error(fmt.Sprintf("failed to run ClickHouse migrations (%s): %%v", action), zap.Error(migrateErr))
		return fmt.Errorf("failed to run ClickHouse migrations (%s): %%w", action, migrateErr)
	}

	lg.Info(fmt.Sprintf("ClickHouse migrations (%s) completed successfully", action))
	return nil
}
