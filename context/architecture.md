# Architecture Context

## Repository and deployables

```text
juntly/
├── frontend/   Next.js App Router application (bootstrap scope)
├── backend/    Go modular-monolith API (created in a later slice)
├── context/    durable sanitized project knowledge
└── supabase/   project-owned PostgreSQL/PostGIS configuration (later slice)
```

The initial bootstrap created `frontend/`. The Clerk frontend identity foundation is implemented in that application, and the health-tracer foundation adds `backend/` plus the versioned OpenAPI contract because both now contain working code. `supabase/` remains absent until its approved persistence slice; empty future-runtime directories remain prohibited.

## Approved request path

```text
browser
  → generated OpenAPI TypeScript client
  → same-origin Next.js BFF/proxy
  → versioned Go JSON REST API
  → domain/application ports
  → Ent persistence/integration adapters
  → project-owned Supabase PostgreSQL/PostGIS
```

The BFF owns browser-safe DTO exposure, Clerk session verification/propagation, input validation, timeouts, correlation IDs, and stable error mapping. It does not pretend browser-visible traffic is secret. The Go API owns marketplace business rules, authorization, entity state transitions, idempotency, persistence, and integrations.

## Frontend boundaries

- `src/app`: routes, layouts, metadata, error/loading boundaries, BFF route composition.
- `src/features/<feature>`: feature UI, typed hooks/services, schemas, and public API.
- `src/components/ui`: Juntly-owned accessible primitives and reviewed React Bits source when actually needed.
- `src/i18n`: routing, request configuration, and locale-safe navigation.
- Components do not issue ad hoc privileged backend requests or import database/Go-internal concerns.
- User-facing copy comes from locale resources, not reusable component literals.

## Backend boundaries

Future modules use domain → application → ports → adapters dependency direction. Transport maps versioned HTTP/OpenAPI contracts to application use cases. Repositories, payment providers, storage, email, notifications, search, and realtime transports implement narrow ports. Independent microservices require measured evidence and a later decision.

## Authentication and authorization

Clerk owns primary email/password identity and session lifecycle. The frontend provides localized Clerk entry routes and resource-local server session enforcement; UI visibility is presentation only. Local browser-facing Next.js routes use the canonical `localhost:4200` origin, because mixing loopback aliases can turn Clerk continuation rewrites into recursive external proxies. Once the API parent is integrated, Go will map each verified Clerk subject uniquely to an opaque internal user and enforce platform role, provider/customer domain role, entitlement, ownership, resource membership, and administrative policy. A real authenticated browser session remains a separate verification gate.

## Data and integration ownership

- PostgreSQL is authoritative; PostGIS handles distance/radius; PostgreSQL full-text search is initial search infrastructure.
- Ent models Go persistence; reviewed deterministic SQL in `supabase/migrations/` is migration/deployment history.
- Exact addresses, contact details, identity documents, messages, proposals, booking locations, and payment data are private.
- Redis is optional and never the source of truth.
- S3-compatible storage is introduced with strict upload types/sizes/ownership and signed access.
- Payment providers sit behind an abstraction; external arrangements generate no Juntly transaction commission.

## Runtime topology

Local development uses a child-owned Supabase stack for PostgreSQL/PostGIS and mail tooling. Future application Compose runs frontend, Go API, and justified adjunct services without a duplicate PostgreSQL container. Staging/production have isolated credentials, data, caches, storage, service identities, and observability. Production hosting is not selected yet.

## Invariants

- Public HTML/metadata/indexes never contain raw phone/WhatsApp numbers or exact private addresses.
- Proposal values are private to the customer and submitting provider; competitors never see them.
- Categories, fees, plans, currencies, and promotion periods are configurable, not UI constants.
- Go validates booking/payment/refund/promotion state transitions and idempotency.
- No commission on external arrangements; contact reveal stays free after authentication.
- Promoted results are labelled and never override relevance/trust completely.
- No custom escrow and no claim that verification guarantees work quality.
- Every protected resource is authorized server-side.

## Vertical tracer for the next foundation slice

The implemented smallest full-stack proof crosses a localized Next.js client
island → same-origin BFF → generated OpenAPI client → Go `/api/v1/health`
transport → framework-independent application service → response mapping. It
returns one matching `X-Request-ID`/body correlation value, fails closed with a
privacy-safe BFF error, and runs locally in frontend/API containers. It does
not claim any Clerk, persistence, marketplace, or payment flow exists.
