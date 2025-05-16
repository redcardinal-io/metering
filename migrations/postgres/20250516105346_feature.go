package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upFeature, downFeature)
}

func upFeature(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		do $$
		 begin
	     -- create feature type enum 
	     if not exists (select 1 from pg_type where typname = 'feature_enum') then
 					create type feature_enum as enum (
 						'static',
 						'metered'
 		      );
 			 end if;

 		  -- create table and indexes
 		  create table if not exists feature (
				id uuid primary key default uuid_generate_v4(),
				name varchar not null,
				slug varchar not null,
				description varchar not null,
				tenant_slug varchar not null,
			  type feature_enum default 'static',
				config jsonb default null,
				created_at timestamp with time zone default now(),
				updated_at timestamp with time zone default now(),
				created_by varchar not null,
				updated_by varchar not null,
				unique (tenant_slug, slug)

 			);

			perform goose_manage_updated_at('feature');
			create index if not exists idx_feature_on_slug on feature (slug);
			create index if not exists idx_feature_on_tenant_slug on feature (tenant_slug);
 		end;
 		$$;
	`)
	return err
}

func downFeature(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		drop table if exists feature;
		drop type if exists feature_enum;
	`)
	return err
}
