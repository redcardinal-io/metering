package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// init registers the up and down migration functions for the "plan" table with goose.
func init() {
	goose.AddMigrationContext(upPlan, downPlan)
}

// upPlan creates the "plan" table with its schema, manages the "updated_at" column, and adds an index on "tenant_slug" if they do not already exist.
func upPlan(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		do $$
		begin
			-- create table and indexes
			create table if not exists plan (
				id uuid primary key default uuid_generate_v4(),
				name varchar not null,
				description text,
				tenant_slug varchar not null,
        created_at timestamp with time zone not null default current_timestamp,
        updated_at timestamp with time zone not null default current_timestamp,
        created_by varchar not null,
        updated_by varchar not null
			);
		  
      perform goose_manage_updated_at('plan');
			create index if not exists idx_plan_tenant_slug on plan(tenant_slug);
		end;
		$$;
	`)
	return err
}

// downPlan drops the "plan" table from the database if it exists.
func downPlan(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		drop table if exists plan;
	`)
	return err
}
