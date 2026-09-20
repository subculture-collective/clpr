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

port_is_available() {
  python3 - "$1" <<'PY'
import socket
import sys

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
    listener.bind(("0.0.0.0", int(sys.argv[1])))
PY
}

select_available_port() {
  local base="$1"
  local offset candidate_slot candidate_port

  for ((offset = 0; offset < 10000; offset++)); do
    candidate_slot=$(((run_slot + offset) % 10000))
    candidate_port=$((base + candidate_slot))
    if port_is_available "$candidate_port" 2>/dev/null; then
      printf '%s\n' "$candidate_port"
      return 0
    fi
  done

  echo "No available hosted test port in range $base-$((base + 9999))" >&2
  return 1
}

# The run number keeps concurrent workflows in separate ranges. Probe from that
# deterministic slot so a long-lived host listener cannot invalidate the run.
TEST_DATABASE_PORT="$(select_available_port 20000)"
TEST_REDIS_PORT="$(select_available_port 30000)"
TEST_OPENSEARCH_PORT="$(select_available_port 40000)"
TEST_OPENSEARCH_METRICS_PORT="$(select_available_port 50000)"
export TEST_DATABASE_PORT TEST_REDIS_PORT TEST_OPENSEARCH_PORT
export TEST_OPENSEARCH_METRICS_PORT

# Compose needs a user-defined network for its service aliases. Keep one
# shared network across runs so a long-lived runner consumes only one subnet.
# The create/inspect fallback is safe when two jobs initialize concurrently.
if ! docker network inspect clpr-hosted-tests >/dev/null 2>&1; then
  docker network create clpr-hosted-tests >/dev/null 2>&1 \
    || docker network inspect clpr-hosted-tests >/dev/null
fi
