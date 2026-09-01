# Deployment and recovery runbook

## Release inputs

Deploy immutable digest-pinned API and frontend images through `compose.production.yaml`. Supply runtime values through the deployment platform's secret store. Never commit an environment file.

Required values are `JUNTLY_API_IMAGE`, `JUNTLY_FRONTEND_IMAGE`, `DATABASE_URL`, Clerk keys and authorized parties, `CONTACT_CHANNEL_ENCRYPTION_KEY`, and the public frontend port.

## Staging promotion

1. Apply all `supabase/migrations/` files in timestamp order to a backed-up staging database.
2. Start the digest-pinned images with `docker compose -f compose.production.yaml up -d`.
3. Require the API `/api/v1/ready` check and frontend health check to pass.
4. Run `JUNTLY_BASE_URL=https://staging.example scripts/smoke.sh` against the API origin.
5. Exercise one localized signed-out discovery journey and one Clerk-authenticated account journey.
6. Record image digests and migration head. Promote those exact digests to production.

## Backup and restore

Run `BACKUP_DIR=<protected-directory> scripts/backup-postgres.sh` before migrations and on the scheduled backup cadence. Move dumps and checksums to encrypted, access-controlled storage with retention appropriate to legal obligations.

Restore only into an isolated target first:

`DATABASE_URL=<isolated-target> BACKUP_FILE=<dump> CONFIRM_RESTORE=RESTORE scripts/restore-postgres.sh`

After restore, apply pending migrations, run readiness and smoke checks, compare bounded record counts, then authorize traffic. Never test restores against production.

## Rollback

Application rollback uses the previous immutable image digests. Database migrations are forward-only. If a migration causes an incident, stop writes, restore into a new database from the verified pre-migration backup, apply the reviewed recovery migration if applicable, validate, and switch the database endpoint under owner approval.

## Monitoring

Alert on sustained `/api/v1/ready` failures, elevated 5xx rate, latency, database connection exhaustion, email-outbox failures, and container restarts. Correlate requests by `X-Request-ID`; do not put message bodies, contact data, tokens, or database URLs in logs.
