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
  create or replace function fn_plan_assignment_insert_trigger()
  returns trigger as $$
  begin
  insert into plan_assignment_history (
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
  ) values (
    'CREATE',
    new.plan_id,
    new.organization_id,
    new.user_id,
    new.valid_from,
    new.valid_until,
    new.created_at,
    new.updated_at,
    new.created_by,
    new.updated_by
  );
  return new;
  end;
  $$ language plpgsql;

  -- Function to handle UPDATE operations
  create or replace function fn_plan_assignment_update_trigger()
  returns trigger as $$
  begin
  insert into plan_assignment_history (
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
  ) values (
    'UPDATE',
    new.plan_id,
    new.organization_id,
    new.user_id,
    new.valid_from,
    new.valid_until,
    new.created_at,
    new.updated_at,
    new.created_by,
    new.updated_by
  );
  return new;
  end;
  $$ LANGUAGE plpgsql;

  -- Function to handle DELETE operations
  create or replace function fn_plan_assignment_delete_trigger()
  returns TRIGGER AS $$
  begin
  insert into plan_assignment_history (
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
    old.plan_id,
    old.organization_id,
    old.user_id,
    old.valid_from,
    old.valid_until,
    old.created_at,
    old.updated_at,
    old.created_by,
    old.updated_by
  );
  return old;
  end;
  $$ language plpgsql;

  -- Create triggers for CREATE, UPDATE, and DELETE operations
  create trigger trg_plan_assignment_insert
  after insert on plan_assignment
  for each row
  execute function fn_plan_assignment_insert_trigger();

  create trigger trg_plan_assignment_update
  after update on plan_assignment
  for each row
  execute function fn_plan_assignment_update_trigger();

  create trigger trg_plan_assignment_delete
  after delete on plan_assignment
  for each row
  execute function fn_plan_assignment_delete_trigger();

  -- create table and indexes
  create table if not exists plan_assignment_history (
    id uuid primary key default uuid_generate_v4(),
    action varchar not null,
    plan_id uuid,
    organization_id varchar default null,
    user_id varchar default null,
    valid_from timestamp with time zone not null,
    valid_until timestamp with time zone default null,
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
	-- drop triggers
  drop trigger if exists trg_plan_assignment_insert on plan_assignment;
  drop trigger if exists trg_plan_assignment_update on plan_assignment;
  drop trigger if exists trg_plan_assignment_delete on plan_assignment;
  drop function if exists fn_plan_assignment_insert_trigger();
  drop function if exists fn_plan_assignment_update_trigger();
  drop function if exists fn_plan_assignment_delete_trigger();

  drop table if exists plan_assignment_history;
  drop type if exists history_action_enum;
  `)
	return err
}
