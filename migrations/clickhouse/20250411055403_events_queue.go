package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upEventsQueue, downEventsQueue)
}

func upEventsQueue(ctx context.Context, tx *sql.Tx) error {
	// load envs for kafka
	brokerList := os.Getenv("RCMETERING_KAFKA_BROKER_LIST")
	if brokerList == "" {
		return fmt.Errorf("RCMETERING_KAFKA_BROKER_LIST is not set")
	}
	topicList := os.Getenv("RCMETERING_KAFKA_TOPIC_LIST")
	if topicList == "" {
		return fmt.Errorf("RCMETERING_KAFKA_TOPIC_LIST is not set")
	}
	groupName := os.Getenv("RCMETERING_KAFKA_GROUP_NAME")
	if groupName == "" {
		return fmt.Errorf("RCMETERING_KAFKA_GROUP_NAME is not set")
	}

	sql := fmt.Sprintf(`
		create table if not exists rc_events_queue(
    		id String not null,
    		type String not null,
    		source String not null,
    		organization String not null,
    		user String not null,
    		timestamp DateTime not null,
    		properties String not null
		)
		engine = Kafka()
		settings
    		kafka_broker_list = '%s',
    		kafka_topic_list = '%s',
    		kafka_group_name = '%s',
    		kafka_format = 'JSONEachRow';
	`, brokerList, topicList, groupName)

	_, err := tx.ExecContext(ctx, sql)
	return err
}

func downEventsQueue(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `drop table if exists rc_events_queue;`)
	return err
}
