# Durable Internal User Mapping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` task-by-task. Every behavior change follows RED → GREEN → REFACTOR.

**Goal:** Persist one opaque Juntly UUID per Clerk-verified subject, expose it through a same-origin reconciliation route, and prove idempotency against disposable Supabase PostgreSQL.

**Architecture:** `feature/internal-user-mapping` consumes the immutable health/API parent (`727f288`) and the committed Clerk foundation (`e154f44`) as a local stacked branch. The browser calls a Next.js BFF with no identity body; the BFF obtains a current Clerk session token, Go verifies it, and a provider-neutral service reconciles the verified subject through Ent into the single `public.internal_users` table.

**Tech Stack:** Next.js 16.3.1, Clerk Next.js 7.7.9, Go 1.26, Clerk Go SDK v2.7.0, Ent v0.14.6, pgx v5.10.0, Supabase CLI 2.78.1, PostgreSQL, OpenAPI 3.1, Vitest.

## Global Constraints

- Preserve `feature/clerk-identity` and `feature/health-tracer`; consume their local commits rather than copying their source.
- The browser must never submit a Clerk subject, internal ID, profile field, role, or entitlement.
- `internal_users` stores only opaque UUID, immutable verified subject, and timestamps. No email, name, provider profile, metadata, or token is persisted.
- SQL in `supabase/migrations/` is the migration authority. Ent reflects SQL; API startup never auto-migrates.
- Bearer verification happens in Go; BFF session validation is defense in depth.
- `JUNTLY_API_ORIGIN`, database URLs, Clerk/Supabase credentials, tokens, and generated keys stay server-only and ignored. Do not print, stage, or commit values.
- The BFF maps backend configuration/upstream/malformed-payload/request-ID failures to privacy-safe documented errors; it never exposes backend topology.
- Preserve `localhost:4200` as the canonical browser auth origin.
- Do not add marketplace, profile, payment, billing, role, entitlement, webhook, or provider-ownership scope.
- No push, PR, remote Supabase action, Clerk write, or mapping-feature commit is authorized by this plan. Task 1 alone is authorized to create the local merge commit required to stack `e154f44` on the health/API base; pause after verification for an explicit delivery decision.

---

### Task 1: Build the local stacked predecessor baseline

**Objective:** Put the already verified health/API and Clerk foundations in the implementation branch without modifying their source commits.

**Files:**
- Merge input: local commit `e154f44` from `feature/clerk-identity`
- Verify: `backend/go.mod`, `frontend/package.json`, `frontend/src/proxy.ts`, `frontend/src/app/api/v1/health/route.ts`

**Interfaces:**
- Consumes: `727f288` as `feature/internal-user-mapping` base and `e154f44` as the Clerk source commit.
- Produces: one clean branch where the Go API and Clerk server helpers are both available.

- [ ] **Step 1: Capture the immutable branch inputs**

Run:

```bash
git status --short --branch
git rev-parse HEAD development feature/health-tracer feature/clerk-identity
git merge-base --is-ancestor 727f288 HEAD
git merge-base --is-ancestor e154f44 HEAD
```

Expected: clean branch, health is an ancestor, Clerk is not yet an ancestor.

- [ ] **Step 2: Merge only the local Clerk foundation**

Run:

```bash
git merge --no-ff e154f44 -m "chore: stack Clerk identity foundation for user mapping"
```

Resolve only overlapping health/Clerk source by preserving both route sets and their tests. Do not alter either predecessor commit or update a remote.

- [ ] **Step 3: Verify the merge result**

Run:

```bash
npm --prefix frontend run verify
cd backend && go test ./...
cd .. && git diff --check && codegraph sync . && codegraph affected .
```

Expected: frontend verification and Go tests pass; whitespace check succeeds; CodeGraph is synchronized.

- [ ] **Step 4: Checkpoint**

Do not commit. Record the merge SHA and exact test results for the later final review.

### Task 2: Add the authoritative Supabase migration and matching Ent schema

**Objective:** Define one forward-only, minimal mapping table and make generated Ent code represent its exact columns and constraints.

