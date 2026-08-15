# Product Context

## Identity

- Name: Juntly
- International tagline: **Local people. Real skills.**
- Portuguese tagline: **Encontra quem sabe fazer.**
- Owner: SourceSensei
- Trust boundary: public source repository; private marketplace/customer data
- Origin: Portugal, designed for international expansion

## Problem and users

Rural communities contain skilled providers who are difficult to discover through fragmented word-of-mouth, social groups, notices, and personal contacts. Juntly gives customers a low-friction way to find nearby people and businesses while giving informal and professional providers a trustworthy place to publish services, communicate, quote, book, build reputation, and optionally transact.

A single account may be a customer, provider, or both. Visitors browse public content; authenticated customers contact and buy; providers publish and manage services; administrators moderate and configure the marketplace.

## Launch geography

Initial focus: Zebreira, Idanha-a-Nova, Penha Garcia, Monsanto, Castelo Branco, and nearby rural communities. Geography, language, currency, and payment-method models must remain configurable for later Portugal, Spain, Europe, urban, and suburban expansion.

## Essential launch outcomes

- Email/password registration and verified account recovery.
- Customer and provider profiles with privacy-safe public data.
- Dynamically administered categories and moderated service listings.
- Location/radius discovery and public SEO-safe listing/provider pages.
- Free authenticated phone/WhatsApp reveal with lead events and abuse controls.
- Internal messaging.
- Quotation requests, private proposals, comparison, and acceptance.
- Basic bookings with Go-validated state transitions.
- Eligible ratings/reviews.
- In-app and email notifications.
- Clearly labelled promoted listings and basic professional subscriptions.
- Administration, moderation, audit records, pt-PT and English; Spanish-ready structure.

## Explicitly deferred until after initial validation

Full protected payments/payouts, MB WAY-compatible payments, advanced verification and analytics, saved searches, web push, dispute automation, institutional dashboards, native mobile apps, and AI-assisted matching. Payment/provider abstractions may be designed earlier, but secondary features must not delay discovery, direct contact, chat, and quotations.

## Non-functional requirements

- Mobile-first, low-bandwidth-friendly, clear for users with limited technical experience.
- WCAG 2.2 AA target where practical.
- GDPR/European privacy by design, data minimization, export/deletion/retention controls.
- pt-PT default with locale-aware dates, times, currencies, numbers, content, notifications, and SEO.
- Configurable categories, fees, plans, localities, languages, currencies, and providers.
- Modular monolith with explicit boundaries; no premature microservices.
- Structured logs/metrics/errors that exclude sensitive payloads.

## Success evidence

The MVP is validated by observable local-market outcomes: active providers per locality/category, successful search-to-contact conversion, timely quotation proposals, accepted proposals/bookings, repeat customers, provider response rate, and low abuse/report/dispute rates. Engineering completion additionally requires tests, builds, migrations, container health, real HTTP/browser probes, and privacy/authorization evidence for each implemented flow.
