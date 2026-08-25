# Taxonomy, Locations, and Provider Profiles Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver database-configurable service taxonomy, source-verifiable Portugal launch locations with PostGIS radius queries, and owner-only provider profiles through Juntly's authenticated full-stack authority chain.

**Architecture:** Reference data is seeded by forward migrations from a deterministic reviewed manifest. Ent schemas model taxonomy, locales, languages, localities, provider profiles, and composite edge schemas; the locality table also has a SQL-generated PostGIS geography point used by a parameterized repository. Go owns provider capability, ownership, validation, transactions, and strict OpenAPI transports; Next.js provides same-origin BFF routes and a localized owner-only onboarding form.

**Tech Stack:** Python 3.12 stdlib source verifier, DGT CAOP 2025 GeoPackage, OpenStreetMap/Nominatim one-time source verification, Go 1.26, Ent 0.14.6, pgx/PostgreSQL/PostGIS, Supabase migrations, OpenAPI 3.1, generated TypeScript, Next.js 16.3.1, Clerk, Vitest, Testing Library.

## Global Constraints

- Taxonomy administration APIs/UI are deferred to Slice 11; Slice 2 taxonomy is database-configurable and read-only through product APIs.
- Provider profiles remain owner-only; no public provider DTO/page/metadata exists until Slice 4.
- A caller must have `providerEnabled=true`; customer capability, paid status, or Clerk metadata never grants provider/admin authority.
- Persist no email, phone, WhatsApp, exact provider address/coordinates, identity documents, verification claims, payment data, Clerk subject/token/session, or internal-user UUID in public/profile responses.
- Browser calls same-origin `/api/v1/...`; `JUNTLY_API_ORIGIN` stays server-only.
- All request/response objects are closed. Unknown properties, explicit nulls, duplicate IDs/codes, unsupported references, invalid bounds, and trailing JSON fail before service mutation.
- Migrations are forward-only and API startup never migrates schema.
- pt-PT is the default locale; English ships; Spanish has exact message-key parity.
- Real PostgreSQL/PostGIS tests use synthetic isolated profile rows and the canonical migration ledger; they must not skip during acceptance.
- Reference data may contain only reviewed public administrative/place facts and attribution. Downloaded CAOP/Nominatim response artifacts remain temporary and are deleted.
- Each behavioral unit follows observed RED → GREEN → REFACTOR, focused/full gates, a reviewed staged allowlist, and a local commit. No push or remote mutation.

---

## Frozen launch reference values

### CAOP source

```text
URL: https://geo2.dgterritorio.gov.pt/caop/CAOP_Continente_2025-gpkg.zip
SHA-256: 87cd67f4b1fbadf23d9324e6fb231ff05531e4db347af36ccc7c6cbabe3ecd1d
Bytes: 111647845
Layers: cont_distritos, cont_municipios, cont_freguesias
```

Administrative rows:

```text
country: PT — Portugal
district: 05 — Castelo Branco
municipality: 0502 — Castelo Branco
municipality: 0505 — Idanha-a-Nova
parish: 050205 — Castelo Branco
parish: 050510 — Penha Garcia
parish: 050518 — União das freguesias de Idanha-a-Nova e Alcafozes
parish: 050520 — União das freguesias de Monsanto e Idanha-a-Velha
parish: 050521 — União das freguesias de Zebreira e Segura
```

### OSM locality centers

```text
castelo-branco | relation 5396187 | 39.8266322, -7.4919318 | parent 050205
idanha-a-nova  | relation 5395738 | 39.9260883, -7.2436356 | parent 050518
zebreira       | node 440173641    | 39.8455920, -7.0703366 | parent 050521
penha-garcia   | relation 5431477 | 40.0422569, -7.0163521 | parent 050510
monsanto       | node 371426674    | 40.0387510, -7.1151133 | parent 050520
```

The manifest records `source=OpenStreetMap`, source type/ID, retrieval date `2026-08-23`, ODbL URL, and attribution `© OpenStreetMap contributors`.

### Initial taxonomy

Top-level slugs and child slugs:

```text
home-repairs: plumbing, electrical-work, construction, small-repairs
home-and-garden: cleaning, gardening
rural-and-transport: agricultural-assistance, transport
care-and-learning: elderly-assistance, animal-care, private-lessons
food-and-technology: meal-preparation, computer-repair
```

