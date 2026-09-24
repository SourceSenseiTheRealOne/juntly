# Local development

Use this guide for the committed marketplace. It does not enable the unpublished redesign or launch-commercial work, and it does not authorize live payments, email, database migrations against production or a deployment.

## Toolchain and checks

Use Node.js `24.13.1` from [`.nvmrc`](../.nvmrc), npm `11.8.0` or a lockfile-compatible version, and the Go toolchain declared in [`backend/go.mod`](../backend/go.mod). Docker Compose is optional for local topology; PostgreSQL is required for complete application journeys.

```bash
git clone https://github.com/SourceSenseiTheRealOne/vila.git
cd vila
npm --prefix frontend ci
npm --prefix frontend run verify
(cd backend && go test ./...)
```

`verify` checks generated-client drift, formatting, unit tests, lint, types, the production build and dependency advisories. A failure in any step is a failure of the command. Historical CI results do not replace a fresh dependency audit. These instructions are not a claim that this presentation update reran the full application suite.

## Frontend preview

Copy [`frontend/.env.example`](../frontend/.env.example) to ignored `frontend/.env.local` in a fresh checkout. Configure a development Clerk instance through its credential tooling, not committed files. Keep `CLERK_SECRET_KEY` and `JUNTLY_API_ORIGIN` server-side; only Clerk's publishable key uses a `NEXT_PUBLIC_` variable.

```bash
npm --prefix frontend run dev:local
```

Open `http://localhost:4200/`. Use `localhost` consistently for Clerk and browser navigation; mixing it with `127.0.0.1` can break authentication continuations. The server-only API origin may still point to `http://127.0.0.1:8080`.

Without the backend, account data and marketplace operations are unavailable. A rendered page or signed-out redirect is not an authenticated acceptance test. Do not invent providers, transactions or statistics to make a preview look populated.

## API and database

Use a dedicated development PostgreSQL target and apply the reviewed [`supabase/migrations/`](../supabase/migrations/) through the documented migration workflow. Never point setup commands at production by convenience. The Supabase project ID remains `juntly`; do not recreate it under the new repository name.

Supply `DATABASE_URL`, Clerk verification material (`CLERK_SECRET_KEY` or `CLERK_JWT_KEY`) and exact `CLERK_AUTHORIZED_PARTIES` through the API process environment. An ignored `.env.local` is a storage convention, not an automatic Go dotenv loader. Contact reveal additionally needs the server-only `JUNTLY_CONTACT_ENCRYPTION_KEY`. Optional media storage has separate server-only configuration.

For a payment-disabled preview, leave **all** of `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `JUNTLY_PUBLIC_ORIGIN` and `JUNTLY_PLATFORM_FEE_BPS` empty/unset. The example file includes a proposed public origin: do not blindly source it. Partial payment configuration fails startup rather than silently enabling a degraded payment flow.

With that development-only environment prepared:

```bash
# Terminal 1, from the repository root:
(cd backend && JUNTLY_API_ADDR=127.0.0.1:8080 go run ./cmd/api)
# Terminal 2:
npm --prefix frontend run dev:local
```

Alternatively, [`compose.yaml`](../compose.yaml) runs the frontend/API pair against an explicitly supplied database. Use `docker compose -p juntly up --build` to preserve the pre-rename Compose project identity. It does not provision the application database for you. Never publish the API or database ports just to make a local test convenient.

## Integration and release checks

Authenticated acceptance requires an approved test account. Payments and storage require their own configured development providers and explicit test procedures. Do not enable email or payment credentials as part of documentation verification.

Feature PRs target `development`, followed by reviewed `staging` and `main` promotions. Image publishing is a manually dispatched workflow; deployment and production/main promotion require owner approval. Use the [deployment/recovery runbook](operations/deployment-recovery.md) for the separate release process.

Older foundation documents may understate the implemented API or describe future milestones. Read the [dated publication status](publication-status.md) together with the current source instead of treating those documents as release evidence.
