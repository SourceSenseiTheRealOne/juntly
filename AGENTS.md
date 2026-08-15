# Juntly Agent Rules

## Start here

Before planning or editing, read `context/README.md` and the linked files in order. `context/product-reference.md` is the primary product and technical source of truth. If a requested change conflicts with it, identify the conflict and obtain an explicit decision; never silently alter business rules.

## Engineering workflow

- Run `codegraph status .`; initialize/sync as required before investigation, use graph-backed exploration/impact, and re-sync after source topology changes. Never commit `.codegraph/`.
- Significant work requires an approved bounded plan and acceptance evidence.
- Behavioral code follows strict RED → GREEN → REFACTOR.
- Keep modules small, typed, layered, and security-first. Do not add speculative infrastructure or empty placeholder packages.
- OpenAPI is the contract authority. Next.js uses a same-origin BFF/proxy; Go remains authoritative for business rules, authorization, ownership, state transitions, idempotency, and persistence.
- Clerk is identity authority; protected Go resources use verified internal user mapping. UI gates are presentation only.
- Never expose private contact details, exact addresses, messages, proposals, identity documents, payment data, credentials, tokens, sessions, or raw private records in public output, metadata, logs, tests, context, or Git.
- User-facing copy belongs in locale resources. pt-PT is default; preserve English and Spanish-ready structure.
- Target WCAG 2.2 AA, mobile-first behavior, low-bandwidth performance, keyboard/focus support, and reduced motion.

## Branches and delivery

The verified initial bootstrap is the one-time direct `main` exception. Thereafter:

```text
feature/* → development → staging → main
```

Use one cohesive branch/PR, squash merge, delete the source branch, and promote the same immutable artifact. Production/main requires SourceSensei approval. Work in this canonical checkout by switching branches; do not create duplicate Juntly folders.

## Verification

Before completion claims, run all configured format, focused/full test, lint, type, build, dependency, runtime HTTP, browser/console, staged-path, and secret/privacy gates relevant to the change. A successful build is not runtime proof. Missing credentials/environments block only their corresponding evidence and are never reinterpreted as success.

## Context maintenance

Update context in the same PR when product rules, architecture, stack, security boundaries, constraints, or workflow change. Temporary task progress belongs in the work queue/issues, not durable context. Secrets/private material remain external and ignored.
