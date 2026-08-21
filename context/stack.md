# Technology Stack

## Selection

- Client-mandated core: Next.js, TypeScript, App Router, Go JSON REST API, OpenAPI, PostgreSQL, Docker.
- UI selection confirmed: selective React Bits plus shadcn/Radix accessible primitives behind Juntly-owned wrappers.
- Authentication confirmed: Clerk email/password identity with internal Go user mapping.
- Data ownership confirmed: project-owned local Supabase PostgreSQL/PostGIS; Ent in Go; reviewed SQL migrations.
- Package manager: npm with committed lockfile because the host pnpm/Corepack shim is broken. No global toolchain mutation.
- Selection date: 2026-08-15.

## Verified bootstrap toolchain

- Node.js `24.13.1` LTS: https://nodejs.org/en/blog/release/v24.13.1
- npm `11.8.0` (installed host tool; lockfile is authoritative).
- Next.js and create-next-app `16.3.1`; stable 16.3 release: https://nextjs.org/blog/next-16-3
- React and React DOM `19.2.8`; React 19.2 release: https://react.dev/blog/2025/10/01/react-19-2
- TypeScript `5.9.3` in strict mode, selected by the official generator. TypeScript 7.0.2 was evaluated but rejected because Next 16.3.1's bundled `typescript-eslint@8.67.0` peer range is below 6.1.
- `next-intl` `4.13.6` for App Router localization: https://next-intl.dev/docs/getting-started/app-router
- Vitest `4.1.10`, Testing Library React `16.3.2`, jest-dom `7.0.1`, and jsdom `29.1.1`. jsdom 30.0.1 was rejected because it requires Node 24.15.0 or newer.
- Prettier `3.9.6` and prettier-plugin-tailwindcss `0.8.1`.

## Approved future baseline

- Frontend: modular Next.js Server Components by default; client components only for browser interactivity.
- API boundary: generated OpenAPI TypeScript client through a same-origin Next.js BFF/proxy to Go.
- Backend: current stable Go, Chi transport, domain/application/port/adapter boundaries, Ent ORM, `log/slog`, OpenTelemetry-compatible observability.
- Data: isolated project-owned local Supabase PostgreSQL with PostGIS and full-text search; one project stack active at a time.
- Redis: only when a measured rate-limit/cache/coordination/realtime requirement is implemented and documented.
- Storage: S3-compatible adapter when uploads are implemented; MinIO or equivalent may provide local proof.
- Payments: compliant marketplace-payment provider behind a provider abstraction; no custom escrow.
- Local deployment: Docker first. Production hosting remains provider-neutral pending a later cost/data-residency/operations decision.

## Current implementation boundary

The current feature branch implements the Next.js shell plus a Clerk frontend identity/session foundation. Its local tests, typecheck, production build, audit, canonical-origin dynamic routes, and signed-out redirect pass. A real authenticated browser session remains separate evidence requiring a test user. Durable Go internal-user mapping, Go/OpenAPI generation, BFF proxying, Supabase, Redis, object storage, background workers, payments, and Docker runtime topology are not yet implemented and must not be represented as working.
