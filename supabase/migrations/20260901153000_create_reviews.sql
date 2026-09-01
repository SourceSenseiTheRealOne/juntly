alter table public.notifications drop constraint notifications_kind;
alter table public.notifications add constraint notifications_kind check (kind in (
  'conversation_started','message_received','conversation_reported','request_published','proposal_received',
  'proposal_accepted','proposal_rejected','booking_created','booking_updated','review_received','review_response'
));
create table public.reviews(
 id uuid primary key default gen_random_uuid(),
 booking_id uuid not null unique references public.bookings(id) on delete restrict,
 customer_internal_user_id uuid not null references public.internal_users(id) on delete restrict,
 provider_internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete restrict,
 rating integer not null,
 body text not null,
 provider_response text,
 verified_booking boolean not null default true,
 state text not null default 'published',
 created_at timestamptz not null default timezone('utc',now()),
 updated_at timestamptz not null default timezone('utc',now()),
 constraint reviews_distinct_participants check(customer_internal_user_id<>provider_internal_user_id),
 constraint reviews_rating check(rating between 1 and 5),
 constraint reviews_body_length check(char_length(btrim(body)) between 10 and 2000),
 constraint reviews_response_length check(provider_response is null or char_length(btrim(provider_response)) between 3 and 1000),
 constraint reviews_state check(state in ('published','hidden'))
);
create index reviews_provider_created_idx on public.reviews(provider_internal_user_id,created_at desc,id) where state='published';
create table public.provider_rating_aggregates(
 provider_internal_user_id uuid primary key references public.provider_profiles(internal_user_id) on delete cascade,
 rating_sum bigint not null default 0,
 review_count integer not null default 0,
 updated_at timestamptz not null default timezone('utc',now()),
 constraint provider_rating_sum check(rating_sum>=0),
 constraint provider_rating_count check(review_count>=0)
);
