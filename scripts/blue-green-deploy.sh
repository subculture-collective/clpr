#!/usr/bin/env bash

# Digest-pinned blue/green deployment with authenticated canary routing,
# automatic rollback, and an externally stored runtime proxy configuration.

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-/opt/clpr}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.blue-green.yml}"
ENV_FILE="${ENV_FILE:-$DEPLOY_DIR/.env}"
TEMPLATE_FILE="${TEMPLATE_FILE:-$DEPLOY_DIR/deploy/Caddyfile.blue-green.template}"
RUNTIME_DIR="${CLPR_RUNTIME_DIR:-/var/lib/clpr/deploy}"
BACKUP_DIR="${BACKUP_DIR:-/var/lib/clpr/deploy-backups}"
CADDY_IMAGE="${CADDY_IMAGE:-caddy:2-alpine@sha256:5f5c8640aae01df9654968d946d8f1a56c497f1dd5c5cda4cf95ab7c14d58648}"
HEALTH_CHECK_RETRIES="${HEALTH_CHECK_RETRIES:-30}"
HEALTH_CHECK_INTERVAL="${HEALTH_CHECK_INTERVAL:-10}"
BAKE_SECONDS="${BAKE_SECONDS:-3600}"
BAKE_CHECK_INTERVAL="${BAKE_CHECK_INTERVAL:-30}"
CANARY_BASE_URL="${CANARY_BASE_URL:-https://clpr.tv}"
CANARY_PATHS="${CANARY_PATHS:-/health /health/ready / /search}"
DEFAULT_ACTIVE_SLOT="${DEFAULT_ACTIVE_SLOT:-blue}"
MONITORING_ENABLED="${MONITORING_ENABLED:-false}"
CRAWLER_DIGEST="${CRAWLER_DIGEST:-}"

ACTIVE_SLOT_FILE="$RUNTIME_DIR/active-slot"
CURRENT_CRAWLER_FILE="$RUNTIME_DIR/crawler-current-digest"
PREVIOUS_CRAWLER_FILE="$RUNTIME_DIR/crawler-previous-digest"
DEPLOY_LOCK_FILE="$RUNTIME_DIR/deploy.lock"

ROLLBACK_ARMED=false
ORIGINAL_SLOT=""
TARGET_SLOT=""
ORIGINAL_CRAWLER_DIGEST=""

log() {
    local level="$1"
    shift
    printf '[%s] %s - %s\n' "$level" "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

die() {
    log ERROR "$*" >&2
    exit 1
}

validate_slot() {
    [[ "$1" == "blue" || "$1" == "green" ]] || die "invalid deployment slot: $1"
}

opposite_slot() {
    validate_slot "$1"
    if [[ "$1" == "blue" ]]; then
        printf 'green\n'
    else
        printf 'blue\n'
    fi
}

validate_digest() {
    [[ "$1" =~ ^sha256:[0-9a-f]{64}$ ]]
}

validate_positive_integer() {
    local name="$1"
    local value="$2"
    [[ "$value" =~ ^[0-9]+$ ]] || die "$name must be a non-negative integer"
}

compose() {
    (
        cd "$DEPLOY_DIR"
        docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
    )
}

load_deployment_env() {
    # The operator-owned .env is already trusted by Docker Compose. Loading it
    # here lets this script validate the exact immutable references it will use.
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
    export CLPR_RUNTIME_DIR="$RUNTIME_DIR"
}

check_prerequisites() {
    command -v docker >/dev/null || die "docker is required"
    docker compose version >/dev/null || die "Docker Compose v2 is required"
    command -v curl >/dev/null || die "curl is required"
    command -v flock >/dev/null || die "flock is required"
    [[ -d "$DEPLOY_DIR" ]] || die "deploy directory does not exist: $DEPLOY_DIR"
    [[ -f "$DEPLOY_DIR/$COMPOSE_FILE" ]] || die "compose file not found: $DEPLOY_DIR/$COMPOSE_FILE"
    [[ -f "$ENV_FILE" ]] || die "operator environment file not found: $ENV_FILE"
    [[ -f "$TEMPLATE_FILE" ]] || die "blue/green Caddy template not found: $TEMPLATE_FILE"

    local deploy_path runtime_path
    deploy_path="$(cd "$DEPLOY_DIR" && pwd)"
    mkdir -p "$RUNTIME_DIR"
    runtime_path="$(cd "$RUNTIME_DIR" && pwd)"
    case "$runtime_path/" in
        "$deploy_path/"*)
            die "CLPR_RUNTIME_DIR must be outside the tracked deploy source"
            ;;
    esac

    load_deployment_env
    local variable value
    for variable in \
        BACKEND_BLUE_DIGEST FRONTEND_BLUE_DIGEST \
        BACKEND_GREEN_DIGEST FRONTEND_GREEN_DIGEST CRAWLER_DIGEST; do
        value="${!variable:-}"
        validate_digest "$value" || die "$variable must be an immutable sha256 digest"
    done

    [[ "${CANARY_TOKEN:-}" =~ ^[A-Za-z0-9._~-]{32,128}$ ]] \
        || die "CANARY_TOKEN must contain 32-128 URL-safe characters"
    validate_positive_integer BAKE_SECONDS "$BAKE_SECONDS"
    validate_positive_integer BAKE_CHECK_INTERVAL "$BAKE_CHECK_INTERVAL"
    (( BAKE_CHECK_INTERVAL > 0 )) || die "BAKE_CHECK_INTERVAL must be greater than zero"
    validate_positive_integer HEALTH_CHECK_RETRIES "$HEALTH_CHECK_RETRIES"
    validate_positive_integer HEALTH_CHECK_INTERVAL "$HEALTH_CHECK_INTERVAL"
}

