# Juntly Project Context

`product-reference.md` is the primary product and technical source of truth. The remaining files make that reference easier to apply; they do not replace or weaken it.

Load in order:

1. `product-reference.md` — complete product and technical requirements.
2. `product.md` — bounded product scope, users, outcomes, and launch evidence.
3. `stack.md` — verified technologies, versions, and approved deviations.
4. `architecture.md` — deployables, module boundaries, data flow, and invariants.
5. `ui.md` — UI system, responsive experience, i18n, and accessibility rules.
6. `security.md` — data classes, trust boundaries, privacy, and controls.
7. `constraints.md` — product and engineering rules implementation may not violate.
8. `workflows.md` — branches, tests, delivery, evidence, and recovery.
9. `decisions.md` — durable project-specific architectural decisions.

Keep this directory sanitized, durable, concise, and synchronized with implementation. Product-rule changes require explicit owner direction and a decision record. Temporary progress belongs in the project work queue/issues, not durable context. Secrets, raw messages, production data, identity documents, and private client material stay outside Git.
