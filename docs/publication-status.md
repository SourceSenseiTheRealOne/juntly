# Publication status and repository identity

Reviewed 24 September 2026. This is a dated evidence boundary, not a claim that the service is publicly launched.

## Published implementation

The presentation review inspected development revision [`4a755a3`](https://github.com/SourceSenseiTheRealOne/vila/commit/4a755a3f3fc3841b38ff01697f45667d28f37a1a). Its file tree matched the then-current staging and main snapshots. The code contains marketplace discovery, provider/listing management, private contact reveal, messaging, quotations, bookings, reviews, entitlements, moderation, administration, booking-payment adapters and listing-media adapters.

Existence of code, unit tests and operational scripts is different from provider acceptance or public deployment. Older README/context text that calls the backend only a health tracer describes an earlier stage; it is not an accurate inventory of this snapshot.

## Outside this publication

Ongoing local work includes the Newsprint visual redesign, fixed-period professional-plan billing, listing promotion checkout, email delivery, quota changes and associated migrations/tests. None of those unpublished changes is included in the presentation commit. Approved launch terms in planning documents do not establish that checkout, renewal, delivery or legal acceptance is ready.

## Public deployment

The intended domain is `somosvila.com`. A public DNS lookup on the review date returned NXDOMAIN, and no working public deployment was verified. The repository therefore shows an architecture diagram rather than a live-demo link or a screenshot implying a launched marketplace.

Legal/operator details, deployment configuration and authenticated/provider acceptance remain launch gates. No active-user, revenue, performance, security-audit or production-readiness claim is made.

## Dependency audit

A lockfile-only `npm audit --package-lock-only --json` against the published frontend manifests on the review date reported 1 critical, 6 high and 2 moderate package findings. The command failed its audit, as expected with these findings; dependency remediation is not part of this documentation slice.

The critical package is Next.js (`16.3.1` in this snapshot), with [GHSA-p293-qw3h-jr36](https://github.com/advisories/GHSA-p293-qw3h-jr36) and [GHSA-2xp9-vwfh-vxw4](https://github.com/advisories/GHSA-2xp9-vwfh-vxw4). High findings include sharp and js-yaml with affected tooling dependants; moderate findings include Vitest tooling. These are package-level audit totals, not a count of distinct exploitable application vulnerabilities. Resolve and reassess them before exposing a server publicly; no clean security-audit claim is made.

## Approved naming boundary

The product and GitHub repository are **Vila** (`SourceSenseiTheRealOne/vila`), with the canonical Lab checkout at `projects/products/vila`. This presentation decision supersedes earlier instructions to preserve the old GitHub repository name.

The following technical identities remain unchanged:

- Go module/import prefix: `github.com/SourceSenseiTheRealOne/juntly/backend`.
- API contract filename and environment names, including `JUNTLY_*`.
- Supabase project ID `juntly`, existing storage/data identities, and `juntly-api` / `juntly-frontend` image names.
- Existing npm package identity and Hermes board `juntly`.

The Lab registry uses the public `vila` slug while retaining the existing board. Repository renaming does not rename or migrate provider resources, database contents or runtime configuration. Historical records retain their original labels.

## Evidence for this update

The presentation slice changes documentation, architecture assets and repository/registration metadata only. Source-tree comparisons and file fingerprints protect the pre-existing application and local drafts. Markdown links must resolve in the published Git tree, not merely to unpublished files on the author's machine.

The current Foundation CI workflow filters by application/configuration paths, so a documentation-only PR does not schedule those application jobs. Do not report skipped-by-path jobs as a new green application run. Full application and live-provider acceptance are separate work.