acquire_deploy_lock() {
    exec 9>"$DEPLOY_LOCK_FILE"
    flock -n 9 || die "another CLPR deployment or rollback is already running"
}

render_runtime_config() {
    local active_slot="$1"
    local canary_slot="$2"
    local output="$3"
    validate_slot "$active_slot"
    validate_slot "$canary_slot"
    [[ "${CANARY_TOKEN:-}" =~ ^[A-Za-z0-9._~-]{32,128}$ ]] \
        || die "CANARY_TOKEN must contain 32-128 URL-safe characters"
    [[ -f "$TEMPLATE_FILE" ]] || die "template not found: $TEMPLATE_FILE"

    local rendered
    rendered="$(<"$TEMPLATE_FILE")"
    rendered="${rendered//__ACTIVE_BACKEND__/clpr-backend-$active_slot:8080}"
    rendered="${rendered//__ACTIVE_FRONTEND__/clpr-frontend-$active_slot:8080}"
    rendered="${rendered//__CANARY_BACKEND__/clpr-backend-$canary_slot:8080}"
    rendered="${rendered//__CANARY_FRONTEND__/clpr-frontend-$canary_slot:8080}"
    rendered="${rendered//__CANARY_SLOT__/$canary_slot}"
    rendered="${rendered//__CANARY_TOKEN__/$CANARY_TOKEN}"

    [[ "$rendered" != *'__ACTIVE_'* && "$rendered" != *'__CANARY_'* ]] \
        || die "runtime proxy template contains unresolved placeholders"
    printf '%s\n' "$rendered" > "$output"
}

install_runtime_config() {
    local active_slot="$1"
    local canary_slot="$2"
    local candidate
    candidate="$(mktemp "$RUNTIME_DIR/Caddyfile.XXXXXX")"
    render_runtime_config "$active_slot" "$canary_slot" "$candidate"

    if ! docker run --rm \
        -v "$candidate:/etc/caddy/Caddyfile:ro" \
        "$CADDY_IMAGE" caddy validate --config /etc/caddy/Caddyfile >/dev/null; then
        rm -f "$candidate"
        die "generated runtime Caddy configuration is invalid"
    fi
    chmod 0640 "$candidate"
    mv -f "$candidate" "$RUNTIME_DIR/Caddyfile"
}

write_active_slot() {
    validate_slot "$1"
    printf '%s\n' "$1" > "$ACTIVE_SLOT_FILE.tmp"
    mv -f "$ACTIVE_SLOT_FILE.tmp" "$ACTIVE_SLOT_FILE"
}

read_active_slot() {
    local slot="$DEFAULT_ACTIVE_SLOT"
    if [[ -s "$ACTIVE_SLOT_FILE" ]]; then
        slot="$(<"$ACTIVE_SLOT_FILE")"
    fi
    validate_slot "$slot"
    printf '%s\n' "$slot"
}

