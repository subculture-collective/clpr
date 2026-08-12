#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
test_service_host=${TEST_SERVICE_HOST:-localhost}
api_port=${E2E_API_PORT:-18088}
api_origin="http://127.0.0.1:${api_port}"
seed_clip_id="00000000-0000-4000-8000-000000000001"
seed_user_id="00000000-0000-4000-8000-000000000002"
playwright_image="${PLAYWRIGHT_IMAGE:-mcr.microsoft.com/playwright:v1.57.0-noble@sha256:3bed4b1a12f2338642f3d8cba28e291deef3c66bd4a964bbeb3e57bbff511dbd}"
log_file="$repo_root/.tmp/backend-e2e.log"
pid_file="$repo_root/.tmp/backend-e2e.pid"
api_binary="$repo_root/.tmp/backend-e2e-api"
browser_container=""

cleanup() {
    if [[ -n "$browser_container" ]]; then
        docker rm -f "$browser_container" >/dev/null 2>&1 || true
    fi
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

(cd "$repo_root/backend" && go build -o "$api_binary" ./cmd/api)

# Seed one deterministic public clip in the disposable test database so the
# browser suite exercises a real repository read instead of only static pages.
docker compose -f "$repo_root/docker-compose.test.yml" exec -T postgres-test \
    psql --username=clpr --dbname=clpr_test --set=ON_ERROR_STOP=1 <<SQL
INSERT INTO users (
    id, twitch_id, username, display_name, role, account_type, account_status
) VALUES (
    '$seed_user_id', 'clpr-release-tester', 'clpr-release-tester',
    'CLPR Release Tester', 'user', 'member', 'active'
)
ON CONFLICT (id) DO UPDATE SET
    is_banned = false,
    account_status = 'active';

INSERT INTO clips (
    id, twitch_clip_id, twitch_clip_url, embed_url, title,
    creator_name, creator_id, broadcaster_name, broadcaster_id,
    game_id, game_name, language, duration, view_count, created_at,
    submitted_by_user_id
) VALUES (
    '$seed_clip_id', 'clpr-release-smoke',
    'https://clips.twitch.tv/clpr-release-smoke',
    'https://clips.twitch.tv/embed?clip=clpr-release-smoke',
    'CLPR release smoke clip', 'Release Tester', 'release-tester',
    'Release Channel', 'release-channel', 'release-game',
    'Release Readiness', 'en', 30, 42, NOW(), '$seed_user_id'
)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    is_removed = false,
    is_hidden = false,
    submitted_by_user_id = EXCLUDED.submitted_by_user_id;
SQL

(
    cd "$repo_root/backend"
    set -a
    # shellcheck disable=SC1091
    source .env.test
    set +a
    PORT="$api_port" GIN_MODE=debug BASE_URL=http://127.0.0.1:5173 \
        DB_HOST="${TEST_DATABASE_HOST:-$test_service_host}" DB_PORT="${TEST_DATABASE_PORT:-5437}" DB_USER=clpr \
        DB_PASSWORD=clpr_password DB_NAME=clpr_test \
        REDIS_HOST="${TEST_REDIS_HOST:-$test_service_host}" REDIS_PORT="${TEST_REDIS_PORT:-6380}" \
        OPENSEARCH_URL="${TEST_OPENSEARCH_URL:-http://$test_service_host:9201}" \
        CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173 \
        RATE_LIMIT_WHITELIST_IPS=127.0.0.1 FEATURE_ANALYTICS=false \
        "$api_binary"
) >"$log_file" 2>&1 &
echo $! >"$pid_file"

for _ in {1..60}; do
    if curl -fsS "$api_origin/health/live" >/dev/null 2>&1; then
        (
            cd "$repo_root/backend"
            set -a
            # shellcheck disable=SC1091
            source .env.test
            set +a
            DB_HOST="${TEST_DATABASE_HOST:-$test_service_host}" DB_PORT="${TEST_DATABASE_PORT:-5437}" DB_USER=clpr \
                DB_PASSWORD=clpr_password DB_NAME=clpr_test \
                OPENSEARCH_URL="${TEST_OPENSEARCH_URL:-http://$test_service_host:9201}" \
                go run ./cmd/backfill-search -batch 100
        )
        curl -fsS -X POST "${TEST_OPENSEARCH_URL:-http://$test_service_host:9201}/clips/_refresh" >/dev/null
        cd "$repo_root/frontend"

        # Keep native Chromium coverage, but use one worker so the 4 GiB
        # hosted runner does not overload Vite or the browser process.
        PLAYWRIGHT_API_BASE_URL="$api_origin" \
            PLAYWRIGHT_SEED_CLIP_ID="$seed_clip_id" \
            PLAYWRIGHT_SEED_GAME_ID="release-game" \
            npx playwright test --project=real-chromium --workers=1

        # Native Firefox cannot create its sandbox namespace in the hosted
        # runner container and the minimal runner image omits libpci. Copy the
        # exact checked-out workspace into the digest-pinned Playwright image;
        # a host bind mount is invalid because the Docker daemon is outside the
        # runner container.
        browser_container="$(docker create --network host --ipc=host \
            --user "$(id -u):$(id -g)" \
            -e HOME=/tmp \
            -e "CI=${CI:-}" \
            -e PLAYWRIGHT_BROWSERS_PATH=/ms-playwright \
            -e "PLAYWRIGHT_API_BASE_URL=$api_origin" \
            -e "PLAYWRIGHT_SEED_CLIP_ID=$seed_clip_id" \
            -e PLAYWRIGHT_SEED_GAME_ID=release-game \
            -w /work/frontend \
            "$playwright_image" \
            sh -lc 'npx playwright test --project=real-firefox --workers=1 && npx playwright test --project=real-webkit --workers=1')"
        docker cp "$repo_root/." "$browser_container:/work"
        docker start --attach "$browser_container"
        docker rm "$browser_container" >/dev/null
        browser_container=""
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
