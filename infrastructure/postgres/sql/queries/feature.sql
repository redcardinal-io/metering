-- name: CreateFeature :one
insert into feature (
  name,
  slug,
  description,
  tenant_slug,
  type,
  config,
  created_by,
  updated_by
) values (
  $1, $2, $3, $4, $5, $6, $7, $8
) returning *;

-- name: GetFeatureByID :one
select * from feature
where id = $1
and tenant_slug = $2;

-- name: GetFeatureBySlug :one
select * from feature
where slug = $1
and tenant_slug = $2;

-- name: ListFeaturesPaginated :many
select * from feature
where tenant_slug = $1
and (sqlc.narg('type')::feature_enum is null or type = sqlc.narg('type')::feature_enum)
order by created_at desc
limit $2
offset $3;

-- name: DeleteFeatureByID :exec
delete from feature
where id = $1
and tenant_slug = $2;

-- name: DeleteFeatureBySlug :exec
delete from feature
where slug = $1
and tenant_slug = $2;

-- name: CountFeatures :one
select count(*) from feature
where tenant_slug = $1
and (sqlc.narg('type')::feature_enum is null or type = sqlc.narg('type')::feature_enum);

-- name: UpdateFeatureByID :one
update feature
set name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    config = coalesce($2, config),
    updated_by = $3
where id = $4
and tenant_slug = $5
returning *;

-- name: UpdateFeatureBySlug :one
update feature
set name = coalesce(sqlc.narg('name'), name),
    description = coalesce($1, description),
    config = coalesce($2, config),
    updated_by = $3
where slug = $4
and tenant_slug = $5
returning *;
