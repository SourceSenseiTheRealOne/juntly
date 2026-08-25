create table public.user_accounts (
  internal_user_id uuid primary key references public.internal_users(id) on delete cascade,
  provider_enabled boolean not null default false,
  onboarding_completed_at timestamptz not null default timezone('utc', now()),
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now())
);