**Files:**
- Create: `supabase/config.toml`
- Create: `supabase/migrations/<UTC timestamp>_create_internal_users.sql`
- Create: `backend/internal/users/migration_contract_test.go`
- Create: `backend/ent/entc.go`
- Create: `backend/ent/schema/internaluser.go`
- Create/generated: `backend/ent/**`
- Modify: `backend/go.mod`, `backend/go.sum`

**Interfaces:**
- Consumes: the approved SQL contract `{ id uuid, clerk_subject text unique, created_at timestamptz, updated_at timestamptz }`.
- Produces: Ent type `InternalUser` with UUID ID, immutable unique subject, immutable creation time, and UTC update time.

- [ ] **Step 1: Write the RED migration contract**

Create `backend/internal/users/migration_contract_test.go` with an assertion that scans migration SQL and requires each of:

```go
func TestCreateInternalUsersMigrationContract(t *testing.T) {
    requireSQL(t, "create table public.internal_users")
    requireSQL(t, "id uuid primary key")
    requireSQL(t, "clerk_subject text not null unique")
    requireSQL(t, "created_at timestamptz not null")
    requireSQL(t, "updated_at timestamptz not null")
    requireSQL(t, "char_length(clerk_subject) between 1 and 255")
}
```

Run:

```bash
cd backend && go test ./internal/users -run TestCreateInternalUsersMigrationContract -count=1
```

Expected: failure because the migration does not exist.

- [ ] **Step 2: Add the minimal forward-only migration**

Create exactly one timestamped `supabase/migrations/*_create_internal_users.sql` containing:

```sql
create extension if not exists pgcrypto;

create table public.internal_users (
  id uuid primary key,
  clerk_subject text not null unique,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint internal_users_clerk_subject_nonempty
    check (char_length(clerk_subject) between 1 and 255)
);

create index internal_users_created_at_idx
  on public.internal_users (created_at);
```

Create Supabase config only through `supabase init` or an equivalent generated project config. Do not create a PostgreSQL service in `compose.yaml`.

- [ ] **Step 3: Verify GREEN for the migration contract**

Run:

```bash
cd backend && go test ./internal/users -run TestCreateInternalUsersMigrationContract -count=1
```

Expected: pass.

- [ ] **Step 4: Write the RED Ent schema contract**

Create `backend/ent/schema/internaluser_test.go` with static/schema assertions covering the exact `InternalUser` fields and no profile/email fields. Run:

```bash
cd backend && go test ./ent/schema -run TestInternalUserSchemaContract -count=1
```

Expected: failure because no Ent schema exists.

- [ ] **Step 5: Add Ent dependencies, schema, and generator**

Pin direct modules in `backend/go.mod`:

```text
entgo.io/ent v0.14.6
github.com/google/uuid
github.com/jackc/pgx/v5 v5.10.0
```

Create `backend/ent/entc.go` with a `go:generate` directive that invokes Ent generation against `./schema`. Implement `InternalUser` so `id` uses `uuid.New`, `clerk_subject` is nonempty, max-255, unique, and immutable, `created_at` is immutable, and `updated_at` uses UTC time. Generate code with:

```bash
cd backend && go generate ./ent
```

- [ ] **Step 6: Verify schema and generated output**

Run:

```bash
cd backend && go test ./ent/schema -run TestInternalUserSchemaContract -count=1
go test ./...
go vet ./...
git diff --check
```

Expected: all pass. Inspect generated names to confirm `internal_users`, `clerk_subject`, `created_at`, and `updated_at` match the migration.

- [ ] **Step 7: Checkpoint**

Do not commit. Run a changed/untracked secret scan; report only paths and detection counts.

### Task 3: Implement provider-neutral, race-safe reconciliation

**Objective:** Reconcile a verified subject to a stable internal user without importing Clerk concepts into domain logic.

**Files:**
- Create: `backend/internal/users/model.go`
- Create: `backend/internal/users/repository.go`
- Create: `backend/internal/users/service.go`
- Create: `backend/internal/users/ent_repository.go`
- Create: `backend/internal/users/service_test.go`
- Create: `backend/internal/users/ent_repository_test.go`

