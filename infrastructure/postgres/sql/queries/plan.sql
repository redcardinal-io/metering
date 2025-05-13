-- name: CreatePlan :one
INSERT INTO plan (
    name,
    description,
    tenant_slug,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPlanByID :one
SELECT * FROM plan
WHERE id = $1
AND tenant_slug = $2;

-- name: ListPlansPaginated :many
SELECT * FROM plan
WHERE tenant_slug = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: DeletePlanByID :exec
DELETE FROM plan
WHERE id = $1
AND tenant_slug = $2;

-- name: CountPlans :one
SELECT count(*) FROM plan
WHERE tenant_slug = $1;

-- name: UpdatePlanByID :one
UPDATE plan
SET name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    updated_by = $4
WHERE id = $2
AND tenant_slug = $3
RETURNING *;


