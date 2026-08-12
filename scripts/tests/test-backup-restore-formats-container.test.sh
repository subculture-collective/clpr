#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT

cat >"$fixture_dir/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

for argument in "$@"; do
    [[ "$argument" != -v ]] || {
        echo "host bind mounts are forbidden in hosted fallback" >&2
        exit 1
    }
done
tar -tf - >"$CLPR_CAPTURED_ARCHIVE"
printf '%s\n' "$*" >"$CLPR_CAPTURED_ARGUMENTS"
EOF
chmod +x "$fixture_dir/docker"

export CLPR_CAPTURED_ARCHIVE="$fixture_dir/archive.txt"
export CLPR_CAPTURED_ARGUMENTS="$fixture_dir/arguments.txt"
PATH="$fixture_dir:$PATH" CLPR_POSTGRES_CLIENT_CONTAINER=1 \
    bash "$repo_root/scripts/test-backup-restore-formats.sh"

for path in \
    scripts/test-backup-restore-formats.sh \
    scripts/restore-drill.sh \
    scripts/validate-backup.sh; do
    grep -Fxq "$path" "$CLPR_CAPTURED_ARCHIVE" || {
        echo "streamed fallback is missing $path" >&2
        exit 1
    }
done

grep -Fq -- '--network host' "$CLPR_CAPTURED_ARGUMENTS"
grep -Fq -- 'postgres:17' "$CLPR_CAPTURED_ARGUMENTS"
echo "backup restore container fallback contract passed"
