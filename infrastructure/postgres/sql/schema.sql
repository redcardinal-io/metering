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

create table if not exists plan_feature (
	id uuid primary key default uuid_generate_v4(),
	plan_id uuid not null references plan(id) on delete cascade,
	feature_id uuid not null references feature(id) on delete cascade,
	created_at timestamp with time zone not null default now(),
	updated_at timestamp with time zone not null default now(),
  created_by varchar not null,
  updated_by varchar not null,
  config jsonb default null
);

create type metered_reset_period_enum as enum (
  'day',
  'week',
  'month',
  'year',
	'custom',
	'rolling',
	'never'
);

create type metered_action_at_limit_enum as enum (
  'none',
  'block',
  'throttle'
);

create table if not exists plan_feature_quota (
  id uuid primary key default uuid_generate_v4(),
  plan_feature_id uuid not null references plan_feature(id) on delete cascade,
  limit_value bigint not null,
  reset_period metered_reset_period_enum not null,
  custom_period_minutes bigint default null,
  action_at_limit metered_action_at_limit_enum not null default 'none',
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now()
);
