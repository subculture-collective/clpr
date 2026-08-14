#!/usr/bin/env bash

set -euo pipefail

lock_name="clpr-hosted-heavy-lock"
lock_image="alpine:3.22"
owner="${GITEA_RUN_ID:-${GITHUB_RUN_ID:-local}}-${GITHUB_JOB:-${GITEA_JOB:-job}}-$$"
deadline=$((SECONDS + ${CLPR_HOSTED_LOCK_WAIT_SECONDS:-14400}))

while (( SECONDS < deadline )); do
  if docker run --detach --name "$lock_name" \
    --label "clpr.hosted-heavy.owner=$owner" \
    "$lock_image" sleep 18000 >/dev/null 2>&1; then
    export CLPR_HOSTED_LOCK_OWNER="$owner"
    if [[ -n "${GITHUB_ENV:-}" ]]; then
      printf 'CLPR_HOSTED_LOCK_OWNER=%s\n' "$owner" >>"$GITHUB_ENV"
    fi
    echo "Acquired hosted heavy lock for $owner"
    exit 0
  fi

  status="$(docker inspect --format '{{.State.Status}}' "$lock_name" 2>/dev/null || true)"
  if [[ -z "$status" || "$status" == exited || "$status" == dead ]]; then
    docker rm --force "$lock_name" >/dev/null 2>&1 || true
    continue
  fi

  echo "Hosted heavy lock is held; waiting for the active job to finish"
  sleep 15
done

echo "Timed out waiting for the hosted heavy lock" >&2
exit 1
