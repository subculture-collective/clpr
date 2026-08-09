#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
gitleaks_bin=${GITLEAKS_BIN:-gitleaks}
fixture_dir=$(mktemp -d)
trap 'rm -rf "$fixture_dir"' EXIT

if grep -Ev '^[[:space:]]*(#|$)' "$repo_root/.gitleaksignore" \
  | grep -Ev '^[0-9a-f]{40}:' >/dev/null; then
  echo "gitleaks policy regression: commit-independent fingerprints are forbidden" >&2
  exit 1
fi

# The only curl exemption is the documented literal placeholder.
printf '%s\n' 'curl -H "Authorization: Bearer YOUR_TOKEN" https://example.invalid' > "$fixture_dir/placeholder.txt"
"$gitleaks_bin" dir --no-banner --config "$repo_root/.gitleaks.toml" "$fixture_dir" >/dev/null

# Assemble a synthetic detector fixture at runtime so the policy test itself is
# not allowlisted. A new default-rule finding must still fail the scan.
printf 'AWS_ACCESS_KEY_ID=AKIA%s\n' 'ABCDEFGHIJKLMNOP' > "$fixture_dir/new-secret.txt"
if "$gitleaks_bin" dir --no-banner --config "$repo_root/.gitleaks.toml" "$fixture_dir" >/dev/null 2>&1; then
  echo "gitleaks policy regression: a newly introduced synthetic secret passed" >&2
  exit 1
fi

# A detector match replacing a historical fixture at the same path and line
# must not inherit that old commit's disposition.
replacement_repo="$fixture_dir/replacement-repository"
mkdir -p "$replacement_repo/docs/testing"
cp "$repo_root/.gitleaks.toml" "$repo_root/.gitleaksignore" "$replacement_repo/"
{
  for _ in $(seq 1 33); do printf '\n'; done
  printf 'STRIPE_SECRET_KEY=sk_test_51%s%s\n' \
    'A1B2C3D4E5F6G7H8I9J0K1L2M3N4' 'O5P6Q7R8S9T0U1V2W3X4Y5Z6'
} > "$replacement_repo/docs/testing/stripe-ci-secrets.md"
git -C "$replacement_repo" init -q
git -C "$replacement_repo" add .
git -C "$replacement_repo" -c user.name=policy-test -c user.email=policy-test@example.invalid \
  commit -qm 'replacement fixture'
if "$gitleaks_bin" git --no-banner --config "$replacement_repo/.gitleaks.toml" "$replacement_repo" >/dev/null 2>&1; then
  echo "gitleaks policy regression: a replacement secret inherited a historical disposition" >&2
  exit 1
fi

echo "Gitleaks allowlist remains narrow and fail-closed."
