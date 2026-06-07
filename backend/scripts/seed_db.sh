#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
MIGRATIONS_DIR="${BACKEND_DIR}/migrations"

ENV_FILE="${ENV_FILE:-${BACKEND_DIR}/.env}"
BASE_SEED_FILE="${BASE_SEED_FILE:-${MIGRATIONS_DIR}/seed.sql}"
COMMENTS_SEED_FILE="${COMMENTS_SEED_FILE:-${MIGRATIONS_DIR}/seed_comments.sql}"
LOAD_TEST_SEED_FILE="${LOAD_TEST_SEED_FILE:-${MIGRATIONS_DIR}/seed_load_test.sql}"
RUN_LOAD_TEST=false
RUN_COMMENTS_AUTO=true
RUN_COMMENTS_EXPLICIT=false

usage() {
    cat <<'EOF'
Usage: seed_db.sh [options]

Options:
    --env-file PATH         Path to .env file to load (default: backend/.env)
    --seed-file PATH        Override base seed file (default: migrations/seed.sql)
    --comments              Explicitly apply comments/nesting seed (migrations/seed_comments.sql)
    --no-comments           Skip comments/nesting seed even if file exists
    --comments-file PATH    Override comments seed file
    --load-test             Also apply load-test seed data (migrations/seed_load_test.sql)
    --load-test-file PATH   Override load-test seed file
  -h, --help             Show this help message

Notes:
- Uses DB_* values from the loaded .env if present; falls back to local defaults.
- Requires psql installed and database reachable.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --env-file)
            ENV_FILE="$2"
            shift
            ;;
        --seed-file)
            BASE_SEED_FILE="$2"
            shift
            ;;
        --comments)
            RUN_COMMENTS_EXPLICIT=true
            RUN_COMMENTS_AUTO=false
            ;;
        --no-comments)
            RUN_COMMENTS_AUTO=false
            RUN_COMMENTS_EXPLICIT=false
            ;;
        --comments-file)
            COMMENTS_SEED_FILE="$2"
            shift
            ;;
        --load-test)
            RUN_LOAD_TEST=true
            ;;
        --load-test-file)
            LOAD_TEST_SEED_FILE="$2"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "[ERROR] Unknown option: $1" >&2
            usage
            exit 1
            ;;
    esac
    shift
done

if ! command -v psql >/dev/null 2>&1; then
    echo "[ERROR] psql is required but not installed or not in PATH" >&2
    exit 1
fi

if [[ -f "${ENV_FILE}" ]]; then
    echo "[INFO] Loading environment from ${ENV_FILE}"
    set -a
    # shellcheck source=/dev/null
    source "${ENV_FILE}"
    set +a
else
    echo "[WARN] Env file not found at ${ENV_FILE}; using existing environment/defaults"
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5436}"
DB_USER="${DB_USER:-clpr}"
DB_PASSWORD="${DB_PASSWORD:-clpr_password}"
DB_NAME="${DB_NAME:-clpr_db}"
DB_URL="${DB_URL:-}" # Optional Postgres connection string

# psql safety flags: stop on first error, ignore ~/.psqlrc, single-transaction per file
PSQL_FLAGS=(-v ON_ERROR_STOP=1 -X -1)

run_seed() {
    local file="$1"
    if [[ ! -f "${file}" ]]; then
        echo "[ERROR] Seed file not found: ${file}" >&2
        exit 1
    fi

    echo "[INFO] Seeding database with ${file}"
    if [[ -n "${DB_URL}" ]]; then
        PGPASSWORD="${DB_PASSWORD}" psql "${DB_URL}" "${PSQL_FLAGS[@]}" -f "${file}"
    else
        PGPASSWORD="${DB_PASSWORD}" psql \
            -h "${DB_HOST}" \
            -p "${DB_PORT}" \
            -U "${DB_USER}" \
            -d "${DB_NAME}" \
            "${PSQL_FLAGS[@]}" \
            -f "${file}"
    fi
}

run_seed "${BASE_SEED_FILE}"

if [[ "${RUN_COMMENTS_EXPLICIT}" == true ]]; then
    run_seed "${COMMENTS_SEED_FILE}"
elif [[ "${RUN_COMMENTS_AUTO}" == true && -f "${COMMENTS_SEED_FILE}" ]]; then
    echo "[INFO] Detected comments seed at ${COMMENTS_SEED_FILE}; applying"
    run_seed "${COMMENTS_SEED_FILE}"
else
    echo "[INFO] Skipping comments seed"
fi

if [[ "${RUN_LOAD_TEST}" == true ]]; then
    run_seed "${LOAD_TEST_SEED_FILE}"
fi

echo "[INFO] Database seed complete"
