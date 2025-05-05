-- name: CreateMeter :one
INSERT INTO meter (
    name,
    slug,
    event_type,
    description,
    value_property,
    properties,
    aggregation,
    tenant_slug,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetMeterByID :one
SELECT * FROM meter
WHERE id = $1
AND tenant_slug = $2;

-- name: GetMeterBySlug :one
SELECT * FROM meter
WHERE slug = $1
AND tenant_slug = $2;

-- name: ListMetersPaginated :many
SELECT * FROM meter
WHERE tenant_slug = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: ListMetersByEventTypes :many
SELECT * FROM meter
WHERE event_type = ANY($1::text[])
AND tenant_slug = $2;


-- name: DeleteMeterByID :exec
DELETE FROM meter
WHERE id = $1 
AND tenant_slug = $2;

-- name: DeleteMeterBySlug :exec
DELETE FROM meter
WHERE slug = $1
AND tenant_slug = $2;

-- name: CountMeters :one
SELECT count(*) FROM meter 
WHERE tenant_slug = $1;

-- name: CountMetersByEventType :one
SELECT count(*) FROM meter
WHERE event_type = $1
AND tenant_slug = $2;

-- name: GetValuePropertiesByEventType :many
SELECT DISTINCT value_property FROM meter
WHERE event_type = $1 AND value_property IS NOT NULL
AND tenant_slug = $2
ORDER BY value_property;

-- name: GetPropertiesByEventType :many
SELECT DISTINCT unnest(properties) as property 
FROM meter
WHERE event_type = $1
AND tenant_slug = $2
ORDER BY property;

-- name: UpdateMeterByID :one
UPDATE meter
SET name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    updated_by = $4
WHERE id = $2
AND tenant_slug = $3
RETURNING *;

-- name: UpdateMeterBySlug :one
UPDATE meter
SET name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    updated_by = $3
WHERE slug = $2
AND tenant_slug = $4
RETURNING *;