**Interfaces:**

```go
type VerifiedIdentity struct { Subject string }

type InternalUser struct {
    ID        uuid.UUID
    CreatedAt time.Time
}

type Repository interface {
    FindBySubject(context.Context, string) (InternalUser, bool, error)
    Create(context.Context, string) (InternalUser, error)
}

type Service interface {
    Reconcile(context.Context, VerifiedIdentity) (InternalUser, bool, error)
}
```

- [ ] **Step 1: Write RED service cases**

In `service_test.go`, use a recording fake repository and require:

```go
func TestReconcileRejectsEmptySubjectBeforeRepositoryAccess(t *testing.T)
func TestReconcileCreatesFirstMapping(t *testing.T)
func TestReconcileReturnsExistingMappingWithoutMutation(t *testing.T)
func TestReconcileReloadsWinnerAfterUniqueConflict(t *testing.T)
func TestReconcileReturnsControlledFailureForRepositoryError(t *testing.T)
```

Run:

```bash
cd backend && go test ./internal/users -run '^TestReconcile' -count=1
```

Expected: failure because the service does not exist.

- [ ] **Step 2: Implement the minimal service**

Add only:

```go
func (s service) Reconcile(ctx context.Context, identity VerifiedIdentity) (InternalUser, bool, error)
```

Validation occurs before `FindBySubject`. It returns an existing row with `false`; it creates a missing row with `true`; it handles only the repository’s typed unique-conflict error by finding the exact same subject once more and returning the winner with `false`.

- [ ] **Step 3: Verify GREEN**

Run the focused service command from Step 1. Expected: all named cases pass.

- [ ] **Step 4: Write the RED Ent repository integration cases**

Create database-backed tests that require:

```go
func TestEntRepositoryFindsAndCreatesByExactSubject(t *testing.T)
func TestReconcileConcurrentSameSubjectProducesOneStableRow(t *testing.T)
```

The test harness must receive only a disposable `TEST_DATABASE_URL` through its environment and skip with an explicit message when absent. The final mandatory integration invocation must set that URL and must not treat a skipped test as proof.

- [ ] **Step 5: Implement the Ent repository**

Implement exact-subject query/create mapping. Translate only PostgreSQL/Ent uniqueness violations into the domain’s typed unique-conflict error. Do not retry arbitrary database errors and do not query by any display field or timestamp.

- [ ] **Step 6: Verify against a disposable database**

Start only the project-owned local Supabase stack, apply the entire migration ledger, supply the disposable test URL through environment, then run:

```bash
cd backend && go test ./internal/users -run 'TestEntRepository|TestReconcileConcurrent' -count=1
```

Expected: one mapping row, one stable UUID, and no duplicate mapping for a synthetic subject. Stop/reset the local stack afterwards without printing connection details.

### Task 4: Verify bearer identity in Go before reconciliation

**Objective:** Derive `VerifiedIdentity` only from a Clerk-verified session token and keep health publicly reachable.

**Files:**
- Create: `backend/internal/authn/verifier.go`
- Create: `backend/internal/authn/clerk_verifier.go`
- Create: `backend/internal/authn/middleware.go`
- Create: `backend/internal/authn/middleware_test.go`
- Modify: `backend/go.mod`, `backend/go.sum`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/internal/httpapi/health_handler.go`

**Interfaces:**

```go
type Verifier interface {
    Verify(context.Context, string) (users.VerifiedIdentity, error)
}

