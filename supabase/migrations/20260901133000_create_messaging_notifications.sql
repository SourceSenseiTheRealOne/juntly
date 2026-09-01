create table public.conversations (
  id uuid primary key default gen_random_uuid(),
  listing_id uuid references public.listings(id) on delete set null,
  customer_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  provider_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  blocked_by_internal_user_id uuid references public.internal_users(id) on delete set null,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint conversations_distinct_participants check (customer_internal_user_id <> provider_internal_user_id),
  constraint conversations_blocker_is_participant check (blocked_by_internal_user_id is null or blocked_by_internal_user_id in (customer_internal_user_id, provider_internal_user_id)),
  unique (listing_id, customer_internal_user_id, provider_internal_user_id)
);
create index conversations_customer_updated_idx on public.conversations (customer_internal_user_id, updated_at desc, id);
create index conversations_provider_updated_idx on public.conversations (provider_internal_user_id, updated_at desc, id);

create table public.conversation_messages (
  id uuid primary key default gen_random_uuid(),
  conversation_id uuid not null references public.conversations(id) on delete cascade,
  sender_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  body text not null,
  created_at timestamptz not null default timezone('utc', now()),
  constraint conversation_messages_body_length check (char_length(btrim(body)) between 1 and 4000)
);
create index conversation_messages_conversation_created_idx on public.conversation_messages (conversation_id, created_at, id);

create table public.conversation_reports (
  id uuid primary key default gen_random_uuid(),
  conversation_id uuid not null references public.conversations(id) on delete cascade,
  message_id uuid references public.conversation_messages(id) on delete set null,
  reporter_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  reason text not null,
  created_at timestamptz not null default timezone('utc', now()),
  constraint conversation_reports_reason_length check (char_length(btrim(reason)) between 5 and 500)
);
create index conversation_reports_created_idx on public.conversation_reports (created_at desc, id);

create table public.notification_preferences (
  internal_user_id uuid primary key references public.internal_users(id) on delete cascade,
  in_app_enabled boolean not null default true,
  email_enabled boolean not null default true,
  updated_at timestamptz not null default timezone('utc', now())
);

create table public.notifications (
  id uuid primary key default gen_random_uuid(),
  recipient_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  kind text not null,
  resource_id uuid,
  in_app_visible boolean not null default true,
  read_at timestamptz,
  created_at timestamptz not null default timezone('utc', now()),
  constraint notifications_kind check (kind in ('conversation_started', 'message_received', 'conversation_reported'))
);
create index notifications_recipient_created_idx on public.notifications (recipient_internal_user_id, created_at desc, id);

create table public.notification_email_outbox (
  id uuid primary key default gen_random_uuid(),
  notification_id uuid not null unique references public.notifications(id) on delete cascade,
  recipient_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  state text not null default 'pending',
  attempt_count integer not null default 0,
  available_at timestamptz not null default timezone('utc', now()),
  created_at timestamptz not null default timezone('utc', now()),
  constraint notification_email_outbox_state check (state in ('pending', 'sent', 'failed')),
  constraint notification_email_outbox_attempt_count check (attempt_count between 0 and 20)
);
create index notification_email_outbox_delivery_idx on public.notification_email_outbox (state, available_at, id);
