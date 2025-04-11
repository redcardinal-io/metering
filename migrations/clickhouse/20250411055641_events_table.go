package clickhouse

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upEventsTable, downEventsTable)
}

func upEventsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.ExecContext(ctx, `
		create table if not exists rc_events(
			id String not null,
			type String not null,
			source String not null,
			source_metadata Map(String, String) not null,
			organization String not null,
			user String not null,
			timestamp DateTime not null,
			properties Map(String, String) not null,
			ingested_at DateTime default now(),
			validation_errors Map(String, String)
		)
		engine = MergeTree
		order by (timestamp);
	`)
	return err
}

func downEventsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.ExecContext(ctx, `drop table if exists rc_events;`)
	return err
}
