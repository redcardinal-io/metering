-- +goose Up
-- +goose StatementBegin
create table if not exists rc_events(
    id String not null,
    type String not null,
    source String not null,
    source_metadata Map(String, String) not null,
    organization String not null,
    user String not null,
    timestamp DateTime64(3) not null,
    properties Map(String, String) not null,
    ingested_at DateTime64(3) not null,
    validation_errors Map(String, String)
)
ENGINE = MergeTree
ORDER BY (timestamp);
-- +goose StatementEnd
-- +goose Down

-- +goose StatementBegin
drop table if exists rc_events;
-- +goose StatementEnd
