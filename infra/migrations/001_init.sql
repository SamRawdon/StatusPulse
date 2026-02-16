create extension if not exists "uuid-ossp";

create table if not exists monitors (
  id uuid primary key default uuid_generate_v4(),
  name text not null,
  url text not null,
  method text not null default 'GET',
  expected_status int not null default 200,
  interval_seconds int not null default 60,
  timeout_ms int not null default 5000,
  enabled boolean not null default true,
  next_run_at timestamptz not null default now(),
  created_at timestamptz not null default now()
);

create table if not exists check_runs (
  id uuid primary key default uuid_generate_v4(),
  monitor_id uuid not null references monitors(id) on delete cascade,
  checked_at timestamptz not null default now(),
  status_code int,
  latency_ms int,
  success boolean not null,
  error_text text
);

create index if not exists idx_check_runs_monitor_time on check_runs (monitor_id, checked_at desc);
