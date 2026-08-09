#!/usr/bin/env bash
set -euo pipefail

output=$(cd frontend && npm run test:e2e:list)
printf '%s\n' "$output"

total=$(sed -n 's/^Total: \([0-9][0-9]*\) tests\{0,1\}.*/\1/p' <<<"$output")
if [[ -z "$total" || "$total" -lt 10 ]]; then
    echo "Expected at least 10 maintained Playwright tests; discovered ${total:-0}" >&2
    exit 1
fi

for project in mocked-chromium real-chromium real-firefox real-webkit; do
    if ! grep -Fq "[$project]" <<<"$output"; then
        echo "Playwright project '$project' discovered no tests" >&2
        exit 1
    fi
done

echo "Playwright manifest verified: $total tests across all required projects"
