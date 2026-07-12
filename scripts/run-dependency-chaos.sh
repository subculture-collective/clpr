#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
compose=(docker compose -f "$repo_root/docker-compose.test.yml")
api_port=${CHAOS_API_PORT:-18089}
api_origin="http://127.0.0.1:${api_port}"
log_file="$repo_root/.tmp/backend-chaos.log"
backend_pid=""

cleanup() {
    "${compose[@]}" start postgres-test redis-test opensearch-test >/dev/null 2>&1 || true
    if [[ -n "$backend_pid" ]]; then
        kill "$backend_pid" 2>/dev/null || true
        wait "$backend_pid" 2>/dev/null || true
    fi
}
trap cleanup EXIT INT TERM

mkdir -p "$repo_root/.tmp"
(
    cd "$repo_root/backend"
    set -a
    # shellcheck disable=SC1091
    source .env.test
    set +a
    PORT="$api_port" GIN_MODE=debug BASE_URL=http://127.0.0.1:5173 \
        DB_HOST=localhost DB_PORT=5437 DB_USER=clpr DB_PASSWORD=clpr_password DB_NAME=clpr_test \
        REDIS_HOST=localhost REDIS_PORT=6380 OPENSEARCH_URL=http://localhost:9201 \
        CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173 RATE_LIMIT_WHITELIST_IPS=127.0.0.1 \
        FEATURE_ANALYTICS=false go run ./cmd/api
) >"$log_file" 2>&1 &
backend_pid=$!

wait_for_status() {
    local expected=$1
    for _ in {1..45}; do
        status=$(curl -sS -o /tmp/clpr-chaos-response.json -w '%{http_code}' "$api_origin/health/ready" || true)
        [[ "$status" == "$expected" ]] && return 0
        sleep 1
    done
    echo "Readiness never returned $expected; last response:" >&2
    cat /tmp/clpr-chaos-response.json >&2 || true
    tail -100 "$log_file" >&2
    return 1
}

wait_for_status 200
grep -q '"degraded_dependencies":\[\]' /tmp/clpr-chaos-response.json

"${compose[@]}" stop opensearch-test >/dev/null
wait_for_status 200
grep -q '"degraded_dependencies":\["opensearch"\]' /tmp/clpr-chaos-response.json
"${compose[@]}" start opensearch-test >/dev/null

"${compose[@]}" stop redis-test >/dev/null
wait_for_status 503
grep -q '"status":"not ready"' /tmp/clpr-chaos-response.json
"${compose[@]}" start redis-test >/dev/null
wait_for_status 200

"${compose[@]}" stop postgres-test >/dev/null
wait_for_status 503
grep -q '"status":"not ready"' /tmp/clpr-chaos-response.json
"${compose[@]}" start postgres-test >/dev/null
wait_for_status 200

echo "Dependency chaos contract passed: optional OpenSearch degrades; required Redis/PostgreSQL fail readiness and recover"
