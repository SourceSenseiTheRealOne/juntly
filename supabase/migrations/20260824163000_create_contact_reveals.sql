create table public.provider_contact_channels (
  id uuid primary key,
  internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete cascade,
  channel text not null,
  ciphertext bytea not null,
  nonce bytea not null,
  key_version text not null,
  enabled boolean not null default false,
  reveal_consent boolean not null default false,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint provider_contact_channels_channel check (channel in ('phone', 'whatsapp')),
  constraint provider_contact_channels_ciphertext_length check (octet_length(ciphertext) between 17 and 1024),
  constraint provider_contact_channels_nonce_length check (octet_length(nonce) = 12),
  constraint provider_contact_channels_key_version_length check (char_length(key_version) between 1 and 32),
  unique (internal_user_id, channel)
);

create index provider_contact_channels_owner_idx
  on public.provider_contact_channels (internal_user_id, enabled, reveal_consent, channel);

create table public.contact_reveal_daily_limits (
  id uuid primary key,
  customer_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  utc_day timestamptz not null,
  successful_count integer not null default 0,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint contact_reveal_daily_limits_day_midnight check (utc_day = date_trunc('day', utc_day)),
  constraint contact_reveal_daily_limits_count check (successful_count between 0 and 10),
  unique (customer_internal_user_id, utc_day)
);

create table public.contact_reveal_events (
  id uuid primary key,
  customer_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  provider_internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete cascade,
  listing_id uuid not null references public.listings(id) on delete cascade,
  channel text not null,
  utc_day timestamptz not null,
  revealed_at timestamptz not null default timezone('utc', now()),
  constraint contact_reveal_events_channel check (channel in ('phone', 'whatsapp')),
  constraint contact_reveal_events_day_midnight check (utc_day = date_trunc('day', utc_day)),
  unique (customer_internal_user_id, listing_id, channel, utc_day)
);

create index contact_reveal_events_listing_created_idx
  on public.contact_reveal_events (listing_id, revealed_at, id);
