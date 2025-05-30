-- name: CreatePlanFeatureQuota :one
insert into plan_feature_quota (
    plan_feature_id,
    limit_value,
    reset_period,
    custom_period_minutes,
    action_at_limit
) values (
    $1, $2, $3, $4, $5
) returning *;

-- name: GetPlanFeatureQuotaByPlanFeatureID :one
select * from plan_feature_quota
where plan_feature_id = $1;

-- name: UpdatePlanFeatureQuota :one
update plan_feature_quota
set
    limit_value = $2,
    reset_period = $3,
    custom_period_minutes = $4,
    action_at_limit = $5,
    updated_at = now()
where plan_feature_id = $1
returning *;

-- name: DeletePlanFeatureQuota :exec
delete from plan_feature_quota
where plan_feature_id = $1;

-- name: CheckMeteredFeature :one
select exists (
    select 1
    from feature f
    join plan_feature pf on f.id = pf.feature_id
    where pf.id = $1
    and f.type = 'metered'
);

