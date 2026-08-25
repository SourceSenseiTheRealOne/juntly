# Juntly MVP Vertical Slices Design

**Status:** Approved by SourceSensei on 2026-08-23

## Goal

Deliver Juntly's Portugal-first marketplace MVP as twelve independently testable vertical slices after the Clerk-to-internal-user mapping foundation is proven in a real browser session.

## Scope and non-goals

Launch scope includes privacy-safe customer/provider accounts, configurable provider supply, public discovery, authenticated contact, chat, quotations, bookings, reviews, promotions, subscriptions, moderation, notifications, analytics, and launch readiness.

This design excludes protected marketplace payments, payouts, refunds, payment dispute automation, MB WAY-compatible payments, advanced verification/analytics, native apps, and AI matching. Those are post-validation work and must not delay discovery, direct contact, chat, or quotations.

## Foundation gate

No marketplace-owned resource may be created before the existing identity foundation is proven end to end:

```text
approved Clerk browser session
→ same-origin Next.js BFF
→ server-side Clerk token retrieval
→ Go verification and authorized-party enforcement
→ durable PostgreSQL internal-user reconciliation
```

The proof returns only status, correlation parity, and response-shape assertions; it never reveals a token, Clerk subject, or opaque internal-user UUID.

## Shared architecture

Every slice follows the same authority chain:

```text
browser
→ localized Next.js Server/Client Components
→ same-origin BFF route
→ generated OpenAPI TypeScript client
→ Go REST transport
→ application service
→ Ent repository
→ project-owned Supabase PostgreSQL/PostGIS
```

- Clerk is identity authority; Go owns marketplace roles, ownership, state transitions, moderation, and persistence.
- Browser/UI state is never authorization.
- OpenAPI remains the public contract authority; generated TypeScript is regenerated for each contract update.
- SQL migrations are forward-only and API startup never migrates production schema.
- Every persisted lifecycle uses a unique owner relation, transactionally written audit event where appropriate, and compare-and-set transitions.
- Every browser-facing upstream failure is normalized to an allowlisted privacy-safe error response with correlation ID.

## Cross-cutting requirements

- Default locale is `pt-PT`; English ships; Spanish remains structurally supported.
- Public pages expose approximate location only and never raw contact details, exact addresses, private messages, proposals, booking data, or credentials.
- Categories, locations, currencies, fees, plans, promotion periods, languages, and feature flags are administrator-configured rather than hardcoded in UI components.
- Every user-owned resource has durable server-side ownership checks.
- All significant behavioral changes use RED → GREEN → REFACTOR with focused, full, integration, and browser evidence.
- Each slice is implemented in the canonical checkout by one writer, reviewed against a frozen candidate, committed locally, then integrated through `development` according to repository policy.

## Delivery slices

### Slice 1 — Onboarding and account capabilities

Extend verified internal users with customer/provider/both capability state and privacy-safe onboarding data. It establishes server-owned role/capability checks and account settings without collecting payment, identity-document, or public contact data.

**Acceptance evidence:** a verified internal user can complete safe onboarding; unauthorized users cannot read or mutate account state; locale-aware account routes work in browser and API tests.

### Slice 2 — Taxonomy, locations, and provider profiles

Add administrator-managed service categories/subcategories, Portugal-ready locality hierarchy, approximate provider location, service areas, travel radius, and provider profile fields.

**Acceptance evidence:** administrators manage taxonomy; providers manage only their own profile/service areas; public responses omit precise address and private contact data; PostGIS/radius behavior is proven against disposable PostgreSQL.

### Slice 3 — Listings, media, and moderation

Add provider-owned listing drafts, review states, publishing/pausing/archiving, pricing/service-mode fields, and a safe storage adapter boundary for media.

**Acceptance evidence:** invalid lifecycle transitions and cross-provider edits fail closed; draft/published/moderation visibility is proven; upload metadata never exposes private storage authority.

### Slice 4 — Public discovery and search

Add SEO-safe public listing/provider pages and proximity-aware discovery by category, text, approximate location, radius, language, availability, and price type.

**Acceptance evidence:** low-bandwidth/mobile browser journey finds relevant nearby published listings; private contact/location fields are absent from public HTML and metadata; sponsored ranking cannot override relevance/trust completely.

### Slice 5 — Contact reveal and lead events

Add authenticated, provider-controlled phone/WhatsApp reveal with abuse controls, rate limiting, consent/configuration checks, and durable lead events.

**Acceptance evidence:** public responses never contain contact values; eligible authenticated requests create one authorized lead event; repeated/abusive requests are controlled; direct external arrangements remain commission-free.

### Slice 6 — Internal chat and notifications

Add participant-authorized conversations/messages, blocking/reporting, bounded attachment metadata, unread state, and in-app/email notification adapters.

**Acceptance evidence:** only conversation participants and authorized moderators access data; messages and reports remain private; notification preferences are honored; sensitive message/attachment contents are excluded from logs.

### Slice 7 — Quotation requests and proposals

Add quotation requests, provider eligibility/matching, private proposals, customer comparison, acceptance/rejection/expiry, and relevant notifications.

**Acceptance evidence:** competitors cannot retrieve each other's proposals; matching is based on server-owned status/category/location/availability; one accepted proposal has one durable outcome.

### Slice 8 — Booking state machine

Add bookings from listings, accepted proposals, or direct agreements with server-authoritative transitions, idempotency, cancellation, and dispute foundations.

**Acceptance evidence:** transitions are compare-and-set and audit-backed; duplicates do not create duplicate bookings/events; private service locations remain inaccessible to non-participants.

### Slice 9 — Reviews and reputation

Add reviews only for completed eligible bookings, provider responses, moderation, and rating aggregates.

**Acceptance evidence:** self-reviews, duplicate booking reviews, and reviews without eligible interaction are rejected; verified-booking status is derived from durable booking state.

### Slice 10 — Promotions and subscriptions

Add administrator-configured professional plans and promotion windows behind server-owned entitlements. Essential marketplace participation remains free.

**Acceptance evidence:** promotions are labelled and bounded; subscription checks do not grant admin or ownership; fee/plan literals are configuration-driven.

### Slice 11 — Administration, moderation, and analytics

Add administrator role enforcement, moderation workflows, audit records, notification preferences, and consent-aware operational metrics.

**Acceptance evidence:** privileged operations are server-authorized and audited; dashboards expose bounded aggregate data; sensitive data never reaches public analytics or logs.

### Slice 12 — Launch hardening and operations

Complete privacy/GDPR controls, accessibility, performance, backup/restore, monitoring, incident handling, Docker/staging deployment, and real end-to-end acceptance journeys.

**Acceptance evidence:** relevant security, accessibility, mobile/slow-network, migration, backup/restore, observability, Docker, HTTP, and browser gates pass; the same immutable artifact is promoted through `development → staging → main` only with owner approval.

## Completion policy

Each slice may finish only after its own contract matrix, focused tests, disposable PostgreSQL proof where persistence changes, full project gates, frozen review, and real runtime/browser evidence pass. A later slice cannot be used to claim an earlier slice is complete.
