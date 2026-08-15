# Product and Engineering Constraints

## Product rules

- One user account may be customer, provider, or both.
- Informal individuals may advertise allowed services without owning a company; account types and verification are represented accurately.
- Authenticated contact reveal is free. Never charge merely to see an enabled provider contact method.
- Users may communicate/pay outside Juntly; no transaction commission applies to external arrangements.
- Platform payment is optional and uses a compliant provider abstraction, never custom escrow.
- Providers never see competing proposals.
- Public pages expose only approximate provider location and no raw private contact data.
- Categories/subcategories, fees, plans, currencies, promotion periods, translations, and feature flags are administered/configurable, not hardcoded in UI components.
- Promotions are clearly labelled and cannot completely override relevance, distance, trust, or quality.
- Verification badges explain what was verified and never claim guaranteed service quality.
- Essential marketplace participation is not blocked behind a professional subscription.

## Architecture rules

- Begin as a modular monolith; independently deployed microservices require measured need and a new decision.
- OpenAPI is the API contract authority. The generated/aligned TypeScript client reaches Go through the same-origin Next.js BFF/proxy.
- Go is authoritative for business rules, authorization, ownership, state transitions, idempotency, persistence, and fee/review eligibility.
- Browser/component code never imports database, Ent, Supabase secret/service-role, payment secrets, or Go-internal concerns.
- Clerk is identity authority; Go uses a unique opaque internal user mapping.
- PostgreSQL/PostGIS is source of truth. Redis is optional ephemeral infrastructure, never canonical data.
- Reviewed SQL migrations are required; no automatic production schema mutation.

## Delivery rules

- Default language is pt-PT; English is supported; Spanish-ready structure is preserved.
- Mobile-first, low-bandwidth behavior and WCAG 2.2 AA are release concerns, not polish-only work.
- Behavioral code follows RED → GREEN → REFACTOR.
- One cohesive feature branch/PR targets `development`; promotion is `development → staging → main`.
- Build/test/runtime/browser evidence is required before completion claims.
- Product, architecture, security, workflow, or constraint changes update context in the same PR.
- Secrets, private data, raw chats, sessions, generated intelligence, and credentials never enter Git.