func RequireVerifiedIdentity(verifier Verifier, next http.Handler) http.Handler
func IdentityFromContext(context.Context) (users.VerifiedIdentity, bool)
```

- [ ] **Step 1: Write RED middleware cases**

Create tests for absent bearer header, malformed scheme, verifier rejection, empty verified subject, and valid subject context propagation. Add an assertion that the downstream handler is not invoked for every invalid case.

Run:

```bash
cd backend && go test ./internal/authn -run TestRequireVerifiedIdentity -count=1
```

Expected: failure because the middleware package is absent.

- [ ] **Step 2: Implement a testable verification boundary**

Add the provider-neutral `Verifier` interface and middleware. The middleware parses only `Authorization: Bearer session-token`, invokes the verifier, rejects invalid outcomes with `401`, and stores a typed identity context value. It does not parse JWT claims itself.

- [ ] **Step 3: Implement and compile-check the Clerk adapter**

Add `github.com/clerk/clerk-sdk-go/v2@v2.7.0`. Use its documented session-verification middleware/API from a dedicated `clerk_verifier.go` adapter, configured only by `CLERK_SECRET_KEY`, optional `CLERK_JWT_KEY`, and exact `CLERK_AUTHORIZED_PARTIES`. Require an actual compile-backed adapter test; do not retain an uncompiled guessed method call. Convert Clerk failures to the local verifier error without embedding SDK diagnostics.

- [ ] **Step 4: Verify GREEN**

Run:

```bash
cd backend && go test ./internal/authn -count=1
go test ./...
go vet ./...
go build ./cmd/api
```

Expected: pass. Health remains public because authentication middleware will be applied only to the reconciliation route in Task 5.

### Task 5: Add the closed reconciliation API contract and handler

**Objective:** Add a body-less, authenticated API that returns only an opaque mapping response.

**Files:**
- Modify: `openapi/juntly-api.v1.yaml`
- Modify/generated: `frontend/src/shared/api/generated/**`
- Create: `backend/internal/httpapi/reconcile_handler.go`
- Create: `backend/internal/httpapi/reconcile_handler_test.go`
- Modify: `backend/internal/httpapi/health_handler.go`

**Interfaces:**

```go
type ReconcileService interface {
    Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

func NewReconcileHandler(service ReconcileService) http.Handler
```

- [ ] **Step 1: Write RED handler cases**

Create tests requiring:

```go
func TestReconcileHandlerReturnsUnauthorizedWithoutVerifiedIdentity(t *testing.T)
func TestReconcileHandlerReturnsOpaqueInternalUser(t *testing.T)
func TestReconcileHandlerReturnsSafeUnavailableForDependencyFailure(t *testing.T)
func TestReconcileHandlerRejectsNonPOSTMethods(t *testing.T)
```

The successful response must include only UUID `id`, RFC3339 `createdAt`, and the correlation header. The failure body must not include the synthetic repository error text.

Run:

```bash
cd backend && go test ./internal/httpapi -run TestReconcileHandler -count=1
```

Expected: failure because the handler does not exist.

- [ ] **Step 2: Add the OpenAPI contract**

Add `POST /api/v1/auth/reconcile` with no request body, `clerkSession` HTTP bearer security, and exact responses:

```yaml
200: InternalUserResponse
401: ErrorResponse with code UNAUTHORIZED
503: ErrorResponse with code SERVICE_UNAVAILABLE
```

`InternalUserResponse` has only `id: uuid` and `createdAt: date-time`. All write schemas remain closed; no caller-controlled identity field exists.

- [ ] **Step 3: Implement handler and route composition**

Implement `NewReconcileHandler` using `IdentityFromContext`, `ReconcileService`, and the existing correlation-header helper. Extend `NewRouter` to register the public health handler unchanged and wrap only `/api/v1/auth/reconcile` with `RequireVerifiedIdentity`.

- [ ] **Step 4: Regenerate and verify GREEN**

Run:

```bash
npm --prefix frontend run codegen
npm --prefix frontend run codegen:check
cd backend && go test ./internal/httpapi -run TestReconcileHandler -count=1
go test ./...
```

Expected: generated client includes the reconciliation endpoint and all focused/full Go tests pass.

### Task 6: Add same-origin, Clerk-aware BFF reconciliation

**Objective:** Reconcile an authenticated session without revealing Go topology or permitting browser-provided identity claims.

**Files:**
- Create: `frontend/src/app/api/v1/auth/reconcile/route.ts`
- Create: `frontend/src/app/api/v1/auth/reconcile/route.test.ts`
- Modify: `frontend/.env.example`
- Modify: `frontend/README.md`
- Modify: `compose.yaml`

**Interfaces:**

```ts
export async function POST(request: Request): Promise<Response>
```

The route consumes `await auth()` from `@clerk/nextjs/server`, a request-local token from `getToken()`, `process.env.JUNTLY_API_ORIGIN`, and the generated OpenAPI SDK. It produces only `InternalUserResponse` or the documented safe error envelope.

- [ ] **Step 1: Write RED BFF cases**

Mock Clerk server auth and upstream fetch. Require:

```ts
it("returns 401 without an authenticated Clerk session", async () => {})
it("returns 401 when the authenticated session has no current token", async () => {})
it("forwards only a server-obtained bearer token and correlation ID", async () => {})
it("maps upstream failure to a topology-safe 503", async () => {})
it("maps malformed upstream response to a topology-safe 503", async () => {})
```

The success case asserts that request headers include `Authorization: Bearer session-token` and `X-Request-ID`, while the browser response body excludes both the bearer token and backend origin.

Run:

```bash
npm --prefix frontend test -- src/app/api/v1/auth/reconcile/route.test.ts
```

Expected: failure because the route is absent.

- [ ] **Step 2: Implement the smallest BFF**

Call `await auth()`. Return `401` before upstream access when not authenticated or no token is available. Forward a current server-obtained token and the validated/generated request ID only to `JUNTLY_API_ORIGIN`. Validate the generated client result structurally before returning its two response fields. Convert all upstream/configuration/validation exceptions to the documented generic `503` without leakable cause text.

- [ ] **Step 3: Add only safe configuration documentation**

Add variable names/placeholders—never values—to `.env.example`, `compose.yaml`, and the frontend README. Document `localhost:4200` as the local browser authority and `JUNTLY_API_ORIGIN` as server-only.

- [ ] **Step 4: Verify GREEN**

Run:

```bash
npm --prefix frontend test -- src/app/api/v1/auth/reconcile/route.test.ts
npm --prefix frontend run verify
docker compose config
```

Expected: all pass with no public backend-origin value in generated browser assets or response fixtures.

### Task 7: Perform durable, runtime, and contract verification

**Objective:** Produce evidence for database durability, API safety, and local authenticated behavior without fabricating a user or exposing credentials.

**Files:**
- Modify only if evidence changes instructions: `frontend/README.md`, `context/architecture.md`, `context/stack.md`
- Verify: all Task 1–6 paths, plus untracked documentation/spec/plan files.

- [ ] **Step 1: Run migration-chain and concurrent integration proof**

Use a disposable local Supabase project/database only. Apply all migrations in order. Run the Task 3 integration test with `TEST_DATABASE_URL` injected only through the process environment. Capture only pass/fail, synthetic row count, UUID equality, and timestamp equality; do not print connection strings or database credentials.

- [ ] **Step 2: Run all local quality gates**

Run:

```bash
codegraph sync .
codegraph affected .
cd backend && go test ./... && go vet ./... && go build ./cmd/api
cd .. && npm --prefix frontend run verify
docker compose config
git diff --check
```

Expected: all commands pass. Any untracked source file is inspected separately because `git diff --check` does not cover it.

- [ ] **Step 3: Run a secret-safe changed-file review**

Inspect staged, modified, and untracked source/docs for credential-like values. Report only path names and detection counts. Confirm ignored local config remains ignored with `git check-ignore` using path names only.

- [ ] **Step 4: Run the optional real-session smoke only when authorized**

With an approved development Clerk test user, start the local API/frontend at canonical `localhost:4200`, establish a real session, and call same-origin `POST /api/v1/auth/reconcile` twice. Evidence is limited to statuses, request IDs, stable opaque-ID equality, stable creation-time equality, and synthetic-safe row count. Do not record email, subject, cookie, bearer, URL, or secret.

If no approved test user/session is available, report this gate as blocked; do not claim it from unit tests.

- [ ] **Step 5: Final review and delivery boundary**

Create a requirement-to-code-to-test matrix against `docs/superpowers/specs/2026-08-21-durable-internal-user-mapping-design.md`. Re-run `git status --short --branch`, `git diff --check`, and inspect every untracked mapping file. Do not commit, push, merge, deploy, or update Kanban terminal status without separate user authorization.
