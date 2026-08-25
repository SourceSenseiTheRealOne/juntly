# User Account Capabilities Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give every verified Juntly internal user an implicit customer capability and an explicit, durable provider capability that can be read and changed only through authenticated same-origin API paths.

**Architecture:** Keep `internal_users` limited to the immutable Clerk-subject mapping. Add a one-to-one `user_accounts` record keyed by `internal_user_id`; it carries only the provider capability and onboarding completion timestamp. Go reconciles the verified identity before account operations, enforces strict request schemas, and owns persistence. Next.js calls the Go API only through a same-origin BFF, then renders a localized account capability card.

**Tech Stack:** Go, Ent, project-owned Supabase PostgreSQL, OpenAPI 3.1, generated TypeScript client, Next.js App Router, Clerk server auth, Vitest, Testing Library.

## Global Constraints

- Do not begin code work until `clerk-authenticated-browser-proof` has passed and the mapping foundation is promoted according to branch policy.
- Customer capability is implicit for every verified internal user; it is not persisted as a mutable role.
- `providerEnabled` is the sole Slice 1 mutable capability. Admin/moderator authority, public provider data, contact values, payment fields, and marketplace profiles are out of scope.
- Persist no Clerk subject, token, email, phone, exact location, profile text, contact data, or session material in `user_accounts`.
- Browser requests remain same-origin `/api/v1/...`; `JUNTLY_API_ORIGIN` stays server-only.
- OpenAPI schemas are closed at every boundary: unknown BFF/Go request fields are rejected before service invocation.
- `internal_users` remains the authoritative Clerk-subject mapping and migration history stays forward-only.
- Every behavioral unit follows RED → GREEN → REFACTOR and is locally committed only after its focused evidence/review gate.

---

## File map

| Path | Responsibility |
|---|---|
| `supabase/migrations/<timestamp>_create_user_accounts.sql` | Forward-only durable one-to-one capability table. |
| `backend/ent/schema/useraccount.go` | Ent schema for the capability record. |
| `backend/ent/**` | Generated Ent model/query/mutation code. |
| `backend/internal/accounts/model.go` | Capability domain DTOs and validation errors. |
| `backend/internal/accounts/repository.go` | Repository contract and controlled errors. |
| `backend/internal/accounts/ent_repository.go` | Ent persistence adapter. |
| `backend/internal/accounts/service.go` | Identity reconciliation plus account read/update service. |
| `backend/internal/accounts/*_test.go` | Unit and disposable PostgreSQL behavior proofs. |
| `backend/internal/httpapi/account_handler.go` | Protected GET/PUT account transport and strict JSON decoding. |
| `backend/internal/httpapi/account_handler_test.go` | HTTP schema/auth/correlation tests. |
| `backend/internal/httpapi/router_test.go` | Router composition regression. |
| `backend/internal/httpapi/openapi_contract_test.go` | OpenAPI endpoint/schema regression. |
| `backend/cmd/api/main.go` | Inject account service into the router. |
| `openapi/juntly-api.v1.yaml` | Closed public account contract. |
| `frontend/src/app/api/v1/me/account/route.ts` | Clerk-aware same-origin BFF GET/PUT bridge. |
| `frontend/src/app/api/v1/me/account/route.test.ts` | BFF auth, strict body, upstream, and privacy tests. |
| `frontend/src/features/account/account-capabilities-card.tsx` | Localized interactive capability UI. |
| `frontend/src/features/account/account-capabilities-card.test.tsx` | Client UI loading/update/error behavior tests. |
| `frontend/src/app/[locale]/account/page.tsx` | Hosts the capability card after server session enforcement. |
| `frontend/src/app/[locale]/account/page.test.tsx` | Protected account page composition regression. |
| `frontend/messages/pt-PT.json`, `frontend/messages/en.json`, `frontend/messages/es.json` | Account capability UI strings. |

## Public contract

