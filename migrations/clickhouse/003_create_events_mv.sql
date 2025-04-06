CREATE MATERIALIZED VIEW IF NOT EXISTS rc_events_mv
TO rc_events
AS
SELECT
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
FROM rc_events_queue;
