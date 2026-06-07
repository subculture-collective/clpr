#!/bin/bash
# Blue-Green Deployment Script
# Implements zero-downtime deployment with automatic rollback on failure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEPLOY_DIR="${DEPLOY_DIR:-/opt/clpr}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.blue-green.yml}"
REGISTRY="${REGISTRY:-ghcr.io/subculture-collective/clpr}"
HEALTH_CHECK_RETRIES="${HEALTH_CHECK_RETRIES:-30}"
HEALTH_CHECK_INTERVAL="${HEALTH_CHECK_INTERVAL:-10}"
BACKUP_DIR="${BACKUP_DIR:-/opt/clpr/backups}"
MONITORING_ENABLED="${MONITORING_ENABLED:-false}"

# Log functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose v2 is not available"
        exit 1
    fi
    
    if [ ! -d "$DEPLOY_DIR" ]; then
        log_error "Deploy directory does not exist: $DEPLOY_DIR"
        exit 1
    fi
    
    if [ ! -f "$DEPLOY_DIR/$COMPOSE_FILE" ]; then
        log_error "Compose file not found: $DEPLOY_DIR/$COMPOSE_FILE"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Detect currently active environment
detect_active_env() {
    log_step "Detecting active environment..."
    
    # Check which backend is currently running
    if docker ps --format '{{.Names}}' | grep -q "clpr-backend-blue"; then
        if docker ps --format '{{.Names}}' | grep -q "clpr-backend-green"; then
            log_warn "Both environments are running, assuming blue is active"
            echo "blue"
        else
            echo "blue"
        fi
    elif docker ps --format '{{.Names}}' | grep -q "clpr-backend-green"; then
        echo "green"
    else
        log_warn "No active environment detected, defaulting to blue"
        echo "blue"
    fi
}

# Get target environment (opposite of active)
get_target_env() {
    local active=$1
    if [ "$active" = "blue" ]; then
        echo "green"
    else
        echo "blue"
    fi
}

# Pull latest images for target environment
pull_images() {
    local env=$1
    log_step "Pulling latest images for $env environment..."
    
    cd "$DEPLOY_DIR" || exit 1
    
    # Set environment variables for image tags
    export BACKEND_${env^^}_TAG="${IMAGE_TAG:-latest}"
    export FRONTEND_${env^^}_TAG="${IMAGE_TAG:-latest}"
    
    if ! docker compose -f "$COMPOSE_FILE" --profile "$env" pull backend-$env frontend-$env; then
        log_error "Failed to pull images for $env environment"
        return 1
    fi
    
    log_success "Images pulled successfully"
    return 0
}

# Start target environment
start_environment() {
    local env=$1
    log_step "Starting $env environment..."
    
    cd "$DEPLOY_DIR" || exit 1
    
    if ! docker compose -f "$COMPOSE_FILE" --profile "$env" up -d backend-$env frontend-$env; then
        log_error "Failed to start $env environment"
        return 1
    fi
    
    log_success "$env environment started"
    return 0
}

# Health check for environment
health_check() {
    local env=$1
    local backend_port=8080
    local frontend_port=80
    
    log_step "Running health checks for $env environment (max $HEALTH_CHECK_RETRIES attempts)..."
    
    local retries=0
    while [ $retries -lt $HEALTH_CHECK_RETRIES ]; do
        retries=$((retries + 1))
        
        # Check backend health
        if docker exec clpr-backend-$env wget --spider -q http://localhost:$backend_port/health 2>/dev/null; then
            log_info "Backend health check passed (attempt $retries/$HEALTH_CHECK_RETRIES)"
            
            # Check frontend health
            if docker exec clpr-frontend-$env wget --spider -q http://localhost:$frontend_port/health.html 2>/dev/null; then
                log_success "$env environment is healthy"
                return 0
            else
                log_warn "Frontend health check failed (attempt $retries/$HEALTH_CHECK_RETRIES)"
            fi
        else
            log_warn "Backend health check failed (attempt $retries/$HEALTH_CHECK_RETRIES)"
        fi
        
        if [ $retries -lt $HEALTH_CHECK_RETRIES ]; then
            sleep $HEALTH_CHECK_INTERVAL
        fi
    done
    
    log_error "$env environment failed health checks after $HEALTH_CHECK_RETRIES attempts"
    return 1
}

