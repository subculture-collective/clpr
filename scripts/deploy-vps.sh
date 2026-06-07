#!/bin/bash
set -euo pipefail

# Clipper VPS Production Deploy Script
# Designed for VPS deployment with external Vault and Caddy
# Usage: ./scripts/deploy-vps.sh [--skip-git] [--no-pull] [--skip-migrations]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Project name for Docker Compose (consistent with Makefile)
PROJECT_NAME="${COMPOSE_PROJECT_NAME:-$(basename "$PROJECT_ROOT")}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Flags
SKIP_GIT=false
NO_PULL=false
SKIP_MIGRATIONS=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-git) SKIP_GIT=true; shift ;;
        --no-pull) NO_PULL=true; shift ;;
        --skip-migrations) SKIP_MIGRATIONS=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

log() { echo -e "${BLUE}[deploy]${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
warn() { echo -e "${YELLOW}!${NC} $1"; }

# =============================================================================
# STAGE 0: VPS Environment Check
# =============================================================================

log "Stage 0: VPS Environment Check"

# Check if we're in the expected location
if [[ ! "$PWD" =~ projects/clpr ]]; then
    warn "Not in expected VPS location (~/projects/clpr)"
    warn "Current location: $PWD"
fi

# Check for external Vault service
if ! docker ps --filter "name=vault" --format "{{.Names}}" | grep -q vault; then
    error "Vault container not running. Expected external Vault service."
    error "Start Vault from ~/projects/vault first."
    exit 1
fi
success "Vault container is running"

# Detect Vault container name for network connectivity
VAULT_CONTAINER=$(docker ps --filter "name=vault" --format "{{.Names}}" | head -1)
success "Found Vault container: $VAULT_CONTAINER"

# Check for external Caddy service
if ! docker ps --filter "name=caddy" --format "{{.Names}}" | grep -q caddy; then
    warn "Caddy container not running (will check after deployment)"
    CADDY_RUNNING=false
else
    CADDY_CONTAINER=$(docker ps --filter "name=caddy" --format "{{.Names}}" | head -1)
    success "Found Caddy container: $CADDY_CONTAINER"
    CADDY_RUNNING=true
fi

# =============================================================================
# STAGE 1: Git Operations (Optional)
# =============================================================================

if [ "$SKIP_GIT" = false ]; then
    log "Stage 1: Git operations"

    current_branch=$(git rev-parse --abbrev-ref HEAD)
    log "Current branch: $current_branch"

    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
        warn "Uncommitted changes detected — deploying with local changes"
    fi

    # Optionally pull latest changes
    if [ "$NO_PULL" = false ]; then
        log "Fetching latest changes..."
        git fetch origin main || warn "Could not fetch from origin"

        if git rev-parse origin/main >/dev/null 2>&1; then
            log "Pulling latest from origin/main..."
            git pull origin main || warn "Could not pull from origin/main"
            success "Updated from origin/main"
        else
            warn "origin/main not found, using local branch"
        fi
    else
        log "Skipping pull (--no-pull flag set)"
    fi
else
    log "Skipping git operations (--skip-git flag set)"
fi

# =============================================================================
# STAGE 2: Docker Prerequisites
# =============================================================================

log "Stage 2: Docker prerequisites"

# Ensure external 'web' network exists (shared with Caddy)
if ! docker network inspect web >/dev/null 2>&1; then
    log "Creating 'web' network..."
    docker network create web
    success "Network 'web' created"
else
    success "Network 'web' exists"
fi

# Check Vault AppRole files
if [ ! -f "vault/approle/role_id" ] || [ ! -f "vault/approle/secret_id" ]; then
    error "Vault AppRole files missing (vault/approle/role_id or vault/approle/secret_id)"
    error "Generate them using:"
    error "  vault read -field=role_id auth/approle/role/clpr-backend/role-id > vault/approle/role_id"
    error "  vault write -field=secret_id -f auth/approle/role/clpr-backend/secret-id > vault/approle/secret_id"
    exit 1
fi
success "Vault AppRole files present"

# =============================================================================
# STAGE 3: Update docker-compose for VPS
# =============================================================================

log "Stage 3: Ensuring VPS-compatible docker-compose configuration"

# Use docker-compose.vps.yml if it exists, otherwise use prod
if [ -f "docker-compose.vps.yml" ]; then
    COMPOSE_FILE="docker-compose.vps.yml"
    success "Using docker-compose.vps.yml"
else
    COMPOSE_FILE="docker-compose.prod.yml"
    success "Using docker-compose.prod.yml"
fi

# =============================================================================
# STAGE 4: Build Images
# =============================================================================

log "Stage 4: Building images"
docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" build --pull backend frontend postgres
success "Images built"

# =============================================================================
# STAGE 5: Start Database & Secrets Infrastructure
# =============================================================================

log "Stage 5: Starting database and secrets infrastructure"

# Stop existing services gracefully
log "Stopping existing services..."
docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down --remove-orphans || true

# Start vault-agent, postgres, and redis
docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d vault-agent postgres redis

# Wait for vault-agent to render secrets
log "Waiting for Vault agent to render secrets..."
max_retries=60
retry=0
while [ $retry -lt $max_retries ]; do
    if docker exec clpr-vault-agent test -s /vault-agent/rendered/backend.env 2>/dev/null && \
       docker exec clpr-vault-agent test -s /vault-agent/rendered/postgres.env 2>/dev/null; then
        success "Vault agent secrets ready"
        break
    fi
    retry=$((retry + 1))
    if [ $((retry % 10)) -eq 0 ]; then
        log "Still waiting for secrets... (${retry}/${max_retries})"
    fi
    sleep 1
done

if [ $retry -eq $max_retries ]; then
    error "Vault agent secrets not ready after ${max_retries}s"
    log "Checking vault-agent logs:"
    docker logs --tail=50 clpr-vault-agent
    exit 1
fi

# Wait for postgres to be healthy
log "Waiting for PostgreSQL to be healthy..."
retry=0
while [ $retry -lt $max_retries ]; do
    if docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" ps postgres | grep -q "healthy"; then
        success "PostgreSQL is healthy"
        break
    fi
    retry=$((retry + 1))
    sleep 1
done

if [ $retry -eq $max_retries ]; then
    error "PostgreSQL not healthy after ${max_retries}s"
    exit 1
fi

# =============================================================================
# STAGE 6: Apply Database Migrations
# =============================================================================

if [ "$SKIP_MIGRATIONS" = false ]; then
    log "Stage 6: Applying database migrations"

    # Run migrations in a one-off container, sourcing Vault credentials
    if docker run --rm \
        --network container:clpr-postgres \
        --volumes-from clpr-vault-agent \
        -v "$PWD/backend/migrations:/migrations:ro" \
        --entrypoint /bin/sh migrate/migrate:latest \
        -c 'set -e; set -a; . /vault-agent/rendered/postgres.env; set +a; \
            migrate -path /migrations \
                    -database "postgresql://clpr:${POSTGRES_PASSWORD}@localhost:5432/clpr_db?sslmode=disable" up'; then
        success "Migrations applied"
    else
        error "Migration failed"
        exit 1
    fi
else
    log "Stage 6: Skipping migrations (--skip-migrations flag set)"
fi

# =============================================================================
# STAGE 7: Start Application Services
# =============================================================================

log "Stage 7: Starting application services"
docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d backend frontend
success "Services started"

# =============================================================================
# STAGE 8: Verify Health
# =============================================================================

log "Stage 8: Verifying service health"

# Wait for healthchecks
max_retries=60
retry=0
backend_healthy=false
frontend_healthy=false

while [ $retry -lt $max_retries ]; do
    if docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" ps backend | grep -q "healthy"; then
        backend_healthy=true
    fi
    if docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" ps frontend | grep -q "healthy"; then
        frontend_healthy=true
    fi

    if [ "$backend_healthy" = true ] && [ "$frontend_healthy" = true ]; then
        success "All services healthy"
        break
    fi

    retry=$((retry + 1))
    if [ $((retry % 10)) -eq 0 ]; then
        log "Waiting for health checks... Backend: $backend_healthy, Frontend: $frontend_healthy (${retry}/${max_retries})"
    fi
    sleep 2
done

if [ $retry -eq $max_retries ]; then
    warn "Services did not report healthy within $((max_retries * 2))s (may still be warming up)"
fi

# =============================================================================
# STAGE 9: Verify Caddy Configuration
# =============================================================================

log "Stage 9: Verifying Caddy reverse proxy"

if [ "$CADDY_RUNNING" = true ]; then
    # Check if Caddy is on the web network
    if docker network inspect web | grep -q "$CADDY_CONTAINER"; then
        success "Caddy is connected to 'web' network"

        # Verify Caddyfile points to the right containers
        log "Checking Caddyfile configuration..."
        if [ -f "Caddyfile" ]; then
            if grep -q "clpr-backend" Caddyfile && grep -q "clpr-frontend" Caddyfile; then
                success "Caddyfile references clpr containers"
            else
                warn "Caddyfile may need updating to reference clpr-backend and clpr-frontend"
            fi
        fi

        # Suggest reloading Caddy
        log "To reload Caddy configuration, run:"
        log "  docker exec $CADDY_CONTAINER caddy reload --config /etc/caddy/Caddyfile"
    else
        warn "Caddy is not connected to 'web' network"
        log "Connect Caddy to 'web' network:"
        log "  docker network connect web $CADDY_CONTAINER"
    fi
else
    warn "Caddy not running. Start Caddy from ~/projects/caddy"
    log "Ensure Caddy Caddyfile includes this configuration for clpr.tv"
fi

# =============================================================================
# STAGE 10: Final Status
# =============================================================================

log "Stage 10: Deployment Status"

# Print container status
log "Container status:"
docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" ps

# Print network connections
log "Network connections (web):"
docker network inspect web --format '{{range .Containers}}{{.Name}} {{end}}' || true

# Test internal connectivity
log "Testing internal backend health:"
if docker exec clpr-backend wget -qO- http://localhost:8080/api/v1/health 2>/dev/null | grep -q "ok\|healthy\|status"; then
    success "Backend responding to health checks"
else
    warn "Backend may not be responding to health checks yet"
fi

# Print URLs
echo ""
success "Deployment complete!"
echo ""
echo "Service endpoints (internal):"
echo "  Backend:  http://clpr-backend:8080 (internal Docker network)"
echo "  Frontend: http://clpr-frontend:80 (internal Docker network)"
echo "  Postgres: postgres://postgres:5432 (internal) / localhost:5436 (external)"
echo ""
echo "Public access (via Caddy):"
echo "  Website:  https://clpr.tv (ensure Caddy is configured and running)"
echo ""
echo "Next steps:"
echo "  1. Verify Caddy is running: docker ps | grep caddy"
echo "  2. Check Caddy config points to clpr-backend and clpr-frontend"
echo "  3. Reload Caddy: docker exec <caddy-container> caddy reload --config /etc/caddy/Caddyfile"
echo "  4. Test https://clpr.tv in browser"
echo ""

# Print recent backend logs (sanitized)
log "Recent backend logs (last 30 lines):"
docker logs --tail=30 clpr-backend 2>&1 | \
    sed -e 's/JWT_PRIVATE_KEY.*/JWT_PRIVATE_KEY=[REDACTED]/' \
        -e 's/JWT_PUBLIC_KEY.*/JWT_PUBLIC_KEY=[REDACTED]/' \
        -e 's/DB_PASSWORD.*/DB_PASSWORD=[REDACTED]/' \
        -e 's/TWITCH_CLIENT_SECRET.*/TWITCH_CLIENT_SECRET=[REDACTED]/' \
        -e 's/password[^[:space:]]*/password=[REDACTED]/gi' || true

echo ""
success "VPS deployment script complete!"
