#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

fail() {
    echo "blue/green contract: $*" >&2
    exit 1
}

bash -n scripts/blue-green-deploy.sh

for forbidden in 'sed -i' 'image prune' 'docker tag' 'IMAGE_TAG' ':latest'; do
    if grep -Fq "$forbidden" scripts/blue-green-deploy.sh docker-compose.blue-green.yml; then
        fail "mutable/destructive deployment operation found: $forbidden"
    fi
done

grep -Fq "BAKE_SECONDS=\"\${BAKE_SECONDS:-3600}\"" scripts/blue-green-deploy.sh \
    || fail "default 60-minute bake window is missing"
grep -Fq "perform_rollback \"\$ORIGINAL_SLOT\" \"\$TARGET_SLOT\"" scripts/blue-green-deploy.sh \
    || fail "automatic rollback path is missing"
grep -Fq 'flock -n 9' scripts/blue-green-deploy.sh \
    || fail "exclusive deployment locking is missing"
grep -Fq 'trap on_exit EXIT' scripts/blue-green-deploy.sh \
    || fail "failure-triggered rollback trap is missing"
# shellcheck disable=SC2016 # Match the literal shell source expression.
grep -Fq 'health_check "$restore_slot"' scripts/blue-green-deploy.sh \
    || fail "rollback readiness check is missing"
grep -Fq 'X-CLPR-Canary' deploy/Caddyfile.blue-green.template \
    || fail "authenticated canary header matcher is missing"
grep -Fq 'header_up -X-CLPR-Canary' deploy/Caddyfile.blue-green.template \
    || fail "canary credential is not removed before proxying"
grep -Fq '/etc/caddy/runtime/Caddyfile' docker-compose.blue-green.yml \
    || fail "Caddy is not configured from external runtime state"

python3 - <<'PY'
import pathlib
import yaml

compose = yaml.safe_load(pathlib.Path("docker-compose.blue-green.yml").read_text())
services = compose["services"]
crawlers = [name for name in services if "crawler" in name]
if crawlers != ["crawler"]:
    raise SystemExit(f"expected exactly one shared crawler service, got {crawlers}")

crawler = services["crawler"]
expected = {
    "user": "10001:10001",
    "read_only": True,
    "cap_drop": ["ALL"],
    "security_opt": ["no-new-privileges:true"],
    "pids_limit": 128,
    "mem_limit": "256m",
    "cpus": 0.5,
}
for key, value in expected.items():
    if crawler.get(key) != value:
        raise SystemExit(f"crawler {key} is {crawler.get(key)!r}, expected {value!r}")
if not any(item.startswith("/tmp:rw,noexec,nosuid") for item in crawler.get("tmpfs", [])):
    raise SystemExit("crawler requires a constrained /tmp tmpfs")

for name, service in services.items():
    image = service.get("image")
    if image and "@" not in image:
        raise SystemExit(f"service {name} has a mutable image reference: {image}")

caddy = services["caddy"]
for key, expected_value in {
    "read_only": True,
    "cap_drop": ["ALL"],
    "security_opt": ["no-new-privileges:true"],
    "pids_limit": 128,
    "mem_limit": "256m",
    "cpus": 0.5,
}.items():
    if caddy.get(key) != expected_value:
        raise SystemExit(f"caddy {key} is not hardened")
if any(str(port).startswith("2019:") for port in caddy.get("ports", [])):
    raise SystemExit("Caddy admin API must not be published on the host")
PY

for dockerfile in backend/Dockerfile backend/Dockerfile.crawler frontend/Dockerfile; do
    grep -Fq 'org.opencontainers.image.revision' "$dockerfile" \
        || fail "$dockerfile lacks an OCI revision label"
    grep -Fq 'org.opencontainers.image.source' "$dockerfile" \
        || fail "$dockerfile lacks an OCI source label"
done

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
token='contract-token-0123456789-ABCDEFGH'

CANARY_TOKEN="$token" \
TEMPLATE_FILE="$repo_root/deploy/Caddyfile.blue-green.template" \
    bash scripts/blue-green-deploy.sh render blue green "$tmp_dir/Caddyfile"

grep -Fq 'clpr-backend-blue:8080' "$tmp_dir/Caddyfile" \
    || fail "rendered active backend is not blue"
grep -Fq 'clpr-frontend-green:8080' "$tmp_dir/Caddyfile" \
    || fail "rendered canary frontend is not green"
grep -Fq "$token" "$tmp_dir/Caddyfile" \
    || fail "rendered canary token is missing"
grep -Fq 'X-CLPR-Served-Slot "green"' "$tmp_dir/Caddyfile" \
    || fail "rendered canary slot assertion is missing"
if grep -Eq '__ACTIVE_|__CANARY_' "$tmp_dir/Caddyfile"; then
    fail "rendered runtime config contains unresolved placeholders"
fi

digest='sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
cat > "$tmp_dir/.env" <<EOF
BACKEND_BLUE_DIGEST=$digest
FRONTEND_BLUE_DIGEST=$digest
BACKEND_GREEN_DIGEST=$digest
FRONTEND_GREEN_DIGEST=$digest
CRAWLER_DIGEST=$digest
POSTGRES_PASSWORD=contract-postgres
JWT_PRIVATE_KEY=contract-jwt
MFA_ENCRYPTION_KEY=contract-mfa
OPERATIONAL_AUTH_TOKEN=contract-operations
TWITCH_CLIENT_ID=contract-twitch
TWITCH_CLIENT_SECRET=contract-twitch-secret
OPENSEARCH_URL=https://opensearch.invalid
CLPR_RUNTIME_DIR=$tmp_dir/runtime
EOF
mkdir -p "$tmp_dir/runtime"

docker compose --env-file "$tmp_dir/.env" \
    -f docker-compose.blue-green.yml --profile green config > "$tmp_dir/compose.yml"

python3 - "$tmp_dir/compose.yml" <<'PY'
import pathlib
import sys
import yaml

compose = yaml.safe_load(pathlib.Path(sys.argv[1]).read_text())
for name, service in compose["services"].items():
    # Locally built PostgreSQL has a Compose-generated cache name; every
    # registry-delivered runtime image must be immutable.
    if "build" in service:
        continue
    image = service.get("image", "")
    if "@sha256:" not in image:
        raise SystemExit(f"rendered service {name} has a mutable image: {image}")
PY

[[ "$(grep -Ec '^  crawler:$' "$tmp_dir/compose.yml")" -eq 1 ]] \
    || fail "rendered compose does not contain exactly one crawler"

echo "blue/green contract: digest, canary, bake, rollback, and compose checks passed"
