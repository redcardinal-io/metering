-- name: CreatePlanFeature :one
insert into plan_feature (
    plan_id,
    feature_id,
    config,
    created_by,
    updated_by
) values (
    $1,
    $2,
    $3,
    $4,
    $5 
) returning
    id as plan_feature_id,
    plan_id,
    feature_id,
    config,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: ListPlanFeaturesByPlan :many
select
    pf.id as plan_feature_id,
    pf.plan_id,
    pf.feature_id,
    pf.config,
    pf.created_at,
    pf.updated_at,
    pf.created_by,
    pf.updated_by,
    f.name as feature_name,
    f.slug as feature_slug,
    f.description as feature_description,
    f.type as feature_type,
    f.config as feature_config,
    f.tenant_slug as feature_tenant_slug
from 
    plan_feature pf
join
    feature f on pf.feature_id = f.id
where
    pf.plan_id = $1
    and (sqlc.narg('feature_type') is null or f.type = sqlc.narg('feature_type'))
order by
    pf.created_at desc;


-- name: UpdatePlanFeatureConfigByPlan :one
update
    plan_feature
set
    config = $1,      
    updated_by = $2  
where
    plan_id = $3      
    and feature_id = $4 
returning
    id AS plan_feature_id,
    plan_id,
    feature_id,
    config,
    created_at,
    updated_at,
    created_by,
    updated_by;

-- name: DeletePlanFeature :exec
delete from plan_feature
where plan_id = $1
and feature_id = $2;
