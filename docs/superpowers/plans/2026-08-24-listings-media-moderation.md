# Listings, Media, and Moderation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver private provider listings with transactional moderation-controlled publication and safe media upload capabilities.

**Architecture:** Extend the existing verified identity → provider capability → Ent/Postgres authority chain with persisted moderator grants, listing lifecycle/audit tables, strict OpenAPI/Go/BFF boundaries, and localized protected owner/moderator UI. All listing state transitions use optimistic revision compare-and-set transactions; media storage is injected behind a capability-safe interface.

**Tech Stack:** Go, Ent, PostgreSQL/Supabase, OpenAPI 3.1, generated TypeScript client, Next.js 16, Clerk server auth, Vitest.

## Global Constraints

- Use the canonical checkout and local commits only; do not push or mutate remotes.
- Use forward-only Supabase migrations and `go generate ./ent`.
- A `moderator` grant is persisted server-side only; no Clerk/browser role authority.
- All new behavioral code follows observed RED → GREEN → REFACTOR.
- Public/browser contracts omit exact addresses, contact data, internal user IDs, Clerk values, storage keys, credentials, and private object URLs.
- Listings are owner-only until Slice 4; no public listing pages/search in this slice.

---

### Task 1: Role and listing persistence contracts

**Files:**
- Create: `backend/ent/schema/platformrole.go`, `backend/ent/schema/listing.go`, `backend/ent/schema/listingevent.go`, `backend/ent/schema/listingmedia.go`
- Create: `supabase/migrations/<timestamp>_create_listing_moderation.sql`
- Modify: `backend/ent/schema/marketplace_reference_test.go`, `backend/internal/users/migration_contract_test.go`

- [ ] Write schema/migration tests requiring role, listing, event/media tables, lifecycle checks, unique role grant, listing owner indexes, revision check, and audit FK.
- [ ] Run targeted Go contracts; expect RED for missing schemas/migration.
- [ ] Add Ent schemas and a timestamped migration with database checks/indexes and no user/listing mock seeds.
- [ ] Run `gofmt -w ent/schema/*.go && go generate ./ent`.
- [ ] Run targeted contracts; expect GREEN.
- [ ] Apply migration to isolated local Supabase and assert table/index/check cardinality without emitting URLs or identities.
- [ ] Commit the frozen schema/migration slice.

### Task 2: Provider listing draft service and transactional repository

**Files:**
- Create: `backend/internal/listings/model.go`, `repository.go`, `service.go`, `ent_repository.go`
- Create: matching unit and PostgreSQL repository tests

- [ ] Write RED tests for provider-only create/list/get/update, profile/category/locality validation, price and service-mode bounds, cross-owner absence, and create audit event.
- [ ] Implement model, authorizer port, strict validation, and Ent transaction that creates/updates listing plus event atomically.
- [ ] Write RED PostgreSQL tests for child/reference failure rollback, exact event count, revision increment, and concurrent draft update conflict.
- [ ] Implement CAS queries and canonical reload with one bounded uniqueness retry only where needed.
- [ ] Run focused unit plus non-skipping PostgreSQL/race tests GREEN.
- [ ] Commit the frozen listing-draft slice.

### Task 3: Moderator authorization and lifecycle transitions

**Files:**
- Create: `backend/internal/moderation/service.go`, `repository.go`, `ent_repository.go`
- Modify: `backend/internal/listings/service.go`, repository/tests

- [ ] Write RED tests proving only persisted moderator grants can review, provider submit/pause/archive is owner-scoped, and invalid transitions do not append events.
- [ ] Implement a server-owned moderator authorizer using reconciled internal users and persisted `platform_roles`.
- [ ] Implement submit/approve/reject/pause/archive as state-and-event atomic CAS transactions with revision requirement and idempotent committed-target reads.
- [ ] Run disposable PostgreSQL concurrent approval/rejection proof: exactly one outcome/event wins.
- [ ] Commit the frozen lifecycle/moderation slice.

### Task 4: Safe media capability boundary

**Files:**
- Create: `backend/internal/listingmedia/model.go`, `service.go`, `storage.go`, tests
- Modify: listing persistence/repository as necessary

- [ ] Write RED tests for owner-only upload intent, content-type/size/count limits, listing-state eligibility, and public DTO redaction of adapter object references/secrets.
- [ ] Implement injected storage adapter interface and disabled production default; only an adapter-created opaque short-lived capability may cross the owner API boundary.
- [ ] Persist pending media metadata in one listing-owned transaction and require owner authorization.
- [ ] Run unit and PostgreSQL ownership/rollback tests GREEN.
- [ ] Commit the frozen media-boundary slice.

### Task 5: OpenAPI, Go handlers, and runtime composition

**Files:**
- Modify: `openapi/juntly-api.v1.yaml`, `backend/cmd/api/main.go`, `backend/internal/httpapi/health_handler.go`
- Create: listing/moderation/media handlers and contract tests

- [ ] Write RED OpenAPI and HTTP tests for every endpoint, strict request bodies, correlation parity, `401/403/409/503`, and closed owner/moderator responses.
- [ ] Add contract schemas/operations, generate TypeScript client, and implement strict Go transport/routers with verified identity middleware.
- [ ] Compose Ent repositories/services in `newAPIHandler` without startup migrations or storage credentials in logs.
- [ ] Run Go focused contracts and frontend codegen/typecheck GREEN.
- [ ] Commit the frozen transport/contract slice.

### Task 6: Same-origin BFF routes

**Files:**
- Create: `frontend/src/app/api/v1/me/listings/**`, `frontend/src/app/api/v1/moderation/listings/**`
- Modify: generated-client consumption tests

- [ ] Write RED Vitest routes that require server Clerk token for protected calls, strict request/response schemas, request-ID parity, 409 preservation, and no topology/role/storage leakage.
- [ ] Implement owner/moderator BFF routes with generated clients and closed allowlisted errors.
- [ ] Run focused routes plus strict TypeScript GREEN.
- [ ] Commit the frozen BFF slice.

### Task 7: Localized owner and moderator UI

**Files:**
- Create: `frontend/src/features/listings/**`, `frontend/src/app/[locale]/account/listings/**`, `frontend/src/app/[locale]/moderation/listings/**`
- Modify: account navigation and `frontend/messages/{pt-PT,en,es}.json`

- [ ] Write RED component/page/i18n tests for draft creation/editing, explicit submit/pause/archive actions, disabled stale/saving controls, media unavailable state, moderator review, and private-data absence.
- [ ] Implement mobile-first protected owner and moderator pages with server protection, localized copy, accessible controls, and optimistic revision handling.
- [ ] Run focused UI tests, full frontend verify, and production route build GREEN.
- [ ] Commit the frozen UI slice.

### Task 8: Slice acceptance and local integration

- [ ] Apply full migration ledger to isolated Supabase.
- [ ] Run real Postgres lifecycle/race tests, complete Go tests/vet/build, frontend verify, Compose config, CodeGraph sync, whitespace and fixture-aware privacy scans.
- [ ] Run real browser proof using approved sessions: provider draft → submit; persisted moderator review → active/rejected; owner pause/archive; no raw storage authority/private fields.
- [ ] Remove synthetic acceptance rows, stop only tracked processes, verify ports are closed, freeze/scan/commit any late fix.
- [ ] Fast-forward local `development` only after all Slice 3 evidence is green and create the next canonical feature branch.