Supported locale/language codes are `pt-PT`, `en`, and `es`.

## File map

### Reference tooling/data

- `reference/portugal/launch-area-2025.json` — deterministic reviewed source manifest.
- `scripts/build_launch_reference.py` — CAOP checksum/SQLite extraction plus policy-compliant Nominatim resolver.
- `scripts/tests/test_build_launch_reference.py` — offline synthetic archive/geocoder contract tests.

### Persistence

- `supabase/migrations/<timestamp>_create_taxonomy_locations_provider_profiles.sql` — reference/profile tables, PostGIS center, constraints, seeds.
- `backend/ent/schema/supportedlocale.go`
- `backend/ent/schema/servicecategory.go`
- `backend/ent/schema/servicecategorytranslation.go` — composite edge schema.
- `backend/ent/schema/spokenlanguage.go`
- `backend/ent/schema/spokenlanguagetranslation.go` — composite edge schema.
- `backend/ent/schema/administrativearea.go`
- `backend/ent/schema/locality.go`
- `backend/ent/schema/providerprofile.go`
- `backend/ent/schema/providerservicelocality.go` — composite edge schema.
- `backend/ent/schema/providerspokenlanguage.go` — composite edge schema.
- `backend/ent/schema/*_test.go` — structural privacy/constraint contracts.
- `backend/ent/**` — generated Ent output.
- `backend/internal/users/migration_contract_test.go` — complete SQL migration/seed contract.

### Go modules

- `backend/internal/reference/model.go`
- `backend/internal/reference/repository.go`
- `backend/internal/reference/sql_repository.go`
- `backend/internal/reference/service.go`
- `backend/internal/reference/*_test.go`
- `backend/internal/provideraccess/service.go`
- `backend/internal/provideraccess/service_test.go`
- `backend/internal/providers/model.go`
- `backend/internal/providers/repository.go`
- `backend/internal/providers/ent_repository.go`
- `backend/internal/providers/service.go`
- `backend/internal/providers/*_test.go`
- `backend/internal/httpapi/reference_handler.go`
- `backend/internal/httpapi/provider_profile_handler.go`
- related handler/router/OpenAPI tests
- `backend/cmd/api/main.go` — dependency composition using the existing DB/Ent client.

### Contract/BFF/UI

- `openapi/juntly-api.v1.yaml`
- `frontend/src/shared/api/generated/**`
- `frontend/src/app/api/v1/catalog/categories/route.ts`
- `frontend/src/app/api/v1/reference/localities/route.ts`
- `frontend/src/app/api/v1/reference/languages/route.ts`
- `frontend/src/app/api/v1/me/provider-profile/route.ts`
- matching route tests
- `frontend/src/features/provider/provider-profile-form.tsx`
- `frontend/src/features/provider/provider-profile-form.test.tsx`
- `frontend/src/app/[locale]/account/provider-profile/page.tsx`
- `frontend/src/app/[locale]/account/provider-profile/page.test.tsx`
- `frontend/src/features/account/account-capabilities-card.tsx` and test
- `frontend/messages/{pt-PT,en,es}.json`
- `frontend/src/i18n/messages.test.ts`

## Task 1: Freeze and verify launch reference data

**Files:**
- Create: `reference/portugal/launch-area-2025.json`
- Create: `scripts/build_launch_reference.py`
- Create: `scripts/tests/test_build_launch_reference.py`
- Modify: `.gitignore` only if a precise source-tool cache path needs exclusion; do not ignore broad data directories.

**Interfaces:**

```python
def verify_caop_archive(path: Path, expected_sha256: str) -> None: ...
def extract_administrative_rows(path: Path, required: tuple[RequiredArea, ...]) -> list[AdministrativeArea]: ...
def resolve_localities(resolve: Callable[[str], list[dict[str, object]]]) -> list[Locality]: ...
def build_manifest(caop_zip: Path, resolve: Callable[[str], list[dict[str, object]]]) -> dict[str, object]: ...
```

- [ ] **Step 1: Write offline RED tests.**

Create a synthetic ZIP containing a SQLite GeoPackage-like database with `cont_distritos`, `cont_municipios`, and `cont_freguesias`. Assert checksum mismatch rejection, missing/duplicate exact row rejection, required code/name/parent extraction, deterministic ordering, exact five OSM element selection, request spacing hook invocation, and manifest output equality. Assert the manifest contains no postal code or source response dump.

