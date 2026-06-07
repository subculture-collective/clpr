#!/bin/bash

# Setup test environment with all required configurations
# This script ensures all skipped tests can run by providing necessary env vars

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Setting up test environment...${NC}"

# Generate a 32-byte key for MFA encryption (required for MFA tests)
if [ -z "$MFA_ENCRYPTION_KEY" ]; then
    export MFA_ENCRYPTION_KEY=$(openssl rand -base64 32)
    echo -e "${GREEN}✓ Generated MFA_ENCRYPTION_KEY${NC}"
else
    echo -e "${GREEN}✓ Using existing MFA_ENCRYPTION_KEY${NC}"
fi

# Stripe webhook secret for webhook signature testing
if [ -z "$TEST_STRIPE_WEBHOOK_SECRET" ]; then
    export TEST_STRIPE_WEBHOOK_SECRET="whsec_test_$(openssl rand -hex 24)"
    echo -e "${GREEN}✓ Generated TEST_STRIPE_WEBHOOK_SECRET${NC}"
else
    echo -e "${GREEN}✓ Using existing TEST_STRIPE_WEBHOOK_SECRET${NC}"
fi

# Set Stripe webhook secret for the main config too
export STRIPE_WEBHOOK_SECRET="${TEST_STRIPE_WEBHOOK_SECRET}"

# OpenSearch/Elasticsearch configuration for semantic search
if [ -z "$OPENSEARCH_URL" ]; then
    export OPENSEARCH_URL="http://localhost:9201"
    echo -e "${GREEN}✓ Set OPENSEARCH_URL=${OPENSEARCH_URL}${NC}"
fi

# Test database configuration
export TEST_DATABASE_HOST="${TEST_DATABASE_HOST:-localhost}"
export TEST_DATABASE_PORT="${TEST_DATABASE_PORT:-5437}"
export TEST_DATABASE_USER="${TEST_DATABASE_USER:-clpr}"
export TEST_DATABASE_PASSWORD="${TEST_DATABASE_PASSWORD:-clpr_password}"
export TEST_DATABASE_NAME="${TEST_DATABASE_NAME:-clpr_test}"

# Compose a full connection URL for tests that consume TEST_DATABASE_URL directly
export TEST_DATABASE_URL="postgres://${TEST_DATABASE_USER}:${TEST_DATABASE_PASSWORD}@${TEST_DATABASE_HOST}:${TEST_DATABASE_PORT}/${TEST_DATABASE_NAME}?sslmode=disable"

# Redis test configuration
export TEST_REDIS_HOST="${TEST_REDIS_HOST:-localhost}"
export TEST_REDIS_PORT="${TEST_REDIS_PORT:-6380}"

echo -e "${GREEN}✓ Test database configured${NC}"

# If previous test runs left migrations dirty, clean them up so migrations can proceed
# This avoids failures like "Dirty database version X. Fix and force version." during test-setup
DB_URL="postgresql://${TEST_DATABASE_USER}:${TEST_DATABASE_PASSWORD}@${TEST_DATABASE_HOST}:${TEST_DATABASE_PORT}/${TEST_DATABASE_NAME}?sslmode=disable"
# Try to detect schema_migrations state; tolerate failures if table doesn't exist yet
if command -v psql >/dev/null 2>&1; then
  DIRTY_STATE=$(PGPASSWORD=${TEST_DATABASE_PASSWORD} psql "${DB_URL}" -t -c "SELECT dirty FROM schema_migrations" 2>/dev/null | tr -d '[:space:]') || true
  CURRENT_VERSION=$(PGPASSWORD=${TEST_DATABASE_PASSWORD} psql "${DB_URL}" -t -c "SELECT version FROM schema_migrations" 2>/dev/null | tr -d '[:space:]') || true
  if [ "${DIRTY_STATE}" = "t" ] || [ "${DIRTY_STATE}" = "true" ]; then
    echo -e "${YELLOW}Detected dirty migration state (version=${CURRENT_VERSION}). Forcing to version ${CURRENT_VERSION}...${NC}"
    if command -v migrate >/dev/null 2>&1; then
      migrate -path migrations -database "${DB_URL}" force "${CURRENT_VERSION}" || true
      echo -e "${GREEN}✓ Forced migration version to ${CURRENT_VERSION}${NC}"
    else
      echo -e "${YELLOW}Warning: 'migrate' CLI not found; cannot force version automatically.${NC}"
    fi
  fi
