#!/bin/bash
set -euo pipefail

# VPS Deployment Verification Script
# Checks all aspects of the deployment to ensure everything is working

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

PASSED=0
FAILED=0
WARNINGS=0

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

check_warn() {
    echo -e "${YELLOW}!${NC} $1"
    ((WARNINGS++))
}

section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Header
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   VPS Deployment Verification                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}"
echo ""

# Check 1: Docker prerequisites
section "1. Docker Prerequisites"

if command -v docker >/dev/null 2>&1; then
    check_pass "Docker is installed"
else
    check_fail "Docker is not installed"
fi

if docker compose version >/dev/null 2>&1; then
    check_pass "Docker Compose is available"
else
    check_fail "Docker Compose is not available"
fi

if docker network inspect web >/dev/null 2>&1; then
    check_pass "Network 'web' exists"
else
    check_fail "Network 'web' does not exist (run: docker network create web)"
fi

# Check 2: Vault infrastructure
section "2. Vault Infrastructure"

if docker ps --format '{{.Names}}' | grep -q vault; then
    VAULT_CONTAINER=$(docker ps --format '{{.Names}}' | grep vault | head -1)
    check_pass "Vault container is running: $VAULT_CONTAINER"
    
    # Check if Vault is on web network
    if docker network inspect web | grep -q "$VAULT_CONTAINER"; then
        check_pass "Vault is connected to 'web' network"
    else
        check_warn "Vault is not connected to 'web' network"
        echo "         Fix: docker network connect web $VAULT_CONTAINER"
    fi
else
    check_fail "Vault container is not running"
fi

if [ -f "vault/approle/role_id" ] && [ -f "vault/approle/secret_id" ]; then
    check_pass "Vault AppRole credentials exist"
else
    check_fail "Vault AppRole credentials missing"
    echo "         Generate with:"
    echo "         vault read -field=role_id auth/approle/role/clpr-backend/role-id > vault/approle/role_id"
    echo "         vault write -field=secret_id -f auth/approle/role/clpr-backend/secret-id > vault/approle/secret_id"
fi

# Check 3: Clipper containers
section "3. Clipper Containers"

COMPOSE_FILE="docker-compose.vps.yml"
if [ ! -f "$COMPOSE_FILE" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
fi

declare -a REQUIRED_CONTAINERS=("clpr-vault-agent" "clpr-postgres" "clpr-redis" "clpr-backend" "clpr-frontend")

for container in "${REQUIRED_CONTAINERS[@]}"; do
    if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        # Check if healthy
        STATUS=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "no-health-check")
        if [ "$STATUS" = "healthy" ]; then
            check_pass "$container is running and healthy"
        elif [ "$STATUS" = "no-health-check" ]; then
            check_pass "$container is running (no health check)"
        else
            check_warn "$container is running but not healthy (status: $STATUS)"
        fi
        
        # Check if on web network (for backend and frontend)
        if [[ "$container" =~ (backend|frontend) ]]; then
            if docker network inspect web | grep -q "$container"; then
                check_pass "$container is connected to 'web' network"
            else
                check_fail "$container is NOT connected to 'web' network"
                echo "         Fix: docker network connect web $container"
            fi
        fi
    else
        check_fail "$container is not running"
    fi
done

# Check 4: Vault secrets rendered
section "4. Vault Secrets"

if docker ps --format '{{.Names}}' | grep -q "clpr-vault-agent"; then
    if docker exec clpr-vault-agent test -f /vault-agent/rendered/backend.env 2>/dev/null; then
        check_pass "Backend secrets rendered"
    else
        check_fail "Backend secrets not rendered"
    fi
    
    if docker exec clpr-vault-agent test -f /vault-agent/rendered/postgres.env 2>/dev/null; then
        check_pass "Postgres secrets rendered"
    else
        check_fail "Postgres secrets not rendered"
    fi
else
    check_warn "vault-agent not running, cannot check secrets"
fi

# Check 5: Service health checks
section "5. Service Health"

if docker ps --format '{{.Names}}' | grep -q "clpr-backend"; then
    if docker exec clpr-backend wget -qO- --timeout=5 http://localhost:8080/api/v1/health 2>/dev/null | grep -q "ok\|healthy\|status"; then
        check_pass "Backend health endpoint responds"
    else
        check_fail "Backend health endpoint not responding"
    fi
