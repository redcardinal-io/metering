package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMeter, downMeter)
}

// upMeter applies the database migration to create the aggregation_enum type and the meter table with associated indexes if they do not already exist.
func upMeter(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		do $$
		begin
			-- check if type exists and create it if it doesn't
			if not exists (select 1 from pg_type where typname = 'aggregation_enum') then
				create type aggregation_enum as enum (
					'count',
					'sum',
					'avg',
					'unique_count',
					'min',
					'max'
				);
			end if;
			
			-- create table and indexes
			create table if not exists meter (
				id uuid primary key default uuid_generate_v4(),
				name varchar not null,
				slug varchar not null,
				event_type varchar not null,
				description text,
				value_property varchar,
				properties text[] not null,
				aggregation aggregation_enum not null,
				tenant_slug varchar not null,
        created_at timestamp with time zone not null default current_timestamp,
        updated_at timestamp with time zone not null default current_timestamp,
        created_by varchar not null,
        updated_by varchar not null,

        unique (tenant_slug, slug)
			);
			
			create index if not exists idx_meter_slug on meter(slug);
			create index if not exists idx_meter_event_type on meter(event_type);
			create index if not exists idx_meter_tenant_slug on meter(tenant_slug);
		end;
		$$;
	`)
	return err
}

func downMeter(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.ExecContext(ctx, `
		drop table if exists meter;
		drop type if exists aggregation_enum;
	`)
	return err
}
