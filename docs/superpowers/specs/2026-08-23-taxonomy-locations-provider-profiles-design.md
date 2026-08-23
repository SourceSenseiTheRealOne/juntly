# Taxonomy, Locations, and Provider Profiles Design

**Status:** Approved by SourceSensei on 2026-08-23

## Goal

Deliver Juntly Slice 2 as a privacy-safe supply foundation: database-configurable service taxonomy, a source-verifiable Portugal launch-area location hierarchy with PostGIS radius behavior, and owner-only provider profiles managed through the existing authenticated BFF → Go → PostgreSQL authority chain.

## Scope decisions

1. Taxonomy is database-configurable and product-readable in Slice 2. Taxonomy administration APIs and UI are deferred to Slice 11, where platform-admin authority is introduced.
2. Location reference data includes a source-verified launch-area subset rather than a schema-only delivery or full Portugal import.
3. Provider profiles remain owner-only until Slice 4 defines and tests a separate public projection.
4. No temporary Clerk metadata, operator secret, paid entitlement, or provider capability is treated as platform-admin authority.

## Authority and request path

```text
browser
→ localized Next.js UI
→ same-origin Next.js BFF
→ generated OpenAPI TypeScript client
→ Go REST transport
→ taxonomy / locations / providers application services
→ Ent and reviewed parameterized PostGIS repositories
→ project-owned Supabase PostgreSQL/PostGIS
```

Clerk remains identity authority. Go resolves the verified Clerk subject to the opaque internal user, checks `providerEnabled`, enforces ownership, validates every referenced category/location/language, and owns transactions. Browser state and hidden controls are presentation only.

## Taxonomy model

### `service_categories`

- `id uuid primary key`
- `parent_id uuid null references service_categories(id)`
- `slug text not null unique`
- `active boolean not null default true`
- `sort_order integer not null`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Rules:

- Only two levels are permitted: category and subcategory.
- A category cannot parent itself.
- Inactive records remain durable but are omitted from new provider/listing choices.
- Slugs are stable lookup identities and are never translated.

### `service_category_translations`

- Composite primary key: `(category_id, locale)`
- `category_id uuid references service_categories(id) on delete cascade`
- `locale text` restricted initially to `pt-PT`, `en`, and `es`
- `name text not null`
- `description text null`

Every seeded active category has all three locale rows. API output uses the requested supported locale and fails closed rather than mixing locales silently.

### Initial category seed

The initial database seed uses product-approved service concepts, not component constants:

- `home-repairs`: plumbing, electrical work, construction, small repairs
- `home-and-garden`: cleaning, gardening
- `rural-and-transport`: agricultural assistance, transport
- `care-and-learning`: elderly assistance, animal care, private lessons
- `food-and-technology`: meal preparation, computer repair

Seed IDs and ordering are deterministic. Later changes are forward migrations until Slice 11 provides server-authorized administration.

## Language reference model

### `spoken_languages`

- `code text primary key` using stable BCP 47 codes
- `active boolean not null default true`
- `sort_order integer not null`

### `spoken_language_translations`

- Composite primary key: `(language_code, locale)`
- Localized display name in `pt-PT`, `en`, and `es`

The initial active values are `pt-PT`, `en`, and `es`. Provider code validates against active database records; UI code does not own the vocabulary.

## Location model

### Administrative hierarchy

`administrative_areas` stores:

- opaque UUID
- `source` (`caop` for the initial hierarchy)
- `source_version` (`2025`)
- stable external administrative code
- `kind`: country, district, municipality, or parish
- official name
- nullable parent UUID
- active flag
- timestamps

The unique identity is `(source, external_code)`. Parent-kind transitions are closed:

```text
country → district → municipality → parish
```

DGT describes CAOP as Portugal's official administrative-boundary record, maintains it, publishes CAOP 2025 as the current version, provides a GeoPackage download, and publishes a PostgreSQL/PostGIS-adapted conceptual model.[1] INE independently lists the relevant Idanha-a-Nova municipal/parish entities, including Penha Garcia, Monsanto and Idanha-a-Velha, and Zebreira and Segura.[2]

### Search localities

`localities` stores:

- opaque UUID
- stable slug
- display name
- parent parish UUID
- PostGIS `geography(Point, 4326)` reference center
- source name
- source element identifier
- source version/retrieval date
- active flag
- timestamps

Initial localities are:

- Castelo Branco
- Idanha-a-Nova
- Zebreira
- Penha Garcia
- Monsanto

