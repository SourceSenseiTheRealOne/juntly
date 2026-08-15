# Juntly frontend

Localized, mobile-first Next.js shell for Juntly.

## Current scope

This bootstrap intentionally provides only:

- pt-PT default routing, English support, and Spanish-ready translations.
- Localized metadata and route boundaries.
- A responsive, accessible product-introduction shell.
- Test, format, lint, type, build, dependency-audit, and CI foundations.

Authentication, listings, search, chat, quotations, bookings, payments, the Go API, and external infrastructure are not implemented yet.

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

For a production runtime probe:

```bash
npm run build
npm run start -- --hostname 127.0.0.1 --port 4200
```

Then open `http://127.0.0.1:4200/`; locale routing redirects to pt-PT by default.

Project-wide architecture and product rules live in `../context/`.
