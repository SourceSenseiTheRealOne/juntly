# Public Discovery and Search Design

**Status:** Approved under SourceSensei’s standing implementation authorization on 2026-08-24

## Goal

Deliver Slice 4: active-listing-only public discovery and SEO-safe listing pages for the Portugal launch area without exposing owner identity, contact data, exact provider location, moderation records, storage metadata, or private lifecycle history.

## Scope decisions

1. Public projection reads only `active` listings. Draft, pending-review, rejected, paused, and archived rows are indistinguishable from absent rows.
2. Public locality is an approximate named launch locality, not a raw coordinate, address, provider service area list, or distance-to-provider claim.
3. Discovery is database-backed and fixed-query: category, text, locality radius, price type, and service-mode filters. Availability is explicitly deferred because it has no durable data model yet.
4. Public pages expose provider display name/type only when a separate allowlisted provider projection is joined from an active listing; no profile page is introduced until the page contract is proved.
5. No sponsored/promoted ranking exists in Slice 4. Organic ordering is deterministic: exact category relevance, locality/radius relevance, active recency, stable ID tie-break.

## Public projection

`PublicListing` contains only:

```text
id, slug, title, description, category name/slug,
primary locality name/slug, price type, nullable price minor, currency,
service modes, provider display name/type, published/updated timestamp
```

The owner profile’s internal UUID, service localities, languages, biography, contact methods, media object references, listing events, review reason, created draft timestamp, and any coordinates remain excluded.

## Discovery contract

```text
GET /api/v1/discovery/listings?locale=pt-PT
  &categoryId=<uuid>
  &q=<2..80 chars>
  &nearLocalityId=<uuid>&radiusKm=<1..200>
  &priceType=fixed|hourly|daily|quote|negotiable
  &serviceMode=travels_to_customer|receives_customer|remote_services

GET /api/v1/public/listings/{listingId}
```

All filters are optional but exact-key allowlisted. Paired locality/radius values are mandatory together. Query text is whitespace-normalized, parameterized, and never injected into SQL identifiers.

## Repository and ordering

The repository joins `listings`, active categories/localities, and provider profiles. Radius uses the existing reviewed PostGIS locality center semantics:

```text
ST_DWithin(listing_locality.center, origin.center, radiusKm * 1000)
ORDER BY category relevance, radius distance, listing.updated_at DESC, listing.id
```

The query never selects provider coordinate/address columns. `active` state is a mandatory predicate, not a UI convention.

## SEO and UI

Public route:

```text
/:locale/listings/[listingId]
```

Discovery route:

```text
/:locale/discover
```

Metadata uses listing title, category/locality names, and a plain description. It includes no contact values, internal identifiers, precise location, review/moderation history, or non-active listing content. Low-bandwidth view uses text-first cards and no mandatory media.

## Acceptance

1. RED→GREEN database tests prove inactive-state exclusion, category/text/filter semantics, radius ordering, deterministic ties, and no private selected columns.
2. Go/OpenAPI/BFF tests prove strict query allowlists, safe `400/503`, and public DTO redaction.
3. Public routes/BFF do not use Clerk tokens.
4. Browser proof finds an active nearby listing and proves draft/pending/rejected/paused/archived routes return no public content.
5. Full Go/PostgreSQL/frontend/Compose/CodeGraph gates, frozen privacy scan, local commit, and cleanup pass.