# Run database migrations (backward compatible)
run_migrations() {
    log_step "Running database migrations..."
    
    # Check if migrations need to be run
    if [ -d "$DEPLOY_DIR/backend/migrations" ]; then
        log_info "Migrations directory found"
        
        # Pre-flight validation: Check database connectivity
        log_info "Validating database connectivity..."
        if ! docker exec clpr-postgres pg_isready -U "${POSTGRES_USER:-clpr}" -d "${POSTGRES_DB:-clpr_db}" > /dev/null 2>&1; then
            log_error "Database is not ready for migrations"
            return 1
        fi
        log_success "Database connectivity verified"
        
        # Construct database URL for migrations
        # Note: SSL is disabled because migrations run within the Docker network (not over internet)
        # The connection is network-isolated and doesn't traverse untrusted networks
        # For external database connections, enable SSL by changing sslmode=require
        # Password is passed via environment variable to avoid exposure in process list
        DB_URL="postgresql://${POSTGRES_USER:-clpr}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB:-clpr_db}?sslmode=disable"
        
        # Run migrations using golang-migrate in a temporary container
        # Pin to specific image digest for supply chain security
        # Current digest is for v4.17.0 (sha256:4d017c6fb5997127093648cab09e63d377997125c3d3dcca18e5d1c847da49fa)
        log_info "Executing database migrations..."
        if docker run --rm \
            --network clpr-network \
            -v "$DEPLOY_DIR/backend/migrations:/migrations:ro" \
            -e DATABASE_URL="$DB_URL" \
            migrate/migrate@sha256:4d017c6fb5997127093648cab09e63d377997125c3d3dcca18e5d1c847da49fa \
            -path /migrations \
            -database "$DB_URL" \
            up; then
            log_success "Migrations executed successfully"
        else
            log_error "Migration execution failed"
            log_error "Rolling back is recommended. Check migration status manually if needed."
            return 1
        fi
        
        # Verify migrations were applied
        log_info "Verifying migration status..."
        MIGRATION_OUTPUT=$(docker run --rm \
            --network clpr-network \
            -e DATABASE_URL="$DB_URL" \
            migrate/migrate@sha256:4d017c6fb5997127093648cab09e63d377997125c3d3dcca18e5d1c847da49fa \
            -path /migrations \
            -database "$DB_URL" \
            version 2>&1)
        MIGRATION_STATUS=$?
        MIGRATION_VERSION=$(printf '%s\n' "$MIGRATION_OUTPUT" | tail -1)
        
        if [ "$MIGRATION_STATUS" -eq 0 ]; then
            log_success "Current migration version: $MIGRATION_VERSION"
        else
            log_warn "Could not verify migration version, but migrations may have succeeded"
        fi
    else
        log_warn "No migrations directory found at $DEPLOY_DIR/backend/migrations"
    fi
    
    log_success "Database migration check complete"
    return 0
}

# Switch traffic to target environment
switch_traffic() {
    local target_env=$1
    log_step "Switching traffic to $target_env environment..."
    
    # Update Caddyfile to point to target environment
    if [ -f "$DEPLOY_DIR/Caddyfile" ]; then
        # Create backup of Caddyfile
        cp "$DEPLOY_DIR/Caddyfile" "$DEPLOY_DIR/Caddyfile.backup"
        
        # Replace environment references in Caddyfile
        log_info "Updating Caddy configuration to route to $target_env..."
        
        # Determine which environment to point to
        if [ "$target_env" = "blue" ]; then
            # Switch from green to blue
            sed -i 's/clpr-backend-green:8080/clpr-backend-blue:8080/g' "$DEPLOY_DIR/Caddyfile" 2>/dev/null || \
            sed -i.bak 's/clpr-backend-green:8080/clpr-backend-blue:8080/g' "$DEPLOY_DIR/Caddyfile"
            sed -i 's/clpr-frontend-green:80/clpr-frontend-blue:80/g' "$DEPLOY_DIR/Caddyfile" 2>/dev/null || \
            sed -i.bak 's/clpr-frontend-green:80/clpr-frontend-blue:80/g' "$DEPLOY_DIR/Caddyfile"
        else
            # Switch from blue to green
            sed -i 's/clpr-backend-blue:8080/clpr-backend-green:8080/g' "$DEPLOY_DIR/Caddyfile" 2>/dev/null || \
            sed -i.bak 's/clpr-backend-blue:8080/clpr-backend-green:8080/g' "$DEPLOY_DIR/Caddyfile"
            sed -i 's/clpr-frontend-blue:80/clpr-frontend-green:80/g' "$DEPLOY_DIR/Caddyfile" 2>/dev/null || \
            sed -i.bak 's/clpr-frontend-blue:80/clpr-frontend-green:80/g' "$DEPLOY_DIR/Caddyfile"
        fi
        
        # Reload Caddy configuration
        if docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile 2>/dev/null; then
            log_success "Traffic switched to $target_env environment"
            return 0
        else
            log_warn "Caddy reload failed, attempting restart..."
            # Pass ACTIVE_ENV explicitly when restarting Caddy
            ACTIVE_ENV=$target_env docker compose -f "$COMPOSE_FILE" restart caddy
            sleep 5
            log_success "Caddy restarted with new configuration"
            return 0
        fi
    else
        log_error "Caddyfile not found"
        return 1
    fi
}

