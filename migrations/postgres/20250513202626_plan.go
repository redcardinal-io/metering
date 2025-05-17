package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// init registers the migration functions for creating and dropping the "plan" table with goose.
func init() {
	goose.AddMigrationContext(upPlan, downPlan)
}

// upPlan creates the "plan" table with its schema, including a custom enum type, and ensures necessary indexes and triggers exist.
func upPlan(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		do $$
		begin
      -- check if type exists and create it if it doesn't
			if not exists (select 1 from pg_type where typname = 'plan_type_enum') then
				create type plan_type_enum as enum (
					'standard',
					'custom'
				);
			end if;

			-- create table and indexes
			create table if not exists plan (
				id uuid primary key default uuid_generate_v4(),
				name varchar not null,
				slug varchar not null,
				description text,
        type plan_type_enum not null,
				tenant_slug varchar not null,
        created_at timestamp with time zone not null default current_timestamp,
        updated_at timestamp with time zone not null default current_timestamp,
        archived_at timestamp with time zone default null,
        created_by varchar not null,
        updated_by varchar not null,

        unique (tenant_slug, slug)
			);
		  
      perform goose_manage_updated_at('plan');
			create index if not exists idx_plan_tenant_slug on plan(tenant_slug);
		end;
		$$;
	`)
	return err
}

// downPlan removes the "plan" table from the database if it exists.
func downPlan(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		drop table if exists plan;
	`)
	return err
}
