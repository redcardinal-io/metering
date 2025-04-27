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
    name varchar not null,
    slug varchar unique not null,
    event_type varchar,
    description text,
    value_property varchar,
    properties text[] not null,
    aggregation aggregation_enum not null,
    created_at timestamp with time zone not null default current_timestamp,
    tenant_slug varchar not null
);
