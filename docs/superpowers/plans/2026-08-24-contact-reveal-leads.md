# Contact Reveal and Lead Events Implementation Plan

**Goal:** Add encrypted, provider-controlled contact reveal with durable lead events and atomic customer/day abuse limits.

**Architecture:** Keep provider contacts in a server-only encrypted channel vault. Compose owner configuration and customer reveal services over verified durable identity, listing state, and PostgreSQL transactional rate/event storage. Public discovery types are not reused for contact disclosure.

## Task 1: Contact vault and durable reveal state

- [ ] Write RED crypto/config tests for 32-byte server-only AES-GCM key parsing, encryption round-trip, and malformed key failure.
- [ ] Add forward-only migration/Ent schemas for contact channels, daily limits, and reveal events with FK indexes/unique constraints.
- [ ] Build transactional repository with conditional daily limit increment and same-day idempotency.
- [ ] Prove owner/channel/event cleanup and concurrency in real PostgreSQL.

## Task 2: Provider contact configuration and reveal services

- [ ] Write RED service tests for provider ownership, consent/enabled state, self-reveal, active-only listing, decrypt-after-policy ordering, daily cap, and event redaction.
- [ ] Add owner configuration service and customer reveal service using verified identity reconciliation.
- [ ] Run focused/race/full backend gates and commit persistence/service slice.

## Task 3: Contract, Go transport, and same-origin BFF

- [ ] Add closed OpenAPI schemas/routes and generated client.
- [ ] Add authenticated Go handlers: owner channel status/config and customer reveal.
- [ ] Add same-origin BFF routes that acquire Clerk token only server-side.
- [ ] Prove public APIs never expose contact fields and commit transport slice.

## Task 4: Private provider UI and authenticated reveal UX

- [ ] Add provider contact configuration page/card with masked status only.
- [ ] Add authenticated reveal action on public listing detail; successful plaintext stays in component state only.
- [ ] Add localized tests, browser proof, full gates, and commit UI slice.

## Task 5: Acceptance and integration

- [ ] Create local synthetic provider/customer channels and reveal flow through real auth/session when available.
- [ ] Prove single event/idempotency/rate cap, public absence, cleanup, ports closed, full gates, privacy scan, commit, and local development fast-forward.
