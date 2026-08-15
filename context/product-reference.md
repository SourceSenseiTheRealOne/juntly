# Juntly — Product and Technical Reference

This document is the primary source of truth for Juntly unless a later explicit instruction changes a requirement and the change is recorded in `decisions.md`. Before major changes, explain how they fit this specification. Prefer pragmatic, maintainable MVP solutions that can scale. Do not silently change business rules; identify conflicts and request a decision.

## 1. Product identity

- **Name:** Juntly
- **Main tagline:** Local people. Real skills.
- **Portuguese tagline:** Encontra quem sabe fazer.

Juntly is a Portuguese company and product with an international brand. It launches in Portugal, initially targeting rural communities such as Zebreira, Idanha-a-Nova, Penha Garcia, Monsanto, Castelo Branco, and surrounding areas.

Although the first market is rural Portugal, the product must support later expansion to other Portuguese regions, Spain, other European countries, urban and suburban communities, additional languages, currencies, and payment methods. It should feel Portuguese in values and origins without being technically restricted to Portugal.

## 2. Product vision

Juntly is a local-services marketplace connecting people who need a service with people who can provide it. Rural communities contain skilled people offering repairs, cleaning, construction, gardening, agricultural work, transportation, food preparation, animal care, elderly assistance, technical support, private lessons, and many other services. Providers are often hard to find because they advertise through word of mouth, Facebook groups, physical notices, or personal contacts.

Juntly centralizes local supply and demand by combining a service directory, direct communication, requests for quotations, booking, optional protected payments, reputation, and professional tools.

The platform must allow users to:

- Discover nearby service providers.
- Publish services they provide.
- Search by service, category, location, distance, price, availability, language, and reputation.
- View provider profiles and service listings.
- Contact providers through internal chat, phone, or WhatsApp.
- Publish requests for quotations.
- Receive and compare private quotations from multiple providers.
- Accept a quotation and create a booking.
- Pay inside Juntly when desired.
- Arrange and pay outside Juntly when preferred.
- Leave ratings and reviews.
- Promote service listings in search results.
- Subscribe to professional provider plans.

Juntly supports informal individual providers and registered professional businesses. A person does not need a company to advertise an allowed service, but the platform distinguishes individuals, verified professionals, and registered businesses.

## 3. Product principles

- **Low friction:** Find and contact providers quickly.
- **Trust:** Profiles, verification, reviews, moderation, and transparent information are essential.
- **Accessibility:** Understandable to users with limited technical experience.
- **Mobile first:** Most rural users are expected to use smartphones.
- **Low-bandwidth friendly:** Core functionality works on slow connections.
- **Local relevance:** Search and recommendations prioritize proximity.
- **User choice:** Users are not forced to pay or communicate exclusively through Juntly.
- **Transparent monetization:** Sponsored results and fees are clearly identified.
- **International scalability:** Locations, currencies, languages, categories, and payment providers are configurable.
- **Pragmatic architecture:** Start with a modular monolith, not unnecessary microservices.

## 4. User roles

A single account may act as customer, provider, or both.

### Visitor

Can browse public listings, search services and locations, view public provider profiles, read public reviews, explore categories, and register/sign in. Cannot reveal private phone numbers or send messages until authenticated.

### Customer

Can contact providers; reveal phone/WhatsApp contact details; start internal conversations; publish quotation requests; receive and compare proposals; accept proposals; create bookings; optionally pay through Juntly; save favourite services/providers; review eligible completed services; and report listings, users, messages, or reviews.

### Service provider

Can create a provider profile; publish/manage multiple services; define service areas and travel distance; configure contact methods/hours; receive messages and quotation opportunities; submit private quotations; accept/manage bookings; receive platform payments/payouts; promote listings; subscribe to a professional plan; and view lead/performance statistics.

### Administrator or moderator

Can manage users, providers, listings, categories, locations, reports, reviews, verification requests, disputes, refunds, commissions, subscription plans, promotions, feature flags, translations, static content, and marketplace analytics. Important administrative actions create audit records.

