package clickhouse

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upEventsMv, downEventsMv)
}

func upEventsMv(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		create materialized view if not exists rc_events_mv
		to rc_events
		as
		select
    		id,
    		type,
    		source,
    		organization,
    		user,
        parseDateTimeBestEffort(replaceOne(timestamp, 'Z', '')) AS timestamp,
    		properties,
    		now() as ingested_at
		from rc_events_queue;
	`)
	return err
}

func downEventsMv(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `drop view if exists rc_events_mv;`)
	return err
}