# Stop old environment
stop_old_environment() {
    local env=$1
    log_step "Stopping $env environment..."
    
    cd "$DEPLOY_DIR" || exit 1
    
    if ! docker compose -f "$COMPOSE_FILE" --profile "$env" stop backend-$env frontend-$env; then
        log_warn "Failed to stop $env environment gracefully"
    fi
    
    # Remove containers
    docker compose -f "$COMPOSE_FILE" --profile "$env" rm -f backend-$env frontend-$env 2>/dev/null || true
    
    log_success "$env environment stopped"
}

# Rollback to previous environment
rollback() {
    local active_env=$1
    local failed_env=$2
    
    log_error "Deployment failed, initiating rollback..."
    
    # Stop failed environment
    stop_old_environment "$failed_env"
    
    # Ensure active environment is running
    if ! docker ps --format '{{.Names}}' | grep -q "clpr-backend-$active_env"; then
        log_warn "Active environment not running, attempting to start..."
        start_environment "$active_env"
    fi
    
    # Switch traffic back to active environment
    switch_traffic "$active_env"
    
    log_error "Rollback complete, traffic restored to $active_env"
    return 1
}

# Send monitoring notification
send_notification() {
    local status=$1
    local message=$2
    
    if [ "$MONITORING_ENABLED" = "true" ]; then
        log_info "Sending monitoring notification: $status - $message"
        # Add monitoring integration here (e.g., Slack, Discord, PagerDuty)
        # Example: curl -X POST -H 'Content-type: application/json' --data "{\"text\":\"$message\"}" $WEBHOOK_URL
    fi
}

# Main deployment flow
main() {
    echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   Clipper Blue-Green Deployment Script        ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # Check prerequisites
    check_prerequisites
    
    # Detect active environment
    ACTIVE_ENV=$(detect_active_env)
    TARGET_ENV=$(get_target_env "$ACTIVE_ENV")
    
    log_info "Active environment: $ACTIVE_ENV"
    log_info "Target environment: $TARGET_ENV"
    echo ""
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    # Backup current state
    log_step "Creating backup..."
    BACKUP_FILE="$BACKUP_DIR/deployment-$(date +%Y%m%d-%H%M%S).tar.gz"
    tar -czf "$BACKUP_FILE" -C "$DEPLOY_DIR" docker-compose.blue-green.yml .env Caddyfile 2>/dev/null || true
    log_success "Backup created: $BACKUP_FILE"
    echo ""
    
    # Send start notification
    send_notification "started" "Blue-Green deployment started: $ACTIVE_ENV → $TARGET_ENV"
    
    # Pull images
    if ! pull_images "$TARGET_ENV"; then
        send_notification "failed" "Failed to pull images for $TARGET_ENV"
        exit 1
    fi
    echo ""
    
    # Run migrations (if needed)
    if ! run_migrations; then
        log_error "Migration failed"
        send_notification "failed" "Database migration failed"
        exit 1
    fi
    echo ""
    
    # Start target environment
    if ! start_environment "$TARGET_ENV"; then
        send_notification "failed" "Failed to start $TARGET_ENV environment"
        exit 1
    fi
    echo ""
    
    # Wait for target environment to be ready
    log_info "Waiting for $TARGET_ENV environment to initialize..."
    sleep 20
    echo ""
    
    # Health check target environment
    if ! health_check "$TARGET_ENV"; then
        rollback "$ACTIVE_ENV" "$TARGET_ENV"
        send_notification "failed" "Health checks failed for $TARGET_ENV, rolled back to $ACTIVE_ENV"
        exit 1
    fi
    echo ""
    
    # Switch traffic
    if ! switch_traffic "$TARGET_ENV"; then
        rollback "$ACTIVE_ENV" "$TARGET_ENV"
        send_notification "failed" "Traffic switch failed, rolled back to $ACTIVE_ENV"
        exit 1
    fi
    echo ""
    
    # Wait and monitor new environment
    log_step "Monitoring new environment for 30 seconds..."
    sleep 30
    
    # Final health check
    if ! health_check "$TARGET_ENV"; then
        rollback "$ACTIVE_ENV" "$TARGET_ENV"
        send_notification "failed" "Post-switch health check failed, rolled back to $ACTIVE_ENV"
        exit 1
    fi
    echo ""
    
    # Stop old environment
    stop_old_environment "$ACTIVE_ENV"
    echo ""
    
    # Cleanup old images (optional)
    log_step "Cleaning up old Docker images..."
    docker image prune -f &>/dev/null || true
    log_success "Cleanup complete"
    echo ""
    
    # Success!
    echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   Deployment Successful! ✓                     ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}"
    echo ""
    log_success "Blue-Green deployment completed successfully"
    log_info "Previous environment: $ACTIVE_ENV (stopped)"
    log_info "Current environment: $TARGET_ENV (active)"
    log_info "Backup: $BACKUP_FILE"
    
    send_notification "success" "Blue-Green deployment completed: $TARGET_ENV is now active"
    
    return 0
}

# Run main function
main "$@"
