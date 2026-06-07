#!/bin/bash
set -euo pipefail

# Clipper Production Deploy Script
# Handles git operations, image builds, migrations, and service startup
# Usage: ./scripts/deploy.sh [--skip-git] [--no-pull]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Flags
SKIP_GIT=false
NO_PULL=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-git) SKIP_GIT=true; shift ;;
        --no-pull) NO_PULL=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

log() { echo -e "${BLUE}[deploy]${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
warn() { echo -e "${YELLOW}!${NC} $1"; }

# =============================================================================
# STAGE 1: Git Operations
# =============================================================================

if [ "$SKIP_GIT" = false ]; then
    log "Stage 1: Git operations"

    current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" != "deploy/production" ]; then
        error "Not on deploy/production branch (currently on $current_branch)"
        exit 1
    fi
    success "On deploy/production branch"

    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        warn "Uncommitted changes detected"
        log "Stashing changes..."
        git stash push -m "auto-stash-before-deploy-$(date +%s)"
        success "Changes stashed"
    fi

    # Check for untracked files
    if [ -n "$(git ls-files --others --exclude-standard)" ]; then
        warn "Untracked files detected (will be left as-is)"
    fi

    # Optionally pull latest main and merge
    if [ "$NO_PULL" = false ]; then
        log "Fetching latest changes..."
        git fetch origin main
        success "Fetched origin/main"

        log "Merging latest main into deploy/production (keeping slim deployment footprint)..."
        if git merge -X ours origin/main -m "Merge branch 'main' into deploy/production (keep deployment-only footprint)"; then
            success "Merged main with 'ours' strategy"
        else
            # Handle modify/delete conflicts: keep deletions from deploy/production
            log "Resolving modify/delete conflicts (keeping deleted state)..."
            git status --porcelain | grep '^DU' | awk '{print $2}' | while read file; do
                git rm "$file" 2>/dev/null || true
            done

            log "Pruning unnecessary docs/test/support files..."
            # Remove test/docs/e2e/support files that got reintroduced
            files_to_remove=(
                "RATE_LIMIT_UI_IMPLEMENTATION.md"
                "backend/TEST_SETUP_GUIDE.md"
                "backend/TEST_SETUP_SUMMARY.md"
                "backend/docs"
                "backend/run-tests-verbose.sh"
                "backend/setup-test-env.sh"
                "backend/test-commands.sh"
                "docs"
                "frontend/E2E_*.md"
                "frontend/FRONTEND_TEST_SETUP_SUMMARY.md"
                "frontend/PLAYWRIGHT_SETUP_GUIDE.md"
                "frontend/e2e"
                "frontend/run-playwright-tests.sh"
                "frontend/setup-e2e-tests.sh"
                "frontend/test-commands.sh"
                "infrastructure/k8s/base/BACKUP_SETUP.md"
                "infrastructure/k8s/base/backup-cronjobs.yaml"
                "infrastructure/k8s/base/postgres-pitr-config.yaml"
            )
            for pattern in "${files_to_remove[@]}"; do
                git rm -rf --ignore-unmatch "$pattern" 2>/dev/null || true
            done

            log "Completing merge..."
            git commit -m "Merge branch 'main' into deploy/production (keep deployment-only footprint)" --no-edit
            success "Merged main successfully (conflicts resolved, slim deployment maintained)"
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

# Ensure external 'web' network exists
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
    exit 1
fi
success "Vault AppRole files present"

# =============================================================================
# STAGE 3: Build Images
# =============================================================================

log "Stage 3: Building images"
docker compose -f docker-compose.prod.yml build --pull --no-cache backend frontend postgres
success "Images built"

# =============================================================================
# STAGE 4: Start Database & Secrets
# =============================================================================

log "Stage 4: Starting database and secrets infrastructure"
docker compose -f docker-compose.prod.yml up -d vault-agent postgres redis

# Wait for postgres and vault-agent
log "Waiting for services to be ready..."
max_retries=30
retry=0
while [ $retry -lt $max_retries ]; do
    if docker exec clpr-vault-agent test -s /vault-agent/rendered/postgres.env 2>/dev/null; then
        success "Vault agent secrets ready"
        break
    fi
    retry=$((retry + 1))
    sleep 1
done

if [ $retry -eq $max_retries ]; then
    error "Vault agent secrets not ready after ${max_retries}s"
    exit 1
fi

# =============================================================================
# STAGE 5: Apply Database Migrations
# =============================================================================

log "Stage 5: Applying database migrations"

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

# =============================================================================
# STAGE 6: Start Application Services
# =============================================================================

log "Stage 6: Starting application services"
docker compose -f docker-compose.prod.yml up -d backend frontend
success "Services started"

# =============================================================================
# STAGE 7: Verify Health
# =============================================================================

log "Stage 7: Verifying service health"

# Wait for healthchecks
max_retries=30
retry=0
while [ $retry -lt $max_retries ]; do
    if docker compose -f docker-compose.prod.yml ps | grep -q "clpr-backend.*healthy"; then
        if docker compose -f docker-compose.prod.yml ps | grep -q "clpr-frontend.*healthy"; then
            success "All services healthy"
            break
        fi
    fi
    retry=$((retry + 1))
    sleep 1
done

if [ $retry -eq $max_retries ]; then
    warn "Services did not report healthy within ${max_retries}s (may still be warming up)"
fi

# Print status
log "Final service status:"
docker compose -f docker-compose.prod.yml ps

# Print latest backend logs
log "Recent backend logs:"
docker logs --tail=20 clpr-backend 2>&1 | sed -e 's/JWT_PRIVATE_KEY.*/JWT_PRIVATE_KEY=[redacted]/' \
                                                 -e 's/JWT_PUBLIC_KEY.*/JWT_PUBLIC_KEY=[redacted]/' \
                                                 -e 's/DB_PASSWORD.*/DB_PASSWORD=[redacted]/'

success "Deployment complete!"
echo ""
echo "Service endpoints:"
echo "  Frontend: http://localhost:80 (or via Caddyfile reverse proxy)"
echo "  Backend:  http://localhost:8080"
echo "  Health:   http://localhost:8080/api/v1/health"
echo ""
