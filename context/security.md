# Security Context

## Data classification

- **Public:** approved listing fields, approximate service area, public provider profile, moderated reviews, categories, static content.
- **Internal:** moderation state, ranking factors, lead/analytics aggregates, operational configuration.
- **Confidential personal:** email, phone/WhatsApp, exact addresses, conversations, proposals, booking details, notification preferences.
- **Restricted:** identity/business documents, authentication/session data, payment/payout identifiers, dispute evidence, administrator audit data.
- **Prohibited in Git/logs/chat:** credentials, tokens, cookies, `.env`, identity documents, production records, raw private messages, payment secrets.

## Trust boundaries

Untrusted browsers and uploads; Next.js BFF; Go API; Clerk; local/remote Supabase PostgreSQL; optional Redis; object storage; email/notification providers; payment provider; CI; and isolated deployment environments. Every network boundary validates input and returns allowlisted output.

## Authentication and authorization

Clerk owns identity and secure email/password sessions. Go maps verified subjects to internal users and enforces role, entitlement, provider/customer domain role, ownership, conversation/booking membership, proposal visibility, moderation powers, and administrative access. Missing/malformed identity fails closed. UI gates are never authorization.

## Privacy controls

- Raw phone/WhatsApp values never appear in public HTML, metadata, search indexes, or unauthenticated responses.
- Contact reveal requires authentication, provider consent/configuration, rate limiting, abuse checks, and a lead event.
- Public location is approximate; precise service addresses are private booking/conversation data.
- Competing proposals are never disclosed.
- Identity documents use encryption/strong access controls, limited administrative roles, access audit, retention/deletion rules, and no public object URLs.
- GDPR requirements include consent, minimization, purpose limitation, export, deletion requests, retention, and cookie controls where required.

## Threat controls

Apply strict schema validation, output encoding, parameterized ORM/query APIs, reviewed raw SQL, CSRF protection where applicable, XSS/injection/SSRF/open-redirect defenses, secure cookies, session revocation, rate limits, anti-harvesting, account blocking/reporting, upload byte/type/dimension limits, malware/content controls where justified, idempotency for financial/booking mutations, and bounded external calls.

Logs use allowlisted structured fields and correlation IDs. Never log authorization headers, cookies, tokens, contact values, addresses, message/proposal bodies, identity documents, payment data, or raw uploads. Important admin/payment/moderation actions create tamper-evident audit records.

## Payments and trust claims

Use an established compliant marketplace-payment provider; never custom legal escrow or bespoke cryptography. Fee calculations are server-owned/configurable and idempotent. Verification/reviews/payment protection are explained accurately; Juntly never guarantees every provider or quality of work.

## Secrets and approval boundaries

Secrets live in approved environment/secret stores and ignored local files, never committed or printed. Human approval is required for production/main promotion, deployment, credentials/external access, paid resources, destructive migrations, remote database operations, refunds outside approved policy, and high-impact moderation/security changes.
