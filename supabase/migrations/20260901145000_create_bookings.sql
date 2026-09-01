alter table public.notifications drop constraint notifications_kind;
alter table public.notifications add constraint notifications_kind check (kind in (
  'conversation_started','message_received','conversation_reported','request_published',
  'proposal_received','proposal_accepted','proposal_rejected','booking_created','booking_updated'
));

create table public.bookings (
  id uuid primary key default gen_random_uuid(),
  customer_internal_user_id uuid not null references public.internal_users(id) on delete restrict,
  provider_internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete restrict,
  source_type text not null,
  source_id uuid,
  idempotency_key text not null,
  state text not null default 'pending_provider_confirmation',
  revision integer not null default 1,
  scheduled_at timestamptz not null,
  private_location text not null,
  agreed_price_minor integer not null,
  currency text not null default 'EUR',
  created_at timestamptz not null default timezone('utc',now()),
  updated_at timestamptz not null default timezone('utc',now()),
  constraint bookings_distinct_participants check(customer_internal_user_id<>provider_internal_user_id),
  constraint bookings_source_type check(source_type in ('proposal','listing','direct')),
  constraint bookings_source_id check((source_type='direct' and source_id is null) or (source_type in ('proposal','listing') and source_id is not null)),
  constraint bookings_idempotency_length check(char_length(idempotency_key) between 8 and 128),
  constraint bookings_state check(state in ('draft','pending_provider_confirmation','confirmed','scheduled','in_progress','completed','cancelled','disputed','refunded')),
  constraint bookings_revision check(revision>=1),
  constraint bookings_location_length check(char_length(btrim(private_location)) between 5 and 500),
  constraint bookings_price check(agreed_price_minor>0),
  constraint bookings_currency check(currency='EUR'),
  unique(customer_internal_user_id,idempotency_key)
);
create index bookings_customer_updated_idx on public.bookings(customer_internal_user_id,updated_at desc,id);
create index bookings_provider_updated_idx on public.bookings(provider_internal_user_id,updated_at desc,id);
create unique index bookings_proposal_source_idx on public.bookings(source_id) where source_type='proposal';

create table public.booking_events (
  id uuid primary key default gen_random_uuid(),
  booking_id uuid not null references public.bookings(id) on delete cascade,
  actor_internal_user_id uuid not null references public.internal_users(id) on delete restrict,
  from_state text,
  to_state text not null,
  revision integer not null,
  reason text,
  created_at timestamptz not null default timezone('utc',now()),
  constraint booking_events_from_state check(from_state is null or from_state in ('draft','pending_provider_confirmation','confirmed','scheduled','in_progress','completed','cancelled','disputed','refunded')),
  constraint booking_events_to_state check(to_state in ('draft','pending_provider_confirmation','confirmed','scheduled','in_progress','completed','cancelled','disputed','refunded')),
  constraint booking_events_revision check(revision>=1),
  constraint booking_events_reason check(reason is null or char_length(btrim(reason)) between 3 and 500),
  unique(booking_id,revision)
);
create index booking_events_booking_created_idx on public.booking_events(booking_id,created_at,id);
