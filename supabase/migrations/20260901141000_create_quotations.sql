alter table public.notifications drop constraint notifications_kind;
alter table public.notifications add constraint notifications_kind check (kind in (
  'conversation_started','message_received','conversation_reported',
  'request_published','proposal_received','proposal_accepted','proposal_rejected'
));
create unique index if not exists notifications_recipient_kind_resource_idx
  on public.notifications (recipient_internal_user_id,kind,resource_id)
  where resource_id is not null;

create table public.quotation_requests (
  id uuid primary key default gen_random_uuid(),
  customer_internal_user_id uuid not null references public.internal_users(id) on delete cascade,
  category_id uuid not null references public.service_categories(id) on delete restrict,
  locality_id uuid not null references public.localities(id) on delete restrict,
  title text not null,
  description text not null,
  budget_minor integer,
  proposal_deadline timestamptz not null,
  state text not null default 'open',
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint quotation_requests_title_length check (char_length(btrim(title)) between 5 and 140),
  constraint quotation_requests_description_length check (char_length(btrim(description)) between 20 and 4000),
  constraint quotation_requests_budget check (budget_minor is null or budget_minor > 0),
  constraint quotation_requests_state check (state in ('open','accepted','closed'))
);
create index quotation_requests_customer_updated_idx on public.quotation_requests (customer_internal_user_id,updated_at desc,id);
create index quotation_requests_opportunity_idx on public.quotation_requests (state,category_id,locality_id,proposal_deadline,id);

create table public.quotation_proposals (
  id uuid primary key default gen_random_uuid(),
  request_id uuid not null references public.quotation_requests(id) on delete cascade,
  provider_internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete cascade,
  price_minor integer not null,
  message text not null,
  available_at timestamptz not null,
  estimated_minutes integer,
  expires_at timestamptz,
  state text not null default 'submitted',
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint quotation_proposals_price check (price_minor > 0),
  constraint quotation_proposals_message_length check (char_length(btrim(message)) between 5 and 2000),
  constraint quotation_proposals_estimate check (estimated_minutes is null or estimated_minutes between 15 and 525600),
  constraint quotation_proposals_state check (state in ('submitted','accepted','rejected','expired')),
  unique (request_id,provider_internal_user_id)
);
create unique index quotation_proposals_one_accepted_idx on public.quotation_proposals (request_id) where state='accepted';
create index quotation_proposals_request_created_idx on public.quotation_proposals (request_id,created_at,id);
create index quotation_proposals_provider_updated_idx on public.quotation_proposals (provider_internal_user_id,updated_at desc,id);
