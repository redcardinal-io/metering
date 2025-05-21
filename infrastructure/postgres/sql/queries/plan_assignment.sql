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
    (organization_id = $2 and $2 is not null) or
    (user_id = $3 and $3 is not null)
);

-- name: UpdateAssignedPlan :one
-- updates the validity period of a plan assignment for either organization or user
update plan_assignment
set valid_until = coalesce($4, valid_until),
    valid_from = coalesce($3, valid_from),
    updated_by = $5
where (plan_id = $1)
and (
    (organization_id = $2 or $2 is null) or
    (user_id = $6 or $6 is null)
)
returning *;
