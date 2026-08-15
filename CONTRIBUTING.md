# Contributing to Juntly

## Before work

1. Read `AGENTS.md` and `context/README.md` in order.
2. Confirm the product flow and any conflicts with `context/product-reference.md`.
3. Sync CodeGraph and inspect affected boundaries.
4. Write/approve a bounded plan for significant work.

## Branches

- Branch from `development` using a cohesive `feature/<slug>`, `fix/<slug>`, `docs/<slug>`, or `chore/<slug>` name.
- Open a PR to `development`.
- Promote reviewed immutable output through `development → staging → main`.
- Use squash merging and delete merged feature branches.
- Never force-push environment branches.

## Development

- Follow RED → GREEN → REFACTOR for behavior.
- Keep business rules in Go application/domain boundaries when the backend exists.
- Keep frontend integration aligned with generated OpenAPI contracts.
- Do not hardcode user-facing copy, categories, fees, plans, currencies, or contact details in components.
- Do not introduce unused dependencies, services, or placeholder modules.
- Update context in the same PR when durable decisions change.

## Verification

Run the affected focused tests first, then every configured project gate. Frontend baseline:

```bash
cd frontend
npm ci
npm run format:check
npm test
npm run lint
npm run typecheck
npm run build
npm audit --audit-level=high
```

Exercise relevant production HTTP/browser flows. Audit the staged manifest and diff for secrets, private data, build output, generated caches, and unrelated files before commit.

## Security and privacy

Never commit/print `.env` values, tokens, cookies, sessions, credentials, identity documents, payment data, private contact details, exact service addresses, messages/proposals, production records, or raw private chats. Report suspected exposure privately to the repository owner rather than opening a public issue with sensitive detail.