- [ ] **Step 2: Run RED.**

```bash
python -m unittest scripts.tests.test_build_launch_reference -v
```

Expected: import failure because `scripts/build_launch_reference.py` is absent.

- [ ] **Step 3: Implement the stdlib verifier.**

Use `urllib.request`, `hashlib`, `zipfile`, `sqlite3`, `tempfile`, and `shutil`. The network CLI uses an identifying `User-Agent`, serial Nominatim requests with at least 1.05 seconds between calls, `limit=3`, `countrycodes=pt`, and exact allowlisted OSM type/ID selection from the frozen values. Always delete the temporary root in `finally`.

- [ ] **Step 4: Write the deterministic manifest.**

The checked-in JSON contains the frozen CAOP/OSM values above, source/license metadata, locale/category/language seeds, and no unreviewed candidates.

- [ ] **Step 5: Run GREEN and reproducibility checks.**

```bash
python -m unittest scripts.tests.test_build_launch_reference -v
python scripts/build_launch_reference.py --verify reference/portugal/launch-area-2025.json
```

The first is offline. The second is a deliberate one-time network verification; it compares bytes to the checked-in manifest and removes temporary files.

- [ ] **Step 6: Commit.**

```bash
git add reference/portugal/launch-area-2025.json scripts/build_launch_reference.py scripts/tests/test_build_launch_reference.py .gitignore
git commit -m "chore: verify launch reference data"
```

## Task 2: Add taxonomy, location, language, and provider-profile schema

**Files:**
- Create the migration and ten Ent schema files listed in the file map.
- Create focused schema tests.
- Modify `backend/internal/users/migration_contract_test.go`.
- Regenerate `backend/ent/**` and `backend/go.sum` if the generator changes it.

**Interfaces and schema rules:**

- `SupportedLocale`: string ID, active, sort order.
- `ServiceCategory`: UUID ID, optional parent UUID, stable slug, active, sort order, timestamps.
- `ServiceCategoryTranslation`: edge schema between category and locale; `field.ID("category_id", "locale")`; name/optional description.
- `SpokenLanguage`: string ID, active, sort order.
- `SpokenLanguageTranslation`: edge schema between language and locale; composite ID.
- `AdministrativeArea`: UUID ID, source/version/external code/kind/name/optional parent/active/timestamps.
- `Locality`: UUID ID, slug/name/parent parish/source/type/source ID/latitude/longitude/active/timestamps. SQL adds generated stored `center geography(Point,4326)` from longitude/latitude.
- `ProviderProfile`: internal-user UUID ID stored as `internal_user_id`; provider type, display name, bio, primary locality ID, radius, three service mode booleans, timestamps.
- `ProviderServiceLocality`: edge schema with composite `(internal_user_id, locality_id)`.
- `ProviderSpokenLanguage`: edge schema with composite `(internal_user_id, language_code)`.

- [ ] **Step 1: Write RED schema/migration tests.**

Assert every table, FK/delete action, composite key, unique slug/code, hierarchy kind check, coordinate/radius/name/bio bounds, at-least-one-service-mode check, generated PostGIS center expression, seed counts/slugs/translations, exact CAOP codes, exact five locality source IDs, and prohibited profile/contact/identity columns.

- [ ] **Step 2: Run RED.**

```bash
cd backend
go test ./ent/schema ./internal/users -run 'Test.*(Taxonomy|Location|Language|ProviderProfile|Migration)' -count=1
```

Expected: missing schema/migration contracts.

- [ ] **Step 3: Add one forward migration.**

Enable PostGIS if absent; create tables/constraints/indexes; insert deterministic UUIDs and manifest values. The API runtime does not execute this migration.

- [ ] **Step 4: Implement Ent schemas and composite edge schemas.**

Use installed Ent `field.ID` on edge schemas. Add inverse/through edges only where needed by typed provider transactions. Do not model the generated `center` column as mutable Ent state.

- [ ] **Step 5: Generate and run GREEN.**

```bash
cd backend
go generate ./ent
go test ./ent/schema ./internal/users -run 'Test.*(Taxonomy|Location|Language|ProviderProfile|Migration)' -count=1
```

- [ ] **Step 6: Apply the complete migration chain to local Supabase.**

