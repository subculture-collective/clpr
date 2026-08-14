#!/usr/bin/env bash

# Resolve test services published by the Docker daemon. Local development uses
# loopback. Gitea/GitHub Actions jobs use a sibling Docker daemon through the
# mounted socket, so published ports live on the job container's host gateway.
if [[ -z "${TEST_SERVICE_HOST:-}" ]]; then
    if [[ "${GITHUB_ACTIONS:-}" == "true" || "${GITEA_ACTIONS:-}" == "true" ]]; then
        TEST_SERVICE_HOST=""
        if command -v ip >/dev/null 2>&1; then
            TEST_SERVICE_HOST="$(ip -4 route show default | awk '/default via / { print $3; exit }')"
        elif command -v python3 >/dev/null 2>&1; then
            TEST_SERVICE_HOST="$(python3 - <<'PY'
import socket
import struct

try:
    with open('/proc/net/route', encoding='ascii') as routes:
        for line in routes:
            fields = line.split()
            if len(fields) >= 3 and fields[1] == '00000000':
                print(socket.inet_ntoa(struct.pack('<L', int(fields[2], 16))))
                break
except (OSError, ValueError):
    pass
PY
)"
        fi
        if [[ ! "$TEST_SERVICE_HOST" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
            echo "Unable to resolve the hosted Docker gateway" >&2
            return 1 2>/dev/null || exit 1
        fi
    else
        TEST_SERVICE_HOST=localhost
    fi
fi

export TEST_SERVICE_HOST
export TEST_DATABASE_HOST="${TEST_DATABASE_HOST:-$TEST_SERVICE_HOST}"
export TEST_REDIS_HOST="${TEST_REDIS_HOST:-$TEST_SERVICE_HOST}"
export TEST_OPENSEARCH_URL="${TEST_OPENSEARCH_URL:-http://$TEST_SERVICE_HOST:${TEST_OPENSEARCH_PORT:-9201}}"
export OPENSEARCH_URL="${OPENSEARCH_URL:-$TEST_OPENSEARCH_URL}"
