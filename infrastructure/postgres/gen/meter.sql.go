// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: meter.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countMeters = `-- name: CountMeters :one
SELECT count(*) FROM meter
`

func (q *Queries) CountMeters(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countMeters)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countMetersByEventType = `-- name: CountMetersByEventType :one
SELECT count(*) FROM meter
WHERE event_type = $1
`

func (q *Queries) CountMetersByEventType(ctx context.Context, eventType pgtype.Text) (int64, error) {
	row := q.db.QueryRow(ctx, countMetersByEventType, eventType)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createMeter = `-- name: CreateMeter :one
INSERT INTO meter (
    name,
    slug,
    event_type,
    description,
    value_property,
    properties,
    aggregation,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, name, slug, event_type, description, value_property, properties, aggregation, created_at, created_by
`

type CreateMeterParams struct {
	Name          string
	Slug          string
	EventType     pgtype.Text
	Description   pgtype.Text
	ValueProperty pgtype.Text
	Properties    []string
	Aggregation   AggregationEnum
	CreatedBy     string
}

func (q *Queries) CreateMeter(ctx context.Context, arg CreateMeterParams) (Meter, error) {
	row := q.db.QueryRow(ctx, createMeter,
		arg.Name,
		arg.Slug,
		arg.EventType,
		arg.Description,
		arg.ValueProperty,
		arg.Properties,
		arg.Aggregation,
		arg.CreatedBy,
	)
	var i Meter
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
		&i.EventType,
		&i.Description,
		&i.ValueProperty,
		&i.Properties,
		&i.Aggregation,
		&i.CreatedAt,
		&i.CreatedBy,
	)
	return i, err
}

const deleteMeterByID = `-- name: DeleteMeterByID :exec
DELETE FROM meter
WHERE id = $1
`

func (q *Queries) DeleteMeterByID(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteMeterByID, id)
	return err
}

const deleteMeterBySlug = `-- name: DeleteMeterBySlug :exec
DELETE FROM meter
WHERE slug = $1
`

func (q *Queries) DeleteMeterBySlug(ctx context.Context, slug string) error {
	_, err := q.db.Exec(ctx, deleteMeterBySlug, slug)
	return err
}

const getEventTypes = `-- name: GetEventTypes :many
SELECT DISTINCT event_type FROM meter
ORDER BY event_type
`

func (q *Queries) GetEventTypes(ctx context.Context) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, getEventTypes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var event_type pgtype.Text
		if err := rows.Scan(&event_type); err != nil {
			return nil, err
		}
		items = append(items, event_type)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMeterByID = `-- name: GetMeterByID :one
SELECT id, name, slug, event_type, description, value_property, properties, aggregation, created_at, created_by FROM meter
WHERE id = $1
`

func (q *Queries) GetMeterByID(ctx context.Context, id pgtype.UUID) (Meter, error) {
	row := q.db.QueryRow(ctx, getMeterByID, id)
	var i Meter
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
		&i.EventType,
		&i.Description,
		&i.ValueProperty,
		&i.Properties,
		&i.Aggregation,
		&i.CreatedAt,
		&i.CreatedBy,
	)
	return i, err
}

const getMeterBySlug = `-- name: GetMeterBySlug :one
SELECT id, name, slug, event_type, description, value_property, properties, aggregation, created_at, created_by FROM meter
WHERE slug = $1
`

func (q *Queries) GetMeterBySlug(ctx context.Context, slug string) (Meter, error) {
	row := q.db.QueryRow(ctx, getMeterBySlug, slug)
	var i Meter
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
		&i.EventType,
		&i.Description,
		&i.ValueProperty,
		&i.Properties,
		&i.Aggregation,
		&i.CreatedAt,
		&i.CreatedBy,
	)
	return i, err
}

const getPropertiesByEventType = `-- name: GetPropertiesByEventType :many
SELECT DISTINCT unnest(properties) as property 
FROM meter
WHERE event_type = $1
ORDER BY property
`

func (q *Queries) GetPropertiesByEventType(ctx context.Context, eventType pgtype.Text) ([]interface{}, error) {
	rows, err := q.db.Query(ctx, getPropertiesByEventType, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []interface{}
	for rows.Next() {
		var property interface{}
		if err := rows.Scan(&property); err != nil {
			return nil, err
		}
		items = append(items, property)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getValuePropertiesByEventType = `-- name: GetValuePropertiesByEventType :many
SELECT DISTINCT value_property FROM meter
WHERE event_type = $1 AND value_property IS NOT NULL
ORDER BY value_property
`

func (q *Queries) GetValuePropertiesByEventType(ctx context.Context, eventType pgtype.Text) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, getValuePropertiesByEventType, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var value_property pgtype.Text
		if err := rows.Scan(&value_property); err != nil {
			return nil, err
		}
		items = append(items, value_property)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listMetersByEventType = `-- name: ListMetersByEventType :many
SELECT id, name, slug, event_type, description, value_property, properties, aggregation, created_at, created_by FROM meter
WHERE event_type = $1
`

func (q *Queries) ListMetersByEventType(ctx context.Context, eventType pgtype.Text) ([]Meter, error) {
	rows, err := q.db.Query(ctx, listMetersByEventType, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Meter
	for rows.Next() {
		var i Meter
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.EventType,
			&i.Description,
			&i.ValueProperty,
			&i.Properties,
			&i.Aggregation,
			&i.CreatedAt,
			&i.CreatedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listMetersPaginated = `-- name: ListMetersPaginated :many
SELECT id, name, slug, event_type, description, value_property, properties, aggregation, created_at, created_by FROM meter
ORDER BY created_at DESC
LIMIT $1
OFFSET $2
`

type ListMetersPaginatedParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListMetersPaginated(ctx context.Context, arg ListMetersPaginatedParams) ([]Meter, error) {
	rows, err := q.db.Query(ctx, listMetersPaginated, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Meter
	for rows.Next() {
		var i Meter
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.EventType,
			&i.Description,
			&i.ValueProperty,
			&i.Properties,
			&i.Aggregation,
			&i.CreatedAt,
			&i.CreatedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
