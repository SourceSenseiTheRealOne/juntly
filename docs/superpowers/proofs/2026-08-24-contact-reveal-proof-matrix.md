# Slice 5 Contact Reveal Proof Matrix

**Branch:** `feature/contact-reveal-leads`
**Baseline:** `56961c8`

| Requirement | Code/Test Evidence | Live Evidence | Status |
|---|---|---|---|
| Provider contact is server-custodied encrypted | AES-256-GCM `crypto.go`; ciphertext/nonce/key-version migration; real channel-store test | Local migration applied | Proven |
| Provider controls enabled/consent state | `provider_service.go`; strict authenticated handler/BFF tests | Authenticated provider page loaded; user confirmed configuration save works | Proven |
| Public APIs/pages contain no contact fields | Closed public discovery contracts, BFF validators, `projection_test.go` | Go public search and BFF detail returned 200 with forbidden JSON keys absent; rendered public page showed only reveal buttons | Proven |
| Customer reveal requires verified identity | Go handler/router and BFF tests require server-side Clerk bearer | Unauthenticated contact BFF compile/runtime probe returned 401 | Proven |
| Policy denial precedes decryption | `reveal_service_test.go` | N/A — unit/domain boundary | Proven |
| Lead event and same-day idempotency | `sql_store_test.go` against local PostgreSQL | Synthetic provider/customer service flow: first and repeat reveal completed; event count 1; daily count 1 | Proven |
| Daily abuse cap | Concurrent real PostgreSQL test: 11 requests → 10 successes, 1 forbidden | N/A — real DB integration proof | Proven |
| Contact plaintext remains transient client state | `contact-reveal-control.tsx` tests and UI scan reject storage/query/cookie use | No contact present in public initial HTML | Proven |
| Separate authenticated customer browser reveal | BFF, Go handler, service/store tests cover flow | No distinct approved Clerk browser customer session was available | **Blocked** |

## Honest acceptance boundary

Source, migrations, contracts, BFFs, provider UI, service-level acceptance, public runtime proof, and cleanup are verified. Final customer-browser acceptance remains blocked only by absence of a distinct authenticated customer session; no identity/session was fabricated to cover that gap.
