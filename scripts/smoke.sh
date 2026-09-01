#!/usr/bin/env sh
set -eu
: "${JUNTLY_BASE_URL:?Set JUNTLY_BASE_URL}"
base=${JUNTLY_BASE_URL%/}
request_id="smoke_$(date -u +%s)"
curl --fail --silent --show-error --max-time 10 -H "X-Request-ID: $request_id" "$base/api/v1/health" >/dev/null
curl --fail --silent --show-error --max-time 10 -H "X-Request-ID: $request_id" "$base/api/v1/ready" >/dev/null
printf '%s\n' 'Public API health and readiness checks passed.'
