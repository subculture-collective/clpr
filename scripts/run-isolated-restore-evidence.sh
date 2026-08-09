#!/usr/bin/env bash

set -euo pipefail

mode="${1:---dry-run}"

if [[ "$mode" == "--dry-run" ]]; then
    cat <<'EOF'
Would restore a protected backup into an explicitly isolated PostgreSQL host,
verify representative tables, require RPO <= 900s and RTO <= 3600s, and call
an isolated restored-application smoke URL. Execution requires:
  TARGET_ENVIRONMENT=isolated
  ALLOW_ISOLATED_RESTORE=true
  RESTORE_TARGET_ID=ephemeral-...
  RESTORE_TARGET_MANIFEST=/protected/provisioned-target.json
  PRODUCTION_POSTGRES_HOSTS=prod-db.example,...
  BACKUP_FILE_OVERRIDE and BACKUP_TIMESTAMP
  PostgreSQL credentials and APPLICATION_SMOKE_URL
EOF
    exit 0
fi
[[ "$mode" == "--execute" ]] || { echo "usage: $0 [--dry-run|--execute]" >&2; exit 2; }
[[ "${TARGET_ENVIRONMENT:-}" == "isolated" ]] || { echo "TARGET_ENVIRONMENT must equal isolated" >&2; exit 1; }
[[ "${ALLOW_ISOLATED_RESTORE:-}" == "true" ]] || { echo "ALLOW_ISOLATED_RESTORE=true is required" >&2; exit 1; }
[[ "${RESTORE_TARGET_ID:-}" == ephemeral-* ]] || { echo "RESTORE_TARGET_ID must start with ephemeral-" >&2; exit 1; }
: "${BACKUP_FILE_OVERRIDE:?BACKUP_FILE_OVERRIDE is required}"
: "${BACKUP_TIMESTAMP:?BACKUP_TIMESTAMP is required}"
: "${POSTGRES_HOST:?POSTGRES_HOST is required}"
: "${POSTGRES_PORT:?POSTGRES_PORT is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${APPLICATION_SMOKE_URL:?APPLICATION_SMOKE_URL is required}"
: "${EVIDENCE_OUTPUT_DIR:?EVIDENCE_OUTPUT_DIR is required}"
: "${RESTORE_TARGET_MANIFEST:?RESTORE_TARGET_MANIFEST is required}"
: "${PRODUCTION_POSTGRES_HOSTS:?PRODUCTION_POSTGRES_HOSTS is required}"
[[ -f "$BACKUP_FILE_OVERRIDE" ]] || { echo "protected backup input does not exist" >&2; exit 1; }
[[ -f "$RESTORE_TARGET_MANIFEST" ]] || { echo "provisioned target manifest does not exist" >&2; exit 1; }

IFS=$'\t' read -r expected_fingerprint expected_role minimum_clips minimum_users < <(
python3 - "$RESTORE_TARGET_MANIFEST" "$RESTORE_TARGET_ID" "$POSTGRES_HOST" "$POSTGRES_PORT" "$POSTGRES_USER" "$APPLICATION_SMOKE_URL" "$PRODUCTION_POSTGRES_HOSTS" <<'PY'
import json
import pathlib
import re
import sys
from urllib.parse import urlparse

manifest_path, target_id, host, port, user, smoke_url, production_hosts = sys.argv[1:]
manifest = json.loads(pathlib.Path(manifest_path).read_text())
if manifest.get("schema_version") != 1:
    raise SystemExit("restore target manifest schema_version must equal 1")
if manifest.get("target_id") != target_id:
    raise SystemExit("RESTORE_TARGET_ID does not match the provisioned target manifest")
if manifest.get("postgres_host", "").lower() != host.lower() or str(manifest.get("postgres_port")) != port:
    raise SystemExit("PostgreSQL endpoint does not match the provisioned target manifest")
role = manifest.get("restricted_role", "")
if role != user or not re.fullmatch(r"[a-z_][a-z0-9_]{0,62}", role):
    raise SystemExit("POSTGRES_USER must equal the manifest's restricted_role")
fingerprint = str(manifest.get("postgres_system_identifier", ""))
if not re.fullmatch(r"[0-9]{10,32}", fingerprint):
    raise SystemExit("manifest postgres_system_identifier is invalid")
denied = {value.strip().lower().rstrip(".") for value in production_hosts.split(",") if value.strip()}
denied.update({"clpr.tv", "www.clpr.tv", "postgres", "clpr-postgres"})
if host.lower().rstrip(".") in denied:
    raise SystemExit("production PostgreSQL endpoints are forbidden for restore drills")
minimum_rows = manifest.get("minimum_rows", {})
clips, users = minimum_rows.get("clips"), minimum_rows.get("users")
if not isinstance(clips, int) or clips < 1 or not isinstance(users, int) or users < 1:
    raise SystemExit("manifest minimum_rows must require at least one clip and user")
smoke = urlparse(smoke_url)
if smoke.scheme not in {"http", "https"} or not smoke.hostname:
    raise SystemExit("APPLICATION_SMOKE_URL must be an absolute URL")
if smoke.hostname.lower() != str(manifest.get("application_smoke_host", "")).lower():
    raise SystemExit("application smoke host does not match the provisioned target manifest")
if smoke.hostname.lower() in {"clpr.tv", "www.clpr.tv"}:
    raise SystemExit("production clpr.tv is forbidden for isolated restore smoke")
print(f"{fingerprint}\t{role}\t{clips}\t{users}")
PY
)

command -v psql >/dev/null || { echo "psql is required" >&2; exit 1; }
target_identity="$(PGPASSWORD="$POSTGRES_PASSWORD" psql -X -A -t -F '|' \
    -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
    -c "SELECT system_identifier::text, current_user, rolsuper, rolcreatedb FROM pg_control_system(), pg_roles WHERE rolname = current_user;")"
IFS='|' read -r actual_fingerprint actual_role actual_superuser actual_createdb <<<"$target_identity"
[[ "$actual_fingerprint" == "$expected_fingerprint" ]] || { echo "PostgreSQL system identifier does not match provisioned target" >&2; exit 1; }
[[ "$actual_role" == "$expected_role" && "$actual_superuser" == f && "$actual_createdb" == t ]] \
    || { echo "restore role must be the provisioned non-superuser CREATEDB role" >&2; exit 1; }

mkdir -p "$EVIDENCE_OUTPUT_DIR"
RELEASE_EVIDENCE_MODE=true \
RESTORE_TARGET_VALIDATED=true \
RESTORE_MIN_CLIPS="$minimum_clips" RESTORE_MIN_USERS="$minimum_users" \
RTO_TARGET_SECONDS=3600 RPO_TARGET_SECONDS=900 \
DRILL_LOG="$EVIDENCE_OUTPUT_DIR/restore-drill.log" \
    bash scripts/restore-drill.sh