reload_proxy() {
    if docker ps --format '{{.Names}}' | grep -Fxq clpr-caddy; then
        docker exec clpr-caddy caddy reload --config /etc/caddy/runtime/Caddyfile
    else
        compose up -d caddy
    fi
}

activate_runtime_config() {
    local active_slot="$1"
    local canary_slot="$2"
    local previous=""
    if [[ -f "$RUNTIME_DIR/Caddyfile" ]]; then
        previous="$(mktemp "$RUNTIME_DIR/Caddyfile.previous.XXXXXX")"
        cp -p "$RUNTIME_DIR/Caddyfile" "$previous"
    fi
    install_runtime_config "$active_slot" "$canary_slot"
    if reload_proxy; then
        [[ -z "$previous" ]] || rm -f "$previous"
        return 0
    fi

    log ERROR "Proxy reload failed; restoring the prior runtime configuration"
    if [[ -n "$previous" ]]; then
        mv -f "$previous" "$RUNTIME_DIR/Caddyfile"
        reload_proxy || true
    else
        rm -f "$RUNTIME_DIR/Caddyfile"
    fi
    return 1
}

pull_candidate_images() {
    local slot="$1"
    log STEP "Pulling immutable $slot candidate and shared crawler images"
    compose --profile "$slot" pull "backend-$slot" "frontend-$slot" crawler
}

start_slot() {
    local slot="$1"
    log STEP "Starting $slot application slot"
    compose --profile "$slot" up -d "backend-$slot" "frontend-$slot"
}

stop_slot() {
    local slot="$1"
    log STEP "Gracefully stopping retained $slot application slot"
    compose --profile "$slot" stop -t 30 "frontend-$slot" "backend-$slot" || true
    compose --profile "$slot" rm -f "frontend-$slot" "backend-$slot" || true
}

health_once() {
    local slot="$1"
    docker exec "clpr-backend-$slot" wget --spider -q http://localhost:8080/health \
        && docker exec "clpr-frontend-$slot" wget --spider -q http://localhost:8080/health.html
}

health_check() {
    local slot="$1"
    local attempt
    for ((attempt = 1; attempt <= HEALTH_CHECK_RETRIES; attempt++)); do
        if health_once "$slot"; then
            log INFO "$slot health check passed ($attempt/$HEALTH_CHECK_RETRIES)"
            return 0
        fi
        if (( attempt < HEALTH_CHECK_RETRIES )); then
            sleep "$HEALTH_CHECK_INTERVAL"
        fi
    done
    log ERROR "$slot failed health checks"
    return 1
}

run_migrations() {
    [[ -d "$DEPLOY_DIR/backend/migrations" ]] || {
        log WARN "No migrations directory found; skipping migration step"
        return 0
    }
    docker exec clpr-postgres pg_isready \
        -U "${POSTGRES_USER:-clpr}" -d "${POSTGRES_DB:-clpr_db}" >/dev/null

    local database_url
    database_url="postgresql://${POSTGRES_USER:-clpr}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB:-clpr_db}?sslmode=disable"
    docker run --rm --network clpr-network \
        -v "$DEPLOY_DIR/backend/migrations:/migrations:ro" \
        migrate/migrate@sha256:4d017c6fb5997127093648cab09e63d377997125c3d3dcca18e5d1c847da49fa \
        -path /migrations -database "$database_url" up
}

canary_smoke() {
    local slot="$1"
    local path headers status
    log STEP "Running authenticated operator canary smoke against $slot"
    for path in $CANARY_PATHS; do
        headers="$(mktemp)"
        status="$(
            printf 'header = "X-CLPR-Canary: %s"\n' "$CANARY_TOKEN" \
                | curl --config - --silent --show-error --output /dev/null \
                    --dump-header "$headers" --write-out '%{http_code}' \
                    "$CANARY_BASE_URL$path"
        )"
        if [[ ! "$status" =~ ^2[0-9][0-9]$ ]] \
            || ! grep -Eiq "^X-CLPR-Served-Slot: $slot\r?$" "$headers"; then
            rm -f "$headers"
            log ERROR "Canary smoke failed for $path (HTTP $status)"
            return 1
        fi
        rm -f "$headers"
    done
}

