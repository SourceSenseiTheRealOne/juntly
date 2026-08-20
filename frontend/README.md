# Juntly frontend

Localized, mobile-first Next.js shell for Juntly.

## Current scope

This bootstrap intentionally provides only:

- pt-PT default routing, English support, and Spanish-ready translations.
- Localized metadata and route boundaries.
- A responsive, accessible product-introduction shell.
- A same-origin `/api/v1/health` BFF route backed by the generated OpenAPI client.
- A localized client-side health indicator.
- Test, format, lint, type, build, dependency-audit, and CI foundations.

Authentication, listings, search, chat, quotations, bookings, payments, and
external infrastructure are not implemented yet. The only Go endpoint is the
foundation health tracer.

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

For a production runtime probe:

```bash
npm run build
npm run start -- --hostname 127.0.0.1 --port 4200
```

Then open `http://127.0.0.1:4200/`; locale routing redirects to pt-PT by default.

For local frontend/API proof, run from the repository root:

```bash
docker compose up --build
```

The compose topology supplies the server-only BFF origin. For a native frontend
runtime, copy `.env.example` to an ignored `.env.local` and start the Go API on
`127.0.0.1:8080`.

Project-wide architecture and product rules live in `../context/`.