```bash
supabase migration up --local
```

Verify the migration ledger and seed counts through process-only DB configuration; do not print the URL.

- [ ] **Step 7: Commit.**

```bash
git add supabase/migrations backend/ent backend/internal/users/migration_contract_test.go backend/go.sum
git commit -m "feat: add provider marketplace reference schema"
```

## Task 3: Implement reference catalog and PostGIS radius service

**Files:**
- Create `backend/internal/reference/{model,repository,sql_repository,service}.go` and tests.

**Interfaces:**

```go
type Repository interface {
    Categories(context.Context, string) ([]Category, error)
    Languages(context.Context, string) ([]Language, error)
    Localities(context.Context, string) ([]Locality, error)
    NearbyLocalities(context.Context, uuid.UUID, int, string) ([]LocalityDistance, error)
    ValidateProfileReferences(context.Context, ProfileReferences) error
}

type Service interface {
    Categories(context.Context, string) ([]Category, error)
    Languages(context.Context, string) ([]Language, error)
    Localities(context.Context, string) ([]Locality, error)
    NearbyLocalities(context.Context, uuid.UUID, int, string) ([]LocalityDistance, error)
}
```

Public DTOs contain UUID/slug/localized names/hierarchy labels and optional integer distance metres. They contain no raw coordinates or provider rows.

- [ ] **Step 1: Write RED service tests.**

Cover supported/unsupported locales, inactive omission, stable sort order, missing origin, radius bounds 1–200, and controlled unavailability.

- [ ] **Step 2: Write RED real-PostgreSQL tests.**

Use at least three synthetic localities: two in range and one out of range. Assert boundary inclusion, ascending distance, UUID tie-break, inactive omission, no raw coordinate fields, and multi-parent category translation correctness.

- [ ] **Step 3: Run RED.**

```bash
cd backend
go test ./internal/reference -count=1
```

Expected: missing package symbols.

- [ ] **Step 4: Implement parameterized SQL repository.**

Reuse the existing `*sql.DB`. Use fixed SQL identifiers and bound values only. Radius SQL uses `ST_DWithin`, `ST_Distance`, active predicates, and `ORDER BY distance_meters, id`. Category queries never apply a global child limit; each category/subcategory contributes once.

- [ ] **Step 5: Run unit and non-skipping PostgreSQL GREEN.**

```bash
cd backend
go test ./internal/reference -count=1
TEST_DATABASE_URL="$DB_URL" go test ./internal/reference -run 'TestSQLRepository|TestRadius' -count=1
```

- [ ] **Step 6: Commit.**

```bash
git add backend/internal/reference
git commit -m "feat: add marketplace reference catalog"
```

## Task 4: Implement provider capability authorization and transactional profiles

**Files:**
- Create `backend/internal/provideraccess/service.go` and test.
- Create `backend/internal/providers/{model,repository,ent_repository,service}.go` and tests.

**Interfaces:**

```go
type InternalUserReconciler interface {
    Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type CapabilityReader interface {
    Get(context.Context, users.VerifiedIdentity) (accounts.Account, error)
}

type ProviderAuthorizer interface {
    RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}

type Service interface {
    Get(context.Context, users.VerifiedIdentity) (*Profile, error)
    Put(context.Context, users.VerifiedIdentity, ReplaceProfile) (Profile, error)
}
```

`ReplaceProfile` contains only approved profile fields and reference IDs/codes. Repository methods accept the already-authorized internal user UUID; no transport can provide it.

- [ ] **Step 1: Write provider-access RED tests.**

Invalid identity, unavailable account, provider-disabled, and provider-enabled cases; prove no repository/profile call occurs before authorization.

- [ ] **Step 2: Write provider service RED tests.**

Cover every scalar bound, trim behavior, provider enum, duplicate/size-constrained locality/language lists, primary-locality membership, at-least-one mode, zero-radius rule, inactive/missing references, owner-only lookup, missing profile, create/update, and controlled failures.

- [ ] **Step 3: Write transaction/concurrency RED tests.**

Against PostgreSQL, prove full replacement commits scalars plus exact child sets, invalid child rolls back all changes, repeat is idempotent, concurrent first PUT yields one owner profile, and one user's ID cannot retrieve/update another's profile.

- [ ] **Step 4: Implement minimal services and Ent repository.**

