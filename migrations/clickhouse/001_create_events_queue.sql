-- +goose Up
-- +goose StatementBegin
create table if not exists rc_events_queue(
    id String not null,
    type String not null,
    source String not null,
    source_metadata Map(String, String) not null,
    organization String not null,
    user String not null,
    timestamp String not null,
    properties Map(String, String) not null,
    ingested_at DateTime64(3) not null,
    validation_errors Map(String, String)
)
engine = Kafka()
settings
    kafka_broker_list = 'redpanda:9092',
    kafka_topic_list = 'rc_events_queue',
    kafka_group_name = 'redcardinal_clickhouse_consumer',
    kafka_format = 'JSONEachRow';
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists rc_events_queue;
-- +goose StatementEnd
