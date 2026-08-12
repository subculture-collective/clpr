#!/usr/bin/env bash

set -euo pipefail

script=scripts/run-real-e2e.sh

bash -n "$script"

grep -Fq 'npx playwright test --project=real-chromium --workers=1' "$script" \
    || { echo "single-worker native Chromium invocation is missing" >&2; exit 1; }
grep -Fq 'npx playwright test --project=real-firefox --workers=1 && npx playwright test --project=real-webkit --workers=1' "$script" \
    || { echo "single-worker containerized Firefox/WebKit invocation is missing" >&2; exit 1; }
grep -Fq 'docker cp "$repo_root/." "$browser_container:/work"' "$script" \
    || { echo "nested-Docker workspace copy is missing" >&2; exit 1; }
grep -Fq 'docker create --network "container:$(hostname)"' "$script" \
    || { echo "browser container must share the runner network namespace" >&2; exit 1; }

if grep -Fq 'npx playwright test --project=real-chromium --project=real-firefox --project=real-webkit' "$script"; then
    echo "real browser engines must not execute concurrently" >&2
    exit 1
fi

echo "real E2E runner contract: hosted browser engines execute sequentially"
