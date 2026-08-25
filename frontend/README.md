# Juntly frontend

Localized, mobile-first Next.js shell for Juntly.

## Current scope

This foundation currently provides:

- pt-PT default routing, English support, and Spanish-ready translations.
- Localized metadata and route boundaries.
- A responsive, accessible product-introduction shell.
- A same-origin `/api/v1/health` BFF route backed by the generated OpenAPI client.
- A same-origin, Clerk-aware `POST /api/v1/auth/reconcile` BFF route that returns only an opaque internal mapping ID and creation time.
- A localized client-side health indicator.
- Clerk account entry, locale-aware in-app authentication forms, and a fail-closed session boundary at `/:locale/account`.
- Test, format, lint, type, build, dependency-audit, and CI foundations.

Provider/customer profiles, listings, search, chat, quotations, bookings, payments, and external infrastructure are not implemented yet. The Go API also has a protected internal-user reconciliation endpoint; authenticated browser proof requires an approved development test user and is recorded separately.

## Requirements

- Node.js `24.13.1` (see the repository `.nvmrc`).
- npm `11.8.0` or a lockfile-compatible npm release.

## Commands

```bash
npm ci
npm run dev
npm test
npm run verify
```

The OpenAPI artifact is checked with:

```bash
npm run codegen:check
```

## Local Clerk configuration

The Clerk CLI links the local checkout and writes credentials only to ignored
`frontend/.env.local`. Copy `frontend/.env.example` only when configuring a
separate development environment; never commit real key values. The app uses
first-party `/:locale/sign-in` and `/:locale/sign-up` routes, and the protected
`/:locale/account` route independently verifies the server-side session.

The source-level integration and signed-out runtime route behavior pass local
verification. Use `localhost` consistently for the browser and Next.js bind
hostname: mixing it with `127.0.0.1` can make Clerk continuation rewrites look
external and recurse through Next's proxy. An authenticated browser session still
requires a real test user; a successful build alone is not auth proof.

For a production runtime probe:

```bash
npm run build
npm run start:local
```

Then open `http://localhost:4200/`; locale routing redirects to pt-PT by default.

For local frontend/API proof, run from the repository root:

```bash
docker compose up --build
```

The compose topology supplies the server-only BFF origin. The API requires
server-only `DATABASE_URL`, Clerk verification material (`CLERK_SECRET_KEY` or
`CLERK_JWT_KEY`), and exact `CLERK_AUTHORIZED_PARTIES`; copy
`backend/.env.example` to ignored backend runtime configuration or set the same
variables in the service environment. For a native frontend runtime, copy
`.env.example` to an ignored `.env.local` and start the Go API on
`127.0.0.1:8080`. Do not add a `NEXT_PUBLIC_` prefix to `JUNTLY_API_ORIGIN` or
any backend/Clerk credential variable.

Project-wide architecture and product rules live in `../context/`.
