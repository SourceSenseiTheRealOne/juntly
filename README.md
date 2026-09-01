# Juntly

**Local people. Real skills.**

**Encontra quem sabe fazer.**

Juntly is a Portuguese local-services marketplace designed to help people discover, contact, compare, book, and build trust with nearby service providers—starting in rural communities and designed for later international expansion.

## Current status

This repository contains the production-oriented marketplace MVP:

- Durable product, architecture, security, UI, workflow, and decision context.
- A localized responsive Next.js frontend shell under `frontend/` after the scaffold commit.
- pt-PT default, English support, and Spanish-ready routing/messages.
- A source-level Clerk frontend identity foundation: localized sign-in/sign-up routes, session-aware navigation, and a server-enforced account route.
- A versioned OpenAPI health contract, generated TypeScript client, same-origin BFF, and narrow Go health API under `backend/`.
- Durable users, provider profiles, listings, discovery, private contact reveal, messaging, quotations, bookings, verified reviews, promotions, subscriptions, moderation, and bounded administration analytics.
- Supabase/PostgreSQL migrations and a local frontend/API Compose topology that connects to an explicitly supplied database.
- A digest-oriented production Compose topology, dependency readiness probe, hardened HTTP/runtime defaults, backup/restore scripts, smoke checks, and operational runbooks.
- Frontend test, format, lint, type, build, dependency-audit, CI, and runtime-verification foundations.

Paid subscriptions and promotions intentionally remain pending until an external payment provider confirms them; free configured entries activate immediately. Production deployment still requires operator-supplied infrastructure, secrets, DNS/TLS, a managed PostgreSQL target, and a real Clerk test account for authenticated acceptance journeys.

## Repository layout

```text
juntly/
├── frontend/   Next.js frontend (created by the scaffold commit)
├── backend/    Go API health-tracer foundation
├── openapi/    versioned API contracts
├── context/    sanitized durable project reference
├── compose.yaml local frontend/API development topology
├── compose.production.yaml hardened immutable-image topology
├── supabase/   ordered PostgreSQL migrations
├── scripts/    backup, restore, and smoke operations
└── AGENTS.md   project operating rules
```

Deployment and recovery procedures live in [`docs/operations/`](docs/operations/).

## Context

Start with [`context/README.md`](context/README.md). The complete source of truth is [`context/product-reference.md`](context/product-reference.md).

## Frontend commands

After the scaffold commit:

```bash
cd frontend
npm ci
npm run verify
npm run dev
```

The production shell uses port `4200` for local handoff/probes when explicitly started with that port.

## Branch model

After the one-time bootstrap:

```text
feature/* → development → staging → main
```

Changes use small reviewed PRs and squash merges. Production/main promotion requires human approval.

## Licensing

No open-source license has been selected. Public repository visibility does not grant reuse, modification, or redistribution rights beyond applicable law.
