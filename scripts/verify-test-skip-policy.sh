#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
inventory="$repo_root/config/test-skip-inventory.tsv"
today=${SKIP_POLICY_DATE:-$(date -u +%F)}
declare -A limits owners expiries actual

while IFS=$'\t' read -r path limit owner reason expiry; do
    [[ -z "$path" || "$path" == \#* ]] && continue
    if [[ -z "$limit" || -z "$owner" || -z "$reason" || -z "$expiry" ]]; then
        echo "Incomplete skip inventory entry for '$path'" >&2
        exit 1
    fi
    limits["$path"]=$limit
    owners["$path"]=$owner
    expiries["$path"]=$expiry
done <"$inventory"

while IFS= read -r match; do
    path=${match%%:*}
    line=${match#*:}; line=${line%%:*}
    source=${match#*:*:}
    [[ "$source" =~ ^[[:space:]]*// ]] && continue
    actual["$path"]=$(( ${actual["$path"]:-0} + 1 ))
done < <(cd "$repo_root" && rg -n '\bt\.Skip(f|Now)?\s*\(' backend --glob '*_test.go')

while IFS= read -r match; do
    path=${match%%:*}
    actual["$path"]=$(( ${actual["$path"]:-0} + 1 ))
done < <(cd "$repo_root" && rg -n '\b(describe|it|test)\.skip\b|\b(xdescribe|xit|xtest)\b' frontend --glob '*.{test,spec}.{ts,tsx,js,jsx}' || true)

failed=0
for path in "${!actual[@]}"; do
    if [[ -z "${limits[$path]+x}" ]]; then
        echo "Unregistered skip site: $path (${actual[$path]})" >&2
        failed=1
        continue
    fi
    if (( actual["$path"] > limits["$path"] )); then
        echo "Skip budget exceeded: $path has ${actual[$path]}, limit ${limits[$path]}" >&2
        failed=1
    fi
done

for path in "${!limits[@]}"; do
    if [[ "${expiries[$path]}" < "$today" ]]; then
        echo "Expired skip exception: $path (${owners[$path]}, ${expiries[$path]})" >&2
        failed=1
    fi
done

(( failed == 0 )) || exit 1
total=0
for count in "${actual[@]}"; do total=$((total + count)); done
echo "Skip policy verified: $total registered sites; no budget increases or expired exceptions"
