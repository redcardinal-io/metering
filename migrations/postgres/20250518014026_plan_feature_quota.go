package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPlanFeatureQuota, downPlanFeatureQuota)
}

func upPlanFeatureQuota(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		do $$
		 begin
	     -- create plan_feature_quota metered_reset_period enum
	     if not exists (select 1 from pg_type where typname = 'metered_reset_period_enum') then
	  			create type metered_reset_period_enum as enum (
	  			  'day',
	  			  'week',
	  			  'month',
	  			  'year',
	  				'custom'
	  				'rolling'
	  				'never'
	  			);
	  	end if;

	  	-- create table and indexes
	  	create table if not exists plan_feature_quota (
	  			id uuid primary key default uuid_generate_v4(),
	  			plan_feature_id uuid not null references plan_feature(id) on delete cascade,
	        metered_limit_value bigint not null,
	        metered_reset_period metered_reset_period_enum not null,
	  			metered_custom_period_minutes bigint default null,
	  			created_at timestamp with time zone not null default now(),
	  			updated_at timestamp with time zone not null default now(),
	  	);
	  	perform goose_manage_updated_at('plan_feature_quota');
	  	create index if not exists idx_plan_feature_quota_plan_feature_id on plan_feature_quota(plan_feature_id);
	  end;
	  $$;
	`)
	return err
}

func downPlanFeatureQuota(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		drop table if exists plan_feature_quota;
		drop type if exists metered_reset_period_enum;
	`)
	return err
}
