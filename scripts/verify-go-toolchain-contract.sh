#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

expected_version="1.26.10"
expected_image="golang:${expected_version}-alpine@sha256:a2012b4a1d1f9f7e9f44e0d39b805a901e856b1dd80b4e6ee4338db547050308"

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