```yaml
GET /api/v1/me/account
  200: AccountCapabilitiesResponse
  401: ErrorResponse
  503: ErrorResponse

PUT /api/v1/me/account
  requestBody: UpdateAccountCapabilitiesRequest
  200: AccountCapabilitiesResponse
  400: ErrorResponse
  401: ErrorResponse
  503: ErrorResponse

AccountCapabilitiesResponse:
  additionalProperties: false
  required: [customerEnabled, providerEnabled, onboardingCompletedAt]
  properties:
    customerEnabled: { type: boolean, const: true }
    providerEnabled: { type: boolean }
    onboardingCompletedAt: { type: string, format: date-time }

UpdateAccountCapabilitiesRequest:
  additionalProperties: false
  required: [providerEnabled]
  properties:
    providerEnabled: { type: boolean }
```

`onboardingCompletedAt` is set once, on the first successful GET or PUT that creates the account record, and remains stable. The account record is created lazily and idempotently after the identity is reconciled. No caller can provide an internal-user ID or Clerk subject.

## Task 0: Foundation gate

**Objective:** Establish the identity prerequisite before any Slice 1 source change.

- [ ] Obtain real browser proof for `POST /api/v1/auth/reconcile` using an approved Clerk test user.
- [ ] Record only sanitized `200`, correlation parity, and opaque UUID/timestamp-shape evidence.
- [ ] Review and promote the mapping foundation through the repository branch policy.
- [ ] Create the Slice 1 feature branch from the approved `development` base in the canonical checkout; do not create a worktree or duplicate checkout.

## Task 1: Add the durable account capability schema

**Files:**
- Create: `supabase/migrations/<timestamp>_create_user_accounts.sql`
- Create: `backend/ent/schema/useraccount.go`
- Modify: `backend/ent/entc.go`
- Regenerate: `backend/ent/**`
- Create: `backend/ent/schema/useraccount_test.go`
- Modify: `backend/internal/users/migration_contract_test.go`

**Interfaces:**

```go
type UserAccount struct {
    InternalUserID      uuid.UUID
    ProviderEnabled     bool
    OnboardingCompletedAt time.Time
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

- [ ] **Step 1: Write failing migration/schema tests**

Assert a `user_accounts` table with `internal_user_id uuid primary key references public.internal_users(id) on delete cascade`, `provider_enabled boolean not null default false`, non-null UTC `onboarding_completed_at`, creation/update timestamps, and no identity/contact columns.

- [ ] **Step 2: Run RED checks**

Run:

```bash
cd backend && go test ./ent/schema ./internal/users -run 'Test.*UserAccount|Test.*Migration' -count=1
```

Expected: failure because the `UserAccount` schema/migration does not exist.

- [ ] **Step 3: Add the forward-only SQL migration and Ent schema**

Use a UUID primary key relation to `internal_users`; do not add a redundant subject/email field. Set `provider_enabled` default false and immutable `internal_user_id`; use UTC time defaults.

- [ ] **Step 4: Regenerate Ent and update the migration contract**

Run the repository-supported Ent generator. Extend the SQL migration contract test to assert table/column/FK/default semantics rather than only migration file existence.

- [ ] **Step 5: Run GREEN checks**

Run the Task 1 command and confirm the schema/migration tests pass.

- [ ] **Step 6: Commit**

```bash
git add supabase/migrations backend/ent backend/internal/users/migration_contract_test.go
git commit -m "feat: add user account capability schema"
```

## Task 2: Build race-safe account service and repository

**Files:**
- Create: `backend/internal/accounts/model.go`
- Create: `backend/internal/accounts/repository.go`
- Create: `backend/internal/accounts/ent_repository.go`
- Create: `backend/internal/accounts/service.go`
- Create: `backend/internal/accounts/service_test.go`
- Create: `backend/internal/accounts/ent_repository_test.go`

**Interfaces:**

```go
type Account struct {
    CustomerEnabled       bool
    ProviderEnabled       bool
    OnboardingCompletedAt time.Time
}

