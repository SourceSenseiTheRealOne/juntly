# Juntly

**Local people. Real skills.**

**Encontra quem sabe fazer.**

Juntly is a Portuguese local-services marketplace designed to help people discover, contact, compare, book, and build trust with nearby service providers—starting in rural communities and designed for later international expansion.

## Current status

This repository is at the foundation stage. The initial delivery contains:

- Durable product, architecture, security, UI, workflow, and decision context.
- A localized responsive Next.js frontend shell under `frontend/` after the scaffold commit.
- pt-PT default, English support, and Spanish-ready routing/messages.
- A source-level Clerk frontend identity foundation: localized sign-in/sign-up routes, session-aware navigation, and a server-enforced account route.
- A versioned OpenAPI health contract, generated TypeScript client, same-origin BFF, and narrow Go health API under `backend/`.
- A local Docker Compose proof for the frontend and API only; no database is included.
- Frontend test, format, lint, type, build, dependency-audit, CI, and runtime-verification foundations.

The Clerk source, local quality gates, canonical-origin runtime routes, signed-out account redirect, and health tracer are verified. An authenticated browser-session journey still requires a real test user. Durable Go internal-user mapping, provider profiles, listings, search, chat, quotations, bookings, reviews, payments, Supabase, Redis, object storage, and production deployment are not implemented yet.

## Repository layout

```text
juntly/
├── frontend/   Next.js frontend (created by the scaffold commit)
├── backend/    Go API health-tracer foundation
├── openapi/    versioned API contracts
├── context/    sanitized durable project reference
├── compose.yaml local frontend/API development topology
└── AGENTS.md   project operating rules
```

`supabase/` will be created only when its approved foundation slice implements it.

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
