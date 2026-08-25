
create extension if not exists pgcrypto;

create table public.internal_users (
  id uuid primary key,
  clerk_subject text not null unique,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint internal_users_clerk_subject_nonempty
    check (char_length(clerk_subject) between 1 and 255)
);

create index internal_users_created_at_idx
  on public.internal_users (created_at);