type Service interface {
    Get(context.Context, users.VerifiedIdentity) (Account, error)
    SetProviderEnabled(context.Context, users.VerifiedIdentity, bool) (Account, error)
}
```

- [ ] **Step 1: Write failing service tests**

Cover: invalid verified identity; first `Get` reconciles identity and creates one account with `CustomerEnabled=true`, `ProviderEnabled=false`; first update enables provider; later update disables it; unknown repository failure maps to controlled unavailability; concurrent first reads/updates produce one account and stable onboarding timestamp.

- [ ] **Step 2: Run RED checks**

Run:

```bash
cd backend && go test ./internal/accounts -run 'TestService' -count=1
```

Expected: failure because `internal/accounts` does not exist.

- [ ] **Step 3: Implement the minimal domain/repository/service layer**

The account service first calls the existing internal-user reconciliation service using only `users.VerifiedIdentity`. The Ent repository finds or inserts a record keyed by internal-user ID; unique conflicts reload the winner. Normalize timestamps to UTC microsecond precision at the persistence boundary.

- [ ] **Step 4: Add disposable PostgreSQL integration tests**

Require `TEST_DATABASE_URL`; use synthetic UUID/subjects; scope cleanup to created rows; prove exact one-to-one account row and concurrent winner reload against the migration-applied local Supabase database.

- [ ] **Step 5: Run GREEN checks**

Run:

```bash
cd backend && go test ./internal/accounts -count=1
TEST_DATABASE_URL="$DB_URL" go test ./internal/accounts -run 'TestEntRepository|TestConcurrent' -count=1
```

Expected: both pass; integration must not skip.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/accounts
git commit -m "feat: add account capability service"
```

## Task 3: Publish the protected Go account API contract

**Files:**
- Modify: `openapi/juntly-api.v1.yaml`
- Create: `backend/internal/httpapi/account_handler.go`
- Create: `backend/internal/httpapi/account_handler_test.go`
- Modify: `backend/internal/httpapi/health_handler.go`
- Modify: `backend/internal/httpapi/router_test.go`
- Modify: `backend/internal/httpapi/openapi_contract_test.go`
- Modify: `backend/cmd/api/main.go`

**Interfaces:**

```go
type AccountService interface {
    Get(context.Context, users.VerifiedIdentity) (accounts.Account, error)
    SetProviderEnabled(context.Context, users.VerifiedIdentity, bool) (accounts.Account, error)
}
```

- [ ] **Step 1: Write failing HTTP/OpenAPI tests**

Assert: unauthenticated GET/PUT return closed JSON `401`; GET returns no identity values; PUT accepts exactly `{"providerEnabled": true|false}`; `{}`, `null`, unknown properties, non-boolean values, trailing JSON, and extra JSON values return closed JSON `400` without service invocation; all responses preserve valid/generated request IDs.

- [ ] **Step 2: Run RED checks**

Run:

```bash
cd backend && go test ./internal/httpapi -run 'Test.*Account|Test.*OpenAPI' -count=1
```

Expected: failure because account routes and schemas are absent.

- [ ] **Step 3: Implement the closed API boundary**

Add the two bearer-protected operations to OpenAPI. Use Go `json.Decoder` with `DisallowUnknownFields()` and explicit end-of-input validation. Reuse only allowlisted error envelopes; map controlled account failures to `503` and invalid payloads to a documented `400` error code added to the closed contract.

- [ ] **Step 4: Generate and verify TypeScript client**

Run:

```bash
cd frontend && npm run codegen && npm run codegen:check
```

- [ ] **Step 5: Run GREEN checks**

Run the Task 3 backend test command and `npm run codegen:check`; both pass.

- [ ] **Step 6: Commit**

```bash
git add openapi backend/internal/httpapi backend/cmd/api frontend/src/shared/api/generated
git commit -m "feat: add protected account capability API"
```

## Task 4: Add same-origin account BFF routes

**Files:**
- Create: `frontend/src/app/api/v1/me/account/route.ts`
- Create: `frontend/src/app/api/v1/me/account/route.test.ts`

- [ ] **Step 1: Write failing BFF tests**

Cover GET and PUT signed-out responses (`401`, no upstream call); server-obtained bearer forwarding; valid response/correlation pass-through; strict browser JSON rejection before upstream; missing upstream origin, upstream errors, malformed data, and correlation mismatch mapping to topology-safe `503` without host/token text.

- [ ] **Step 2: Run RED checks**

Run:

```bash
cd frontend && npm test -- src/app/api/v1/me/account/route.test.ts
```

