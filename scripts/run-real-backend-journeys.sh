#!/usr/bin/env bash

set -euo pipefail

mode="${1:---dry-run}"
manifest="${OPERATOR_JOURNEY_MANIFEST:-config/release/operator-journeys.json}"
node scripts/validate-operator-journeys.mjs "$manifest"

if [[ "$mode" == "--dry-run" ]]; then
    cat <<'EOF'
Would execute all required operator journeys in real Chromium, Firefox, and
WebKit using disposable user, moderator, and administrator storage states.
Execution requires TARGET_ENVIRONMENT=staging, STAGING_BASE_URL, three
E2E_*_STORAGE_STATE protected files, and an executable OPERATOR_JOURNEY_RUNNER.
EOF
    exit 0
fi
[[ "$mode" == "--execute" ]] || { echo "usage: $0 [--dry-run|--execute]" >&2; exit 2; }
[[ "${TARGET_ENVIRONMENT:-}" == "staging" ]] || { echo "TARGET_ENVIRONMENT must equal staging" >&2; exit 1; }
: "${STAGING_BASE_URL:?STAGING_BASE_URL is required}"
python3 - "$STAGING_BASE_URL" <<'PY'
import sys
from urllib.parse import urlparse
url = urlparse(sys.argv[1])
if url.scheme != "https" or not url.hostname:
    raise SystemExit("STAGING_BASE_URL must be an absolute HTTPS URL")
if url.hostname.lower().rstrip(".") in {"clpr.tv", "www.clpr.tv"}:
    raise SystemExit("production clpr.tv is forbidden for operator browser journeys")
PY
for variable in E2E_USER_STORAGE_STATE E2E_MODERATOR_STORAGE_STATE E2E_ADMIN_STORAGE_STATE; do
    value="${!variable:-}"
    [[ -f "$value" ]] || { echo "$variable must name a protected storage-state file" >&2; exit 1; }
done
[[ -x "${OPERATOR_JOURNEY_RUNNER:-}" ]] || { echo "OPERATOR_JOURNEY_RUNNER must be an executable protected runner" >&2; exit 1; }
"$OPERATOR_JOURNEY_RUNNER" "$manifest"
