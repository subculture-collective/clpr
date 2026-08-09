#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
if ! command -v psql >/dev/null; then
    exec docker run --rm --network host \
        -v "$repo_root:/repo:ro" \
        -e TEST_DATABASE_HOST -e TEST_DATABASE_PORT -e TEST_DATABASE_USER \
        -e TEST_DATABASE_PASSWORD -e DRILL_LOG=/tmp/clpr-restore-format-contract.log \
        postgres:17 bash /repo/scripts/test-backup-restore-formats.sh
fi
export DRILL_LOG="${DRILL_LOG:-/tmp/clpr-restore-format-contract.log}"
# shellcheck disable=SC1091
source "$repo_root/scripts/restore-drill.sh"
# shellcheck disable=SC1091
source "$repo_root/scripts/validate-backup.sh"

export POSTGRES_HOST="${TEST_DATABASE_HOST:-localhost}"
export POSTGRES_PORT="${TEST_DATABASE_PORT:-5437}"
export POSTGRES_USER="${TEST_DATABASE_USER:-clpr}"
export POSTGRES_PASSWORD="${TEST_DATABASE_PASSWORD:-clpr_password}"

suffix=${RANDOM}_$$
source_db="restore_source_${suffix}"
plain_db="restore_plain_${suffix}"
custom_db="restore_custom_${suffix}"
plain_dump=$(mktemp /tmp/clpr-plain-dump.XXXXXX.sql.gz)
custom_dump=$(mktemp /tmp/clpr-custom-dump.XXXXXX.dump.gz)

cleanup() {
    for database in "$source_db" "$plain_db" "$custom_db"; do
        PGPASSWORD="$POSTGRES_PASSWORD" dropdb --if-exists -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$database" >/dev/null 2>&1 || true
    done
    rm -f "$plain_dump" "$custom_dump"
}
trap cleanup EXIT

for database in "$source_db" "$plain_db" "$custom_db"; do
    PGPASSWORD="$POSTGRES_PASSWORD" createdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$database"
done
PGPASSWORD="$POSTGRES_PASSWORD" psql -v ON_ERROR_STOP=1 -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$source_db" \
    -c "CREATE TABLE restore_contract (id integer PRIMARY KEY, value text NOT NULL); INSERT INTO restore_contract VALUES (1, 'verified');" >/dev/null

PGPASSWORD="$POSTGRES_PASSWORD" pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -Fp "$source_db" | gzip > "$plain_dump"
validate_local_backup_file "$plain_dump"
BACKUP_FILE="$plain_dump" TEST_DB="$plain_db" restore_dump_file >/dev/null

PGPASSWORD="$POSTGRES_PASSWORD" pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -Fc "$source_db" | gzip > "$custom_dump"
validate_local_backup_file "$custom_dump"
BACKUP_FILE="$custom_dump" TEST_DB="$custom_db" restore_dump_file >/dev/null

for database in "$plain_db" "$custom_db"; do
    value=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -At -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$database" -c "SELECT value FROM restore_contract WHERE id = 1")
    [[ "$value" == "verified" ]]
done

echo "Plain-SQL and custom-format gzip restore contracts passed"
