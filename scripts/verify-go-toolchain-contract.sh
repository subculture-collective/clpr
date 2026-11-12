#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

expected_version="1.26.11"
expected_image="golang:${expected_version}-alpine@sha256:ae52b5bc6ed045a9ed26ab03dc19327a802a9cc9aa8b27ed57f5c6b60579b0bb"

fail() {
  echo "Go toolchain contract: $*" >&2
  exit 1
}

grep -Fxq "toolchain go${expected_version}" backend/go.mod \
  || fail "backend/go.mod must select Go ${expected_version}"

for dockerfile in backend/Dockerfile backend/Dockerfile.crawler; do
  grep -Fxq "FROM ${expected_image} AS builder" "$dockerfile" \
    || fail "$dockerfile must use the approved Go ${expected_version} builder digest"
done

if grep -Eq '^(toolchain go|FROM golang:)1\.26\.[0-9]([^0-9]|$)' backend/go.mod backend/Dockerfile backend/Dockerfile.crawler; then
  fail "vulnerable Go 1.26 patch release remains in an executable build input"
fi

echo "Go toolchain contract: verified ${expected_version}"
