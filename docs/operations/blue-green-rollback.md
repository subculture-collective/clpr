---
title: "Blue Green Rollback"
summary: "This document describes the rollback procedures for blue/green deployments in Clipper."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Blue/Green Deployment Rollback Procedures

This document describes the rollback procedures for blue/green deployments in Clipper.

## Table of Contents

- [Overview](#overview)
- [When to Rollback](#when-to-rollback)
- [Automatic Rollback](#automatic-rollback)
- [Manual Rollback](#manual-rollback)
- [Emergency Rollback](#emergency-rollback)
- [Post-Rollback Steps](#post-rollback-steps)
- [Prevention](#prevention)

## Overview

Blue/green deployment enables instant rollback by maintaining two production environments. If issues are detected in the new (target) environment, traffic can be immediately switched back to the previous (active) environment.

### Rollback Types

1. **Automatic Rollback**: Triggered by failed health checks during deployment
2. **Manual Rollback**: Initiated by operators when issues are detected
3. **Emergency Rollback**: Immediate traffic switch for critical issues

## When to Rollback

Initiate a rollback when:

- ✗ Health checks fail after deployment
- ✗ Error rates spike above threshold (>1% of requests)
- ✗ Response times exceed SLA (p95 > 500ms)
- ✗ Critical functionality is broken
- ✗ Database queries are failing
- ✗ Security vulnerabilities discovered
- ✗ User-reported critical bugs

Do NOT rollback for:

- ✓ Minor UI issues that don't affect functionality
- ✓ Non-critical feature bugs
- ✓ Performance optimizations needed
- ✓ Cosmetic issues

## Automatic Rollback

The deployment script (`blue-green-deploy.sh`) includes automatic rollback that triggers when:

1. Health checks fail during deployment
2. Post-deployment monitoring detects issues
3. Traffic switch fails

### How Automatic Rollback Works

```bash
# During deployment, if health check fails:
1. Stop the target environment (new version)
2. Ensure source environment (old version) is running
3. Verify source environment health
4. Switch traffic back to source environment
5. Log rollback event
6. Send monitoring notification
```

### Monitoring Automatic Rollback

```bash
# Check deployment logs
tail -f /var/log/clpr/deployment.log

# Watch for rollback messages
grep "Rollback" /var/log/clpr/deployment.log
```

## Manual Rollback

### Prerequisites

- SSH access to production server
- Sudo privileges (if needed)
- Knowledge of currently active environment

### Quick Manual Rollback (Method 1: Using Script)

```bash
# 1. SSH to production server
ssh deploy@production-server

# 2. Navigate to deployment directory
cd /opt/clpr

# 3. Check current state
docker ps --filter "name=clpr-backend"

# 4. Identify active environment
# If clpr-backend-green is running, rollback to blue
# If clpr-backend-blue is running, rollback to green

# 5. Execute rollback script (switches traffic and restarts old environment)
./scripts/rollback-blue-green.sh

# 6. Verify rollback
./scripts/health-check.sh
```

### Manual Rollback (Method 2: Manual Steps)

```bash
# 1. Determine current active environment
CURRENT_ENV=$(docker ps --format '{{.Names}}' | grep -o 'clpr-backend-\(blue\|green\)' | head -1 | cut -d'-' -f3)
echo "Current active: $CURRENT_ENV"

# 2. Determine target environment (opposite)
if [ "$CURRENT_ENV" = "blue" ]; then
    TARGET_ENV="green"
else
    TARGET_ENV="blue"
fi
echo "Rolling back to: $TARGET_ENV"

# 3. Ensure target environment is running
docker compose -f docker-compose.blue-green.yml --profile $TARGET_ENV up -d backend-$TARGET_ENV frontend-$TARGET_ENV

# 4. Wait for target to be healthy
sleep 20
docker exec clpr-backend-$TARGET_ENV wget --spider -q http://localhost:8080/health

# 5. Switch Caddy configuration to target environment
export ACTIVE_ENV=$TARGET_ENV
docker compose -f docker-compose.blue-green.yml restart caddy

# 6. Verify traffic is flowing to target
curl -f http://localhost/health
curl -f http://localhost/api/v1/health

# 7. Monitor for 1-2 minutes
watch -n 5 'curl -s http://localhost/health | jq'

# 8. Stop the problematic environment
docker compose -f docker-compose.blue-green.yml --profile $CURRENT_ENV stop backend-$CURRENT_ENV frontend-$CURRENT_ENV
```

### Verify Rollback Success

```bash
# Check active containers
docker ps --filter "name=clpr"

# Test endpoints
curl -f http://localhost/health
curl -f http://localhost/api/v1/health

# Check logs for errors
docker compose logs --tail=50 backend-blue
docker compose logs --tail=50 backend-green

# Monitor metrics (if enabled)
# - Response times
# - Error rates
# - Active connections
```

## Emergency Rollback

For critical production issues requiring immediate action:

### One-Command Emergency Rollback

```bash
# Create this as an alias or script for emergency use
# Replace TARGET_ENV with blue or green depending on which is the stable version

# Emergency rollback to blue
export ACTIVE_ENV=blue && \
docker compose -f /opt/clpr/docker-compose.blue-green.yml up -d backend-blue frontend-blue && \
docker compose -f /opt/clpr/docker-compose.blue-green.yml restart caddy && \
echo "Emergency rollback to blue completed"

# Emergency rollback to green  
export ACTIVE_ENV=green && \
docker compose -f /opt/clpr/docker-compose.blue-green.yml --profile green up -d backend-green frontend-green && \
docker compose -f /opt/clpr/docker-compose.blue-green.yml restart caddy && \
echo "Emergency rollback to green completed"
```

### Caddy-Only Traffic Switch (Fastest)

If both environments are running, instantly switch traffic:

```bash
# Switch to blue
export ACTIVE_ENV=blue
docker compose -f /opt/clpr/docker-compose.blue-green.yml restart caddy

# Switch to green
export ACTIVE_ENV=green
docker compose -f /opt/clpr/docker-compose.blue-green.yml restart caddy

# Verify
curl http://localhost/health
```

## Post-Rollback Steps

After completing a rollback:

### 1. Verify System Health

```bash
# Run comprehensive health checks
./scripts/health-check.sh

# Check error logs
docker compose logs --tail=100 backend-blue | grep -i error
docker compose logs --tail=100 frontend-blue | grep -i error

# Test critical user flows
# - Login
# - Browse clips
# - Submit clips
# - Comments
```

### 2. Notify Stakeholders

```bash
# Send notification to team
# Include:
# - Rollback reason
# - Current active environment
# - Time of rollback
# - Next steps

# Example Slack message:
"🔄 Rollback completed
Environment: blue (stable)
Reason: Health checks failed in green deployment
Time: $(date)
Status: All systems operational
Next: Investigating deployment failure"
```

### 3. Investigate Root Cause

```bash
# Collect logs from failed environment
docker compose logs backend-green > /tmp/backend-green-failure.log
docker compose logs frontend-green > /tmp/frontend-green-failure.log

# Check deployment logs
cat /var/log/clpr/deployment.log

# Review recent changes
git log --oneline -10

# Analyze database state
docker exec clpr-postgres psql -U clpr -d clpr_db -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 5;"
```

### 4. Document Incident

Create incident report with:
- Timestamp of deployment
- Timestamp of rollback
- Symptoms observed
- Rollback method used
- Root cause (when identified)
- Prevention measures

### 5. Plan Fix

- Fix the issue in development
- Test thoroughly in staging
- Review changes carefully
- Plan new deployment with monitoring

## Prevention

### Pre-Deployment Checklist

- [ ] Run all tests locally (`make test`)
- [ ] Deploy to staging first
- [ ] Run migration compatibility check
- [ ] Review database migrations for backward compatibility
- [ ] Verify health check endpoints work
- [ ] Test rollback procedure in staging
- [ ] Ensure monitoring is active
- [ ] Have rollback plan ready

### Database Migration Safety

For backward-compatible migrations:

```sql
-- ✓ SAFE: Add new column with default
ALTER TABLE clips ADD COLUMN featured BOOLEAN DEFAULT false;

-- ✓ SAFE: Create new table
CREATE TABLE IF NOT EXISTS tags (...);

-- ✓ SAFE: Add index
CREATE INDEX CONCURRENTLY idx_clips_featured ON clips(featured);

-- ✗ UNSAFE: Drop column (breaks old version)
-- ALTER TABLE clips DROP COLUMN old_field;

-- ✗ UNSAFE: Rename column (breaks old version)
-- ALTER TABLE clips RENAME COLUMN old_name TO new_name;
```

### Monitoring During Deployment

```bash
# Watch key metrics during and after deployment:

# 1. Error rates
watch -n 5 'curl -s http://localhost:8080/health/stats | jq .error_rate'

# 2. Response times  
watch -n 5 'curl -s http://localhost:8080/health/stats | jq .response_time_p95'

# 3. Active connections
watch -n 5 'curl -s http://localhost:8080/health/stats | jq .active_connections'

# 4. Database connections
docker exec clpr-postgres psql -U clpr -d clpr_db -c "SELECT count(*) FROM pg_stat_activity;"
```

## Troubleshooting

### Both Environments Are Down

```bash
# 1. Check Docker daemon
sudo systemctl status docker

# 2. Check compose file
docker compose -f docker-compose.blue-green.yml config

# 3. Start blue environment (default)
docker compose -f docker-compose.blue-green.yml up -d backend-blue frontend-blue

# 4. Check logs for errors
docker compose logs --tail=100
```

### Health Checks Failing After Rollback

```bash
# 1. Verify containers are running
docker ps

# 2. Check container health
docker inspect clpr-backend-blue | jq '.[0].State.Health'

# 3. Test health endpoint directly
docker exec clpr-backend-blue wget -O- http://localhost:8080/health

# 4. Check database connectivity
docker exec clpr-postgres pg_isready
```

### Traffic Not Switching

```bash
# 1. Check Caddy status
docker ps | grep caddy

# 2. Check Caddy configuration
docker exec clpr-caddy caddy environ

# 3. Verify ACTIVE_ENV variable
docker exec clpr-caddy env | grep ACTIVE_ENV

# 4. Manually reload Caddy
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile

# 5. Restart Caddy if needed
docker compose -f docker-compose.blue-green.yml restart caddy
```

## Support

For assistance with rollbacks:

1. Check logs: `/var/log/clpr/deployment.log`
2. Run diagnostics: `./scripts/preflight-check.sh`
3. Contact: <ops-team@clpr.app>
4. Escalate: CTO (for critical issues)

## Related Documentation

- [Blue/Green Deployment Guide](./BLUE_GREEN_DEPLOYMENT.md)
- [Deployment Procedures](./deployment.md)
- [Runbook](./runbook.md)
- [Database Migrations](../backend/migrations/README.md)
