#!/usr/bin/env bash

set -euo pipefail

mode="${1:---dry-run}"
[[ "$mode" == "--dry-run" || "$mode" == "--check" ]] \
    || { echo "usage: $0 [--dry-run|--check]" >&2; exit 2; }

node scripts/validate-operator-journeys.mjs
bash scripts/seed-release-identities.sh --dry-run
bash scripts/run-real-backend-journeys.sh --dry-run
bash scripts/run-stripe-test-evidence.sh --dry-run
bash scripts/run-release-load-profiles.sh --dry-run
bash scripts/run-isolated-restore-evidence.sh --dry-run
bash scripts/run-staging-rollback-evidence.sh --dry-run

if [[ "$mode" == "--dry-run" ]]; then
    cat <<'EOF'
Dry-run contracts passed. This is not release evidence. Hosted controls,
disposable identity state, Stripe test mode, staging load, isolated restore,
rollback, JWT exposure review, key-rotation determination, and the full-history
secret-scan disposition remain fail-closed until protected operator runs supply
passing evidence.
EOF
    exit 0
fi

# The hosted check is intentionally read-only and fails until branch controls
# and secret names are actually configured. It never reads secret values.
bash scripts/verify-gitea-release-controls.sh

for variable in \
    RELEASE_CANDIDATE_MANIFEST GITLEAKS_REDACTED_REPORT JWT_REVIEW_RECORD \
    E2E_USER_STORAGE_STATE E2E_MODERATOR_STORAGE_STATE E2E_ADMIN_STORAGE_STATE; do
    value="${!variable:-}"
    [[ -s "$value" ]] || { echo "$variable must identify a non-empty protected artifact" >&2; exit 1; }
done

echo "Hosted/operator artifact preflight passed; provider and staging harnesses must still execute separately."