The containing administrative hierarchy comes from CAOP 2025. Locality center points are a reviewed one-time OpenStreetMap extraction because villages/towns are place features rather than CAOP administrative boundaries. The import tool uses the public Nominatim service only during development, serially at no more than one request per second, with an identifying User-Agent, and caches the reviewed result; it is never a product runtime dependency.[3] Committed provenance retains OSM type/element ID and the UI/reference API includes OpenStreetMap attribution because OSM data is ODbL-licensed and requires credit.[4]

No downloaded national archive, Nominatim response dump, or generated temporary database is committed. The reviewed launch seed manifest is small, deterministic, source-attributed product reference data—not mock marketplace data.

### Radius query

The locations repository uses a reviewed parameterized SQL query:

```text
ST_DWithin(candidate.center, origin.center, radius_meters)
ORDER BY ST_Distance(candidate.center, origin.center), candidate.id
```

Rules:

- Radius is an integer from 1 to 200 kilometres at public/application boundaries.
- Only active localities participate.
- Origin must be an active locality.
- The UUID tie-breaker makes equal-distance ordering deterministic.
- Real PostgreSQL tests include at least three localities, multiple in-range results, one out-of-range result, boundary behavior, and deterministic ordering.

## Provider profile model

### `provider_profiles`

- `internal_user_id uuid primary key references user_accounts(internal_user_id) on delete cascade`
- `display_name text not null`, trimmed length 2–100
- `provider_type text not null`: individual, professional, or business
- `bio text not null`, trimmed maximum 1,000 characters
- `primary_locality_id uuid references localities(id)`
- `max_travel_distance_km integer not null`, range 0–200
- `travels_to_customer boolean not null`
- `receives_customer boolean not null`
- `remote_services boolean not null`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

At least one service mode is true. A zero travel radius is allowed only when a non-travel mode remains enabled.

### `provider_service_localities`

- Composite primary key: `(internal_user_id, locality_id)`
- Profile FK and locality FK with cascade/restrict behavior appropriate to ownership/reference data
- The list contains 1–20 unique active localities and must contain the primary locality.

### `provider_spoken_languages`

- Composite primary key: `(internal_user_id, language_code)`
- The list contains 1–10 unique active language codes.

### Privacy boundary

Slice 2 does not collect or return:

- email, phone, WhatsApp, or preferred contact method
- exact address or provider-specific coordinates
- identity/business documents
- verification claims
- payment/subscription data
- listings, portfolio media, reviews, response statistics, or availability
- Clerk subject, token/session values, or internal-user UUID

A provider profile has no public route or publication state in Slice 2. Slice 4 must create an explicit allowlisted public DTO and SEO/privacy tests before any provider data becomes public.

## Provider capability and ownership

- A verified internal user must have `providerEnabled=true` before profile GET/PUT is available.
- A provider-disabled account receives `403 FORBIDDEN` and no profile row is created.
- The caller cannot submit an owner/internal-user ID.
- GET and PUT always derive ownership from verified identity.
- Cross-user lookup and mutation are absent from the public contract.
- Concurrent first PUT operations resolve through the unique owner key and one transactional winner.

## Application service and transaction

The provider service consumes:

```go
type AccountAuthorizer interface {
    RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}
```

The service validates the full replacement request before opening a transaction. One transaction:

1. loads or creates the owner profile;
2. replaces scalar profile fields;
3. replaces service-locality links;
4. replaces spoken-language links;
5. rereads and returns the canonical owner-only profile.

Any invalid reference or failed child replacement rolls back the complete mutation. A repeat with identical input is idempotent in durable state. No raw request body is logged or persisted.

## API contract

### Reference reads

```text
GET /api/v1/catalog/categories?locale=pt-PT
GET /api/v1/reference/localities?locale=pt-PT
GET /api/v1/reference/localities?locale=pt-PT&nearLocalityId=<uuid>&radiusKm=<1..200>
GET /api/v1/reference/languages?locale=pt-PT
```

These endpoints expose only active reference rows, stable public UUIDs/slugs, localized names, hierarchy labels, distance where requested, and required OSM attribution metadata. They expose no provider/user records.

### Owner profile

```text
GET /api/v1/me/provider-profile
PUT /api/v1/me/provider-profile
```

GET returns either a closed nullable profile envelope or the current complete owner-only profile. PUT is strict full replacement and accepts exactly:

