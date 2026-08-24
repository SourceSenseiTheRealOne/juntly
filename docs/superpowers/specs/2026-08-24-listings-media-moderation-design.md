# Listings, Media, and Moderation Design

**Status:** Approved under SourceSensei’s standing implementation authorization on 2026-08-24

## Goal

Deliver Slice 3: provider-owned listing drafts, auditable moderation-controlled publication, pause/archive controls, and a storage-provider-neutral media upload boundary. This slice creates supply for Slice 4 without exposing listings, provider details, contact data, or storage authority publicly.

## Scope decisions

1. A persisted `moderator` role is introduced only for listing review. It is assigned by direct server-side operational data; there is no browser role assignment, Clerk metadata role, administrator UI, or general platform administration before Slice 11.
2. Every listing starts `draft`. Providers may submit and pause/archive only their own listings. Only an authorized moderator may transition a `pending_review` listing to `active` or `rejected`.
3. Listings stay owner-only in Slice 3. Slice 4 must define the public projection and search visibility explicitly.
4. Media storage is an adapter boundary. Browser responses may contain an opaque, short-lived upload capability returned by the configured adapter, but never bucket names, storage credentials, signing keys, provider account identifiers, or private object URLs.
5. Listings use EUR minor units; no floating-point prices. Quote and negotiable listings carry no price amount; fixed/hourly/daily listings require a positive amount.

## Authority chain

```text
browser
→ localized Next BFF
→ generated OpenAPI client
→ Go verified identity / moderator gate
→ listing application service
→ Ent transaction + Postgres audit event
→ project-owned Supabase PostgreSQL
```

Go derives owner and moderator authority from the verified identity and persisted records. Browser input never supplies owner IDs, role claims, review state, storage keys, or audit actors.

## Persisted model

### `platform_roles`

- `internal_user_id uuid references internal_users(id)`
- `role text` restricted initially to `moderator`
- `granted_at timestamptz`
- composite primary key `(internal_user_id, role)`

Slice 3 exposes no role mutation API. Operations may seed/grant a moderator through audited server-side data until Slice 11 owns broader role administration.

### `listings`

- opaque UUID primary key
- `internal_user_id` owner FK
- `category_id` active category FK
- `primary_locality_id` active locality FK
- title 2–140 characters; description 20–4,000 characters
- `price_type`: `fixed`, `hourly`, `daily`, `quote`, `negotiable`
- nullable `price_minor`; `currency` fixed to `EUR`
- service-mode booleans (`travels_to_customer`, `receives_customer`, `remote_services`)
- lifecycle state: `draft`, `pending_review`, `active`, `rejected`, `paused`, `archived`
- optimistic `revision`, creation/update timestamps

A listing requires an existing provider-enabled owner profile. Its service modes must be a subset of that profile’s enabled modes. The listing locality must be one of its owner profile’s service localities. This prevents a listing from silently expanding provider reach.

### `listing_events`

Each create or lifecycle transition writes one normalized event in the same transaction:

- listing ID, actor internal-user ID, event type, previous and next state, revision, timestamp;
- review rejection reason only when bounded to 500 characters;
- no raw request payload, contact value, credential, or storage response.

### `listing_media`

- opaque UUID primary key
- listing ID, bounded ordinal, safe content type, byte size, SHA-256 checksum, opaque adapter object reference, state (`pending_upload`, `ready`, `deleted`), timestamps

The database never stores a storage secret. `object_reference` is not exposed to browser/public DTOs.

## Lifecycle

```text
draft → pending_review → active
                 └────→ rejected
active → paused → pending_review
active|paused|draft|rejected → archived
```

- Owner edits only `draft`, `rejected`, or `paused` records; an edit of rejected/paused resets it to `draft` and records an event.
- Owner submits `draft` or `rejected` to `pending_review`.
- Moderator approves/rejects only `pending_review` through a compare-and-set revision transaction.
- Owner pauses only `active`; owner archives any non-archived listing.
- Archived records are immutable.
- A repeated target transition returns the committed listing without duplicating the event; incompatible or cross-owner transition fails closed.

## API and UI boundary

Owner endpoints:

```text
GET  /api/v1/me/listings
POST /api/v1/me/listings
GET  /api/v1/me/listings/{listingId}
PUT  /api/v1/me/listings/{listingId}
POST /api/v1/me/listings/{listingId}/submit
POST /api/v1/me/listings/{listingId}/pause
POST /api/v1/me/listings/{listingId}/archive
POST /api/v1/me/listings/{listingId}/media/upload-intents
```

Moderator endpoints:

```text
GET  /api/v1/moderation/listings?state=pending_review
POST /api/v1/moderation/listings/{listingId}/approve
POST /api/v1/moderation/listings/{listingId}/reject
```

All mutations require the expected `revision` in the strict JSON body. Unknown properties, nulls, second JSON values, bounds failures, unauthorized roles, and lifecycle conflicts fail closed with correlated allowlisted errors. The BFF validates exact response shapes and keeps Go origin/bearer credentials server-side.

Localized owner UI is `/[locale]/account/listings`; moderator review is a protected, non-discoverable internal route pending Slice 11 navigation work. Neither is publicly indexable in Slice 3.

## Acceptance evidence

1. migration, Ent-schema, lifecycle, and role contract tests go RED then GREEN;
2. real PostgreSQL proves atomic state/event writes, CAS rejection, idempotent retries, owner isolation, and concurrent moderation only yields one outcome;
3. media adapter tests prove no storage secret/object reference reaches API/BFF DTOs;
4. OpenAPI generation, Go handler, BFF, and localized UI tests pass;
5. browser proof shows provider draft → submit and authorized moderator approve/reject with only sanitized state/count/correlation evidence;
6. full Go race/vet/build, frontend verify, Compose, source/privacy scans, frozen review, local commit, and cleanup pass.

## Explicit deferrals

Public discovery/SEO, contact reveal, chat, quotations, bookings, reviews, paid promotion, full administration UI, generic role management, actual production storage provider configuration, and payments remain in their assigned later slices.
