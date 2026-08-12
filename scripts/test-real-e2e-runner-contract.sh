#!/usr/bin/env bash

set -euo pipefail

script=scripts/run-real-e2e.sh

bash -n "$script"

grep -Fq 'for project in real-chromium real-firefox real-webkit; do' "$script" \
    || { echo "real browser engines must execute sequentially on the hosted runner" >&2; exit 1; }
grep -Fq 'npx playwright test --project="$project" --workers=1' "$script" \
    || { echo "single-worker sequential browser project invocation is missing" >&2; exit 1; }

if grep -Fq 'npx playwright test --project=real-chromium --project=real-firefox --project=real-webkit' "$script"; then
    echo "real browser engines must not execute concurrently" >&2
    exit 1
fi

echo "real E2E runner contract: hosted browser engines execute sequentially"
