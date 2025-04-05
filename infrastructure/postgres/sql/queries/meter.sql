-- name: CreateMeter :one
INSERT INTO meter (
    slug,
    event_type,
    description,
    value_property,
    properties,
    aggregation,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetMeterByID :one
SELECT * FROM meter
WHERE id = $1
LIMIT 1;

-- name: GetMeterBySlug :one
SELECT * FROM meter
WHERE slug = $1
LIMIT 1;

-- name: ListMeters :many
SELECT * FROM meter
ORDER BY created_at DESC;

-- name: ListMetersPaginated :many
SELECT * FROM meter
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: ListMetersByEventType :many
SELECT * FROM meter
WHERE event_type = $1
ORDER BY created_at DESC;

-- name: ListMetersByEventTypeAndValueProperty :many
SELECT * FROM meter
WHERE event_type = $1 AND value_property = $2
ORDER BY created_at DESC;

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

-- name: CheckMeterSlugExists :one
SELECT EXISTS (
    SELECT 1 FROM meter
    WHERE slug = $1
) AS exists;

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
