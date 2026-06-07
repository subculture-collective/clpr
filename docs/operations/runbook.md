---
title: "Operations Runbook"
summary: "Operational procedures and commands for managing Clipper in production."
tags: ["operations", "runbook", "ops"]
area: "deployment"
status: "stable"
owner: "team-ops"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["runbook", "ops procedures"]
---

# Operations: Runbook

Operational procedures and commands for managing Clipper in production.

## Monitoring Dashboards

Access comprehensive monitoring through Grafana dashboards (default: <http://localhost:3000>):

- **[System Health](../../monitoring/dashboards/system-health.json)** - CPU, memory, disk, network metrics
- **[API Performance](../../monitoring/dashboards/api-performance.json)** - Request rates, latency, errors
- **[Database](../../monitoring/dashboards/database.json)** - PostgreSQL performance and connections
- **[Redis Cache](../../monitoring/dashboards/redis.json)** - Cache hit rates, memory usage
- **[Kubernetes Cluster](../../monitoring/dashboards/kubernetes.json)** - Pod, node, and workload health
- **[Resource Quotas](../../monitoring/dashboards/resource-quotas.json)** - Namespace quotas and limits
- **[Application Overview](../../monitoring/dashboards/app-overview.json)** - High-level SLO compliance
- **[Background Jobs](../../monitoring/dashboards/background-jobs.json)** - Job execution and queue health
- **[Webhook Monitoring](../../monitoring/dashboards/webhook-monitoring.json)** - Webhook delivery status

See [Dashboard README](../../monitoring/dashboards/README.md) for full documentation.

## Common Tasks

### Check Service Health

```bash
# All services
kubectl get pods -n clpr

# Specific service
kubectl describe pod backend-xyz -n clpr

# Logs
kubectl logs -f deployment/backend -n clpr
```

### Database Operations

```bash
# Connect to database
psql $POSTGRES_URL

# Check connection count
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Kill long-running query
psql -c "SELECT pg_terminate_backend(PID);"

# Run migrations
kubectl exec -it backend-pod -- make migrate-up
```

### Search Operations

```bash
# Cluster health
curl https://opensearch.clpr.app/_cluster/health

# Reindex from PostgreSQL
kubectl exec -it backend-pod -- go run cmd/backfill-search/main.go

# Force refresh
curl -X POST https://opensearch.clpr.app/_refresh
```

### Cache Operations

```bash
# Connect to Redis
kubectl exec -it redis-pod -- redis-cli

# Clear cache
redis-cli FLUSHDB

# Check memory usage
redis-cli INFO memory
```

### Scaling

```bash
# Scale backend pods
kubectl scale deployment backend --replicas=5 -n clpr

# Horizontal Pod Autoscaler
kubectl autoscale deployment backend --cpu-percent=70 --min=3 --max=10 -n clpr
```

### Deployments

```bash
# Deploy new version
kubectl set image deployment/backend backend=clpr:v1.2.3 -n clpr

# Check rollout status
kubectl rollout status deployment/backend -n clpr

# Rollback
kubectl rollout undo deployment/backend -n clpr
```

## Deployment Testing & Validation

### Deployment Scripts Testing Harness

The deployment testing harness validates deployment scripts (deploy, rollback, blue-green) in safe DRY_RUN and MOCK modes before production use.

#### Running the Harness

**Basic Usage:**

```bash
# Run in MOCK mode (no actual Docker commands)
cd scripts
./test-deployment-harness.sh

# Run with verbose logging
VERBOSE=true ./test-deployment-harness.sh

# Custom test results directory
TEST_RESULTS_DIR=/tmp/my-tests ./test-deployment-harness.sh
```

**Environment Variables:**

- `DRY_RUN` (default: `true`) - Enable dry-run mode
- `MOCK` (default: `true`) - Use mock Docker/curl commands
- `VERBOSE` (default: `false`) - Enable detailed logging
- `TEST_RESULTS_DIR` (default: `/tmp/deployment-harness-results`) - Output directory

#### What the Harness Tests

The harness validates:

1. **Validation Checks** - Scripts verify prerequisites (Docker, docker-compose, directories)
2. **Backup Mechanisms** - Deployment creates proper backups before changes
3. **Rollback Logic** - Rollback scripts can restore from backups
4. **Environment Detection** - Blue-green deployment detects active environment
5. **Error Handling** - Scripts use `set -e` and proper exit codes
6. **Health Checks** - Deployments verify service health post-deployment
7. **DRY_RUN Support** - Rotation scripts support safe dry-run mode

#### Interpreting Test Results

**Success Output:**

```
=== Test Summary ===
Tests Run: 10
Tests Passed: 10
Tests Failed: 0

Test results saved to: /tmp/deployment-harness-results
=== All Tests Passed ===
```

**Failure Output:**

```
[FAIL] deploy.sh backup mechanism

Failed Tests:
  - deploy.sh backup mechanism

Tests Run: 10
Tests Passed: 9
Tests Failed: 1
```

**Test Artifacts:**

All test results are saved to `TEST_RESULTS_DIR`:
- `mock-commands.log` - Log of all mocked Docker/curl commands (if `VERBOSE=true`)
- `deploy-dry-run.log` - Output from deploy script test
- `mock-deploy/` - Mock deployment directory with docker-compose files

#### Troubleshooting Test Failures

| Failure | Cause | Resolution |
|---------|-------|------------|
| "missing docker validation" | Script doesn't check for Docker | Add `command_exists docker` check |
| "missing backup mechanism" | No backup creation before deploy | Add `BACKUP_TAG` and `docker tag` logic |
| "missing confirmation prompt" | Rollback has no safety prompt | Add `read -p` confirmation |
| "missing 'set -e'" | Script doesn't exit on errors | Add `set -e` at script top |
| "missing exit codes" | No explicit exit statements | Add `exit 0` (success) and `exit 1` (failure) |

### Rollback Drills

Periodic rollback drills ensure deployment reversibility and validate disaster recovery procedures.

#### Running a Rollback Drill

**DRY_RUN Mode (Safe):**

```bash
# Test rollback process without actual containers
cd scripts
DRY_RUN=true ./rollback-drill.sh
```

**LIVE Mode (Creates Real Containers):**

```bash
# Full drill with actual Docker containers
DRY_RUN=false ./rollback-drill.sh

# With automatic cleanup
DRY_RUN=false CLEANUP=true ./rollback-drill.sh
```

**Environment Variables:**

- `DRY_RUN` (default: `true`) - Safe simulation mode
- `DRILL_DIR` (default: `/tmp/rollback-drill`) - Drill workspace
- `ENVIRONMENT` (default: `drill`) - Environment identifier
- `CLEANUP` (default: `false`) - Auto-cleanup after drill

#### Drill Phases

The rollback drill executes these phases:

1. **Setup** - Create drill environment with docker-compose files
2. **Initial State Capture** - Snapshot current state (containers, images, config)
3. **Deployment Simulation** - Deploy "v2" with backup creation
4. **Deployment Verification** - Verify v2 is running and healthy
5. **Rollback Execution** - Rollback to v1 using backup
6. **Rollback Verification** - Verify v1 restored and healthy
7. **Clean State Verification** - Compare final state with initial state
8. **Data Integrity Check** - Verify no data loss or corruption

#### Interpreting Drill Results

The drill generates a detailed report at `$DRILL_DIR/state/drill-report.txt`:

**Successful Drill:**

```
=== Rollback Drill Report ===
Date: 2026-01-29 02:00:00
Environment: drill
DRY_RUN: true

Overall Result: PASSED

✓ All verification checks passed
✓ Rollback mechanism working correctly
✓ Clean state achieved post-rollback
✓ Data integrity maintained

Recommendation: Deployment rollback procedures are operational.
```

**Failed Drill:**

```
Overall Result: FAILED

✗ Some verification checks failed
✗ Review logs for details

Recommendation: Investigate failures before production rollback.
```

#### Drill Schedule

**Automated Schedule:**
- **Frequency:** Weekly (every Monday at 2 AM UTC)
- **Execution:** GitHub Actions workflow `.github/workflows/deployment-tests.yml`
- **Mode:** Both DRY_RUN and LIVE modes

**Manual Execution:**
- Run before major deployments
- After infrastructure changes
- When validating disaster recovery plans
- Include in commit message: `[rollback-drill]`

#### Troubleshooting Drill Failures

| Phase Failure | Possible Cause | Action |
|---------------|----------------|--------|
| Setup | Docker not available | Verify Docker installation |
| Deployment Simulation | Compose file errors | Check docker-compose.drill.yml syntax |
| Deployment Verification | Health checks fail | Increase health check timeout |
| Rollback Execution | Backup tag missing | Verify backup creation in deployment phase |
| Clean State Verification | Orphaned containers | Check for container cleanup logic |
| Data Integrity | Config files missing | Verify state file preservation |

### CI/CD Integration

The deployment tests run automatically in GitHub Actions:

#### Workflow Triggers

- **Harness Tests:** Push/PR to main/develop with deployment script changes
- **Rollback Drills:** 
  - Weekly schedule (Monday 2 AM UTC)
  - Manual via workflow_dispatch
  - Commit message containing `[rollback-drill]`

#### Viewing CI Results

1. **GitHub Actions Tab:** View workflow runs
2. **Artifacts:** Download test results (retained 30 days)
   - `deployment-harness-results` - Harness test logs
   - `rollback-drill-results` - Drill reports and state files
3. **Summary:** Check workflow summary for quick overview

#### CI Failure Response

When deployment tests fail in CI:

1. **Check Workflow Logs:** Identify which job failed
2. **Download Artifacts:** Get detailed logs and reports
3. **Reproduce Locally:** Run failing test locally
4. **Fix Issues:** Update scripts based on failures
5. **Re-run:** Push fix and verify tests pass

**Example:**

```bash
# Download artifacts from failed CI run
gh run download <run-id>

# Review harness results
cat deployment-harness-results/harness-output.log

# Reproduce locally
cd scripts
MOCK=true ./test-deployment-harness.sh

# Fix issues in scripts
vim deploy.sh

# Test again
./test-deployment-harness.sh
```

### Best Practices

#### Before Production Deployment

1. **Run Harness:** Verify all deployment scripts pass tests
   ```bash
   cd scripts && ./test-deployment-harness.sh
   ```

2. **Run Drill (DRY_RUN):** Validate rollback procedures
   ```bash
   DRY_RUN=true ./rollback-drill.sh
   ```

3. **Review Artifacts:** Check logs and reports for warnings

4. **Staging Rehearsal:** Use `staging-rehearsal.sh` for full end-to-end test
   ```bash
   ./staging-rehearsal.sh
   ```

#### Regular Maintenance

- **Weekly:** Automated rollback drills via CI
- **Monthly:** Manual full drill in staging environment
- **Quarterly:** Review and update test scenarios
- **After Changes:** Run harness when modifying deployment scripts

#### Emergency Rollback

If production deployment fails:

1. **Don't Panic:** Rollback procedures are tested weekly
2. **Use Backup Tag:** Check deployment logs for backup tag
3. **Execute Rollback:** Run rollback script with backup tag
   ```bash
   ./rollback.sh backup-20260129-120000
   ```
4. **Verify Health:** Check all services are healthy
5. **Post-Mortem:** Analyze what went wrong and update tests

### Database Backups

```bash
# Manual backup
pg_dump $POSTGRES_URL > backup_$(date +%Y%m%d).sql

# Restore
psql $POSTGRES_URL < backup_20251130.sql
```

## Incident Scenarios

### High Error Rate

1. Check logs: `kubectl logs -f deployment/backend`
2. Check metrics: Grafana dashboard
3. Recent deploy? Rollback: `kubectl rollout undo`
4. Database issue? Check connections, slow queries
5. External API down? Enable circuit breaker

### High Latency

1. Check p95/p99 metrics
2. Database slow? Check `pg_stat_statements`
3. Cache cold? Warm up or increase TTL
4. Scale up: `kubectl scale deployment backend --replicas=N`

### Database Connection Exhaustion

1. Check active connections
2. Kill idle/long-running queries
3. Increase connection pool size (restart required)
4. Add read replicas if read-heavy

### OpenSearch Cluster Red

1. Check cluster health
2. Identify problematic indices
3. Delete/recreate if needed
4. Reindex from PostgreSQL

### Out of Disk Space

1. Check disk usage: `df -h`
2. Clear old logs, backups
3. Increase volume size (cloud provider)
4. Add log rotation policy

## Maintenance Windows

Planned maintenance:
1. Announce in advance (status page, email)
2. Enable maintenance mode (static page)
3. Run migrations, upgrades
4. Test thoroughly
5. Re-enable traffic
6. Monitor for 30 minutes

---

Related: [[monitoring|Monitoring]] · [[infra|Infrastructure]] · [[deployment|Deployment]]

[[../index|← Back to Index]]
