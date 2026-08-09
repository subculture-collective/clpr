#!/usr/bin/env bash

set -euo pipefail

login="${TEA_LOGIN:-patrickfanella}"
repository="${GITEA_REPOSITORY:-subculture-collective/clpr}"

command -v tea >/dev/null || { echo "tea CLI is required" >&2; exit 1; }
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
tea api -l "$login" '/version' > "$tmp_dir/version.json"
tea api -l "$login" -r "$repository" '/repos/{owner}/{repo}/actions/runners' > "$tmp_dir/runners.json"

python3 - "$tmp_dir/version.json" "$tmp_dir/runners.json" <<'PY'
import json
import pathlib
import sys

version = json.loads(pathlib.Path(sys.argv[1]).read_text()).get("version", "0.0.0")
parts = tuple(int(part) for part in version.split(".")[:2])
if parts < (1, 27):
    raise SystemExit(f"Gitea 1.27+ is required, found {version}")
runners = json.loads(pathlib.Path(sys.argv[2]).read_text()).get("runners", [])
online = [runner for runner in runners if runner.get("status") == "online"]
if not online:
    raise SystemExit("No repository-visible online Gitea Actions runner is available")
print(f"Gitea {version} Actions runtime passed with {len(online)} online runner(s)")
PY
