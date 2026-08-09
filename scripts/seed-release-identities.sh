#!/usr/bin/env bash

set -euo pipefail

mode="${1:---dry-run}"
run_id="${RELEASE_RUN_ID:-release-dry-run}"

[[ "$run_id" =~ ^[a-z0-9-]{6,40}$ ]] || {
    echo "RELEASE_RUN_ID must contain 6-40 lowercase letters, digits, or hyphens" >&2
    exit 1
}

if [[ "$mode" == "--dry-run" ]]; then
    printf 'Would seed disposable roles for run %s: user, moderator, administrator\n' "$run_id"
    printf 'Execution requires TARGET_ENVIRONMENT=staging, ALLOW_DISPOSABLE_IDENTITY_MUTATION=true, and PostgreSQL credentials.\n'
    exit 0
fi
[[ "$mode" == "--execute" ]] || { echo "usage: $0 [--dry-run|--execute]" >&2; exit 2; }
[[ "${TARGET_ENVIRONMENT:-}" == "staging" ]] || { echo "TARGET_ENVIRONMENT must equal staging" >&2; exit 1; }
[[ "${ALLOW_DISPOSABLE_IDENTITY_MUTATION:-}" == "true" ]] || { echo "ALLOW_DISPOSABLE_IDENTITY_MUTATION=true is required" >&2; exit 1; }
: "${POSTGRES_HOST:?POSTGRES_HOST is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
[[ "${STAGING_DATABASE_CONFIRMATION:-}" == "$POSTGRES_DB" ]] \
    || { echo "STAGING_DATABASE_CONFIRMATION must exactly match POSTGRES_DB" >&2; exit 1; }

PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -v ON_ERROR_STOP=1 -v run_id="$run_id" \
    -h "$POSTGRES_HOST" -p "${POSTGRES_PORT:-5432}" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
INSERT INTO users (twitch_id, username, display_name, email, role, account_type, account_status)
VALUES
  ('release-' || :'run_id' || '-user', 'release_' || replace(:'run_id', '-', '_') || '_user', 'Release User', NULL, 'user', 'member', 'active'),
  ('release-' || :'run_id' || '-moderator', 'release_' || replace(:'run_id', '-', '_') || '_moderator', 'Release Moderator', NULL, 'moderator', 'member', 'active'),
  ('release-' || :'run_id' || '-administrator', 'release_' || replace(:'run_id', '-', '_') || '_administrator', 'Release Administrator', NULL, 'admin', 'member', 'active')
ON CONFLICT (twitch_id) DO UPDATE SET
  role = EXCLUDED.role,
  account_status = 'active',
  is_banned = false,
  updated_at = NOW();
SQL

echo "Disposable staging roles seeded for $run_id; OAuth storage states remain operator-supplied protected inputs."
