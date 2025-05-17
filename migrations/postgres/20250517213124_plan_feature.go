package postgres

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPlanFeature, downPlanFeature)
}

func upPlanFeature(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		create table if not exists plan_feature (
			id uuid primary key default uuid_generate_v4(),
			plan_id uuid not null,
			feature_id uuid not null,
			created_at timestamp with time zone not null default current_timestamp,
			updated_at timestamp with time zone not null default current_timestamp,
		  created_by varchar not null,
		  updated_by varchar not null,
		  config jsonb default null
		);

		perform goose_manage_updated_at('plan_feature');
		create index if not exists idx_plan_feature_plan_id on plan_feature(plan_id);
		create index if not exists idx_plan_feature_feature_id on plan_feature(feature_id);
	`)
	return err
}

func downPlanFeature(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		drop table if exists plan_feature;
	`)
	return err
}