## 5. Provider profiles

A provider profile supports:

- Display name and profile photograph/business logo.
- Individual, professional, or registered-business account type.
- Short biography and languages spoken.
- Main approximate location, areas served, and maximum travel radius.
- Whether the provider travels to customers, receives customers, or works remotely.
- Phone, WhatsApp, and internal-chat availability.
- Preferred contact method and contact hours.
- Email, phone, identity, professional, and business verification statuses.
- Business details when applicable.
- Average rating/review count and completed platform bookings.
- Response rate and typical response time.
- Active service listings and portfolio images.
- Account creation date.

Private information must never be exposed unintentionally. A precise home address is not public unless explicitly configured as a public business location.

## 6. Service listings

A provider may publish multiple listings. Each supports:

- Title, detailed description, category, and subcategory.
- Photographs and optional video.
- Starting price and fixed, hourly, daily, quotation-only, or negotiable pricing.
- Currency.
- Service location and travel radius.
- Customer-location, provider-location, on-site, or remote service.
- Availability and estimated duration.
- Languages supported and information required from the customer.
- Cancellation conditions.
- Draft, pending-review, active, paused, rejected, or archived status.
- Organic or promoted status.
- Creation and update timestamps.

Categories/subcategories are dynamically administered and never hardcoded in the interface. Examples include cleaning, garden maintenance, plumbing, electrical work, construction, small repairs, agricultural assistance, transportation, meal preparation, computer repair, animal care, elderly assistance, private lessons, beauty, and wellness.

## 7. Location model

Support country; district/region; municipality; parish; village/town/city; postal code; latitude/longitude; provider travel radius; and customer-provider distance. Providers may serve multiple areas. Search prioritizes nearby relevant providers while allowing radius expansion. Normally only approximate provider locations are public; exact service addresses belong to private bookings and conversations.

## 8. Search and discovery

Support ordinary phrases such as “plumber near Zebreira,” “garden cleaning,” “someone to repair a water heater,” or “transport to Castelo Branco.”

Filters include category/subcategory, location/distance, price type, availability, rating, verification, individual/business, on-site/remote, and language.

Organic ranking considers text relevance, distance, availability, rating, verified reviews, response rate, recent activity, listing completeness, provider verification, and reliability. Promotions may boost ranking but are clearly labelled Sponsored/Promoted and never completely override relevance, distance, trust, or quality.

## 9. Direct contact system

Customers may contact providers through Juntly chat, phone, or WhatsApp. Public pages must not contain raw phone numbers in HTML, metadata, or search-engine indexes.

To reveal phone/WhatsApp, a customer must register/sign in, click the reveal/contact action, pass rate-limit/abuse checks, and generate a contact/lead analytics event. Providers control enabled methods, contact hours, and whether only chat is accepted.

Contact information remains free after authentication. Customers do not pay merely to see contact information. When users negotiate/pay outside Juntly, no transaction commission is charged; Juntly may monetize through promotions and professional subscriptions.

## 10. Internal chat

Authenticated chat supports real-time or near-real-time messages, history, unread indicators, text, image attachments, optional documents, blocking/reporting, notification preferences, and links to related listings, requests, proposals, or bookings. A conversation may belong to a listing, quotation request, proposal, booking, or general enquiry.

## 11. Quotation request system

Quotation requests are a core launch feature. A customer request contains title/description, category/subcategory, location, preferred date/range, urgency, photos/documents, optional budget, fixed/flexible budget, required travel distance, preferred contact method, and proposal deadline.

A request may be open to eligible nearby providers, sent to selected providers, or created from a listing. Publishing identifies eligible providers by category, location/radius, availability, and account status, then notifies them. Providers submit private proposals; customers compare proposals/profiles, accept one, reject proposals, contact providers, or close the request.

