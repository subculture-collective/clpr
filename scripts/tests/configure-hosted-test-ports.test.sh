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

rendered="$(docker compose -f docker-compose.test.yml config)"
for expected in \
  'name: clpr-test-3573' \
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
  grep -Fq 'source scripts/configure-hosted-test-ports.sh' "$workflow" \
    || fail "$workflow does not isolate its browser test services"
done

echo "hosted test port contract passed"
