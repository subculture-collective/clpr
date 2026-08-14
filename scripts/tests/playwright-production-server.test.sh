#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

fail() {
  echo "Playwright production server contract: $*" >&2
  exit 1
}

config=frontend/playwright.config.ts
grep -Fq "npm run build && npm run preview" "$config" \
  || fail "browser gates do not exercise a production build"
grep -Fq -- "--strictPort" "$config" \
  || fail "browser gates may silently attach to an unrelated server"
grep -Fq "reuseExistingServer: false" "$config" \
  || fail "browser gates may reuse stale local application state"
grep -Fq "serviceWorkers: 'block'" "$config" \
  || fail "production service workers may bypass Playwright API routes"
if grep -Fq "npm run dev" "$config"; then
  fail "browser gates still depend on the cold-transform development server"
fi

echo "Playwright production server contract passed"
