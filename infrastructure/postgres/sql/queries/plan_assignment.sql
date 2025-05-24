-- name: AssignPlan :one
-- assigns a plan to either an organization or a user based on which id is provided
insert into plan_assignment (
    plan_id,
    organization_id,
    user_id,
    valid_from,
    valid_until,
    created_by,
    updated_by
) values (
$1, $2, $3, $4, $5, $6, $7
) returning *;

-- name: TerminateAssignedPlan :exec
-- removes a plan assignment for either an organization or user
delete from plan_assignment
where plan_id = $1
and (
    (organization_id = $2 or $2 is null) and 
    (user_id = $3 or $3 is null)
);

-- name: UpdateAssignedPlan :one
-- updates the validity period of a plan assignment for either organization or user
update plan_assignment
set valid_until = coalesce(sqlc.narg('valid_until'), valid_until),
    valid_from = coalesce(sqlc.narg('valid_from'), valid_from),
    updated_by = $2
where (plan_id = $1)
and (
    (organization_id = $3 or $3 is null) and
    (user_id = $4 or $4 is null)
)
returning *;

-- name: ListAssignmentsPaginated :many
SELECT *
FROM plan_assignment
WHERE (
    (organization_id = $1 or $1 is null) and
    (user_id = $2 or $2 is null)
)
AND (plan_id = $7 or $7 is null)
AND (valid_from >= $5 or $5 is null)
AND (valid_until <= $6 or $6 is null)
ORDER BY created_at DESC
LIMIT $3
OFFSET $4;

-- name: CountAssignments :one
SELECT count(*)
FROM plan_assignment
WHERE (
    (organization_id = $1 or $1 is null) and 
    (user_id = $2 or $2 is null)
)
AND (plan_id = $3 or $3 is null)
AND (valid_from >= $4 or $4 is null)
AND (valid_until <= $5 or $5 is null);

-- name: ListAllAssignmentsPaginated :many
SELECT
    pa.id,
    pa.plan_id,
    pa.organization_id,
    pa.user_id,
    pa.valid_from,
    pa.valid_until,
    pa.created_at,
    pa.updated_at,
    pa.created_by,
    pa.updated_by
FROM plan_assignment pa
INNER JOIN plan p ON pa.plan_id = p.id
WHERE p.tenant_slug = $1
AND p.archived_at IS NULL
ORDER BY pa.created_at DESC
LIMIT $2
OFFSET $3;

-- name: CountAllAssignments :one
SELECT count(pa.*)
FROM plan_assignment pa
INNER JOIN plan p ON pa.plan_id = p.id
WHERE p.tenant_slug = $1
AND p.archived_at IS NULL;

-- name: ListAssignmentsHistoryPaginated :many
SELECT *
FROM plan_assignment_history
WHERE (
    (organization_id = $1 or $1 is null) and
    (user_id = $2 or $2 is null)
)
AND (plan_id = $3 or $3 is null)
AND (valid_from < $4 or $4 is null)
AND (valid_from >= $5 or $5 is null)
AND (valid_until < $6 or $6 is null)
AND (valid_until >= $7 or $7 is null)
AND (action = $10)
ORDER BY created_at DESC
LIMIT $8
OFFSET $9;

-- name: CountAssignmentsHistory :one
SELECT count(*)
FROM plan_assignment_history
WHERE (
    (organization_id = $1 or $1 is null) and
    (user_id = $2 or $2 is null)
)
AND (plan_id = $3 or $3 is null)
AND (valid_from < $4 or $4 is null)
AND (valid_from >= $5 or $5 is null)
AND (valid_until < $6 or $6 is null)
AND (valid_until >= $7 or $7 is null)
AND (action = $8);