Use one Ent transaction for profile and edge-schema replacement. Reload canonical state before commit; normalize timestamps to UTC microseconds. Retry only the documented unique-owner conflict by loading the winner; do not retry validation/FK errors.

- [ ] **Step 5: Run GREEN.**

```bash
cd backend
go test ./internal/provideraccess ./internal/providers -count=1
TEST_DATABASE_URL="$DB_URL" go test -race ./internal/providers -run 'TestEntRepository|TestConcurrent|TestTransaction' -count=1
```

- [ ] **Step 6: Commit.**

```bash
git add backend/internal/provideraccess backend/internal/providers
git commit -m "feat: add owner-only provider profiles"
```

## Task 5: Publish Go/OpenAPI reference and profile contracts

**Files:**
- Modify `openapi/juntly-api.v1.yaml`.
- Create reference/profile handlers and tests.
- Modify router/OpenAPI contract tests and `backend/cmd/api/main.go`.
- Regenerate `frontend/src/shared/api/generated/**`.

**Operations:**

```text
GET /api/v1/catalog/categories
GET /api/v1/reference/localities
GET /api/v1/reference/languages
GET /api/v1/me/provider-profile
PUT /api/v1/me/provider-profile
```

Reference operations accept locale; localities optionally accept paired `nearLocalityId` and `radiusKm`. Profile GET returns `{ "profile": null }` before onboarding or a closed owner-only profile. PUT accepts the exact spec request.

- [ ] **Step 1: Write RED OpenAPI/handler/router tests.**

Cover strict query/body parsing, locale/radius pairing, null/unknown/duplicate/bounds cases, no service call on invalid input, public reference routes, protected profile routes, `401`, `403 FORBIDDEN`, `400`, `503`, correlation parity, and output privacy.

- [ ] **Step 2: Run RED.**

```bash
cd backend
go test ./internal/httpapi -run 'Test.*(Category|Locality|Language|ProviderProfile|OpenAPI|Router)' -count=1
```

- [ ] **Step 3: Implement handlers and dependency composition.**

Use `url.Values` exact allowlisting for reference queries and `json.Decoder.DisallowUnknownFields()` plus second-value EOF checks for profile PUT. Add only `INVALID_REQUEST`, `FORBIDDEN`, and existing safe codes.

- [ ] **Step 4: Extend OpenAPI closed schemas and regenerate.**

```bash
cd frontend
npm run codegen
npm run codegen:check
```

- [ ] **Step 5: Run GREEN and cross-layer compile.**

```bash
cd backend && go test ./internal/httpapi ./cmd/api -count=1
cd ../frontend && npm run typecheck
```

- [ ] **Step 6: Commit.**

```bash
git add openapi backend/internal/httpapi backend/cmd/api frontend/src/shared/api/generated
git commit -m "feat: add provider reference and profile API"
```

## Task 6: Add same-origin BFF routes

**Files:**
- Create four BFF route files and matching tests listed in the file map.

- [ ] **Step 1: Write RED reference BFF tests.**

Prove locale/query allowlisting, generated-client invocation, exact public response parsing, correlation parity, attribution preservation, and topology-safe failure.

- [ ] **Step 2: Write RED provider BFF tests.**

Prove signed-out/no-token `401`, server-obtained bearer only, exact PUT object/arrays, null/unknown/duplicate/bound rejection before upstream, preserved `403`, and generic upstream/malformed/correlation `503` without origin/token/private data.

- [ ] **Step 3: Run RED.**

```bash
cd frontend
npm test -- src/app/api/v1/catalog/categories/route.test.ts src/app/api/v1/reference/localities/route.test.ts src/app/api/v1/reference/languages/route.test.ts src/app/api/v1/me/provider-profile/route.test.ts
```

- [ ] **Step 4: Implement thin generated-client BFF routes.**

No reference route imports Clerk. Profile routes use `await auth()` and `getToken()` only. Extract shared request-ID/error helpers only after all focused tests are green and duplication is exact.

- [ ] **Step 5: Run GREEN and full typecheck.**

