create type aggregation_enum as enum (
    'count',
    'sum',
    'avg',
    'unique_count',
    'min',
    'max'
);

create table "meter" (
  id uuid primary key default uuid_generate_v4(),
	name varchar not null,
	slug varchar not null,
	event_type varchar not null,
	description text,
	value_property varchar,
	properties text[] not null,
	aggregation aggregation_enum not null,
	tenant_slug varchar not null,
	created_at timestamp with time zone not null default current_timestamp,
	updated_at timestamp with time zone not null default current_timestamp,
	created_by varchar not null,
	updated_by varchar not null,

  unique (tenant_slug, slug)
);
