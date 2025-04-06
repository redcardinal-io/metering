CREATE TABLE IF NOT EXISTS rc_events_queue(
    id String NOT NULL,
    type String NOT NULL,
    source String NOT NULL,
    source_metadata Map(String, String) NOT NULL,
    organization String NOT NULL,
    user String NOT NULL,
    timestamp String NOT NULL,
    properties Map(String, String) NOT NULL,
    ingested_at DateTime64(3) NOT NULL,
    validation_errors Map(String, String)
)
ENGINE = Kafka()
SETTINGS
    kafka_broker_list = '${RCMETERING_KAFKA_BOOTSTRAP_SERVERS}',
    kafka_topic_list = '${RCMETERING_KAFKA_RAW_EVENTS_TOPIC}',
    kafka_group_name = '${RCMETERING_KAFKA_CLICKHOUSE_CONSUMER_GROUP}',
    kafka_format = 'JSONEachRow'
