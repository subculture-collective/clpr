---
title: "Blue Green Deployment"
summary: "Complete guide to implementing and using blue/green deployments for zero-downtime releases in Clipper."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Blue/Green Deployment Guide

Complete guide to implementing and using blue/green deployments for zero-downtime releases in Clipper.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Deployment Process](#deployment-process)
- [Traffic Switching](#traffic-switching)
- [Database Migrations](#database-migrations)
- [Monitoring](#monitoring)
- [Rollback](#rollback)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

## Overview

Blue/green deployment is a release strategy that maintains two identical production environments:

- **Blue Environment**: Currently serving production traffic
- **Green Environment**: New version being deployed and tested

### Benefits

✓ **Zero Downtime**: Switch traffic instantly between environments  
✓ **Instant Rollback**: Revert to previous version in seconds  
✓ **Safe Testing**: Verify new version in production before switching traffic  
✓ **Reduced Risk**: Old version remains running during deployment  
✓ **Easy Comparison**: A/B test between versions if needed

### Architecture

```
                    ┌─────────────────┐
                    │  Caddy Proxy    │
                    │  (Port 80/443)  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Active: Blue   │ ◄─── ACTIVE_ENV variable
                    └────────┬────────┘
                             │
            ┌────────────────┴───────────────┐
            │                                │
    ┌───────▼────────┐              ┌───────▼────────┐
    │  Blue Env      │              │  Green Env     │
    │  (Current)     │              │  (New)         │
    ├────────────────┤              ├────────────────┤
    │ Backend:8080   │              │ Backend:8080   │
    │ Frontend:80    │              │ Frontend:80    │
    └───────┬────────┘              └───────┬────────┘
            │                               │
            └───────────┬───────────────────┘
                        │
                ┌───────▼────────┐
                │  PostgreSQL    │
                │  Redis         │
                └────────────────┘
```

## Prerequisites

### Infrastructure

- Docker Engine 24+ with Compose v2
- 4GB RAM minimum (2GB per environment)
- 20GB disk space minimum
- Ubuntu 22.04 LTS or similar

### Software

```bash
# Check Docker version
docker --version  # Should be 24.0+
docker compose version  # Should be v2.0+

# Check available resources
free -h  # At least 4GB free RAM
df -h    # At least 20GB free disk
```

### Access

- SSH access to production server
- GitHub Container Registry access
- Database credentials
- Environment variables configured

## Setup

### 1. Initial Configuration

```bash
# Clone repository
git clone https://git.subcult.tv/subculture-collective/clpr.git
cd clpr

# Copy environment file
cp .env.production.example .env

# Edit with your values
nano .env
```

### 2. Environment Variables

Create `.env` file with required variables:

```bash
# Database
POSTGRES_DB=clpr_db
POSTGRES_USER=clpr
POSTGRES_PASSWORD=<secure-password>

# Image Tags (optional, defaults to 'latest')
BACKEND_BLUE_TAG=latest
FRONTEND_BLUE_TAG=latest
BACKEND_GREEN_TAG=latest
FRONTEND_GREEN_TAG=latest

# Application
BASE_URL=https://clpr.tv
CORS_ALLOWED_ORIGINS=https://clpr.tv

# Twitch API
TWITCH_CLIENT_ID=<your-client-id>
TWITCH_CLIENT_SECRET=<your-client-secret>

# Active Environment
ACTIVE_ENV=blue
```

### 3. Initial Deployment

```bash
# Start shared services (database, redis)
docker compose -f docker-compose.blue-green.yml up -d postgres redis

# Wait for database to be ready
sleep 10

# Start blue environment (default)
docker compose -f docker-compose.blue-green.yml up -d backend-blue frontend-blue

# Start Caddy proxy
docker compose -f docker-compose.blue-green.yml up -d caddy

# Verify
docker ps
curl http://localhost/health
```

## Deployment Process

### Automated Deployment

Use the provided script for automated deployments:

```bash
# Run blue-green deployment
./scripts/blue-green-deploy.sh

# With custom image tag
IMAGE_TAG=v1.2.3 ./scripts/blue-green-deploy.sh

# With monitoring enabled
MONITORING_ENABLED=true ./scripts/blue-green-deploy.sh
```

The script automatically:
1. Detects active environment
2. Pulls new images for target environment
3. Runs database migrations
4. Starts target environment
5. Performs health checks
6. Switches traffic
7. Monitors for issues
8. Stops old environment on success
9. Rolls back on failure

### Manual Deployment

For step-by-step control:

```bash
# 1. Detect active environment
ACTIVE_ENV=$(docker ps --format '{{.Names}}' | grep -o 'backend-\(blue\|green\)' | head -1 | cut -d'-' -f2)
echo "Active: $ACTIVE_ENV"

# 2. Set target environment
if [ "$ACTIVE_ENV" = "blue" ]; then
    TARGET_ENV="green"
else
    TARGET_ENV="blue"
fi
echo "Target: $TARGET_ENV"

# 3. Pull latest images
export BACKEND_${TARGET_ENV^^}_TAG=latest
export FRONTEND_${TARGET_ENV^^}_TAG=latest
docker compose -f docker-compose.blue-green.yml --profile $TARGET_ENV pull

# 4. Start target environment
docker compose -f docker-compose.blue-green.yml --profile $TARGET_ENV up -d backend-$TARGET_ENV frontend-$TARGET_ENV

# 5. Wait for startup
sleep 30

# 6. Health check
docker exec clpr-backend-$TARGET_ENV wget --spider -q http://localhost:8080/health
docker exec clpr-frontend-$TARGET_ENV wget --spider -q http://localhost:80/health.html

# 7. Switch traffic
export ACTIVE_ENV=$TARGET_ENV
docker compose -f docker-compose.blue-green.yml up -d caddy

# 8. Verify
curl http://localhost/health
curl http://localhost/api/v1/health

# 9. Monitor (watch for 2-5 minutes)
watch -n 5 'curl -s http://localhost/health | jq'

# 10. Stop old environment
docker compose -f docker-compose.blue-green.yml --profile $ACTIVE_ENV stop
```

## Traffic Switching

### Using Caddy

Traffic is controlled by the `ACTIVE_ENV` environment variable:

```bash
# Switch to blue
export ACTIVE_ENV=blue
docker compose -f docker-compose.blue-green.yml up -d caddy

# Switch to green
export ACTIVE_ENV=green
docker compose -f docker-compose.blue-green.yml up -d caddy

# Verify current environment
curl -s http://localhost/health | jq .environment
```

### Gradual Rollout (Advanced)

For gradual traffic migration, modify Caddyfile to use weighted load balancing:

```caddyfile
# Example: 90% blue, 10% green
handle /api/* {
    reverse_proxy {
        to clpr-backend-blue:8080 clpr-backend-blue:8080 \
           clpr-backend-blue:8080 clpr-backend-blue:8080 \
           clpr-backend-blue:8080 clpr-backend-blue:8080 \
           clpr-backend-blue:8080 clpr-backend-blue:8080 \
           clpr-backend-blue:8080 clpr-backend-green:8080
        lb_policy round_robin
    }
}
```

## Database Migrations

### Backward Compatible Migrations

For zero-downtime deployments, migrations MUST be backward compatible:

#### Safe Operations

```sql
-- ✓ Add new table
CREATE TABLE new_feature (...);

-- ✓ Add nullable column
ALTER TABLE clips ADD COLUMN new_field TEXT;

-- ✓ Add column with default
ALTER TABLE clips ADD COLUMN featured BOOLEAN DEFAULT false;

-- ✓ Create index (use CONCURRENTLY)
CREATE INDEX CONCURRENTLY idx_clips_featured ON clips(featured);

-- ✓ Add constraints (if data already validates)
ALTER TABLE clips ADD CONSTRAINT check_positive_votes CHECK (upvotes >= 0);
```

#### Unsafe Operations (Require Two-Phase Migration)

```sql
-- ✗ Drop table (old version still uses it)
DROP TABLE old_table;

-- ✗ Drop column (old version still references it)
ALTER TABLE clips DROP COLUMN old_field;

-- ✗ Rename column (old version uses old name)
ALTER TABLE clips RENAME COLUMN old_name TO new_name;

-- ✗ Change column type (may break old version)
ALTER TABLE clips ALTER COLUMN score TYPE BIGINT;
```

### Two-Phase Migration Strategy

For breaking changes, use a two-phase approach:

#### Phase 1: Additive Changes (Deploy with old code still working)

```sql
-- Add new column alongside old one
ALTER TABLE clips ADD COLUMN new_featured_at TIMESTAMPTZ;

-- Backfill data
UPDATE clips SET new_featured_at = featured_date WHERE featured = true;

-- Create index
CREATE INDEX CONCURRENTLY idx_clips_new_featured ON clips(new_featured_at);
```

Deploy new code that writes to BOTH old and new columns.

#### Phase 2: Remove Old (After new version is stable)

```sql
-- Remove old column
ALTER TABLE clips DROP COLUMN featured_date;

-- Rename new column if needed
ALTER TABLE clips RENAME COLUMN new_featured_at TO featured_at;
```

### Migration Verification

```bash
# Check migration compatibility
./scripts/check-migration-compatibility.sh

# Test migrations in staging first
cd backend
migrate -path migrations -database "$STAGING_DB_URL" up

# Verify in staging before production
```

## Monitoring

### Health Check Endpoints

Available health endpoints:

```bash
# Basic health
curl http://localhost/health

# API health with version
curl http://localhost/api/v1/health

# Readiness check (database + redis)
curl http://localhost/health/ready

# Liveness check
curl http://localhost/health/live

# Cache stats
curl http://localhost/health/cache
```

### Key Metrics to Monitor

During and after deployment, monitor:

1. **Error Rate**: Should stay below 0.1%
2. **Response Time**: p95 < 200ms, p99 < 500ms
3. **Throughput**: Requests per second
4. **Database Connections**: Should be stable
5. **Memory Usage**: No memory leaks
6. **CPU Usage**: Should not spike dramatically

### Monitoring Commands

```bash
# Watch container stats
docker stats clpr-backend-blue clpr-backend-green

# Monitor logs for errors
docker compose logs -f --tail=100 | grep -i error

# Check response times
time curl http://localhost/api/v1/clips

# Monitor database connections
docker exec clpr-postgres psql -U clpr -d clpr_db -c \
  "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"

# Check Redis stats
docker exec clpr-redis redis-cli INFO stats
```

### Integration with Monitoring Tools

Configure monitoring notifications:

```bash
# Set in .env or export
export MONITORING_ENABLED=true
export WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Run deployment with monitoring
./scripts/blue-green-deploy.sh
```

## Rollback

See [Blue/Green Rollback Procedures](./BLUE_GREEN_ROLLBACK.md) for detailed rollback documentation.

### Quick Rollback

```bash
# Automatic rollback script
./scripts/rollback-blue-green.sh

# Or manual one-liner
export ACTIVE_ENV=blue && docker compose -f docker-compose.blue-green.yml up -d caddy
```

## Testing

### Pre-Deployment Testing

```bash
# 1. Test in staging
cd /opt/clpr-staging
./scripts/blue-green-deploy.sh

# 2. Run smoke tests
./scripts/staging-rehearsal.sh

# 3. Check migration compatibility
./scripts/check-migration-compatibility.sh

# 4. Verify health checks
./scripts/health-check.sh
```

### Post-Deployment Validation

```bash
# Automated validation
./scripts/validate-deployment.sh

# Manual validation
# 1. Check all critical endpoints
curl http://localhost/api/v1/clips
curl http://localhost/api/v1/users/me
curl http://localhost/api/v1/health

# 2. Test user flows
# - Login
# - Browse clips
# - Submit clip
# - Vote on clip
# - Comment

# 3. Check logs for errors
docker compose logs --tail=500 | grep -i error

# 4. Monitor metrics for 15 minutes
watch -n 30 'curl -s http://localhost/health/stats'
```

## Troubleshooting

### Common Issues

#### Issue: Health Checks Failing

```bash
# Check container logs
docker compose logs backend-green

# Test health endpoint directly
docker exec clpr-backend-green wget -O- http://localhost:8080/health

# Check database connectivity
docker exec clpr-backend-green psql -h postgres -U clpr -d clpr_db -c "SELECT 1"
```

#### Issue: High Memory Usage

```bash
# Check memory per container
docker stats --no-stream

# Check for memory leaks
docker exec clpr-backend-green ps aux --sort=-%mem | head

# Restart if needed
docker compose restart backend-green
```

#### Issue: Traffic Not Switching

```bash
# Verify ACTIVE_ENV
docker exec clpr-caddy env | grep ACTIVE_ENV

# Check Caddy config
docker exec clpr-caddy caddy environ

# Reload Caddy
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile

# Restart Caddy
docker compose restart caddy
```

### Debug Commands

```bash
# View all running containers
docker ps --filter "name=clpr"

# Check network connectivity
docker exec clpr-backend-blue ping -c 3 postgres
docker exec clpr-backend-green ping -c 3 redis

# Inspect container health
docker inspect clpr-backend-blue | jq '.[0].State.Health'

# View Caddy logs
docker compose logs caddy --tail=100

# Test from inside container
docker exec -it clpr-backend-blue sh
wget -O- http://localhost:8080/health
```

## Best Practices

1. **Always Test in Staging First**: Never deploy directly to production
2. **Monitor During Deployment**: Watch metrics for at least 15 minutes
3. **Keep Old Environment Running**: Don't stop it immediately after switch
4. **Document Changes**: Keep deployment logs and notes
5. **Practice Rollbacks**: Test rollback procedure regularly
6. **Backward Compatible Migrations**: Never break old version
7. **Gradual Rollout**: Consider canary or percentage-based rollout for major changes
8. **Health Checks**: Ensure comprehensive health check coverage
9. **Automated Testing**: Run smoke tests before switching traffic
10. **Communication**: Notify team before, during, and after deployments

## Related Documentation

- [Rollback Procedures](./BLUE_GREEN_ROLLBACK.md)
- [Database Migrations](../backend/migrations/README.md)
- [Deployment Guide](./deployment.md)
- [Runbook](./runbook.md)
- [Monitoring](./monitoring.md)

## Support

For issues or questions:
- Slack: #ops-deployments
- Email: <ops-team@clpr.app>
- On-call: Check PagerDuty rotation
