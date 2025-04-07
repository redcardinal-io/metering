-- +goose Up
-- +goose StatementBegin
create materialized view if not exists rc_events_mv
to rc_events
as
select
    id,
    type,
    source,
    source_metadata,
    organization,
    user,
    toDateTime64(timestamp, 3) as timestamp,
    properties,
    ingested_at,
    validation_errors
from rc_events_queue;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop view if exists rc_events_mv;
-- +goose StatementEnd
