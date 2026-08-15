# Project Workflows

## Repository bootstrap

The verified initial context and frontend shell establish `main` as a one-time exception. `development` and `staging` are then created at the same exact SHA. Parent Coding Lab submodule/registry registration is deferred until the dirty parent repository is clean; this does not authorize unmanaged duplicate Juntly checkouts.

## Branches and review

- Feature/fix target: `development`.
- Promotion: `development → staging → main`.
- One cohesive slice per branch/PR; squash merge and delete the source branch.
- Main/production promotion requires SourceSensei approval.
- Work in the one canonical Juntly checkout and switch branches; no duplicate Juntly folders.

## Task execution

1. Load `AGENTS.md` and context in the documented order.
2. Understand the affected product flow/entities and identify specification conflicts.
3. Run/sync CodeGraph and inspect affected boundaries before broad file guessing.
4. Write a bounded plan with acceptance evidence for significant work.
5. Use strict RED → GREEN → REFACTOR for behavior.
6. Run CodeGraph affected analysis, focused tests, and all configured gates.
7. Freeze exact bytes, audit staged paths/secrets, and review at exact SHA.
8. Open a small PR to the legal base, resolve findings, squash merge, verify, and delete branch.

## Frontend bootstrap gates

From `frontend/`: `npm ci`, format check, focused/full tests, ESLint, strict typecheck, production build, high-severity dependency audit, production HTTP probes, browser/console/accessibility review. A build without runtime proof is not live verification.

## Future full-stack gates

Go tests/race/vet/build; OpenAPI deterministic generation; migration reset/equivalence; Docker image/build/health; local Supabase/PostGIS connectivity; BFF-to-Go vertical tracer; signed-out fail-closed auth; credentialed auth separately; upload/privacy/authorization negative tests; runtime/browser proof.

## Environments and delivery

Local Docker first. Production platform remains undecided until cost, data residency, payment provider, operations, backup, and observability needs are evaluated. Staging/production use isolated credentials, data stores, caches, storage, and service identities. Build once and promote the same immutable artifact.

## Recovery

Clone the child repository, use the pinned Node/toolchain, install with the lockfile, restore secrets only from approved external stores, and run the complete verification gate. Future database recovery uses committed Supabase config/migrations/seed plus approved backups; generated credentials and local data remain machine-local.

## Context maintenance

Update durable context only for changed product rules, architecture, security boundaries, stack decisions, or workflows. Temporary task progress belongs in the project queue/issues. Save reusable procedures as skills, not context.
