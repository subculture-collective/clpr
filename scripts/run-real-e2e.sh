#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
api_port=${E2E_API_PORT:-18088}
api_origin="http://127.0.0.1:${api_port}"
log_file="$repo_root/.tmp/backend-e2e.log"
pid_file="$repo_root/.tmp/backend-e2e.pid"

cleanup() {
    if [[ -f "$pid_file" ]]; then
        kill "$(cat "$pid_file")" 2>/dev/null || true
        wait "$(cat "$pid_file")" 2>/dev/null || true
        rm -f "$pid_file"
    fi
}
trap cleanup EXIT INT TERM

if curl -fsS "$api_origin/health/live" >/dev/null 2>&1; then
    echo "Port $api_port already serves a healthy API; refusing to test an unowned process" >&2
    exit 1
fi

mkdir -p "$repo_root/.tmp"
(
    cd "$repo_root/backend"
    set -a
    # shellcheck disable=SC1091
    source .env.test
    set +a
    PORT="$api_port" GIN_MODE=debug BASE_URL=http://127.0.0.1:5173 \
        DB_HOST=localhost DB_PORT=5437 DB_USER=clpr \
        DB_PASSWORD=clpr_password DB_NAME=clpr_test \
        REDIS_HOST=localhost REDIS_PORT=6380 \
        OPENSEARCH_URL=http://localhost:9201 \
        CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173 \
        RATE_LIMIT_WHITELIST_IPS=127.0.0.1 FEATURE_ANALYTICS=false \
        go run ./cmd/api
) >"$log_file" 2>&1 &
echo $! >"$pid_file"

for _ in {1..60}; do
    if curl -fsS "$api_origin/health/live" >/dev/null 2>&1; then
        cd "$repo_root/frontend"
        PLAYWRIGHT_API_BASE_URL="$api_origin" npm run test:e2e:real
        exit 0
    fi
    if ! kill -0 "$(cat "$pid_file")" 2>/dev/null; then
        echo "Backend exited before becoming ready" >&2
        tail -100 "$log_file" >&2
        exit 1
    fi
    sleep 1
done

echo "Backend did not become ready within 60 seconds" >&2
tail -100 "$log_file" >&2
exit 1
