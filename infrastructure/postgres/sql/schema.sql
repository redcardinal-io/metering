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

create type feature_enum as enum (
	'static',
	'metered'
);

create table if not exists feature (
	id uuid primary key default uuid_generate_v4(),
	name varchar not null,
	slug varchar not null,
	description varchar default null,
	tenant_slug varchar not null,
  type feature_enum not null default 'static',
	config jsonb default null,
	created_at timestamp with time zone default now(),
	updated_at timestamp with time zone default now(),
	created_by varchar not null,
	updated_by varchar not null,
	unique (tenant_slug, slug)
);

create table if not exists plan_assignment (
	id uuid primary key default uuid_generate_v4(),
	plan_id uuid not null,
	organization_id string default null,
	user_id string default null,
	valid_from timestamp with time zone not null,
	valid_until timestamp with time zone default null,
	created_at timestamp with time zone not null default current_timestamp,
	updated_at timestamp with time zone not null default current_timestamp,
	created_by varchar not null,
	updated_by varchar not null,

	FOREIGN KEY (plan_id) REFERENCES plan(id)
	ON DELETE CASCADE,

	CONSTRAINT only_one_entity CHECK (
	(organization_id IS NULL AND user_id IS NOT NULL)
	OR
	(organization_id IS NOT NULL AND user_id IS NULL)
	)
);

create table if not exists plan_assignment_history (
    id uuid primary key default uuid_generate_v4(),
    plan_assignment_id uuid,
    action varchar not null,
    plan_id uuid,
    organization_id uuid default null,
    user_id uuid default null,
    valid_from timestamp with time zone not null,
    valid_until timestamp with time zone not null,
    created_at timestamp with time zone not null default current_timestamp,
    updated_at timestamp with time zone not null default current_timestamp,
    created_by varchar not null,
    updated_by varchar not null
);
