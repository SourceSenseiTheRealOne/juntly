# Engineering notes

This describes the committed implementation reviewed for the September 2026 presentation update. It is a source-level account, not a production acceptance report. [Publication boundaries](publication-status.md) distinguish it from ongoing local work.

## A backend that owns the rules

The browser talks to same-origin Next.js route handlers. Those handlers adapt requests through the generated OpenAPI client; the Go API verifies identity and decides access. This keeps service origins and server credentials out of the browser while allowing one API contract across TypeScript and Go.

Clerk supplies verified identity, which the backend maps to an internal user. Ownership and moderator checks belong in backend services and persistence paths. Account navigation and hidden buttons only help users find the appropriate actions.

Sources: [`frontend/src/app/api/v1/`](../frontend/src/app/api/v1/), [`backend/internal/authn/`](../backend/internal/authn/), [`backend/internal/users/`](../backend/internal/users/), [`openapi/juntly-api.v1.yaml`](../openapi/juntly-api.v1.yaml).

## Private contact access

Provider contacts are sealed with AES-256-GCM. The reveal service reconciles the caller, asks the store to authorize and reserve access, and decrypts only after that succeeds. A transaction checks the provider's sharing consent, locks the caller's per-day counter, enforces its limit and deduplicates the resulting lead.

This ordering prevents denied requests from reaching decryption. The store commits the reservation before decryption, so database authorization and response delivery are not one atomic operation. Tests explicitly exercise denial before decryption and authorized reveal behavior.

Sources: [`reveal_service.go`](../backend/internal/contactreveal/reveal_service.go), [`sql_store.go`](../backend/internal/contactreveal/sql_store.go), [`crypto.go`](../backend/internal/contactreveal/crypto.go), [`reveal_service_test.go`](../backend/internal/contactreveal/reveal_service_test.go).

## Quotations under concurrent requests

Accepting a proposal is a state transition, not a sequence of independent browser writes. The SQL store locks the quotation request, verifies customer ownership and open state, updates the chosen proposal, rejects competing proposals and writes in-app notifications before committing. Concurrent attempts share the request lock, so they cannot both accept different proposals against the same open state.

The corresponding booking and review modules have their own ownership and lifecycle rules; a submitted review is not automatically proof of a verified service.

Sources: [`quotations/sql_store.go`](../backend/internal/quotations/sql_store.go), [`bookings/`](../backend/internal/bookings/), [`reviews/`](../backend/internal/reviews/).

## External payments are a separate failure boundary

Booking payments use a stored order and a Stripe gateway interface. The store prepares an order before the external Checkout request; the returned session is attached in a subsequent transaction. Retries use stable order-derived provider idempotency keys instead of assuming a network request succeeded exactly once.

Webhook signatures are verified before processing. Receipt deduplication, order transitions and payment-event records happen transactionally. A browser redirect is not payment confirmation. This architecture still requires provider acceptance tests and reconciliation of failure paths; source tests do not prove live payments, refunds or payouts.

Platform-plan billing and promotion checkout are separate from booking Connect payments. The launch-specific implementation remains outside this published slice.

Sources: [`payments/service.go`](../backend/internal/payments/service.go), [`payments/sql_store.go`](../backend/internal/payments/sql_store.go), [`payments/stripe_gateway.go`](../backend/internal/payments/stripe_gateway.go).

## Storage and operations

Ent handles generated persistence code alongside explicit SQL for transitions that need transaction and lock control. Reviewed PostgreSQL migrations remain the schema authority; startup is not permission to mutate the schema. Listing-media adapters separate upload reservation, storage access and finalization.

The production Compose definition expects immutable image references. Readiness probes, backup/restore scripts and privacy incident notes exist, but they do not establish that a production installation has been deployed or restored successfully.

Sources: [`backend/ent/`](../backend/ent/), [`supabase/migrations/`](../supabase/migrations/), [`backend/internal/listingmedia/`](../backend/internal/listingmedia/), [`compose.production.yaml`](../compose.production.yaml), [`operations/`](operations/).

## Trade-offs and remaining evidence

The browser/API split requires contract generation and two runtimes, in exchange for an explicit business-authority boundary. SQL transactions make local state changes atomic, but cannot make an external payment provider and PostgreSQL one transaction. Encryption reduces accidental contact disclosure; it does not replace authorization, key custody or safe operational access.

No load, availability, conversion or production-scale claim is made here. Public launch needs legal/operator details, an approved deployment, current dependency checks and authenticated end-to-end journeys with the configured providers.
