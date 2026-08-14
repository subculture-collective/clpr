#!/usr/bin/env bash

# Scope Docker Compose resources and published ports to one hosted workflow run.
# This file is sourced so the exports remain available to convergence and its
# teardown trap.
hosted_run_number="${1:-${GITEA_RUN_NUMBER:-${GITHUB_RUN_NUMBER:-}}}"
if [[ ! "$hosted_run_number" =~ ^[0-9]+$ ]]; then
  echo "A numeric hosted workflow run number is required" >&2
  return 1 2>/dev/null || exit 1
fi

run_slot=$((10#$hosted_run_number % 10000))
export TEST_RUN_SLOT="$run_slot"
export COMPOSE_PROJECT_NAME="clpr-test-$run_slot"
export TEST_DATABASE_PORT="$((20000 + run_slot))"
export TEST_REDIS_PORT="$((30000 + run_slot))"
export TEST_OPENSEARCH_PORT="$((40000 + run_slot))"
export TEST_OPENSEARCH_METRICS_PORT="$((50000 + run_slot))"
