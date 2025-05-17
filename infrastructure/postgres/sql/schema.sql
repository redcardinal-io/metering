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

create type plan_type_enum as enum (
	'standard',
	'custom'
);

create table if not exists plan (
	id uuid primary key default uuid_generate_v4(),
	name varchar not null,
	slug varchar not null,
	description text,
	type plan_type_enum not null,
	tenant_slug varchar not null,
	created_at timestamp with time zone not null default current_timestamp,
	updated_at timestamp with time zone not null default current_timestamp,
	archived_at timestamp with time zone default null,
	created_by varchar not null,
	updated_by varchar not null,

	unique (tenant_slug, slug)
);