current_crawler_digest() {
    local digest=""
    if [[ -s "$CURRENT_CRAWLER_FILE" ]]; then
        digest="$(<"$CURRENT_CRAWLER_FILE")"
    elif docker inspect clpr-crawler >/dev/null 2>&1; then
        digest="$(docker inspect --format '{{.Config.Image}}' clpr-crawler | awk -F@ '{print $2}')"
    fi
    if validate_digest "$digest"; then
        printf '%s\n' "$digest"
    else
        printf '%s\n' "$CRAWLER_DIGEST"
    fi
}

reconcile_crawler() {
    local digest="$1"
    validate_digest "$digest" || die "refusing to start crawler without an immutable digest"
    log STEP "Reconciling the single shared crawler"
    CRAWLER_DIGEST="$digest" compose up -d --no-deps --force-recreate crawler
    [[ "$(docker inspect --format '{{.State.Running}}' clpr-crawler)" == "true" ]]
}

write_crawler_state() {
    local current="$1"
    local previous="$2"
    validate_digest "$current" || die "invalid current crawler digest"
    validate_digest "$previous" || die "invalid previous crawler digest"
    printf '%s\n' "$current" > "$CURRENT_CRAWLER_FILE.tmp"
    printf '%s\n' "$previous" > "$PREVIOUS_CRAWLER_FILE.tmp"
    mv -f "$CURRENT_CRAWLER_FILE.tmp" "$CURRENT_CRAWLER_FILE"
    mv -f "$PREVIOUS_CRAWLER_FILE.tmp" "$PREVIOUS_CRAWLER_FILE"
}

perform_rollback() {
    local restore_slot="$1"
    local failed_slot="$2"
    local crawler_digest="$3"
    log ERROR "Rolling traffic back from $failed_slot to $restore_slot"
    start_slot "$restore_slot" || return 1
    health_check "$restore_slot" || {
        log ERROR "Rollback slot $restore_slot is not ready; traffic was not switched"
        return 1
    }
    activate_runtime_config "$restore_slot" "$failed_slot"
    write_active_slot "$restore_slot"
    if validate_digest "$crawler_digest"; then
        reconcile_crawler "$crawler_digest" || log ERROR "Crawler restoration requires operator attention"
        write_crawler_state "$crawler_digest" "${CRAWLER_DIGEST:-$crawler_digest}"
    fi
    stop_slot "$failed_slot"
    ROLLBACK_ARMED=false
    log ERROR "Automatic rollback complete; $restore_slot is active"
}

on_interrupt() {
    local exit_code=$?
    trap - INT TERM
    exit "$exit_code"
}
trap on_interrupt INT TERM

on_exit() {
    local exit_code=$?
    trap - EXIT INT TERM
    if (( exit_code != 0 )) && [[ "$ROLLBACK_ARMED" == true ]]; then
        if ! perform_rollback "$ORIGINAL_SLOT" "$TARGET_SLOT" "$ORIGINAL_CRAWLER_DIGEST"; then
            log ERROR "Automatic rollback failed; immediate operator intervention is required"
        fi
    fi
    exit "$exit_code"
}
trap on_exit EXIT

bake_candidate() {
    local slot="$1"
    local remaining="$BAKE_SECONDS"
    log STEP "Retaining the previous slot for a ${BAKE_SECONDS}s bake window"
    while (( remaining > 0 )); do
        local interval="$BAKE_CHECK_INTERVAL"
        (( interval > remaining )) && interval="$remaining"
        sleep "$interval"
        remaining=$((remaining - interval))
        health_once "$slot" || return 1
        log INFO "$slot healthy; ${remaining}s remain in bake window"
    done
}

send_notification() {
    local status="$1"
    shift
    if [[ "$MONITORING_ENABLED" == true ]]; then
        log INFO "notification[$status]: $*"
    fi
}

backup_runtime_state() {
    mkdir -p "$BACKUP_DIR"
    local destination
    destination="$BACKUP_DIR/$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$destination"
    for file in Caddyfile active-slot crawler-current-digest crawler-previous-digest; do
        [[ -f "$RUNTIME_DIR/$file" ]] && cp -p "$RUNTIME_DIR/$file" "$destination/$file"
    done
    printf '%s\n' "$destination"
}

