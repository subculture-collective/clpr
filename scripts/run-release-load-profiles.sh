#!/usr/bin/env bash

set -euo pipefail

readonly K6_IMAGE='grafana/k6:0.57.0@sha256:70af91f86cd8e142e0544a4edaf79835a80033f71974b92edd5ac36fd4442a7b'
mode="${1:---dry-run}"

if [[ "$mode" == "--dry-run" ]]; then
    cat <<'EOF'
Would run staging-only k6 profiles against one fixture revision:
  baseline: 5 total VUs for 1 minute
  stress: ramp to 25, then 75 total VUs, then recover
  soak: 10 total VUs for 30 minutes
Execution requires TARGET_ENVIRONMENT=staging, ALLOW_STAGING_LOAD=true,
STAGING_BASE_URL, CLIP_ID, STAGING_AUTH_TOKEN, STAGING_ADMIN_TOKEN, and an
operator-owned EVIDENCE_OUTPUT_DIR.
EOF
    exit 0
fi
[[ "$mode" == "--execute" ]] || { echo "usage: $0 [--dry-run|--execute]" >&2; exit 2; }
[[ "${TARGET_ENVIRONMENT:-}" == "staging" ]] || { echo "TARGET_ENVIRONMENT must equal staging" >&2; exit 1; }
[[ "${ALLOW_STAGING_LOAD:-}" == "true" ]] || { echo "ALLOW_STAGING_LOAD=true is required" >&2; exit 1; }
: "${STAGING_BASE_URL:?STAGING_BASE_URL is required}"
: "${CLIP_ID:?CLIP_ID is required}"
: "${STAGING_AUTH_TOKEN:?STAGING_AUTH_TOKEN is required}"
: "${STAGING_ADMIN_TOKEN:?STAGING_ADMIN_TOKEN is required}"
: "${EVIDENCE_OUTPUT_DIR:?EVIDENCE_OUTPUT_DIR is required}"

python3 - "$STAGING_BASE_URL" <<'PY'
import sys
from urllib.parse import urlparse
url = urlparse(sys.argv[1])
if url.scheme != "https" or not url.hostname:
    raise SystemExit("STAGING_BASE_URL must be an absolute HTTPS URL")
if url.hostname.lower() in {"clpr.tv", "www.clpr.tv"}:
    raise SystemExit("production clpr.tv is forbidden for release load tests")
PY

mkdir -p "$EVIDENCE_OUTPUT_DIR"
[[ -d "$EVIDENCE_OUTPUT_DIR" && -w "$EVIDENCE_OUTPUT_DIR" ]] \
    || { echo "EVIDENCE_OUTPUT_DIR must be writable" >&2; exit 1; }

for profile in baseline stress soak; do
    echo "Running $profile release profile"
    docker run --rm \
        -v "$PWD/backend/tests/load/release.js:/release.js:ro" \
        -v "$EVIDENCE_OUTPUT_DIR:/evidence" \
        -e "BASE_URL=$STAGING_BASE_URL" \
        -e "CLIP_ID=$CLIP_ID" \
        -e "AUTH_TOKEN=$STAGING_AUTH_TOKEN" \
        -e "ADMIN_TOKEN=$STAGING_ADMIN_TOKEN" \
        -e "PROFILE=$profile" \
        -e "SOAK_DURATION=${SOAK_DURATION:-30m}" \
        "$K6_IMAGE" run --summary-export "/evidence/k6-$profile-summary.json" /release.js
done
