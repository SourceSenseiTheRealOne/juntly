# Project Decisions

## ADR-001: Public independent child repository

- Status: accepted
- Date: 2026-08-15
- Decision: Juntly is a public repository owned by SourceSensei and located at the canonical Coding Lab child path. Parent submodule/registry registration is deferred until the dirty parent is clean.
- Consequences: public source/context must contain no private marketplace data or credentials. Parent onboarding remains a separate verified action.

## ADR-002: Frontend/backend repository topology

- Status: accepted
- Date: 2026-08-15
- Decision: Use `frontend/` for Next.js and create `backend/` for Go only in the API-foundation slice.
- Alternatives: `apps/web`/`apps/api` workspace; Next.js at repository root.
- Consequences: clear familiar deployable boundaries in one repository; avoid empty placeholder directories.

## ADR-003: OpenAPI-first Next.js BFF

- Status: accepted
- Date: 2026-08-15
- Decision: Browser clients use a generated OpenAPI TypeScript client against a same-origin Next.js BFF/proxy, which reaches the authoritative Go REST API.
- Alternatives: tRPC over Go REST; direct browser-to-Go calls.
- Consequences: one contract authority and safer cookie/CORS/upstream boundaries, with proxy code that must remain thin and allowlisted.

## ADR-004: Clerk identity with internal Go users

- Status: accepted
- Date: 2026-08-15
- Decision: Clerk owns email/password identity/session lifecycle; Go maps verified subjects uniquely to opaque internal users and enforces authorization.
- Alternatives: custom Go password/session implementation; provider-neutral deferral.
- Consequences: smaller credential attack surface; provider integration remains isolated from domain policy.

## ADR-005: Project-owned Supabase PostgreSQL/PostGIS

- Status: accepted
- Date: 2026-08-15
- Decision: Local PostgreSQL/PostGIS is owned by Juntly’s Supabase project; application Compose must not duplicate PostgreSQL. Ent is the Go ORM and reviewed SQL migrations are authoritative history.
- Consequences: Coding Lab-compliant isolation and geospatial support; local startup coordinates Supabase and application services.

## ADR-006: Local-Docker-first provider-neutral deployment

- Status: accepted
- Date: 2026-08-15
- Decision: Verify the full topology locally with Docker before selecting production hosting.
- Consequences: no premature cloud coupling; production selection remains blocked on an explicit later decision.

## ADR-007: npm host-tooling deviation

- Status: accepted
- Date: 2026-08-15
- Decision: Use npm and a committed lockfile because the host pnpm/Corepack shim resolves an invalid path. Do not mutate the global toolchain merely to satisfy the generic default.
- Consequences: reproducible installs without machine-level repair; revisit only through a reviewed package-manager migration.

## ADR-008: React Bits plus shadcn/Radix

- Status: accepted
- Date: 2026-08-15
- Decision: Use selected React Bits treatments and shadcn/Radix accessible primitives behind Juntly-owned components/tokens. Add them only when a real feature needs them.
- Consequences: visual polish without sacrificing accessible interaction or carrying unused dependencies.

## ADR-009: Three-environment promotion model

- Status: accepted
- Date: 2026-08-15
- Decision: After the one-time bootstrap, features target `development` and promote through `staging` to `main` using reviewed PRs and squash merges.
- Consequences: safer release evidence than direct feature-to-main merges; all work remains in the canonical checkout by branch switching.

## ADR-010: Frontend-only initial bootstrap

- Status: accepted
- Date: 2026-08-15
- Decision: The first repository delivery contains durable context and a verified localized Next.js shell only.
- Alternatives: complete Go/OpenAPI/Docker/Supabase tracer in the initial commit.
- Consequences: matches the requested starting scope. Documentation and reports must not call the full-stack foundation complete until the later vertical tracer exists.

## ADR-011: Vila user-facing brand with stable technical identifiers

- Status: accepted
- Date: 2026-09-02
- Decision: Present the marketplace publicly as Vila and use `https://somosvila.com` as its canonical future public origin. Preserve Juntly repository, API, environment-variable, database, migration, and package identifiers unless a separate infrastructure migration is approved.
- Consequences: customer-facing copy, metadata, and legal surfaces use Vila while deployed technical contracts remain backwards-compatible.

## ADR-012: Advance protected Stripe marketplace payments

- Status: accepted
- Date: 2026-09-02
- Decision: Implement optional protected payments now through Stripe-hosted Checkout and Connect destination charges, with server-owned EUR amounts and commission, durable payment/refund/dispute records, signed idempotent webhooks, and hosted provider onboarding. Cards and MB WAY are requested when eligible; external arrangements remain allowed and commission-free.
- Consequences: payment routes fail closed without complete server configuration and a payout-ready provider. Production activation remains blocked on rotated secrets, production Clerk, Connect and payment-method enablement, legal operator details, test-mode transaction/refund/dispute evidence, monitoring, and explicit deployment approval. No custom escrow or browser-owned financial authority is introduced.
