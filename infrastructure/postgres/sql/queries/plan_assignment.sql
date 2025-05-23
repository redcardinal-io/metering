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
$1, $2, $3, $4, sqlc.narg('valid_until'), $5, $6
) returning *;

-- name: TerminateAssignedPlan :exec
-- removes a plan assignment for either an organization or user
delete from plan_assignment
where plan_id = $1
and (
    (organization_id = $2 and $2 is not null) or
    (user_id = $3 and $3 is not null)
);

-- name: UpdateAssignedPlan :one
-- updates the validity period of a plan assignment for either organization or user
update plan_assignment
set valid_until = coalesce(sqlc.narg('valid_until'), valid_until),
    valid_from = coalesce(sqlc.narg('valid_from'), valid_from),
    updated_by = $2
where (plan_id = $1)
and (
    (organization_id = $3 or $3 is null) or
    (user_id = $4 or $4 is null)
)
returning *;