else
    check_warn "Backend not running, cannot check health"
fi

if docker ps --format '{{.Names}}' | grep -q "clpr-postgres"; then
    if docker exec clpr-postgres pg_isready -U clpr -d clpr_db >/dev/null 2>&1; then
        check_pass "PostgreSQL is ready"
    else
        check_fail "PostgreSQL is not ready"
    fi
else
    check_warn "PostgreSQL not running"
fi

if docker ps --format '{{.Names}}' | grep -q "clpr-redis"; then
    if docker exec clpr-redis redis-cli ping >/dev/null 2>&1; then
        check_pass "Redis is responding"
    else
        check_fail "Redis is not responding"
    fi
else
    check_warn "Redis not running"
fi

# Check 6: Caddy reverse proxy
section "6. Caddy Reverse Proxy"

if docker ps --format '{{.Names}}' | grep -q caddy; then
    CADDY_CONTAINER=$(docker ps --format '{{.Names}}' | grep caddy | head -1)
    check_pass "Caddy container is running: $CADDY_CONTAINER"
    
    # Check if Caddy is on web network
    if docker network inspect web | grep -q "$CADDY_CONTAINER"; then
        check_pass "Caddy is connected to 'web' network"
    else
        check_fail "Caddy is NOT connected to 'web' network"
        echo "         Fix: docker network connect web $CADDY_CONTAINER"
    fi
    
    # Check if Caddy can reach backend
    if docker ps --format '{{.Names}}' | grep -q "clpr-backend"; then
        if docker exec "$CADDY_CONTAINER" wget -qO- --timeout=5 http://clpr-backend:8080/api/v1/health 2>/dev/null >/dev/null; then
            check_pass "Caddy can reach clpr-backend"
        else
            check_fail "Caddy cannot reach clpr-backend"
        fi
    fi
    
    # Check if Caddy can reach frontend
    if docker ps --format '{{.Names}}' | grep -q "clpr-frontend"; then
        if docker exec "$CADDY_CONTAINER" wget -qO- --timeout=5 http://clpr-frontend:80/ 2>/dev/null >/dev/null; then
            check_pass "Caddy can reach clpr-frontend"
        else
            check_fail "Caddy cannot reach clpr-frontend"
        fi
    fi
    
    # Check if port 80 and 443 are bound
    if docker port "$CADDY_CONTAINER" | grep -q "80/tcp"; then
        check_pass "Caddy is listening on port 80"
    else
        check_warn "Caddy is not listening on port 80"
    fi
    
    if docker port "$CADDY_CONTAINER" | grep -q "443/tcp"; then
        check_pass "Caddy is listening on port 443"
    else
        check_warn "Caddy is not listening on port 443"
    fi
else
    check_warn "Caddy container is not running"
    echo "         Start Caddy with: ./scripts/caddy-setup.sh start"
fi

# Check 7: External access (if possible)
section "7. External Access"

# Try to check if domain resolves
if command -v dig >/dev/null 2>&1; then
    if dig +short clpr.tv | grep -q '^[0-9]'; then
        check_pass "clpr.tv DNS resolves"
    else
        check_warn "clpr.tv DNS does not resolve (or dig not available)"
    fi
else
    check_warn "dig not available, cannot check DNS"
fi

# Try to access the site (only if we're on the VPS)
if command -v curl >/dev/null 2>&1; then
    if curl -sSf -m 5 http://localhost/api/v1/health >/dev/null 2>&1; then
        check_pass "HTTP access to localhost works"
    else
        check_warn "Cannot access http://localhost/api/v1/health"
    fi
else
    check_warn "curl not available, cannot test HTTP access"
fi

# Summary
section "Summary"
echo ""
echo "  Passed:   $PASSED"
echo "  Failed:   $FAILED"
echo "  Warnings: $WARNINGS"
echo ""

if [ $FAILED -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   All checks passed! Deployment looks good.   ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}"
    exit 0
elif [ $FAILED -eq 0 ]; then
    echo -e "${YELLOW}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║   Deployment OK with warnings.                ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║   Deployment has issues that need fixing.     ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════╝${NC}"
    exit 1
fi
