create table public.provider_payment_accounts (
  internal_user_id uuid primary key references public.provider_profiles(internal_user_id) on delete cascade,
  stripe_account_id text not null unique,
  details_submitted boolean not null default false,
  charges_enabled boolean not null default false,
  payouts_enabled boolean not null default false,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint provider_payment_accounts_stripe_id check (stripe_account_id ~ '^acct_[A-Za-z0-9]+$')
);

create table public.payment_orders (
  id uuid primary key default gen_random_uuid(),
  booking_id uuid not null references public.bookings(id) on delete restrict,
  customer_internal_user_id uuid not null references public.internal_users(id) on delete restrict,
  provider_internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete restrict,
  idempotency_key text not null,
  state text not null default 'pending_checkout',
  gross_minor integer not null,
  platform_fee_minor integer not null,
  provider_net_minor integer not null,
  currency text not null default 'EUR',
  stripe_checkout_session_id text unique,
  stripe_payment_intent_id text unique,
  stripe_invoice_id text,
  stripe_refund_id text unique,
  paid_at timestamptz,
  refunded_at timestamptz,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  unique(booking_id),
  unique(customer_internal_user_id, idempotency_key),
  constraint payment_orders_state check (state in ('pending_checkout','checkout_created','processing','paid','failed','refund_pending','refunded','disputed','dispute_won','dispute_lost','cancelled')),
  constraint payment_orders_amounts check (gross_minor > 0 and platform_fee_minor >= 0 and platform_fee_minor < gross_minor and provider_net_minor = gross_minor - platform_fee_minor),
  constraint payment_orders_currency check (currency = 'EUR'),
  constraint payment_orders_idempotency check (char_length(idempotency_key) between 8 and 128),
  constraint payment_orders_checkout_id check (stripe_checkout_session_id is null or stripe_checkout_session_id ~ '^cs_[A-Za-z0-9_]+$'),
  constraint payment_orders_intent_id check (stripe_payment_intent_id is null or stripe_payment_intent_id ~ '^pi_[A-Za-z0-9]+$'),
  constraint payment_orders_invoice_id check (stripe_invoice_id is null or stripe_invoice_id ~ '^in_[A-Za-z0-9]+$'),
  constraint payment_orders_refund_id check (stripe_refund_id is null or stripe_refund_id ~ '^re_[A-Za-z0-9]+$')
);
create index payment_orders_customer_updated_idx on public.payment_orders(customer_internal_user_id, updated_at desc, id);
create index payment_orders_provider_updated_idx on public.payment_orders(provider_internal_user_id, updated_at desc, id);
create index payment_orders_state_updated_idx on public.payment_orders(state, updated_at desc, id);

create table public.payment_events (
  id uuid primary key default gen_random_uuid(),
  payment_order_id uuid not null references public.payment_orders(id) on delete cascade,
  event_type text not null,
  from_state text,
  to_state text not null,
  provider_object_id text,
  created_at timestamptz not null default timezone('utc', now()),
  constraint payment_events_type check (event_type in ('checkout_created','processing','paid','failed','refund_requested','refunded','dispute_opened','dispute_won','dispute_lost','cancelled')),
  constraint payment_events_state check (to_state in ('pending_checkout','checkout_created','processing','paid','failed','refund_pending','refunded','disputed','dispute_won','dispute_lost','cancelled')),
  constraint payment_events_from_state check (from_state is null or from_state in ('pending_checkout','checkout_created','processing','paid','failed','refund_pending','refunded','disputed','dispute_won','dispute_lost','cancelled')),
  constraint payment_events_provider_id check (provider_object_id is null or char_length(provider_object_id) between 3 and 255)
);
create index payment_events_order_created_idx on public.payment_events(payment_order_id, created_at, id);

create table public.stripe_webhook_receipts (
  stripe_event_id text primary key,
  event_type text not null,
  provider_object_id text not null,
  outcome text not null,
  processed_at timestamptz not null default timezone('utc', now()),
  constraint stripe_webhook_receipts_event_id check (stripe_event_id ~ '^evt_[A-Za-z0-9]+$'),
  constraint stripe_webhook_receipts_lengths check (char_length(event_type) between 3 and 100 and char_length(provider_object_id) between 3 and 255 and char_length(outcome) between 2 and 100)
);

create table public.payment_disputes (
  stripe_dispute_id text primary key,
  payment_order_id uuid not null references public.payment_orders(id) on delete cascade,
  stripe_charge_id text not null,
  amount_minor integer not null,
  currency text not null default 'EUR',
  state text not null,
  reason text not null,
  opened_at timestamptz not null,
  closed_at timestamptz,
  updated_at timestamptz not null default timezone('utc', now()),
  constraint payment_disputes_id check (stripe_dispute_id ~ '^dp_[A-Za-z0-9]+$'),
  constraint payment_disputes_charge check (stripe_charge_id ~ '^ch_[A-Za-z0-9]+$'),
  constraint payment_disputes_amount check (amount_minor > 0),
  constraint payment_disputes_currency check (currency = 'EUR'),
  constraint payment_disputes_state check (state in ('needs_response','under_review','won','lost','warning_closed')),
  constraint payment_disputes_reason check (char_length(reason) between 2 and 100)
);
create index payment_disputes_order_updated_idx on public.payment_disputes(payment_order_id, updated_at desc, stripe_dispute_id);
