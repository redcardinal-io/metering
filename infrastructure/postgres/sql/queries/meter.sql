-- name: CreateMeter :one
INSERT INTO meter (
    name,
    slug,
    event_type,
    description,
    value_property,
    properties,
    aggregation,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetMeterByID :one
SELECT * FROM meter
WHERE id = $1;

-- name: GetMeterBySlug :one
SELECT * FROM meter
WHERE slug = $1;


-- name: ListMetersCursorPaginated :many
SELECT * FROM meter
WHERE 
    CASE WHEN @use_cursor::boolean THEN 
        (created_at, id) < (@cursor_time, @cursor_id::uuid)
    ELSE 
        TRUE 
    END
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: ListMetersCursorPaginatedByEventType :many
SELECT * FROM meter
WHERE event_type = $1 and 
    CASE WHEN @use_cursor::boolean THEN 
        (created_at, id) < (@cursor_time, @cursor_id::uuid)
    ELSE 
        TRUE 
    END
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: DeleteMeterByID :exec
DELETE FROM meter
WHERE id = $1;

-- name: DeleteMeterBySlug :exec
DELETE FROM meter
WHERE slug = $1;

-- name: CountMeters :one
SELECT count(*) FROM meter;

-- name: CountMetersByEventType :one
SELECT count(*) FROM meter
WHERE event_type = $1;

-- name: GetEventTypes :many
SELECT DISTINCT event_type FROM meter
ORDER BY event_type;

-- name: GetValuePropertiesByEventType :many
SELECT DISTINCT value_property FROM meter
WHERE event_type = $1 AND value_property IS NOT NULL
ORDER BY value_property;

-- name: GetPropertiesByEventType :many
SELECT DISTINCT unnest(properties) as property 
FROM meter
WHERE event_type = $1
ORDER BY property;