A proposal contains proposed price, breakdown/notes, available date/time, estimated duration, message, included/excluded work, optional attachments, expiration, and whether protected Juntly payment is supported. Proposals are private; providers never see competitors’ proposals. Accepting a proposal may create a booking, after which the customer chooses protected platform payment or a direct external arrangement.

## 12. Booking workflow

Bookings originate from listings, accepted quotations, or direct chat agreements. Statuses: Draft, Pending provider confirmation, Confirmed, Scheduled, In progress, Completed, Cancelled, Disputed, Refunded.

A booking references customer, provider, applicable listing/request/proposal, scheduled date, private service location, agreed price, fees, payment status, cancellation information, completion confirmation, and review eligibility. The Go API validates transitions. Both parties receive clear status notifications.

## 13. Payments and transaction protection

Juntly payment is optional but encouraged through transparent benefits: payment records, support, verified reviews, and dispute handling. Payments sit behind a provider abstraction for later expansion/change.

Eventually support cards, MB WAY or compatible Portuguese payment provider, other European methods, marketplace split payments, provider payouts, refunds, platform commission, customer-protection/service fees, and payment/payout status tracking.

For Juntly payments: show customers full breakdown before payment; show providers expected payout; calculate fees automatically; link payment to booking; enable verified-review eligibility after successful completion; provide refund/dispute workflows; and apply configured payout rules.

Do not implement custom legal escrow. Use compliant marketplace-payment functionality from an established provider. All fees/percentages are administrator-configurable and never hardcoded.

## 14. Reviews and reputation

A review supports overall rating, written feedback, optional punctuality/communication/quality/value ratings, provider response, moderation status, and verified-booking indicator. Completed platform-booking reviews are marked Verified.

Contact-only reviews may come later but must be distinguished and abuse-resistant. Users cannot self-review, duplicate-review a booking, or review without an eligible interaction. Administrators may moderate fraudulent/abusive/unlawful reviews without manipulating legitimate criticism.

## 15. Provider verification

Verification may cover email, phone, identity, business, professional credentials, and address/service area. Verification never means Juntly guarantees work quality; badge meaning is explained accurately. Verification may affect organic ranking but is not sold as a fake trust badge.

## 16. Monetization

Juntly uses a hybrid model.

### Free access

At launch, account creation, basic listings, search, authenticated contact reveal, internal chat, quotation requests, and a reasonable number of proposals are free.

### Promoted listings

Providers may promote listings for configurable periods. Example assumptions are about €2/day or €6–€10/week, always configurable. Promotions are clearly labelled.

### Professional subscription

Example assumption: €7.99–€14.99/month. Potential benefits: more active listings/photos, professional presentation, advanced availability calendar, analytics, reduced transaction commission, promotion credits, priority support, portfolio features, reusable quotation templates, and business-profile features. Essential participation is not paywalled.

### Transaction commission

Juntly may charge about 8–12% on platform-processed payments, configurable by plan, category, country, launch promotion, or payment method. No commission applies to external arrangements.

### Customer protection fee

A small transparent service fee may apply when customers choose protected platform payment.

### Institutional partnerships

Future revenue may include service packages for municipalities, parish councils, rural associations, cooperatives, local-development organizations, and business associations. Advertising is not an initial priority.

## 17. Notifications

Deliver in-app and email notifications, later optional web push, SMS, or WhatsApp. Events include messages, nearby requests, proposals, proposal decisions, booking updates, payments/payouts, reviews, moderation, expiring promotions, subscription changes, verification, and disputes. Users configure non-essential notifications.

## 18. Administration and moderation

Administration supports users, providers, listings, categories, locations, listing approval/moderation, reports, review moderation, identity/business verification, quotation/booking inspection, payment/payout/refund monitoring, disputes, promotions/subscriptions, configurable fees/plans, feature flags, translations/static content, audit logs, and marketplace analytics.

## 19. Analytics

