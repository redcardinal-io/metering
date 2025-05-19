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

-- name: UnAssignPlanToOrgByPlanId :exec
DELETE FROM plan_assignment
WHERE plan_id = $1
AND organization_id = $2;

-- name: UnAssignPlanToUserByPlanId :exec
DELETE FROM plan_assignment
WHERE plan_id = $1
AND user_id = $2;

-- name: UpdateOrgsValidFrom :one
UPDATE plan_assignment
SET valid_from = $4,
    updated_by = $3
WHERE plan_id = $1
AND organization_id = $2
RETURNING *;

-- name: UpdateUsersValidFrom :one
UPDATE plan_assignment
SET valid_from = $4,
    updated_by = $3
WHERE plan_id = $1
AND user_id = $2
RETURNING *;

-- name: UpdateOrgsValidUntil :one
UPDATE plan_assignment
SET valid_until = $4,
    updated_by = $3
WHERE plan_id = $1
AND organization_id = $2
RETURNING *;

-- name: UpdateUsersValidUntil :one
UPDATE plan_assignment
SET valid_until = $4,
    updated_by = $3
WHERE plan_id = $1
AND user_id = $2
RETURNING *;


