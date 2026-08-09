#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
gitleaks_bin=${GITLEAKS_BIN:-gitleaks}
report_dir=$(mktemp -d)
trap 'rm -rf "$report_dir"' EXIT

cd "$repo_root"

"$gitleaks_bin" git \
  --redact=100 \
  --report-format json \
  --report-path "$report_dir/history.json" \
  --config .gitleaks.toml \
  .

"$gitleaks_bin" dir \
  --redact=100 \
  --report-format json \
  --report-path "$report_dir/tree.json" \
  --config .gitleaks.toml \
  .

node scripts/audit-jwt-secret-history.mjs
echo "Secret history and current tree are clean."