Track search, listing/provider views, contact reveal, WhatsApp clicks, chat start, quotation creation, proposal submission/acceptance, booking creation, payment completion, review submission, promotion, and subscription lifecycle.

Metrics include active providers/locality, active listings/category, search-to-contact conversion, time to first proposal, proposals/request, provider response rate, booking conversion, repeat customers, gross transaction value, promotion/subscription/commission revenue, and report/cancellation/refund/dispute rates. Analytics respects consent and European privacy requirements.

## 20. Internationalization

Default language: Portuguese from Portugal (`pt-PT`). Architecture supports English, Spanish, and later languages. User-facing strings are not hardcoded inside components. Support locale-specific dates, time zones, currencies, numbers, translated categories, static pages, notification templates, and SEO metadata. User-generated descriptions need no automatic MVP translation, but the model must not prevent later translation.

## 21. Accessibility and experience

Use clear language, large touch targets, readable typography, simple navigation, short multistep forms, clear confirmations, strong contrast, keyboard/screen-reader support, helpful empty states, useful errors, minimal unnecessary animation, and fast loading. Target WCAG 2.2 AA where practical.

## 22. Technical stack

### Frontend

Use Next.js, TypeScript, App Router, responsive mobile-first UI, server rendering where useful, PWA capabilities where appropriate, reusable components/design tokens, proper internationalization, and typed API integration.

The frontend eventually includes public marketplace, authentication, customer/provider dashboards, chat, quotations, bookings, payments, and protected administration. Public listing/provider pages are SEO-optimized without exposing private contact data.

### Backend

Use Go, JSON REST, OpenAPI, modular monolith, clear domain/application/transport/infrastructure boundaries, background jobs as needed, and WebSockets or SSE where appropriate.

Suggested modules: authentication, users, provider profiles, locations, categories, listings, search, contacts/leads, chat, quotation requests, proposals, bookings, payments, reviews, verification, promotions, subscriptions, notifications, reports, administration, and analytics events. Do not start with independent microservices; preserve extraction-ready module boundaries only where useful.

### Database and infrastructure

Use PostgreSQL as relational source of truth; consider PostGIS for distance/radius and PostgreSQL full-text search initially. Use migrations, transactional booking/payment consistency, audit timestamps, and appropriate soft deletion. Redis may support caching, rate limiting, temporary state, and realtime coordination. Use S3-compatible storage, transactional email, background workers, structured logging, metrics, and error monitoring. Add a dedicated search engine only when PostgreSQL is insufficient.

## 23. Docker requirements

The complete local environment must run with Docker. Provide separate Dockerfiles for Next.js and Go. The local topology eventually runs frontend, API, PostgreSQL, Redis when used, S3-compatible local storage when used, and local email testing when useful.

Under the Coding Lab’s accepted project-specific infrastructure decision, the project-owned local Supabase stack owns PostgreSQL/PostGIS and local mail services rather than duplicating PostgreSQL in application Compose. Application Compose runs frontend/API and justified adjunct services against that database boundary.

Use environment configuration, health checks, named persistent volumes, internal networking, explicit ports, and development/production-appropriate builds. Never store secrets in images or committed environment files; include a safe `.env.example` when runtime configuration begins.

## 24. Frontend and API integration

The Go API is authoritative for business data/rules. Next.js communicates through a typed client generated from or aligned with OpenAPI. The approved project boundary is browser → same-origin Next.js BFF/proxy → Go REST API, preserving OpenAPI as the contract authority without duplicating tRPC schemas.

The API provides consistent errors, pagination/filtering/sorting, validation, authentication/authorization, idempotency for booking/payment/refund/promotion purchases, versioning, correlation IDs, secure uploads, and rate limiting. Critical business rules are not duplicated in frontend components.

## 25. Authentication and security

Initial authentication supports email/password, email verification, password reset, and secure session management. Magic links, social login, and phone auth may come later. Clerk is the approved identity authority; Go maps verified Clerk subjects to opaque internal users and enforces domain authorization.

