# Incident and privacy operations

## Incident handling

1. Assign an incident lead and timestamp the start.
2. Contain access or disable the affected capability without deleting evidence.
3. Rotate exposed credentials through their provider and restart only the affected services.
4. Preserve sanitized request IDs, audit records, image digests, and deployment metadata. Do not copy private messages or credentials into tickets.
5. Determine affected data subjects and legal notification obligations with the data controller.
6. Restore service from verified artifacts, monitor, and write a blameless follow-up with concrete controls.

## Data-subject requests

Authenticate the requester through Clerk and reconcile the durable internal user before processing access, correction, portability, or deletion requests. Record request date, scope, decision, completion date, and legal-retention exceptions in restricted operational records. Export only the requester's data through an encrypted transfer. Never send exports through logs or public object URLs.

Deletion is a reviewed workflow, not an unauthenticated endpoint. Revoke sessions, remove or anonymize marketplace profile data where legally permitted, preserve required financial/security/audit records under documented retention, and verify that backups expire through normal retention rather than being selectively edited.

## Privacy defaults

- Contact channels remain encrypted at rest and are revealed only through the authorized endpoint.
- Conversations, proposals, bookings, reports, and private locations remain participant- or moderator-scoped.
- Administrative metrics are aggregate-only and capped.
- Analytics must not include message bodies, contact channels, private locations, Clerk tokens, or stable cross-site identifiers.
- Notification preferences are honored before in-app or email-outbox creation.

## Quarterly exercises

Run a restore drill, credential-rotation drill, administrator-access review, data-subject-request tabletop, and incident communication exercise. Record only sanitized outcomes and owners.
