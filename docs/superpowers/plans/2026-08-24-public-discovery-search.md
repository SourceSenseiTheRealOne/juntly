# Public Discovery and Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development. Steps use checkbox syntax.

**Goal:** Publish active-only discovery and SEO-safe public listing pages with deterministic locality-aware ranking.

**Architecture:** Add a separate public projection service/repository over active listings, extend OpenAPI and generated clients, add no-auth same-origin BFF routes, then deliver text-first localized discovery/detail pages. Owner/moderator DTOs are never reused as public DTOs.

**Tech Stack:** Go, Ent/PostgreSQL/PostGIS, OpenAPI 3.1, generated TypeScript client, Next.js, next-intl, Vitest.

## Global Constraints

- No Clerk/browser authentication on public discovery routes.
- Active listing state is mandatory for every public query.
- Public DTOs omit owner IDs, contacts, coordinates, service areas, media references, audit/moderation data.
- Radius uses existing PostGIS locality centers and parameterized `ST_DWithin`.
- Organic deterministic ordering only; promotions remain deferred.

### Task 1: Public discovery service and PostGIS repository

- [ ] Write RED unit/PostgreSQL tests for active-only state, category/text/price/service filters, locality pairing, radius ordering, stable tie-break, and private-column absence.
- [ ] Add `backend/internal/discovery` model, service, repository, and parameterized SQL query.
- [ ] Prove all reference filters against isolated Supabase/PostGIS.
- [ ] Commit the frozen discovery service slice.

### Task 2: OpenAPI, Go transport, and public BFF

- [ ] Write RED OpenAPI/handler/BFF tests for public collection/detail operations, strict query keys, safe errors, and closed DTOs.
- [ ] Add discovery contract, generated client, Go handlers, and no-auth BFF routes.
- [ ] Prove no Clerk token or topology is used/exposed.
- [ ] Commit the frozen transport slice.

### Task 3: Localized discovery and SEO-safe pages

- [ ] Write RED component/page/i18n tests for discovery filters, text cards, active detail page, metadata redaction, and non-active not-found behavior.
- [ ] Add `/:locale/discover` and `/:locale/listings/[listingId]` pages.
- [ ] Run low-bandwidth/mobile text-first browser proof.
- [ ] Commit the frozen UI slice.

### Task 4: Acceptance and local integration

- [ ] Run full PostgreSQL/race/Go/frontend/Compose/CodeGraph gates.
- [ ] Create local acceptance listings through the real owner/moderator flow; prove discovery finds only active listing.
- [ ] Remove synthetic acceptance records, stop ports, freeze/scan, commit, and fast-forward local development.
