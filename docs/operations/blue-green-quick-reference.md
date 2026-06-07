---
title: "Blue Green Quick Reference"
summary: "Quick reference card for blue/green deployment operations."
tags: ["operations","quick-reference"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Blue/Green Deployment Quick Reference

Quick reference card for blue/green deployment operations.

## Quick Commands

### Deploy New Version

```bash
# Standard deployment
./scripts/blue-green-deploy.sh

# Specific version
IMAGE_TAG=v1.2.3 ./scripts/blue-green-deploy.sh

# With monitoring
MONITORING_ENABLED=true ./scripts/blue-green-deploy.sh
```

### Rollback

```bash
# Quick rollback (with confirmations)
./scripts/rollback-blue-green.sh

# Auto rollback (skip confirmations)
./scripts/rollback-blue-green.sh --yes
```

### Health Checks

```bash
# All services
./scripts/health-check.sh

# Specific service
curl http://localhost/health
curl http://localhost/api/v1/health
```

### Check Active Environment

```bash
# Via Docker
docker ps --filter "name=clpr-backend" --format "{{.Names}}" | grep -o "blue\|green"

# Via Caddy
docker exec clpr-caddy env | grep ACTIVE_ENV
```

### Manual Traffic Switch

```bash
# Switch to blue
export ACTIVE_ENV=blue
docker compose -f docker-compose.blue-green.yml up -d caddy

# Switch to green
export ACTIVE_ENV=green
docker compose -f docker-compose.blue-green.yml up -d caddy
```

## Deployment Checklist

### Pre-Deployment

- [ ] Review changes in staging
- [ ] Check migration compatibility: `./scripts/check-migration-compatibility.sh`
- [ ] Backup current state: `./scripts/backup.sh`
- [ ] Verify images are available in registry
- [ ] Check disk space: `df -h` (need 5GB+ free)
- [ ] Check memory: `free -h` (need 2GB+ free)
- [ ] Notify team of deployment window

### During Deployment

- [ ] Run deployment: `./scripts/blue-green-deploy.sh`
- [ ] Monitor health checks (automatic)
- [ ] Watch logs: `docker compose logs -f --tail=100`
- [ ] Check error rates
- [ ] Verify critical endpoints work

### Post-Deployment

- [ ] Test critical user flows
  - [ ] Login/logout
  - [ ] Browse clips
  - [ ] Submit clip
  - [ ] Vote on clip
  - [ ] Post comment
- [ ] Monitor metrics for 15 minutes
- [ ] Check error logs
- [ ] Verify database connections stable
- [ ] Document deployment (version, time, issues)
- [ ] Notify team of completion

### Rollback Checklist

- [ ] Identify issue requiring rollback
- [ ] Run rollback: `./scripts/rollback-blue-green.sh`
- [ ] Verify rollback success
- [ ] Test critical functionality
- [ ] Document incident
- [ ] Investigate root cause
- [ ] Plan fix

## Environment Status

### Check Running Containers

```bash
docker ps --filter "name=clpr"
```

### Check Container Health

```bash
# All containers
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Specific container health
docker inspect clpr-backend-blue | jq '.[0].State.Health'
```

### View Logs

```bash
# All services
docker compose logs -f

# Specific environment
docker compose logs -f backend-blue frontend-blue
docker compose logs -f backend-green frontend-green

# Last N lines
docker compose logs --tail=100

# Filter by error
docker compose logs | grep -i error
```

## Common Issues & Solutions

### Issue: Health Checks Failing

**Symptoms**: Deployment fails with "Health check failed"

**Solutions**:
```bash
# Check container logs
docker compose logs backend-green --tail=100

# Test health endpoint directly
docker exec clpr-backend-green wget -O- http://localhost:8080/health

# Check database connectivity
docker exec clpr-postgres pg_isready

# Check Redis
docker exec clpr-redis redis-cli ping
```

### Issue: Out of Disk Space

**Symptoms**: "no space left on device"

**Solutions**:
```bash
# Check disk space
df -h

# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Remove old backups
rm -rf /opt/clpr/backups/*-old
```

### Issue: Port Already in Use

**Symptoms**: "port is already allocated"

**Solutions**:
```bash
# Check what's using the port
sudo lsof -i :80
sudo lsof -i :443

# Stop conflicting service
sudo systemctl stop nginx  # or apache2

# Or change port in docker-compose
```

### Issue: Image Pull Failed

**Symptoms**: "failed to pull image"

**Solutions**:
```bash
# Check Docker login
docker login ghcr.io

# Manually pull image
docker pull ghcr.io/subculture-collective/clpr/backend:latest

# Check network connectivity
ping ghcr.io
```

### Issue: Both Environments Running

**Symptoms**: High memory usage, unexpected behavior

**Solutions**:
```bash
# Identify active environment
ACTIVE=$(docker ps --filter "name=clpr-backend" --format "{{.Names}}" | grep -o "blue\|green" | head -1)

# Determine which to stop
if [ "$ACTIVE" = "blue" ]; then
  docker compose --profile green stop
else
  docker compose stop backend-blue frontend-blue
fi
```

## Monitoring Commands

### Real-Time Metrics

```bash
# Container stats
docker stats clpr-backend-blue clpr-backend-green

# System resources
htop  # or top

# Network connections
netstat -an | grep :80
```

### Health Endpoints

```bash
# Main health
curl http://localhost/health | jq

# Detailed backend health
curl http://localhost/api/v1/health | jq

# Readiness (DB + Redis)
curl http://localhost/health/ready | jq

# Liveness
curl http://localhost/health/live | jq

# Cache stats
curl http://localhost/health/cache | jq
```

### Database Queries

```bash
# Connection count
docker exec clpr-postgres psql -U clpr -d clpr_db -c \
  "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"

# Active queries
docker exec clpr-postgres psql -U clpr -d clpr_db -c \
  "SELECT pid, now() - query_start as duration, query FROM pg_stat_activity WHERE state = 'active';"

# Database size
docker exec clpr-postgres psql -U clpr -d clpr_db -c \
  "SELECT pg_size_pretty(pg_database_size('clpr_db'));"
```

## Emergency Procedures

### Complete System Restart

```bash
# Stop all
docker compose -f docker-compose.blue-green.yml --profile green down
docker compose -f docker-compose.blue-green.yml down

# Start fresh
docker compose -f docker-compose.blue-green.yml up -d postgres redis
sleep 10
docker compose -f docker-compose.blue-green.yml up -d backend-blue frontend-blue
sleep 20
docker compose -f docker-compose.blue-green.yml up -d caddy
```

### Emergency Rollback (One-Liner)

```bash
# To blue
export ACTIVE_ENV=blue && docker compose -f /opt/clpr/docker-compose.blue-green.yml up -d backend-blue frontend-blue caddy

# To green
export ACTIVE_ENV=green && docker compose -f /opt/clpr/docker-compose.blue-green.yml --profile green up -d backend-green frontend-green caddy
```

### Get Help

```bash
# Show script help
./scripts/blue-green-deploy.sh --help
./scripts/rollback-blue-green.sh --help

# Check system logs
journalctl -u docker -n 100

# Check deployment logs
tail -f /var/log/clpr/deployment.log
```

## Testing

### Pre-Production Testing

```bash
# Run test suite
./scripts/test-blue-green-deployment.sh

# Test in staging
cd /opt/clpr-staging
./scripts/blue-green-deploy.sh
```

### Smoke Tests

```bash
# Critical endpoints
curl -f http://localhost/health || echo "FAILED"
curl -f http://localhost/api/v1/health || echo "FAILED"
curl -f http://localhost/api/v1/clips?limit=10 || echo "FAILED"

# Response time test
time curl -s http://localhost/api/v1/clips > /dev/null
```

## Key Metrics to Monitor

### During Deployment (Watch for 15 min)

- **Error Rate**: Should stay < 0.1%
- **Response Time**: p95 < 200ms, p99 < 500ms
- **Request Rate**: Should be stable
- **Database Connections**: Should be stable (< 50)
- **Memory Usage**: Should be stable
- **CPU Usage**: Should normalize after 5 minutes

### Alert Thresholds

- Error rate > 1% → Investigate immediately
- Response time p95 > 500ms → Performance issue
- Response time p99 > 1000ms → Critical performance issue
- Database connections > 80 → Connection pool issue
- Memory usage > 90% → Memory leak or under-provisioned

## Useful Aliases

Add to `~/.bashrc` or `~/.bash_aliases`:

```bash
# Blue/Green deployment aliases
alias bgdeploy='cd /opt/clpr && ./scripts/blue-green-deploy.sh'
alias bgrollback='cd /opt/clpr && ./scripts/rollback-blue-green.sh'
alias bghealth='cd /opt/clpr && ./scripts/health-check.sh'
alias bgtest='cd /opt/clpr && ./scripts/test-blue-green-deployment.sh'
alias bglogs='cd /opt/clpr && docker compose logs -f --tail=100'
alias bgstatus='docker ps --filter "name=clpr" --format "table {{.Names}}\t{{.Status}}"'
```

## Documentation Links

- **[Complete Guide](./BLUE_GREEN_DEPLOYMENT.md)** - Full documentation
- **[Rollback Procedures](./BLUE_GREEN_ROLLBACK.md)** - Detailed rollback steps
- **[Deployment Guide](./deployment.md)** - General deployment info
- **[Runbook](./runbook.md)** - Operational procedures

## Support

- **Slack**: #ops-deployments
- **Email**: <ops-team@clpr.app>
- **On-Call**: Check PagerDuty rotation
- **Emergency**: Escalate to CTO

---

**Last Updated**: 2025-12-16  
**Maintained by**: Operations Team
