-- +goose Up
-- +goose StatementBegin
create type aggregation_enum as enum (
    'count',
    'sum',
    'avg',
    'unique_count',
    'min',
    'max'
);

create table "meter" (
    id uuid primary key default uuid_generate_v4 () ,
    slug varchar unique not null,
    event_type varchar unique not null,
    description text,
    value_property varchar,
    group_by text[] not null,
    aggregation aggregation_enum not null,
    created_at timestamp with time zone not null default current_timestamp,
    created_by varchar not null
);

create index idx_meter_slug on meter(slug);
create index idx_meter_event_type on meter(event_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists "meter";
drop type if exists aggregation_enum;
-- +goose StatementEnd
