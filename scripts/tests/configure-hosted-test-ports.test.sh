#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

fail() {
  echo "hosted test port contract: $*" >&2
  exit 1
}

# shellcheck disable=SC1091 # The file under test intentionally exports values.
source scripts/configure-hosted-test-ports.sh 3573

[[ "$TEST_RUN_SLOT" == 3573 ]] || fail "unexpected run slot"
[[ "$COMPOSE_PROJECT_NAME" == clpr-test-3573 ]] || fail "project is not run-scoped"
[[ "$TEST_DATABASE_PORT" == 23573 ]] || fail "database port is not run-scoped"
[[ "$TEST_REDIS_PORT" == 33573 ]] || fail "Redis port is not run-scoped"
[[ "$TEST_OPENSEARCH_PORT" == 43573 ]] || fail "OpenSearch port is not run-scoped"
[[ "$TEST_OPENSEARCH_METRICS_PORT" == 53573 ]] || fail "OpenSearch metrics port is not run-scoped"

python3 - <<'PY' &
import socket
import time

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
    listener.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    listener.bind(("127.0.0.1", 23573))
    listener.listen()
    time.sleep(30)
PY
listener_pid=$!
trap 'kill "$listener_pid" 2>/dev/null || true; wait "$listener_pid" 2>/dev/null || true' EXIT

for _ in {1..50}; do
  if ! port_is_available 23573 2>/dev/null; then
    break
  fi
  sleep 0.02
done
! port_is_available 23573 2>/dev/null || fail "failed to reserve the collision-test port"

collision_database_port="$({
  # shellcheck disable=SC1091 # Exercise a second run in an isolated subshell.
  source scripts/configure-hosted-test-ports.sh 3573
  printf '%s' "$TEST_DATABASE_PORT"
})"
[[ "$collision_database_port" == 23574 ]] \
  || fail "database port selection did not skip an occupied host port"

kill "$listener_pid"
wait "$listener_pid" 2>/dev/null || true
trap - EXIT

rendered="$(docker compose -f docker-compose.test.yml config)"
for expected in \
  'name: clpr-test-3573' \
  'name: clpr-hosted-tests' \
  'external: true' \
  'container_name: clpr-postgres-test-3573' \
  'container_name: clpr-redis-test-3573' \
  'container_name: clpr-opensearch-test-3573' \
  'published: "23573"' \
  'published: "33573"' \
  'published: "43573"' \
  'published: "53573"'; do
  grep -Fq "$expected" <<<"$rendered" || fail "compose output is missing: $expected"
done

for workflow in \
  .gitea/workflows/release-gates.yml \
  .gitea/workflows/source-convergence.yml \
  .gitea/workflows/immutable-candidate.yml; do
  ! grep -Fq 'clpr-hosted-heavy' "$workflow" \
    || fail "$workflow still relies on unsupported hosted locking"
  grep -Fq 'source scripts/configure-hosted-test-ports.sh' "$workflow" \
    || fail "$workflow does not isolate its browser test services"
done

grep -A4 '^  image-security:' .gitea/workflows/release-gates.yml \
  | grep -Fq 'needs: browser' \
  || fail "release image security must wait for the browser gate"
grep -Fq 'workflow_dispatch:' .gitea/workflows/source-convergence.yml \
  || fail "source convergence must remain manually dispatchable"
grep -Fq '  pull_request:' .gitea/workflows/source-convergence.yml \
  || fail "source convergence must remain a required pull-request gate"
! grep -Eq '^  push:' .gitea/workflows/source-convergence.yml \
  || fail "source convergence must not duplicate post-merge release work"

echo "hosted test port contract passed"