Requirements: secure session revocation; HTTP-only secure cookies or justified alternative; CSRF where applicable; rate limiting; input validation/output encoding; permission checks on every protected resource; upload validation/limits; abuse reporting/account blocking; audit logs; contact privacy; and automated-harvesting prevention.

## 26. Privacy and legal requirements

Design for GDPR/European requirements: clear consent, privacy policy, terms, cookie controls where needed, data export, deletion requests, retention, minimization, secure personal-data storage, provider consent for contact display, and clear platform-mediated/external-transaction distinction. Identity documents require stronger protection and limited administrative access. Never claim Juntly guarantees every provider/service; explain verification, reviews, payment protection, and dispute support accurately.

## 27. Core domain relationships

- A User may be customer, provider, or both.
- A User may own one ProviderProfile.
- A ProviderProfile owns multiple ServiceListings.
- A customer creates a QuotationRequest.
- A QuotationRequest receives multiple private Proposals.
- One accepted Proposal may create one Booking.
- A Booking may create one or more PaymentTransactions.
- A completed eligible Booking allows a Review.
- A Conversation may reference a listing, request, proposal, or booking.
- Contact reveal creates a LeadEvent.
- A provider may have a Subscription.
- A listing may have time-limited Promotions.
- A Report may target a user, listing, review, message, conversation, or booking.

## 28. Important workflows

- **Contact:** listing view → authentication → reveal → lead event → phone, WhatsApp, or chat.
- **Quotation:** request → provider matching → notifications → private proposals → acceptance → booking or direct contact.
- **Protected booking:** listing/accepted proposal → booking → payment → completion → payout → verified review.
- **External arrangement:** listing/quotation → direct contact → external agreement → no commission.
- **Promotion:** provider selects listing/period → payment → activation/ranking boost → automatic expiration.
- **Subscription:** plan → payment → entitlements → renewal/cancellation/expiration.

## 29. MVP priorities

The first launch validates whether local supply/demand can connect successfully.

Essential launch scope:

- Registration/authentication.
- Customer/provider profiles.
- Dynamic categories.
- Listing creation/moderation.
- Location/radius search.
- Public provider/listing pages.
- Authenticated phone/WhatsApp reveal.
- Internal messaging.
- Quotation requests and private proposals.
- Proposal comparison/acceptance.
- Basic bookings.
- Ratings/reviews.
- In-app/email notifications.
- Promoted listings.
- Basic professional subscriptions.
- Administration/moderation.
- Portuguese and English localization, Spanish prepared.

Potential immediate follow-ups after validation: full protected payments/payouts, MB WAY-compatible payments, advanced verification/analytics, saved searches, web push, dispute automation, institutional dashboards, native apps, and AI-assisted matching/quotation suggestions.

Secondary features must not delay validation of discovery, direct contact, chat, and quotation requests.

## 30. Development standards

- Understand the affected product flow first.
- Identify affected entities/modules.
- Present an implementation plan before significant work.
- Keep architecture modular but simple.
- Define API contracts before coupling frontend to incomplete endpoints.
- Use migrations for database changes.
- Add unit/integration tests for business rules.
- Add end-to-end tests for critical flows.
- Update documentation when behaviour changes.
- Do not silently alter requirements.

Important tests cover authentication, authorization, listing permissions, contact privacy, geographical search, quotation visibility, proposal privacy, proposal acceptance, booking transitions, payment idempotency, review eligibility, promotion expiration, subscription entitlements, and administration permissions.

## 31. Final product definition

Juntly is neither only an online directory nor only a payment marketplace. It is a local-service network combining discovery, direct contact, internal communication, quotation requests, provider competition, booking, optional protected payment, reputation, promotion, and professional tools.

Its advantage is reducing the difficulty of finding trusted local services where existing platforms have limited coverage. Preserve direct human communication while adding enough digital structure for trust, discovery, convenience, and local economic opportunity. Always maintain this balance when proposing, designing, or implementing features.
