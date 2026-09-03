# Deployment and recovery runbook

## Release inputs

Deploy immutable digest-pinned API and frontend images through `compose.production.yaml`. Supply runtime values through the deployment platform's secret store. Never commit an environment file.

Required values are `JUNTLY_API_IMAGE`, `JUNTLY_FRONTEND_IMAGE`, `DATABASE_URL`, production Clerk keys and authorized parties, `JUNTLY_CONTACT_ENCRYPTION_KEY`, a rotated `STRIPE_SECRET_KEY`, the Dashboard-created `STRIPE_WEBHOOK_SECRET`, canonical HTTPS `JUNTLY_PUBLIC_ORIGIN`, reviewed `JUNTLY_PLATFORM_FEE_BPS`, and the public frontend port. Never reuse a credential pasted into chat, an issue, CI logs, or source control.

The canonical future public origin is `https://somosvila.com`. Name the Stripe event destination `Vila Production Payments` and configure its endpoint as `https://somosvila.com/api/v1/payments/webhooks/stripe`. Subscribe only to the allowlisted payment and connected-account events documented below; do not subscribe to SetupIntent lifecycle events.

Production Clerk must authorize `https://somosvila.com`; local development must authorize `http://localhost:4200`. Google OAuth is configured in Clerk for both origins. Keep the OAuth client secret and Clerk secret only in their provider/deployment secret stores.

Use the manually dispatched `Publish immutable images` GitHub workflow to build a reviewed ref. It publishes `linux/amd64` API and frontend images to GHCR with the resolved full commit SHA as the only tag, disables mutable provenance/SBOM side artifacts, and records each registry digest in the workflow summary. Supply the resulting `image@sha256:...` references to `compose.production.yaml`; do not deploy a branch tag.

## Staging promotion

1. Apply all `supabase/migrations/` files in timestamp order to a backed-up staging database.
2. Start the digest-pinned images with `docker compose -f compose.production.yaml up -d`.
3. Require the API `/api/v1/ready` check and frontend health check to pass.
4. Run `JUNTLY_BASE_URL=https://staging.example scripts/smoke.sh` against the API origin.
5. Exercise one localized signed-out discovery journey and one Clerk-authenticated account journey.
6. Complete Stripe Connect onboarding with a test provider. Require `details_submitted`, `charges_enabled`, and `payouts_enabled` before checkout.
7. In Stripe test mode, exercise one card payment and one eligible MB WAY payment. Verify gross, fee, provider net, invoice reference, and one idempotent signed webhook receipt in PostgreSQL.
8. Exercise one refund and one synthetic dispute event. Confirm duplicate webhook delivery creates no duplicate payment event.
9. Confirm Stripe Tax registrations/settings, invoice branding, Connect platform profile, MB WAY enablement, legal operator identity, VAT/tax details, terms, privacy, and cancellation/refund wording with the responsible business/legal owner.
10. Record image digests and migration head. Promote those exact digests to production only after explicit approval.

## Backup and restore

Run `BACKUP_DIR=<protected-directory> scripts/backup-postgres.sh` before migrations and on the scheduled backup cadence. Move dumps and checksums to encrypted, access-controlled storage with retention appropriate to legal obligations.

Restore only into an isolated target first:

`DATABASE_URL=<isolated-target> BACKUP_FILE=<dump> CONFIRM_RESTORE=RESTORE scripts/restore-postgres.sh`

After restore, apply pending migrations, run readiness and smoke checks, compare bounded record counts, then authorize traffic. Never test restores against production.

## Rollback

Application rollback uses the previous immutable image digests. Database migrations are forward-only. If a migration causes an incident, stop writes, restore into a new database from the verified pre-migration backup, apply the reviewed recovery migration if applicable, validate, and switch the database endpoint under owner approval.

## Monitoring

Alert on sustained `/api/v1/ready` failures, elevated 5xx rate, latency, database connection exhaustion, email-outbox failures, container restarts, Stripe webhook failures/retries, payment orders stuck in `processing` or `refund_pending`, open disputes, Connect accounts losing charge/payout capability, and payout failures visible in Stripe. Correlate requests by `X-Request-ID`; do not put message bodies, contact data, Stripe payloads/signatures, tokens, or database URLs in logs.

## Payment activation gate

Keep production checkout unavailable until every item below is true:

- the previously exposed live secret is revoked and a replacement exists only in the deployment secret store;
- Stripe Connect is activated for the platform and each provider completes Stripe-hosted onboarding;
- the public webhook points to `https://somosvila.com/api/v1/payments/webhooks/stripe` and subscribes to `checkout.session.completed`, `checkout.session.async_payment_succeeded`, `checkout.session.async_payment_failed`, `charge.refunded`, `charge.dispute.created`, `charge.dispute.closed`, and `account.updated`;
- cards and MB WAY are enabled for EUR where the Stripe account and transaction are eligible;
- automatic tax, tax-ID collection, invoices, platform fee, refund policy, dispute ownership, payout schedule, and reserve exposure are reviewed;
- production Clerk, database backups, legal operator details, privacy contacts, and monitoring are configured;
- test-mode payment, payout, refund, dispute, replay, and recovery evidence has been recorded.
