# Vila Supabase Media Implementation Plan

> **For agentic workers:** Execute inline with `superpowers:executing-plans` and strict RED → GREEN → REFACTOR in the canonical checkout.

**Goal:** Replace unavailable listing storage with a private Supabase upload and validated publication flow.

**Architecture:** Clerk/Go own identity and listing authorization. Go mints a path-specific, non-overwriting signed upload capability for a private quarantine bucket. Raw objects never become public merely because an upload completed. Finalization must verify the stored object, sanitize image metadata, persist a ready record and serve images only for authorized owners/moderators or active approved listings.

**Tech Stack:** Existing Go HTTP adapter boundary, Ent/PostgreSQL, Supabase Storage REST, Next.js BFF, generated OpenAPI.

## Constraints

- Supabase Storage is owner-approved; no new remote accounts, buckets, secrets or paid resources are authorized.
- Server key never reaches browser, logs or errors; only narrow signed upload capability may be returned.
- Allow JPEG, PNG, WebP; existing 10 MiB limit. Reject path escape, cross-origin returned URLs, redirects, oversized/malformed responses.
- Do not enable raw uploads as public listing photos. Keep runtime unavailable until finalization/read controls are implemented and verified.

## Task 1: Authorize before minting upload capabilities

Modify `backend/internal/listingmedia/repository.go`, `ent_repository.go`, `service.go`, `service_test.go`; extend `backend/internal/listings/listingmedia_repository_test.go`.

- [ ] RED: authorized provider uploading to another owner's/non-editable listing must cause zero storage calls.
- [ ] GREEN: add `RequireEditable(context.Context, uuid.UUID, uuid.UUID) error` to the repository; call it before storage. Keep the existing reservation ownership/state check as a second guard.
- [ ] Verify service negative tests and actual PostgreSQL/Ent owner isolation.

## Task 2: Private Supabase signed-upload adapter

Create `backend/internal/listingmedia/supabase_storage.go` and `supabase_storage_test.go`.

- [ ] RED: `NewSupabaseStorage(SupabaseStorageConfig)` returns `Storage`; `CreateUploadReservation` POSTs `/storage/v1/object/upload/sign/<bucket>/pending/<media-uuid>` with server Authorization/apikey and `x-upsert: false`, then exposes only the exact path-scoped PUT URL and Content-Type.
- [ ] GREEN: strict origin/bucket config, bounded response and timeout, deny redirects, exact response path and token validation; no key-bearing request/error serialization.
- [ ] Verify loopback HTTP contract, invalid config, cross-origin/wrong-path/no-token responses, redirect refusal and invalid upload before provider I/O.

## Task 3: Finalization, delivery and runtime wiring

Trace the existing owner listing, moderation and public discovery models before adding contracts. Finalization must verify byte count, SHA-256, supported image decoding and dimensions; re-encode to remove EXIF/private metadata, write to a separate server-only immutable key, then atomically mark ready under editable-listing authorization. Add protected owner/moderator media reads and sanitized approved-public reads. Implement Next BFF/upload controls and generated OpenAPI together, then wire complete opt-in configuration.

Acceptance: real local upload → byte verification → ready record → moderated public rendering; cross-owner/non-public reads fail closed; replay cannot replace a verified object. Remote provider configuration and real browser proof remain explicit activation gates, not assumed from unit tests.
