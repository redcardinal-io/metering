package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	_ "github.com/redcardinal-io/metering/migrations/postgres"
	"github.com/spf13/cobra"
)

var (
	dbString string
)

func init() {
	fmt.Printf("DATABASE_URL from environment: '%s'\n", os.Getenv("DATABASE_URL"))
	migrateCmd.Flags().StringVarP(&dbString, "db", "d", os.Getenv("DATABASE_URL"), "Database connection string (or set DATABASE_URL env var)")

	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dbString == "" {
			return fmt.Errorf("database connection string is required")
		}

		db, err := sql.Open("pgx", dbString)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}

		// Run migrations
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("failed to set dialect: %w", err)
		}

		// Run all migrations up
		if err := goose.Up(db, "migrations/postgres"); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Println("Migrations completed successfully")
		return nil
	},
}
