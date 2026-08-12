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

node <<'NODE'
const fs = require('node:fs');
const yaml = require('js-yaml');

const compose = yaml.load(fs.readFileSync('docker-compose.blue-green.yml', 'utf8'));
const services = compose.services;
const crawlers = Object.keys(services).filter((name) => name.includes('crawler'));
if (JSON.stringify(crawlers) !== JSON.stringify(['crawler'])) {
    throw new Error(`expected exactly one shared crawler service, got ${JSON.stringify(crawlers)}`);
}

const crawler = services.crawler;
const expected = {
    user: '10001:10001',
    read_only: true,
    cap_drop: ['ALL'],
    security_opt: ['no-new-privileges:true'],
    pids_limit: 128,
    mem_limit: '256m',
    cpus: 0.5,
};
for (const [key, value] of Object.entries(expected)) {
    if (JSON.stringify(crawler[key]) !== JSON.stringify(value)) {
        throw new Error(`crawler ${key} is ${JSON.stringify(crawler[key])}, expected ${JSON.stringify(value)}`);
    }
}
if (!(crawler.tmpfs || []).some((item) => item.startsWith('/tmp:rw,noexec,nosuid'))) {
    throw new Error('crawler requires a constrained /tmp tmpfs');
}

for (const [name, service] of Object.entries(services)) {
    const image = service.image;
    if (image && !image.includes('@')) {
        throw new Error(`service ${name} has a mutable image reference: ${image}`);
    }
}

const caddy = services.caddy;
for (const [key, value] of Object.entries({
    read_only: true,
    cap_drop: ['ALL'],
    security_opt: ['no-new-privileges:true'],
    pids_limit: 128,
    mem_limit: '256m',
    cpus: 0.5,
})) {
    if (JSON.stringify(caddy[key]) !== JSON.stringify(value)) {
        throw new Error(`caddy ${key} is not hardened`);
    }
}
if ((caddy.ports || []).some((port) => String(port).startsWith('2019:'))) {
    throw new Error('Caddy admin API must not be published on the host');
}
NODE

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

node - "$tmp_dir/compose.yml" <<'NODE'
const fs = require('node:fs');
const yaml = require('js-yaml');

const compose = yaml.load(fs.readFileSync(process.argv[2], 'utf8'));
for (const [name, service] of Object.entries(compose.services)) {
    // Locally built PostgreSQL has a Compose-generated cache name; every
    // registry-delivered runtime image must be immutable.
    if ('build' in service) continue;
    const image = service.image || '';
    if (!image.includes('@sha256:')) {
        throw new Error(`rendered service ${name} has a mutable image: ${image}`);
    }
}
NODE

[[ "$(grep -Ec '^  crawler:$' "$tmp_dir/compose.yml")" -eq 1 ]] \
    || fail "rendered compose does not contain exactly one crawler"

echo "blue/green contract: digest, canary, bake, rollback, and compose checks passed"
