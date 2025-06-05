-- name: CreatePlan :one
INSERT INTO plan (
    name,
    description,
    slug,
    type,
    tenant_slug,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPlanByID :one
SELECT * FROM plan
WHERE id = $1
AND tenant_slug = $2;

-- name: GetPlanBySlug :one
SELECT * FROM plan
WHERE slug = $1
AND tenant_slug = $2;

-- name: ListPlansPaginated :many
SELECT * FROM plan
WHERE tenant_slug = $1
and (sqlc.narg('type')::plan_type_enum is null or type = sqlc.narg('type')::plan_type_enum)
and (sqlc.narg('archived')::boolean is null or archived_at is not null = sqlc.narg('archived')::boolean)
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: DeletePlanByID :exec
DELETE FROM plan
WHERE id = $1
AND tenant_slug = $2;

-- name: DeletePlanBySlug :exec
DELETE FROM plan
WHERE slug = $1
AND tenant_slug = $2;

-- name: ArchivePlanByID :one
UPDATE plan
SET archived_at = CURRENT_TIMESTAMP,
    updated_by = $3
WHERE id = $1
AND tenant_slug = $2
RETURNING *;

-- name: ArchivePlanBySlug :one
UPDATE plan
SET archived_at = CURRENT_TIMESTAMP,
    updated_by = $3
WHERE slug = $1
AND tenant_slug = $2
RETURNING *;

-- name: UnArchivePlanByID :one
UPDATE plan
SET archived_at = null,
    updated_by = $3
WHERE id = $1
AND tenant_slug = $2
RETURNING *;

-- name: UnArchivePlanBySlug :one
UPDATE plan
SET archived_at = null,
    updated_by = $3
WHERE slug = $1
AND tenant_slug = $2
RETURNING *;

-- name: CountPlans :one
SELECT count(*) FROM plan
WHERE tenant_slug = $1
AND (sqlc.narg('type')::plan_type_enum is null or type = sqlc.narg('type')::plan_type_enum);

-- name: UpdatePlanByID :one
UPDATE plan
SET name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    updated_by = $4
WHERE id = $2
AND tenant_slug = $3
RETURNING *;

-- name: UpdatePlanBySlug :one
UPDATE plan
SET name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    updated_by = $3
WHERE slug = $2
AND tenant_slug = $4
RETURNING *;
