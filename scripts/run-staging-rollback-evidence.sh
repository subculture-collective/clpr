#!/usr/bin/env bash

set -euo pipefail

mode="${1:---dry-run}"

if [[ "$mode" == "--dry-run" ]]; then
    cat <<'EOF'
Would execute a staging blue/green canary, graceful drain, forced rollback,
and post-rollback smoke using exact candidate digests. Execution requires
TARGET_ENVIRONMENT=staging, ALLOW_STAGING_ROLLBACK_DRILL=true,
STAGING_BASE_URL, STAGING_DEPLOY_DIR, CANARY_TOKEN, and protected digest inputs.
No production hostname is accepted.
EOF
    exit 0
fi
[[ "$mode" == "--execute" ]] || { echo "usage: $0 [--dry-run|--execute]" >&2; exit 2; }
[[ "${TARGET_ENVIRONMENT:-}" == "staging" ]] || { echo "TARGET_ENVIRONMENT must equal staging" >&2; exit 1; }
[[ "${ALLOW_STAGING_ROLLBACK_DRILL:-}" == "true" ]] || { echo "ALLOW_STAGING_ROLLBACK_DRILL=true is required" >&2; exit 1; }
: "${STAGING_BASE_URL:?STAGING_BASE_URL is required}"
: "${STAGING_DEPLOY_DIR:?STAGING_DEPLOY_DIR is required}"
: "${CANARY_TOKEN:?CANARY_TOKEN is required}"

python3 - "$STAGING_BASE_URL" <<'PY'
import sys
from urllib.parse import urlparse
url = urlparse(sys.argv[1])
if url.scheme != "https" or not url.hostname:
    raise SystemExit("STAGING_BASE_URL must be an absolute HTTPS URL")
if url.hostname.lower() in {"clpr.tv", "www.clpr.tv"}:
    raise SystemExit("production clpr.tv is forbidden for rollback drills")
PY

DEPLOY_DIR="$STAGING_DEPLOY_DIR" CANARY_BASE_URL="$STAGING_BASE_URL" \
BAKE_SECONDS="${DRILL_BAKE_SECONDS:-60}" \
    bash scripts/blue-green-deploy.sh deploy
DEPLOY_DIR="$STAGING_DEPLOY_DIR" CANARY_BASE_URL="$STAGING_BASE_URL" \
    bash scripts/blue-green-deploy.sh rollback
curl --fail --silent --show-error "$STAGING_BASE_URL/health/ready" >/dev/null
echo "Staging canary, drain, forced rollback, and post-rollback smoke passed"
