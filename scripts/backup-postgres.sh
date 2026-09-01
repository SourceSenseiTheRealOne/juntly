#!/usr/bin/env sh
set -eu
: "${DATABASE_URL:?Set DATABASE_URL}"
backup_dir=${BACKUP_DIR:-./backups}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
umask 077
mkdir -p "$backup_dir"
target="$backup_dir/juntly-$stamp.dump"
pg_dump --dbname="$DATABASE_URL" --format=custom --no-owner --no-privileges --file="$target"
pg_restore --list "$target" >/dev/null
sha256sum "$target" >"$target.sha256"
printf '%s\n' "$target"
