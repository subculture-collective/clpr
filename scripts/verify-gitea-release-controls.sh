#!/usr/bin/env bash

set -euo pipefail

login="${TEA_LOGIN:-patrickfanella}"
repository="${GITEA_REPOSITORY:-subculture-collective/clpr}"
manifest="${OPERATOR_JOURNEY_MANIFEST:-config/release/operator-journeys.json}"

command -v tea >/dev/null || { echo "tea CLI is required" >&2; exit 1; }
command -v python3 >/dev/null || { echo "python3 is required" >&2; exit 1; }

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

# These are read-only API calls. Secret values are neither requested nor
# returned by Gitea; only configured secret names are compared.
tea api -l "$login" -r "$repository" \
    '/repos/{owner}/{repo}/branch_protections/main' > "$tmp_dir/protection.json"
tea api -l "$login" -r "$repository" \
    '/repos/{owner}/{repo}/actions/secrets' > "$tmp_dir/secrets.json"

python3 - "$tmp_dir/protection.json" "$tmp_dir/secrets.json" "$manifest" <<'PY'
import json
import pathlib
import sys

protection = json.loads(pathlib.Path(sys.argv[1]).read_text())
secret_payload = json.loads(pathlib.Path(sys.argv[2]).read_text())
manifest = json.loads(pathlib.Path(sys.argv[3]).read_text())
secrets = secret_payload if isinstance(secret_payload, list) else secret_payload.get("secrets", [])
secret_names = {item["name"] for item in secrets}
failures = []

requirements = {
    "branch_name": "main",
    "enable_push": True,
    "enable_status_check": True,
    "dismiss_stale_approvals": True,
    "block_on_rejected_reviews": True,
    "block_on_outdated_branch": True,
    "block_admin_merge_override": True,
}
for key, expected in requirements.items():
    if protection.get(key) != expected:
        failures.append(f"main protection {key} must equal {expected!r}")
if int(protection.get("required_approvals") or 0) < 1:
    failures.append("main protection must require at least one approval")
required_contexts = {
    "Complete local-equivalent source gate",
    "Secret history gate",
    "Documentation and OpenAPI gate",
    "Browser gate (Chromium, Firefox, WebKit)",
    "Image scan and SBOM gate",
    "Operator and recovery source contracts",
}
configured_contexts = set(protection.get("status_check_contexts") or [])
for context in sorted(required_contexts - configured_contexts):
    failures.append(f"main protection is missing required status context: {context}")

for name in manifest.get("required_secret_names", []):
    if name not in secret_names:
        failures.append(f"missing Gitea Actions secret name: {name}")

if failures:
    print("Gitea release controls failed:\n- " + "\n- ".join(failures), file=sys.stderr)
    raise SystemExit(1)
print(f"Gitea release controls passed; {len(secret_names)} required secret names are configured")
PY
