# Juntly

**Local people. Real skills.**

**Encontra quem sabe fazer.**

Juntly is a Portuguese local-services marketplace designed to help people discover, contact, compare, book, and build trust with nearby service providers—starting in rural communities and designed for later international expansion.

## Current status

This repository is at the foundation stage. The initial delivery contains:

- Durable product, architecture, security, UI, workflow, and decision context.
- A localized responsive Next.js frontend shell under `frontend/` after the scaffold commit.
- pt-PT default, English support, and Spanish-ready routing/messages.
- Frontend test, format, lint, type, build, audit, CI, and runtime-verification foundations.

It does **not** yet implement accounts, provider profiles, listings, search, chat, quotations, bookings, reviews, payments, Go/OpenAPI, Clerk, Supabase, Redis, object storage, Docker, or production deployment.

## Repository layout

```text
juntly/
├── frontend/   Next.js frontend (created by the scaffold commit)
├── context/    sanitized durable project reference
└── AGENTS.md   project operating rules
```

`backend/` and `supabase/` will be created only when the approved API-foundation slice implements them.

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