```json
{
  "displayName": "string",
  "providerType": "individual | professional | business",
  "bio": "string",
  "primaryLocalityId": "uuid",
  "serviceLocalityIds": ["uuid"],
  "maxTravelDistanceKm": 25,
  "travelsToCustomer": true,
  "receivesCustomer": false,
  "remoteServices": false,
  "languageCodes": ["pt-PT"]
}
```

Unknown properties, omitted required fields, explicit nulls, duplicate IDs/codes, invalid bounds, inactive/missing references, primary locality omission, and invalid service-mode combinations return correlated `400 INVALID_REQUEST` without service invocation. Authentication failures return `401`, provider-disabled access returns `403`, and dependency failures return generic `503`.

## Same-origin BFF

The Next.js BFF mirrors the Go/OpenAPI boundaries:

- obtains Clerk session/token server-side;
- never accepts browser ownership/subject/token fields;
- strictly validates PUT before generated-client invocation;
- forwards bearer and request ID only;
- validates exact upstream response keys and correlation parity;
- maps topology/upstream/malformed failures to generic `503`;
- preserves `401`, `403`, and local `400` allowlisted envelopes.

Reference endpoints are same-origin and use generated clients but require no Clerk bearer.

## Localized UI

Protected route:

```text
/:locale/account/provider-profile
```

The account capability page links to it only when provider capability is enabled. The page remains protected server-side and owner-only.

The mobile-first form provides:

- short profile fields;
- database-loaded provider type copy, languages, primary locality, and service localities;
- three accessible service-mode controls;
- bounded travel-radius input;
- loading, empty, save, success, validation, and safe retry states;
- keyboard/focus support and 44×44px minimum controls;
- pt-PT and English copy with Spanish structural parity;
- visible OpenStreetMap attribution near locality reference use.

No browser component renders or stores internal owner IDs, precise addresses, contact details, or upstream diagnostics.

## Source import and reproducibility

A developer-only source tool:

1. downloads CAOP 2025 GeoPackage to an OS temporary path;
2. verifies a pinned SHA-256 recorded in the source manifest;
3. extracts only required district/municipality/parish codes and parent relations;
4. queries the five locality centers serially under the Nominatim policy;
5. writes a deterministic reviewed manifest containing only approved names, codes, points, source IDs, versions, retrieval dates, and attribution;
6. deletes downloaded/temp artifacts after verification.

The application and tests never call Nominatim or download CAOP at runtime. Updating source versions requires a reviewed forward migration and regenerated manifest.

## Error and logging policy

- Public errors contain only allowlisted code/message/request ID.
- Logs may contain route, status, duration, operation, and correlation ID.
- Logs must not include profile biography, display name, token, subject, internal user ID, exact input arrays, database URL, source response bodies, or raw coordinates tied to a user.
- Reference import failures report source stage and count only, never downloaded payload fragments.

## TDD and acceptance evidence

Every task follows RED → GREEN → REFACTOR. Slice completion requires:

1. migration/schema contracts and complete forward migration chain;
2. real PostgreSQL/PostGIS hierarchy and radius tests with multiple localities;
3. provider-disabled and cross-owner negative tests;
4. transactional profile create/update/child-replacement/concurrency proof;
5. strict Go and BFF contract matrices for unknown/null/duplicate/bound violations;
6. OpenAPI regeneration and drift check;
7. localized component/page/i18n tests;
8. full Go test/race/vet/build and frontend verify gates;
9. Compose config and whitespace/secret/privacy scans;
10. live reference HTTP evidence;
11. real approved Clerk browser profile create → read → replace → read proof showing only status, correlation parity, response shape, service-area/language counts, capability state, and stable owner identity without emitting values;
12. frozen candidate review and local commit before the next slice.

## Explicit deferrals

- Taxonomy mutation APIs/UI and platform-admin role: Slice 11.
- Public provider DTO/pages and SEO: Slice 4.
- Listings/category ownership: Slice 3.
- Contact methods/hours and contact reveal: Slice 5.
- Verification, documents, moderation UI, analytics, subscriptions, promotions, and payments: later approved slices.

## Sources

[1] https://www.dgterritorio.gov.pt/atividades/cartografia/cartografia-tematica/caop — Carta Administrativa Oficial de Portugal — DGT
[2] https://www.ine.pt/ngt_server/attachfileu.jsp?look_parentBoui=456019385&att_display=n&att_download=y — 2025 Entidades do Setor Institucional das Administrações Públicas — INE
[3] https://operations.osmfoundation.org/policies/nominatim — Nominatim Usage Policy — OpenStreetMap Foundation
[4] https://www.openstreetmap.org/copyright — Copyright and License — OpenStreetMap
