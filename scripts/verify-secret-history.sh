#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
gitleaks_bin=${GITLEAKS_BIN:-gitleaks}
report_dir=$(mktemp -d)
trap 'rm -rf "$report_dir"' EXIT

cd "$repo_root"

if [[ "${SECRET_TREE_ONLY:-0}" != 1 ]]; then
  "$gitleaks_bin" git \
    --redact=100 \
    --report-format json \
    --report-path "$report_dir/history.json" \
    --config .gitleaks.toml \
    .
fi

# Scan exactly the candidate source set: tracked files plus non-ignored
# untracked files. Gitleaks directory mode does not honor .gitignore by itself,
# so scanning the worktree directly would inspect this gate's own .tmp output.
tree_dir="$report_dir/tree"
mkdir -p "$tree_dir"
while IFS= read -r -d '' path; do
  [[ -f "$path" ]] || continue
  cp --parents --preserve=mode,timestamps -- "$path" "$tree_dir"
done < <(git ls-files --cached --others --exclude-standard -z)

"$gitleaks_bin" dir \
  --redact=100 \
  --report-format json \
  --report-path "$report_dir/tree.json" \
  --config .gitleaks.toml \
  --gitleaks-ignore-path "$repo_root/.gitleaksignore" \
  "$tree_dir"

node scripts/audit-jwt-secret-history.mjs
echo "Secret history and current tree are clean."
