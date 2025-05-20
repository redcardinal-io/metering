package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// init registers the migration functions for creating and dropping the "plan_assignment" table with goose.
func init() {
	goose.AddMigrationContext(upPlanAssignment, downPlanAssignment)
}

// upPlanAssignment creates the "plan_assignment" table with its schema, ensures necessary indexes and triggers exist.
func upPlanAssignment(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		do $$
		begin
			-- create table and indexes
			create table if not exists plan_assignment (
				id uuid primary key default uuid_generate_v4(),
				plan_id uuid not null,
				organization_id uuid default null,
				user_id uuid default null,
        valid_from timestamp with time zone not null,
        valid_until timestamp with time zone default null,
        created_at timestamp with time zone not null default current_timestamp,
        updated_at timestamp with time zone not null default current_timestamp,
        created_by varchar not null,
        updated_by varchar not null,

        unique (plan_id, organization_id, user_id),

        FOREIGN KEY (plan_id) REFERENCES plan(id)
        ON DELETE CASCADE,

        CONSTRAINT only_one_entity CHECK (
          (organization_id IS NULL AND user_id IS NOT NULL)
          OR
          (organization_id IS NOT NULL AND user_id IS NULL)
        )
			);
      perform goose_manage_updated_at('plan_assignment');
			create index if not exists idx_plan_assignment_organization on plan_assignment(organization_id);
			create index if not exists idx_plan_assignment_user on plan_assignment(user_id);
		end;
		$$;
	`)
	return err
}

// downPlanAssignment removes the "plan_assignment" table from the database if it exists.
func downPlanAssignment(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		drop table if exists plan_assignment;
	`)
	return err
}
