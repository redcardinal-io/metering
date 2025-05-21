package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// init registers the migration functions for creating and dropping the "plan_assignment_history" table with goose.
func init() {
	goose.AddMigrationContext(upPlanAssignmentHistory, downPlanAssignmentHistory)
}

// upPlanAssignmentHistory creates the "plan_assignment_history" table with its schema, ensures necessary indexes and triggers exist.
func upPlanAssignmentHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
  -- Function to handle INSERT operations
  CREATE OR REPLACE FUNCTION fn_plan_assignment_insert_trigger()
  RETURNS TRIGGER AS $$
  BEGIN
  INSERT INTO plan_assignment_history (
    action,
    plan_id,
    organization_id,
    user_id,
    valid_from,
    valid_until,
    created_at,
    updated_at,
    created_by,
    updated_by
  ) VALUES (
    'INSERT',
    NEW.plan_id,
    NEW.organization_id,
    NEW.user_id,
    NEW.valid_from,
    NEW.valid_until,
    NEW.created_at,
    NEW.updated_at,
    NEW.created_by,
    NEW.updated_by
  );
  RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  -- Function to handle UPDATE operations
  CREATE OR REPLACE FUNCTION fn_plan_assignment_update_trigger()
  RETURNS TRIGGER AS $$
  BEGIN
  INSERT INTO plan_assignment_history (
    action,
    plan_id,
    organization_id,
    user_id,
    valid_from,
    valid_until,
    created_at,
    updated_at,
    created_by,
    updated_by
  ) VALUES (
    'UPDATE',
    NEW.plan_id,
    NEW.organization_id,
    NEW.user_id,
    NEW.valid_from,
    NEW.valid_until,
    NEW.created_at,
    NEW.updated_at,
    NEW.created_by,
    NEW.updated_by
  );
  RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  -- Function to handle DELETE operations
  CREATE OR REPLACE FUNCTION fn_plan_assignment_delete_trigger()
  RETURNS TRIGGER AS $$
  BEGIN
  INSERT INTO plan_assignment_history (
    action,
    plan_id,
    organization_id,
    user_id,
    valid_from,
    valid_until,
    created_at,
    updated_at,
    created_by,
    updated_by
  ) VALUES (
    'DELETE',
    OLD.plan_id,
    OLD.organization_id,
    OLD.user_id,
    OLD.valid_from,
    OLD.valid_until,
    OLD.created_at,
    OLD.updated_at,
    OLD.created_by,
    OLD.updated_by
  );
  RETURN OLD;
  END;
  $$ LANGUAGE plpgsql;

  -- Create triggers for INSERT, UPDATE, and DELETE operations
  CREATE TRIGGER trg_plan_assignment_insert
  AFTER INSERT ON plan_assignment
  FOR EACH ROW
  EXECUTE FUNCTION fn_plan_assignment_insert_trigger();

  CREATE TRIGGER trg_plan_assignment_update
  AFTER UPDATE ON plan_assignment
  FOR EACH ROW
  EXECUTE FUNCTION fn_plan_assignment_update_trigger();

  CREATE TRIGGER trg_plan_assignment_delete
  AFTER DELETE ON plan_assignment
  FOR EACH ROW
  EXECUTE FUNCTION fn_plan_assignment_delete_trigger();

  -- check if history_action exists and create it if it doesn't
  do $$
  begin
  if not exists (select 1 from pg_type where typname = 'history_action_enum') then
  create type history_action_enum as enum (
    'INSERT',
    'UPDATE',
    'DELETE'
  );
  end if;
  end
  $$;

  -- create table and indexes
  create table if not exists plan_assignment_history (
    id uuid primary key default uuid_generate_v4(),
    action history_action_enum not null,
    plan_id uuid,
    organization_id string default null,
    user_id string default null,
    valid_from timestamp with time zone not null,
    valid_until timestamp with time zone not null,
    created_at timestamp with time zone not null default current_timestamp,
    updated_at timestamp with time zone not null default current_timestamp,
    created_by varchar not null,
    updated_by varchar not null
  );
  create index if not exists idx_plan_assignment_history_organization on plan_assignment_history(organization_id, action);
  create index if not exists idx_plan_assignment_history_user on plan_assignment_history(user_id, action);
  `)
	return err
}

// downPlanAssignment removes the "plan_assignment" table from the database if it exists.
func downPlanAssignmentHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
  drop table if exists plan_assignment_history;
  `)
	return err
}
