#!/usr/bin/env sh
set -eu
: "${DATABASE_URL:?Set the target DATABASE_URL}"
: "${BACKUP_FILE:?Set BACKUP_FILE}"
: "${CONFIRM_RESTORE:?Set CONFIRM_RESTORE to RESTORE}"
if [ "$CONFIRM_RESTORE" != RESTORE ]; then
  printf '%s\n' 'Restore confirmation rejected' >&2
  exit 2
fi
sha256sum --check "$BACKUP_FILE.sha256"
pg_restore --dbname="$DATABASE_URL" --clean --if-exists --no-owner --no-privileges --single-transaction "$BACKUP_FILE"
printf '%s\n' 'Restore completed; run migrations and smoke checks before traffic.'
