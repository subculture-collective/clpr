#!/usr/bin/env bash

set -euo pipefail

readonly version='v0.15.1'
readonly expected_sha256='9afdf0f48064fdcd8a8815ca6d2d9fcd43a0ef4c2edf9da23d6f7bf46a764eb8'
readonly archive_name="lychee-${version}-x86_64-unknown-linux-gnu.tar.gz"
readonly download_url="https://github.com/lycheeverse/lychee/releases/download/${version}/${archive_name}"

install_dir=${1:-${RUNNER_TEMP:-/tmp}/clpr-lychee-bin}
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT

curl --fail --silent --show-error --location \
    --proto '=https' --tlsv1.2 \
    "$download_url" --output "$work_dir/$archive_name"
printf '%s  %s\n' "$expected_sha256" "$work_dir/$archive_name" | sha256sum --check --status
tar -xzf "$work_dir/$archive_name" -C "$work_dir"
install -d -m 0755 "$install_dir"
install -m 0755 "$work_dir/lychee" "$install_dir/lychee"

if [[ -n "${GITHUB_PATH:-}" ]]; then
    printf '%s\n' "$install_dir" >> "$GITHUB_PATH"
fi

"$install_dir/lychee" --version
