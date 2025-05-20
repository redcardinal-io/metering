-- name: AssignPlanToOrg :one
INSERT INTO plan_assignment (
    plan_id,
    organization_id,
    valid_from,
    valid_until,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: AssignPlanToUser :one
INSERT INTO plan_assignment (
    plan_id,
    user_id,
    valid_from,
    valid_until,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UnAssignPlanToOrg :exec
DELETE FROM plan_assignment
WHERE plan_id = $1
AND organization_id = $2;

-- name: UnAssignPlanToUser :exec
DELETE FROM plan_assignment
WHERE plan_id = $1
AND user_id = $2;

-- name: UpdateOrgsValidFromAndUntil :one
UPDATE plan_assignment
SET valid_until = $5,
    valid_from = $4,
    updated_by = $3
WHERE plan_id = $1
AND organization_id = $2
RETURNING *;

-- name: UpdateUsersValidFromAndUntil :one
UPDATE plan_assignment
SET valid_until = $5,
    valid_from = $4,
    updated_by = $3
WHERE plan_id = $1
AND user_id = $2
RETURNING *;

