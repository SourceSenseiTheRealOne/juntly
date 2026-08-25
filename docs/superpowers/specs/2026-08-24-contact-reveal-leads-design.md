# Contact Reveal and Lead Events Design

**Status:** Approved under SourceSensei’s standing implementation authorization on 2026-08-24.

## Goal

Deliver Slice 5: an authenticated customer can reveal a provider-controlled phone or WhatsApp contact channel for an active listing. Every successful reveal is rate-limited and creates a durable lead event. Public listing/discovery responses remain permanently contact-free.

## Policy vocabulary

| Concept | Values / source of truth |
|---|---|
| Viewer identity | Verified Clerk identity reconciled to immutable internal user ID |
| Listing eligibility | Durable `listings.state = active` only |
| Channel | `phone`, `whatsapp` |
| Provider control | Per-channel enabled flag + explicit reveal consent stored server-side |
| Contact custody | Server-only encrypted ciphertext, nonce, key version; never profile/public DTOs |
| Lead event | Customer/listing/provider/channel/timestamp only; never plaintext contact |
| Abuse limit | Maximum 10 successful reveals per customer per UTC day; at most one event per customer/listing/channel/UTC day |

No subscription tier, provider metadata, browser role, client-supplied internal ID, Clerk subject, or UI state grants reveal authority.

## Storage

### `provider_contact_channels`

One row per `(internal_user_id, channel)` with:

```text
internal_user_id
channel
ciphertext
nonce
key_version
enabled
reveal_consent
created_at
updated_at
```

Contact values use AES-256-GCM. The 32-byte key is decoded only from server-side `JUNTLY_CONTACT_ENCRYPTION_KEY`; configuration fails closed when the key is missing/malformed. Ciphertext and nonce are never returned by any HTTP/BFF/UI route or logged.

### `contact_reveal_daily_limits`

```text
customer_internal_user_id
utc_day
successful_count
```

A conditional UPSERT increments only below 10 in one transaction. This provides durable, race-safe UTC-day rate enforcement without a browser or Redis authority.

### `contact_reveal_events`

```text
id
customer_internal_user_id
provider_internal_user_id
listing_id
channel
utc_day
revealed_at
```

Unique `(customer_internal_user_id, listing_id, channel, utc_day)` makes repeated same-channel reveals idempotent per UTC day: return the same allowed contact but create no duplicate lead event/count increment.

## Reveal transaction

```text
verified Clerk identity
→ reconcile durable customer ID
→ reject self-reveal
→ load active listing and owner
→ load enabled+consented provider channel
→ acquire customer/day limit atomically
→ check same customer/listing/channel/day event
→ decrypt contact only after all authorization/rate checks
→ append event + increment daily count in one transaction
→ return channel + contact to authenticated caller
```

For an existing same-day idempotent event, decrypt/return the channel but do not increment or insert another event. Invalid listing, inactive listing, absent/disabled/unconsented channel, self-reveal, or exhausted rate limit returns a generic permission-safe error and never decrypts.

## APIs and BFFs

Owner-only configuration:

```text
GET/PUT /api/v1/me/contact-channels
```

Uses verified identity → provider capability → durable owner ID. PUT is closed, bounded, validates contact format per channel, and accepts explicit `enabled` and `revealConsent` booleans. It never returns contact plaintext after write; GET returns channel statuses only.

Authenticated customer reveal:

```text
POST /api/v1/listings/{listingId}/contact-reveals
{ "channel": "phone" | "whatsapp" }
```

Browser path remains same-origin:

```text
POST /api/v1/listings/{listingId}/contact-reveals
```

The BFF obtains Clerk bearer server-side. Public discovery/detail APIs stay unauthenticated and contractually exclude all contact fields.

## Errors

- `401 UNAUTHORIZED`: no verified identity.
- `403 FORBIDDEN`: authenticated but non-eligible, self-reveal, unconsented/disabled channel, or rate limit reached. No reason distinguishes provider settings from anti-abuse policy.
- `404 NOT_FOUND`: non-public/missing listing.
- `400 INVALID_REQUEST`: malformed closed body/channel/config request.
- `503 SERVICE_UNAVAILABLE`: vault/persistence dependency/configuration unavailable.

## UI

1. Provider account adds a private contact-channel configuration card with enabled/consent state and masked status only.
2. Public detail retains no contact value in HTML. A contact reveal control is a follow-up authenticated client action; signed-out users are sent to sign-in, not given a public fallback contact value.
3. Successful reveal is rendered only in the authenticated viewer session and never inserted into page metadata, query strings, local storage, or public cache.

## Acceptance

1. RED→GREEN unit/PostgreSQL proofs cover encryption round-trip, malformed key fail-closed, disabled/consent/self/non-active denial before decrypt, idempotent same-day event, 10/day cap under concurrency, and no plaintext in event rows.
2. OpenAPI/Go/BFF tests prove closed bodies, server-side bearer acquisition, private contact status owner-only, public contract redaction, and generic forbidden errors.
3. Browser proof saves a provider channel, reveals it as a different authenticated customer, proves one lead event, repeats without duplicate event, and confirms public page/HTML has no contact value.
4. Synthetic records, decrypted values, and local processes are removed before final commit/integration.