deploy() {
    check_prerequisites
    mkdir -p "$RUNTIME_DIR"
    acquire_deploy_lock

    ORIGINAL_SLOT="$(read_active_slot)"
    TARGET_SLOT="$(opposite_slot "$ORIGINAL_SLOT")"
    ORIGINAL_CRAWLER_DIGEST="$(current_crawler_digest)"
    local backup
    backup="$(backup_runtime_state)"
    log INFO "Active slot: $ORIGINAL_SLOT; candidate slot: $TARGET_SLOT"
    log INFO "Runtime state backup: $backup"
    send_notification started "$ORIGINAL_SLOT -> $TARGET_SLOT"

    pull_candidate_images "$TARGET_SLOT"
    run_migrations
    start_slot "$TARGET_SLOT"
    if ! health_check "$TARGET_SLOT"; then
        stop_slot "$TARGET_SLOT"
        die "candidate slot failed pre-promotion health checks"
    fi

    # Public traffic remains on the original slot; only possession of the
    # operator token routes these requests to the inactive candidate.
    activate_runtime_config "$ORIGINAL_SLOT" "$TARGET_SLOT"
    write_active_slot "$ORIGINAL_SLOT"
    if ! canary_smoke "$TARGET_SLOT"; then
        stop_slot "$TARGET_SLOT"
        activate_runtime_config "$ORIGINAL_SLOT" "$ORIGINAL_SLOT"
        die "authenticated canary smoke failed"
    fi

    # Atomically promote while retaining the prior containers and digests.
    # Arm rollback before the first promotion mutation. The EXIT trap restores
    # the previous slot if either proxy activation or state persistence fails.
    ROLLBACK_ARMED=true
    activate_runtime_config "$TARGET_SLOT" "$ORIGINAL_SLOT"
    write_active_slot "$TARGET_SLOT"

    if ! reconcile_crawler "$CRAWLER_DIGEST"; then
        perform_rollback "$ORIGINAL_SLOT" "$TARGET_SLOT" "$ORIGINAL_CRAWLER_DIGEST"
        die "candidate crawler failed to start; rollback completed"
    fi
    write_crawler_state "$CRAWLER_DIGEST" "$ORIGINAL_CRAWLER_DIGEST"

    if ! bake_candidate "$TARGET_SLOT"; then
        perform_rollback "$ORIGINAL_SLOT" "$TARGET_SLOT" "$ORIGINAL_CRAWLER_DIGEST"
        die "candidate failed during bake; rollback completed"
    fi

    ROLLBACK_ARMED=false
    stop_slot "$ORIGINAL_SLOT"
    activate_runtime_config "$TARGET_SLOT" "$TARGET_SLOT"
    send_notification success "$TARGET_SLOT active after ${BAKE_SECONDS}s bake"
    log INFO "Deployment complete; $TARGET_SLOT is active"
    log INFO "Candidate images were not retagged or pruned"
}

manual_rollback() {
    check_prerequisites
    mkdir -p "$RUNTIME_DIR"
    acquire_deploy_lock
    local active restore crawler_restore
    active="$(read_active_slot)"
    restore="$(opposite_slot "$active")"
    crawler_restore="${CRAWLER_DIGEST}"
    if [[ -s "$PREVIOUS_CRAWLER_FILE" ]]; then
        crawler_restore="$(<"$PREVIOUS_CRAWLER_FILE")"
    fi
    start_slot "$restore"
    health_check "$restore" || die "rollback slot $restore is not healthy"
    perform_rollback "$restore" "$active" "$crawler_restore"
}

usage() {
    cat <<'EOF'
Usage:
  blue-green-deploy.sh deploy
  blue-green-deploy.sh rollback
  blue-green-deploy.sh render <active-slot> <canary-slot> <output-file>

The deploy command requires immutable digest variables and CANARY_TOKEN in the
operator environment file. Runtime state defaults to /var/lib/clpr/deploy and
must remain outside the tracked deployment source.
EOF
}

case "${1:-deploy}" in
    deploy)
        deploy
        ;;
    rollback)
        manual_rollback
        ;;
    render)
        [[ $# -eq 4 ]] || {
            usage >&2
            exit 2
        }
        render_runtime_config "$2" "$3" "$4"
        ;;
    --help|-h|help)
        usage
        ;;
    *)
        usage >&2
        exit 2
        ;;
esac
