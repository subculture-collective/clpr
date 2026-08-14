#!/usr/bin/env bash

set -euo pipefail

lock_name="clpr-hosted-heavy-lock"
owner="${CLPR_HOSTED_LOCK_OWNER:-}"

[[ -n "$owner" ]] || exit 0
current_owner="$(docker inspect --format '{{index .Config.Labels "clpr.hosted-heavy.owner"}}' "$lock_name" 2>/dev/null || true)"
if [[ "$current_owner" == "$owner" ]]; then
  docker rm --force "$lock_name" >/dev/null
  echo "Released hosted heavy lock for $owner"
fi
