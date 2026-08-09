#!/bin/sh
set -eu

child_pid=""
shutdown() {
    if [ -n "$child_pid" ]; then
        kill "$child_pid" 2>/dev/null || true
        wait "$child_pid" 2>/dev/null || true
    fi
    exit 0
}
trap shutdown INT TERM

while true; do
    echo "=== $(date -Iseconds) Starting scrape ==="
    if [ -n "${SCRAPE_BROADCASTERS:-}" ]; then
        ./scraper -min-views "$SCRAPE_MIN_VIEWS" -max-age-days "$SCRAPE_MAX_AGE_DAYS" -batch-size "$SCRAPE_BATCH_SIZE" -broadcasters "$SCRAPE_BROADCASTERS" &
    else
        ./scraper -min-views "$SCRAPE_MIN_VIEWS" -max-age-days "$SCRAPE_MAX_AGE_DAYS" -batch-size "$SCRAPE_BATCH_SIZE" &
    fi
    child_pid=$!
    wait "$child_pid" || true
    child_pid=""

    echo "=== Sleeping ${SCRAPE_INTERVAL}s ==="
    sleep "$SCRAPE_INTERVAL" &
    child_pid=$!
    wait "$child_pid" || true
    child_pid=""
done
