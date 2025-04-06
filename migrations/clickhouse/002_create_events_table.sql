CREATE TABLE IF NOT EXISTS rc_events(
    id String NOT NULL,
    type String NOT NULL,
    source String NOT NULL,
    source_metadata Map(String, String) NOT NULL,
    organization String NOT NULL,
    user String NOT NULL,
    timestamp DateTime64(3) NOT NULL,
    properties Map(String, String) NOT NULL,
    ingested_at DateTime64(3) NOT NULL,
    validation_errors Map(String, String)
)
ENGINE = MergeTree
ORDER BY (timestamp);
