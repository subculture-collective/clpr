#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

expected_version="1.26.6"
expected_image="golang:${expected_version}-alpine@sha256:af8d6740070b8906d12eae1c3e3ea0957fb63f492051ea05e354c38ef9fe88df"

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

if rg -q '1\.26\.5' backend/go.mod backend/Dockerfile backend/Dockerfile.crawler; then
  fail "vulnerable Go 1.26.5 release remains in an executable build input"
fi

echo "Go toolchain contract: verified ${expected_version}"
