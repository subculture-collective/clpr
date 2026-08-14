#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

required_paths=(
  "Taskfile.yml"
  "Makefile"
  "package-lock.json"
  "frontend/package-lock.json"
  "backend/go.mod"
  "backend/run-tests-verbose.sh"
  "backend/setup-test-env.sh"
  "backend/tests/integration/premium"
  "docker-compose.test.yml"
)

for path in "${required_paths[@]}"; do
  if [[ ! -e "$path" ]]; then
    echo "task contract: missing required path: $path" >&2
    exit 1
  fi
done

task_list="$(task --list)"
required_tasks=(
  "contract:"
  "build:"
  "lint:"
  "test:backend:"
  "test:frontend:"
  "test:e2e:"
  "test:integration:"
  "openapi:validate:"
)

for task_name in "${required_tasks[@]}"; do
  if ! grep -Fq "$task_name" <<<"$task_list"; then
    echo "task contract: missing public task: $task_name" >&2
    exit 1
  fi
done

unsupported_references=(
  "test-commands.sh"
  "test-seed-opensearch.sh"
  "test-seed-e2e.sh"
  "setup-e2e-tests.sh"
  "frontend/test-commands.sh"
  "backend/tests/load/"
  "backend/tests/security/"
  "infrastructure/k8s/bootstrap/"
  "infrastructure/k8s/overlays/"
)

for reference in "${unsupported_references[@]}"; do
  if grep -Fq "$reference" Taskfile.yml; then
    echo "task contract: unsupported path is still advertised: $reference" >&2
    exit 1
  fi
done

bash scripts/verify-go-toolchain-contract.sh

echo "task contract: verified"
