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

	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  "Run database migrations for PostgreSQL, ClickHouse, or both",
}

var migratePgCmd = &cobra.Command{
	Use:   "pg",
	Short: "Run PostgreSQL database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPostgresMigrations()
	},
}

var migrateChCmd = &cobra.Command{
	Use:   "ch",
	Short: "Run ClickHouse database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClickHouseMigrations()
	},
}

var migrateAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run both PostgreSQL and ClickHouse migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPostgresMigrations(); err != nil {
			return err
		}
		return runClickHouseMigrations()
	},
}

func runPostgresMigrations() error {
	if pgDbString == "" {
		return fmt.Errorf("PostgreSQL database connection string is required")
	}

	lg.Info("Running PostgreSQL migrations...")
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

	if err := goose.Up(db, pgMigrationsDir); err != nil {
		lg.Error("failed to run PostgreSQL migrations: %v", zap.Error(err))
		return fmt.Errorf("failed to run PostgreSQL migrations: %w", err)
	}

	lg.Info("PostgreSQL migrations completed successfully")
	return nil
}

func runClickHouseMigrations() error {
	if chDbString == "" {
		return fmt.Errorf("ClickHouse database connection string is required")
	}

	lg.Info("Running ClickHouse migrations...")
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

	if err := goose.Up(db, chMigrationsDir); err != nil {
		lg.Error("failed to run ClickHouse migrations: %v", zap.Error(err))
		return fmt.Errorf("failed to run ClickHouse migrations: %w", err)
	}

	lg.Info("ClickHouse migrations completed successfully")
	return nil
}