Expected: failure because the account BFF route does not exist.

- [ ] **Step 3: Implement GET and PUT BFF handlers**

Use server-side `auth()` and `getToken()` only. Validate the browser request as an exact object with the sole `providerEnabled` boolean key before calling the generated client. Forward only the bearer and correlation ID upstream. Never expose `JUNTLY_API_ORIGIN`, Clerk details, or upstream topology.

- [ ] **Step 4: Run GREEN checks**

Run the Task 4 command and confirm all account BFF tests pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/api/v1/me/account
 git commit -m "feat: add same-origin account capability BFF"
```

## Task 5: Render localized account capability controls

**Files:**
- Create: `frontend/src/features/account/account-capabilities-card.tsx`
- Create: `frontend/src/features/account/account-capabilities-card.test.tsx`
- Modify: `frontend/src/app/[locale]/account/page.tsx`
- Modify: `frontend/src/app/[locale]/account/page.test.tsx`
- Modify: `frontend/messages/pt-PT.json`
- Modify: `frontend/messages/en.json`
- Modify: `frontend/messages/es.json`
- Modify: `frontend/src/i18n/messages.test.ts`

- [ ] **Step 1: Write failing component/page/i18n tests**

Cover: loading state; fetched implicit customer capability; provider switch update; disabled controls while saving; controlled error state; no internal ID or provider token rendered; page still enforces `requireAuthenticatedUser(locale)`; exact message keys exist in all three locale files.

- [ ] **Step 2: Run RED checks**

Run:

```bash
cd frontend && npm test -- src/features/account/account-capabilities-card.test.tsx src/app/[locale]/account/page.test.tsx src/i18n/messages.test.ts
```

Expected: failure because the capability card and messages do not exist.

- [ ] **Step 3: Implement minimal accessible UI**

Render an account-only card with semantic status text and one labelled provider-capability toggle. Use the same-origin BFF only; capture the local update request identity before awaiting and ignore stale completions. Keep browser-visible copy in locale JSON; do not display Clerk subject, internal UUID, email, contact data, or marketplace profile controls.

- [ ] **Step 4: Run GREEN checks**

Run the Task 5 command and confirm all focused tests pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/account frontend/src/app/[locale]/account frontend/messages frontend/src/i18n/messages.test.ts
git commit -m "feat: add account capability onboarding UI"
```

## Task 6: Perform end-to-end Slice 1 acceptance

**Files:**
- Modify only if evidence exposes a verified defect; otherwise none.

- [ ] **Step 1: Apply the forward migration to disposable local Supabase**

Use the project-owned local Supabase workflow. Keep database URLs/process credentials in environment only and never print them.

- [ ] **Step 2: Run backend gates**

```bash
cd backend && go test ./... && go vet ./... && go build -o "$LOCALAPPDATA/Temp/juntly-api.exe" ./cmd/api
```

- [ ] **Step 3: Run frontend gates**

```bash
cd frontend && npm run verify
```

- [ ] **Step 4: Run topology and whitespace gates**

```bash
cd .. && docker compose config >/dev/null && git diff --check && git diff --cached --check
```

- [ ] **Step 5: Run live runtime/browser proof**

Start the Go API against disposable Supabase and start Next.js with a server-only local API origin. With an approved real Clerk session, prove browser GET and provider-capability PUT through the same-origin BFF, then repeat GET to prove durable state. Record only response status, correlation parity, `customerEnabled=true`, provider boolean state, and timestamp-shape assertions.

- [ ] **Step 6: Freeze/review/deliver the slice**

Synchronize CodeGraph, freeze the intended path set, scan staged paths for secrets, request a read-only contract/security review, commit the frozen candidate locally, and update the slice task only after every acceptance gate passes.

## Risks and decisions

- The current Clerk browser proof is operationally blocked by local session clock skew/preview-session availability. Slice 1 code must not be started until that proof is genuinely green.
- Provider capability is intentionally a narrow boolean. Public provider data, service areas, verification, listings, and contact controls begin in later slices.
- No external email/storage provider is selected in Slice 1. Notification/media adapters start only in their respective slices.
