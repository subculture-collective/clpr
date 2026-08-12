#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

scripts=(
    scripts/release-operator-preflight.sh
    scripts/run-isolated-restore-evidence.sh
    scripts/run-real-backend-journeys.sh
    scripts/run-release-load-profiles.sh
    scripts/run-staging-rollback-evidence.sh
    scripts/run-stripe-test-evidence.sh
    scripts/seed-release-identities.sh
    scripts/verify-gitea-release-controls.sh
    scripts/restore-drill.sh
)
for script in "${scripts[@]}"; do bash -n "$script"; done

node scripts/validate-operator-journeys.mjs
bash scripts/release-operator-preflight.sh --dry-run >/dev/null

if STRIPE_MODE=live STRIPE_TEST_SECRET_KEY=sk_live_forbidden \
    bash scripts/run-stripe-test-evidence.sh --execute >/dev/null 2>&1; then
    echo "Stripe harness accepted live mode/key" >&2
    exit 1
fi

tmp_dir="$(mktemp -d)"
k6_container=""
cleanup() {
    [[ -z "$k6_container" ]] || docker rm -f "$k6_container" >/dev/null 2>&1 || true
    rm -rf "$tmp_dir"
}
trap cleanup EXIT
if TARGET_ENVIRONMENT=staging ALLOW_STAGING_LOAD=true \
    STAGING_BASE_URL=https://clpr.tv CLIP_ID=fixture \
    STAGING_AUTH_TOKEN=redacted STAGING_ADMIN_TOKEN=redacted \
    EVIDENCE_OUTPUT_DIR="$tmp_dir" \
    bash scripts/run-release-load-profiles.sh --execute >/dev/null 2>&1; then
    echo "Load harness accepted the production hostname" >&2
    exit 1
fi

if TARGET_ENVIRONMENT=production ALLOW_DISPOSABLE_IDENTITY_MUTATION=true \
    RELEASE_RUN_ID=release-contract \
    bash scripts/seed-release-identities.sh --execute >/dev/null 2>&1; then
    echo "Identity seeder accepted production" >&2
    exit 1
fi

if TARGET_ENVIRONMENT=staging STAGING_BASE_URL=https://clpr.tv \
    bash scripts/run-real-backend-journeys.sh --execute >/dev/null 2>&1; then
    echo "Browser journey harness accepted the production hostname" >&2
    exit 1
fi

python3 - "$tmp_dir/restore-target.json" <<'PY'
import json
import pathlib
import sys
pathlib.Path(sys.argv[1]).write_text(json.dumps({
    "schema_version": 1,
    "target_id": "ephemeral-contract",
    "postgres_host": "prod-db.example",
    "postgres_port": 5432,
    "postgres_system_identifier": "1234567890123456789",
    "restricted_role": "restore_drill",
    "application_smoke_host": "restore-app.example",
    "minimum_rows": {"clips": 1, "users": 1},
}))
PY
touch "$tmp_dir/backup.sql.gz"
restore_error="$(TARGET_ENVIRONMENT=isolated ALLOW_ISOLATED_RESTORE=true \
    RESTORE_TARGET_ID=ephemeral-contract \
    RESTORE_TARGET_MANIFEST="$tmp_dir/restore-target.json" \
    PRODUCTION_POSTGRES_HOSTS=prod-db.example \
    BACKUP_FILE_OVERRIDE="$tmp_dir/backup.sql.gz" BACKUP_TIMESTAMP=2026-08-09T00:00:00Z \
    POSTGRES_HOST=prod-db.example POSTGRES_PORT=5432 POSTGRES_USER=restore_drill \
    POSTGRES_PASSWORD=redacted APPLICATION_SMOKE_URL=https://restore-app.example/health \
    EVIDENCE_OUTPUT_DIR="$tmp_dir/evidence" \
    bash scripts/run-isolated-restore-evidence.sh --execute 2>&1 || true)"
grep -Fq 'production PostgreSQL endpoints are forbidden' <<<"$restore_error" \
    || { echo "Restore harness did not reject a production database endpoint" >&2; exit 1; }

unsafe_prefix_error="$(TEST_DB_PREFIX='unsafe;drop' DRILL_LOG="$tmp_dir/restore.log" \
    bash scripts/restore-drill.sh 2>&1 || true)"
grep -Fq 'safe lowercase PostgreSQL identifier' <<<"$unsafe_prefix_error" \
    || { echo "Restore drill accepted an unsafe database identifier prefix" >&2; exit 1; }

for required in \
    'postgres_system_identifier' \
    'RESTORE_TARGET_VALIDATED=true' \
    'validate_migrations' \
    'Backup timestamp is in the future' \
    'Full recovery RTO'; do
    grep -Fq "$required" scripts/run-isolated-restore-evidence.sh scripts/restore-drill.sh \
        || { echo "Restore evidence contract is missing: $required" >&2; exit 1; }
done

python3 - <<'PY'
import json
import pathlib
import yaml

for path in pathlib.Path("release-evidence/templates").glob("*.json"):
    record = json.loads(path.read_text())
    if record.get("status") != "not_executed":
        raise SystemExit(f"{path} must remain explicitly non-passing")
    for key, value in record.items():
        if isinstance(value, bool) and value is not False:
            raise SystemExit(f"{path}: {key} must default false")

yaml.safe_load(pathlib.Path(".gitea/workflows/operator-preflight.yml").read_text())
PY

for profile in baseline stress soak; do
    k6_container="$(docker create \
        grafana/k6:0.57.0@sha256:70af91f86cd8e142e0544a4edaf79835a80033f71974b92edd5ac36fd4442a7b \
        inspect -e "PROFILE=$profile" --execution-requirements /release.js)"
    docker cp "$repo_root/backend/tests/load/release.js" "$k6_container:/release.js"
    docker start --attach "$k6_container" > "$tmp_dir/k6-$profile.json"
    docker rm "$k6_container" >/dev/null
    k6_container=""
done

python3 - "$tmp_dir" <<'PY'
import json
import pathlib
import sys

root = pathlib.Path(sys.argv[1])
baseline = json.loads((root / "k6-baseline.json").read_text())["scenarios"]["release_traffic"]
stress = json.loads((root / "k6-stress.json").read_text())["scenarios"]["release_traffic"]
soak = json.loads((root / "k6-soak.json").read_text())["scenarios"]["release_traffic"]
if (baseline["vus"], baseline["duration"]) != (5, "1m0s"):
    raise SystemExit("baseline profile must be 5 total VUs for 1 minute")
if [stage["target"] for stage in stress["stages"]] != [25, 75, 0]:
    raise SystemExit("stress profile must ramp to 25, 75, then recover to 0")
if (soak["vus"], soak["duration"]) != (10, "30m0s"):
    raise SystemExit("soak profile must be 10 total VUs for 30 minutes")
PY

echo "operator preflight contracts passed"
