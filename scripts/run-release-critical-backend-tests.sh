#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
results=${RELEASE_CRITICAL_RESULTS:-"$repo_root/.tmp/release-critical-tests.json"}
mkdir -p "$(dirname "$results")"

cd "$repo_root/backend"
set -a
# shellcheck disable=SC1091
source .env.test
set +a

set +e
go test -count=1 -json ./internal/handlers ./internal/middleware ./pkg/jwt | tee "$results" >/dev/null
test_status=${PIPESTATUS[0]}
set -e
(( test_status == 0 )) || exit "$test_status"

go test -count=1 -tags=integration ./internal/repository -run TestWebhookRetryClaimsAreExclusiveAcrossWorkers
bash "$repo_root/scripts/test-backup-restore-formats.sh"

if grep -q '"Action":"skip"' "$results"; then
    echo "Release-critical backend suite skipped tests:" >&2
    grep '"Action":"skip"' "$results" >&2
    exit 1
fi

echo "Release-critical backend suites passed without skips"