```bash
cd frontend
npm test -- src/app/api/v1/catalog/categories/route.test.ts src/app/api/v1/reference/localities/route.test.ts src/app/api/v1/reference/languages/route.test.ts src/app/api/v1/me/provider-profile/route.test.ts
npm run typecheck
```

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/app/api/v1/catalog frontend/src/app/api/v1/reference frontend/src/app/api/v1/me/provider-profile
git commit -m "feat: add provider profile BFF"
```

## Task 7: Add localized owner-only provider onboarding UI

**Files:**
- Create provider form/page/tests.
- Modify account capability card/test to expose the onboarding link only when enabled.
- Modify three locale JSON files and message test.

- [ ] **Step 1: Write RED page/form/i18n tests.**

Cover protected dynamic page, provider-disabled `403`/safe state, reference loading, empty profile, all fields, accessible grouped controls, primary locality inclusion, duplicate prevention, client validation, saving lock, stale completion rejection, successful create/update, controlled error/retry, no identity/contact/coordinates in DOM, OSM attribution, pt-PT/en/es key parity, and 44×44px control classes.

- [ ] **Step 2: Run RED.**

```bash
cd frontend
npm test -- src/features/provider/provider-profile-form.test.tsx src/app/'[locale]'/account/provider-profile/page.test.tsx src/features/account/account-capabilities-card.test.tsx src/i18n/messages.test.ts
```

- [ ] **Step 3: Implement minimal UI.**

Use only same-origin BFF fetches. Keep public source options in local state and profile mutation generation separate. Disable all dismissal/submit/toggle surfaces during save; apply a completion only when its generation still owns the rendered profile.

- [ ] **Step 4: Run GREEN, lint, and typecheck.**

```bash
cd frontend
npm test -- src/features/provider/provider-profile-form.test.tsx src/app/'[locale]'/account/provider-profile/page.test.tsx src/features/account/account-capabilities-card.test.tsx src/i18n/messages.test.ts
npm run lint
npm run typecheck
```

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/features/provider frontend/src/features/account/account-capabilities-card* frontend/src/app/'[locale]'/account/provider-profile frontend/messages frontend/src/i18n/messages.test.ts
git commit -m "feat: add provider profile onboarding"
```

## Task 8: Complete Slice 2 acceptance and delivery

- [ ] **Step 1: Reapply the complete local migration ledger.**

Start/reuse only Juntly's isolated Supabase stack and apply migrations. Verify reference seed counts/source IDs without printing DB configuration.

- [ ] **Step 2: Run non-skipping database/race gates.**

```bash
TEST_DATABASE_URL="$DB_URL" go test -race ./internal/reference ./internal/providers -count=1
```

- [ ] **Step 3: Run complete backend gates.**

```bash
cd backend
go test ./...
go vet ./...
go build -o "$LOCALAPPDATA/Temp/juntly-api-slice2.exe" ./cmd/api
```

- [ ] **Step 4: Run complete frontend/topology gates.**

```bash
cd frontend && npm run verify
cd .. && docker compose config >/dev/null && git diff --check && git diff --cached --check
```

- [ ] **Step 5: Run live public reference proof.**

Prove category/language/locality responses and radius ordering through same-origin BFF. Record only status, correlation parity, counts/slugs/source attribution, in-range ordering, and absence of coordinates/private fields.

- [ ] **Step 6: Run real authenticated provider journey.**

With the approved Clerk development session and process-only local clock skew only if Windows remains unsynchronized:

```text
provider capability enabled
→ empty provider profile GET
→ strict PUT create
→ GET same owner state
→ PUT replacement
→ GET replaced state
→ fresh page render
```

Record only status/correlation/schema, provider type, service-mode booleans, radius, language/service-area counts, stable owner/profile timestamp properties, and no raw identity/contact/coordinate values.

- [ ] **Step 7: Freeze and review.**

Synchronize CodeGraph; stage the explicit Slice 2 allowlist; run fixture-aware secret/privacy scans; record binary digest/path count. Request one independent read-only contract/security review if the Codex environment is available. If unavailable for the already documented local cache/MCP reason, record that honestly and perform the repository-required direct review without claiming independence.

- [ ] **Step 8: Commit final acceptance correction only if needed.**

Any runtime-found defect gets a new RED test, minimal fix, affected/full gates, new digest, and local commit. Otherwise make no empty commit.

- [ ] **Step 9: Cleanup and close.**

Stop tracked frontend/API processes, remove temporary binaries/source downloads, verify project proof ports closed, preserve the Juntly Supabase stack only if immediately needed for the next slice, and require a clean worktree before marking Slice 2 complete.