fi

# JWT secrets for testing
if [ -z "$JWT_SECRET" ]; then
    export JWT_SECRET="test_jwt_secret_$(openssl rand -hex 16)"
    echo -e "${GREEN}✓ Generated JWT_SECRET${NC}"
fi

if [ -z "$JWT_REFRESH_SECRET" ]; then
    export JWT_REFRESH_SECRET="test_jwt_refresh_secret_$(openssl rand -hex 16)"
    echo -e "${GREEN}✓ Generated JWT_REFRESH_SECRET${NC}"
fi

# Twitch credentials (can use test values)
export TWITCH_CLIENT_ID="${TWITCH_CLIENT_ID:-test_client_id}"
export TWITCH_CLIENT_SECRET="${TWITCH_CLIENT_SECRET:-test_client_secret}"

# Session secret
if [ -z "$SESSION_SECRET" ]; then
    export SESSION_SECRET=$(openssl rand -base64 32)
    echo -e "${GREEN}✓ Generated SESSION_SECRET${NC}"
fi

# CDN Configuration for failover testing
export CDN_ENABLED="${CDN_ENABLED:-true}"
export CDN_FAILOVER_MODE="${CDN_FAILOVER_MODE:-true}"
export CDN_PRIMARY_URL="${CDN_PRIMARY_URL:-http://cdn-test.local:8081}"
export CDN_FALLBACK_URL="${CDN_FALLBACK_URL:-http://origin-test.local:8082}"
echo -e "${GREEN}✓ Set CDN failover configuration${NC}"

# Create .env.test file for persistence
cat > .env.test <<EOF
# Auto-generated test environment configuration
# Generated on $(date)

# MFA Configuration
MFA_ENCRYPTION_KEY=${MFA_ENCRYPTION_KEY}

# Stripe Configuration
TEST_STRIPE_WEBHOOK_SECRET=${TEST_STRIPE_WEBHOOK_SECRET}
STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}

# OpenSearch Configuration
OPENSEARCH_URL=${OPENSEARCH_URL}

# Database Configuration
TEST_DATABASE_HOST=${TEST_DATABASE_HOST}
TEST_DATABASE_PORT=${TEST_DATABASE_PORT}
TEST_DATABASE_USER=${TEST_DATABASE_USER}
TEST_DATABASE_PASSWORD=${TEST_DATABASE_PASSWORD}
TEST_DATABASE_NAME=${TEST_DATABASE_NAME}
TEST_DATABASE_URL=${TEST_DATABASE_URL}

# Redis Configuration
TEST_REDIS_HOST=${TEST_REDIS_HOST}
TEST_REDIS_PORT=${TEST_REDIS_PORT}

# JWT Configuration
JWT_SECRET=${JWT_SECRET}
JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}

# Session Configuration
SESSION_SECRET=${SESSION_SECRET}

# Twitch Configuration
TWITCH_CLIENT_ID=${TWITCH_CLIENT_ID}
TWITCH_CLIENT_SECRET=${TWITCH_CLIENT_SECRET}

# CDN Configuration for testing
CDN_ENABLED=${CDN_ENABLED}
CDN_FAILOVER_MODE=${CDN_FAILOVER_MODE}
CDN_PRIMARY_URL=${CDN_PRIMARY_URL}
CDN_FALLBACK_URL=${CDN_FALLBACK_URL}
EOF

echo -e "${GREEN}✓ Created .env.test file${NC}"
echo ""
echo -e "${YELLOW}Test environment is ready!${NC}"
echo -e "${YELLOW}To use this environment in your shell, run:${NC}"
echo -e "  ${GREEN}source backend/.env.test${NC}"
echo ""
echo -e "${YELLOW}Or source it before running tests:${NC}"
echo -e "  ${GREEN}set -a; source backend/.env.test; set +a; make test INTEGRATION=1 E2E=1${NC}"
