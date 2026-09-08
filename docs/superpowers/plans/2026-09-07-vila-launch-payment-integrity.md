# Vila Launch Payment Integrity Implementation Plan

> **For agentic workers:** Use `superpowers:executing-plans` inline in the canonical checkout. User approved the complete launch-completion sequence and selected Supabase Storage + Resend.

**Goal:** Prevent payment webhooks and delayed refund responses from corrupting durable financial state before completing the remaining launch integrations.

**Architecture:** Keep verified raw Stripe events at the gateway boundary and enforce persisted processor identity, exact money, authorization and monotonic state transitions inside PostgreSQL transactions. Keep browser DTOs unchanged; no credentials or live provider operations are needed for this slice.

**Tech Stack:** Go, database/sql + pgx, PostgreSQL, existing Stripe HTTP adapter, Go tests.

## Global Constraints

- Vila is the public brand; preserve internal Juntly identifiers.
- No Heroku deployment, paid resources, live payments, external credential changes, or remote database mutation.
- Canonical checkout, one writer, RED → GREEN → REFACTOR. No commits/pushes without renewed explicit authorization.
- Supabase Storage and Resend are the approved next integrations, not proof of configured provider accounts.
- Retain Portugal/EUR, optional commission-free external arrangements, and the existing exact administrator identity boundary.

## Approved completion sequence

1. Payment correctness and recovery.
2. Real Supabase listing uploads, verification and publication.
3. Resend email delivery with durable claiming/retry and notification preferences.
4. Paid plan and promotion Checkout/activation/renewal/cancellation.
5. Customer/provider/admin journey verification, privacy/accessibility and launch operations.

Each subsequent subsystem gets a bounded plan after tracing its current contracts. Existing future exclusions (advanced verification/analytics, native apps, AI matching, institutional dashboards) remain exclusions.

## Task 1: Refund completion cannot regress state

Files: `backend/internal/payments/sql_store.go`, new `backend/internal/payments/refund_integrity_integration_test.go`.

Interface: existing `Store.AttachRefund(context.Context, uuid.UUID, uuid.UUID, RefundResult) (Order, error)`.

- [ ] RED: PostgreSQL tests prepare a paid order, deliver a full refund webhook first, then attach the HTTP refund response. Assert the order stays refunded and replay adds no audit event. Test unauthorized actor and a different refund ID on pending state.
- [ ] GREEN: authorize in the attachment transaction, lock and inspect the order, allow only paid/dispute-won → refund-pending; treat matching pending/refunded attachment as idempotent, retaining terminal state.
- [ ] Verify using `go test -count=1 ./internal/payments -run RefundIntegrity -v` with an isolated test database.

## Task 2: Reconcile verified events against durable Checkout and money

Files: `backend/internal/payments/stripe_gateway.go`, `stripe_gateway_test.go`, `sql_store.go`, `sql_store_integration_test.go`, new `webhook_integrity_integration_test.go`.

Interface: existing `ProviderEvent.AmountMinor`, `Currency`, `ProviderObjectID`, `OrderID`, `PaymentIntentID`.

- [ ] RED: project `amount_total` and currency; wrong/missing amount, currency, persisted Checkout ID or PaymentIntent must not change order or create a processed receipt. Partial refunds must not mark the whole order refunded. Contradictory event-ID replay must conflict.
- [ ] GREEN: validate bounded verified projection fields; resolve Checkout by its persisted session ID rather than trusting metadata alone; compare metadata/amount/currency and existing PaymentIntent under the row lock. Check duplicate receipt identity/outcome and make same-state replay a no-op without duplicate financial events.
- [ ] Verify exact paid/replay/refund tests in PostgreSQL and gateway signature-negative tests.

## Task 3: Verification and handoff into uploads

- [ ] Run focused race tests, full Go tests, vet/build, CodeGraph affected, and diff whitespace checks.
- [ ] Keep provider-side activation explicitly blocked until rotated credentials, legal operator/fee/tax decisions and real test-mode payment/refund/payout evidence exist.
- [ ] Record source evidence separately from external launch proof; do not call the complete product launch-ready after this slice alone.
