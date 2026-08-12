#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
resolver="$repo_root/scripts/configure-test-service-host.sh"
fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT

fail() {
    echo "test service host contract: $*" >&2
    exit 1
}

cat >"$fixture_dir/ip" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' 'default via 172.24.0.1 dev eth0'
EOF
chmod +x "$fixture_dir/ip"

local_values="$({
    unset GITHUB_ACTIONS GITEA_ACTIONS CI TEST_SERVICE_HOST TEST_DATABASE_HOST
    unset TEST_REDIS_HOST TEST_OPENSEARCH_URL OPENSEARCH_URL
    # shellcheck disable=SC1090
    source "$resolver"
    printf '%s|%s|%s|%s\n' \
        "$TEST_SERVICE_HOST" "$TEST_DATABASE_HOST" "$TEST_REDIS_HOST" \
        "$TEST_OPENSEARCH_URL"
})"
[[ "$local_values" == 'localhost|localhost|localhost|http://localhost:9201' ]] || \
    fail "local defaults changed: $local_values"

hosted_values="$(PATH="$fixture_dir:$PATH" GITHUB_ACTIONS=true bash -c '
    source "$1"
    printf "%s|%s|%s|%s\n" \
        "$TEST_SERVICE_HOST" "$TEST_DATABASE_HOST" "$TEST_REDIS_HOST" \
        "$TEST_OPENSEARCH_URL"
' _ "$resolver")"
[[ "$hosted_values" == '172.24.0.1|172.24.0.1|172.24.0.1|http://172.24.0.1:9201' ]] || \
    fail "hosted gateway was not propagated: $hosted_values"

override_values="$(
    TEST_SERVICE_HOST=services.internal \
    TEST_DATABASE_HOST=postgres.internal \
    TEST_REDIS_HOST=redis.internal \
    TEST_OPENSEARCH_URL=https://search.internal \
    bash -c '
        source "$1"
        printf "%s|%s|%s|%s\n" \
            "$TEST_SERVICE_HOST" "$TEST_DATABASE_HOST" "$TEST_REDIS_HOST" \
            "$TEST_OPENSEARCH_URL"
    ' _ "$resolver"
)"
[[ "$override_values" == 'services.internal|postgres.internal|redis.internal|https://search.internal' ]] || \
    fail "explicit overrides were not preserved: $override_values"

echo "test service host contract passed"
