create table public.platform_roles (
  id uuid primary key,
  internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  role text not null,
  granted_at timestamptz not null default timezone('utc', now()),
  constraint platform_roles_role check (role in ('moderator')),
  unique (internal_user_id, role)
);

create table public.listings (
  id uuid primary key,
  internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete cascade,
  category_id uuid not null references public.service_categories(id) on delete restrict,
  primary_locality_id uuid not null references public.localities(id) on delete restrict,
  title text not null,
  description text not null,
  price_type text not null,
  price_minor integer,
  currency text not null default 'EUR',
  travels_to_customer boolean not null default false,
  receives_customer boolean not null default false,
  remote_services boolean not null default false,
  state text not null default 'draft',
  revision integer not null default 1,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint listings_title_length check (char_length(btrim(title)) between 2 and 140),
  constraint listings_description_length check (char_length(btrim(description)) between 20 and 4000),
  constraint listings_price_type check (price_type in ('fixed', 'hourly', 'daily', 'quote', 'negotiable')),
  constraint listings_price_minor check (
    (price_type in ('fixed', 'hourly', 'daily') and price_minor is not null and price_minor > 0)
    or (price_type in ('quote', 'negotiable') and price_minor is null)
  ),
  constraint listings_currency check (currency = 'EUR'),
  constraint listings_service_mode check (travels_to_customer or receives_customer or remote_services),
  constraint listings_state check (state in ('draft', 'pending_review', 'active', 'rejected', 'paused', 'archived')),
  constraint listings_revision check (revision >= 1)
);

create index listings_owner_state_updated_idx
  on public.listings (internal_user_id, state, updated_at desc, id);
create index listings_state_updated_idx
  on public.listings (state, updated_at desc, id);
create index listings_category_locality_state_idx
  on public.listings (category_id, primary_locality_id, state, id);

create table public.listing_events (
  id uuid primary key,
  listing_id uuid not null references public.listings(id) on delete cascade,
  actor_internal_user_id uuid not null references public.internal_users(id) on delete restrict,
  event_type text not null,
  from_state text,
  to_state text not null,
  revision integer not null,
  reason text,
  created_at timestamptz not null default timezone('utc', now()),
  constraint listing_events_type check (event_type in ('created', 'updated', 'submitted', 'approved', 'rejected', 'paused', 'archived')),
  constraint listing_events_from_state check (from_state is null or from_state in ('draft', 'pending_review', 'active', 'rejected', 'paused', 'archived')),
  constraint listing_events_to_state check (to_state in ('draft', 'pending_review', 'active', 'rejected', 'paused', 'archived')),
  constraint listing_events_revision check (revision >= 1),
  constraint listing_events_reason_length check (reason is null or char_length(btrim(reason)) between 1 and 500),
  constraint listing_events_reason_only_rejection check (reason is null or event_type = 'rejected')
);

create unique index listing_events_listing_revision_idx
  on public.listing_events (listing_id, revision);
create index listing_events_listing_created_idx
  on public.listing_events (listing_id, created_at, id);

create table public.listing_media (
  id uuid primary key,
  listing_id uuid not null references public.listings(id) on delete cascade,
  ordinal integer not null,
  content_type text not null,
  byte_size bigint not null,
  checksum_sha256 text not null,
  object_reference text not null,
  state text not null default 'pending_upload',
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint listing_media_ordinal check (ordinal between 1 and 10),
  constraint listing_media_content_type check (content_type in ('image/jpeg', 'image/png', 'image/webp')),
  constraint listing_media_byte_size check (byte_size between 1 and 10485760),
  constraint listing_media_checksum_sha256 check (checksum_sha256 ~ '^[0-9a-f]{64}$'),
  constraint listing_media_object_reference_length check (char_length(object_reference) between 1 and 512),
  constraint listing_media_state check (state in ('pending_upload', 'ready', 'deleted')),
  unique (listing_id, ordinal)
);

create index listing_media_listing_state_ordinal_idx
  on public.listing_media (listing_id, state, ordinal);
