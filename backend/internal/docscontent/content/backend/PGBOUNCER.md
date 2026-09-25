---
title: "PGBOUNCER"
summary: "PgBouncer is deployed as a connection pooler between the Clipper backend and PostgreSQL database. It provides efficient connection pooling in transaction mode, reducing connection overhead and improvi"
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# PgBouncer Connection Pool Configuration

## Overview

PgBouncer is deployed as a connection pooler between the Clipper backend and PostgreSQL database. It provides efficient connection pooling in transaction mode, reducing connection overhead and improving scalability.

## Architecture

```
[Backend Pods] → [PgBouncer Service] → [PostgreSQL StatefulSet]
     2-10            2 replicas              1 replica
```

## Configuration Details

### Pool Settings

Based on load testing results and application requirements:

- **Pool Mode**: `transaction` - Allows connection reuse between queries
- **Min Pool Size**: `10` - Minimum connections maintained to database
- **Max Pool Size**: `50` - Maximum connections per database/user pair
- **Reserve Pool Size**: `5` - Additional connections for emergencies
- **Max Client Connections**: `500` - Total clients that can connect

### Rationale

The pool sizing is based on:
1. **Current Application Pool**: 25 max connections (from `pkg/database/database.go`)
2. **HPA Scaling**: 2-10 backend replicas under load
3. **Headroom**: 50 max connections provides 2x headroom for peak load
4. **Load Test Data**: No connection exhaustion observed during testing with typical workloads

### Timeouts

- **Server Idle Timeout**: 600s (10 minutes)
- **Server Lifetime**: 3600s (1 hour) - connections recycled periodically
- **Query Wait Timeout**: 120s - maximum time client waits for connection
- **Server Connect Timeout**: 15s - timeout for connecting to PostgreSQL

## Deployment

### Prerequisites

1. Kubernetes cluster with PostgreSQL deployed
2. `postgres-secret` with POSTGRES_PASSWORD key (required for PgBouncer metrics exporter)
3. Prometheus for metrics collection
4. Grafana for dashboard visualization

### Installation Steps

1. **Create PgBouncer Secret (Production)**:
   ```bash
   # IMPORTANT: Use SCRAM-SHA-256 authentication (not MD5) for production
   # Ensure PostgreSQL is configured with password_encryption = 'scram-sha-256'
   # and the user is created with a SCRAM-SHA-256 password in PostgreSQL
   
   # Create the user in PostgreSQL first if not exists:
   # kubectl exec postgres-0 -- psql -U postgres -c \
   #   "CREATE ROLE clpr WITH LOGIN PASSWORD 'your-strong-password';"
   
   # For PgBouncer with SCRAM-SHA-256, use auth_query or auth_file with plaintext
   # (PgBouncer will handle the SCRAM exchange with PostgreSQL)
   # Note: The userlist.txt format for SCRAM is: "username" "password"
   
   kubectl create secret generic pgbouncer-secret \
     --from-literal=userlist.txt='\"clpr\" \"your-strong-password\"'
   
   # Update ConfigMap to use scram-sha-256 auth_type
   kubectl patch configmap pgbouncer-config \
     -p '{"data":{"auth_type":"scram-sha-256"}}'
   ```

2. **Deploy PgBouncer**:
   ```bash
   cd backend/k8s
   
   # Deploy ConfigMap
   kubectl apply -f pgbouncer-configmap.yaml
   
   # Deploy Secret (ensure it's populated first)
   kubectl apply -f pgbouncer.yaml
   
   # Deploy PodDisruptionBudget
   kubectl apply -f pdb-pgbouncer.yaml
   ```

3. **Verify Deployment**:
   ```bash
   # Check pods are running
   kubectl get pods -l app=pgbouncer
   
   # Check service
   kubectl get svc pgbouncer
   
   # Verify metrics endpoint
   kubectl port-forward svc/pgbouncer 9127:9127
   curl http://localhost:9127/metrics
   
   # Test connection
   kubectl run -it --rm psql --image=postgres:17 -- \
     psql -h pgbouncer -p 6432 -U clpr -d clpr_db -c "SELECT version();"
   ```

4. **Update Backend to Use PgBouncer**:
   ```bash
   # Edit backend ConfigMap to point to pgbouncer instead of postgres
   kubectl edit configmap backend-config
   # Change: DB_HOST: "pgbouncer"
   # Change: DB_PORT: "6432"
   
   # Restart backend deployment
   kubectl rollout restart deployment/clpr-backend
   kubectl rollout status deployment/clpr-backend
   ```

5. **Configure Prometheus** (if not using in-cluster service discovery):
   ```yaml
   # Add to prometheus.yml scrape_configs
   - job_name: 'pgbouncer'
     static_configs:
       - targets: ['pgbouncer:9127']
     scrape_interval: 10s
   ```

6. **Import Grafana Dashboard**:
   ```bash
   # Dashboard file: monitoring/dashboards/pgbouncer-pool.json
   # Import via Grafana UI or ConfigMap
   ```

## Monitoring

### Key Metrics

1. **Active Client Connections**: `pgbouncer_pools_cl_active`
2. **Server Connections**: `pgbouncer_pools_sv_active`, `pgbouncer_pools_sv_idle`
3. **Waiting Clients**: `pgbouncer_pools_cl_waiting` (should be 0 under normal load)
4. **Query Rate**: `rate(pgbouncer_stats_queries_total[5m])`
5. **Wait Time**: `rate(pgbouncer_stats_waiting_duration_microseconds[5m])`

### Alerts

Configured in `monitoring/alerts.yml`:

- **PgBouncerPoolExhaustion**: Active connections >= 48 (96% utilization)
- **PgBouncerHighWaitTime**: Average wait time > 50ms
- **PgBouncerWaitingClients**: > 10 clients waiting for connections
- **PgBouncerDown**: Service unavailable
- **PgBouncerHighErrors**: Connection errors > 1/sec
- **PgBouncerHighUtilization**: Pool utilization > 80%

### Dashboard

Access the PgBouncer dashboard in Grafana:
- **URL**: `<grafana-url>/d/pgbouncer-pool`
- **Panels**: Client connections, server connections, pool size, query rate, wait time, errors

## Load Testing

### Validation Script

Run load tests to verify no connection exhaustion:

```bash
cd backend/tests/load

# Run baseline test with PgBouncer
./run_all_benchmarks.sh

# Monitor PgBouncer metrics during test
kubectl port-forward svc/pgbouncer 9127:9127 &
watch -n 1 'curl -s http://localhost:9127/metrics | grep pgbouncer_pools'

# Check for waiting clients (should be 0)
kubectl logs -l app=pgbouncer --tail=100 | grep waiting
```

### Expected Behavior

Under load (100 concurrent users):
- **Client Connections**: < 100 active
- **Server Connections**: 10-50 range, scaling with load
- **Waiting Clients**: 0
- **Wait Time**: < 10ms average
- **Errors**: 0

## Rollback Procedure

If issues occur with PgBouncer, follow this rollback process:

### Immediate Rollback (Emergency)

```bash
# 1. Revert backend to direct PostgreSQL connection
kubectl edit configmap backend-config
# Change: DB_HOST: "postgres"
# Change: DB_PORT: "5432"

# 2. Restart backend immediately
kubectl rollout restart deployment/clpr-backend

# 3. Verify backend health
kubectl get pods -l app=clpr-backend
kubectl logs -f deployment/clpr-backend

# 4. Test database connectivity
curl https://clpr.tv/health/ready
```

### Graceful Rollback

```bash
# 1. Scale down PgBouncer to reduce load
kubectl scale deployment pgbouncer --replicas=0

# 2. Update backend configuration
kubectl edit configmap backend-config
# Change: DB_HOST: "postgres"
# Change: DB_PORT: "5432"

# 3. Rolling restart of backend
kubectl rollout restart deployment/clpr-backend
kubectl rollout status deployment/clpr-backend

# 4. Verify functionality
kubectl run -it --rm test-db --image=postgres:17 -- \
  psql -h postgres -p 5432 -U clpr -d clpr_db -c "SELECT COUNT(*) FROM clips;"

# 5. Remove PgBouncer resources (if needed)
kubectl delete -f backend/k8s/pgbouncer.yaml
kubectl delete -f backend/k8s/pdb-pgbouncer.yaml
kubectl delete -f backend/k8s/pgbouncer-configmap.yaml
```

### Post-Rollback Verification

```bash
# Check database connection pool stats from application
curl https://clpr.tv/health/stats

# Monitor direct PostgreSQL connections
kubectl exec postgres-0 -- psql -U clpr -d clpr_db \
  -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'clpr_db';"

# Run health checks
kubectl get pods
kubectl get svc
kubectl logs -f deployment/clpr-backend --tail=50
```

## Tuning Guidelines

### When to Increase Pool Size

Increase `default_pool_size` and `max_db_connections` if you observe:
- Consistent waiting clients (`pgbouncer_pools_cl_waiting > 0`)
- High pool utilization (> 90%)
- Wait times > 50ms
- Backend scaling beyond 10 replicas

### When to Decrease Pool Size

Decrease `default_pool_size` if you observe:
- Low utilization (< 20%) for extended periods
- High number of idle server connections
- PostgreSQL max_connections limit concerns

### Configuration Update Process

```bash
# 1. Edit ConfigMap
kubectl edit configmap pgbouncer-config

# 2. Restart PgBouncer pods (picks up new config)
kubectl rollout restart deployment/pgbouncer

# 3. Update hardcoded values in monitoring (if you changed max_db_connections)
# - Update monitoring/dashboards/pgbouncer-pool.json line 284 (pool utilization calculation)
# - Update monitoring/alerts.yml line 1260 (PgBouncerHighUtilization alert)
# Both use hardcoded value of 50 for max pool size

# 4. Monitor metrics for 15-30 minutes
kubectl port-forward svc/pgbouncer 9127:9127
# Watch metrics in Grafana dashboard

# 4. Run load test to validate
cd backend/tests/load && ./run_all_benchmarks.sh
```

## Troubleshooting

### Connection Refused Errors

```bash
# Check PgBouncer pods
kubectl get pods -l app=pgbouncer
kubectl logs -l app=pgbouncer

# Verify service
kubectl get svc pgbouncer
kubectl describe svc pgbouncer

# Test connectivity
kubectl run -it --rm test-pgbouncer --image=postgres:17 -- \
  psql -h pgbouncer -p 6432 -U clpr -d clpr_db -c "SELECT 1;"
```

### Authentication Errors

```bash
# Verify secret is correct
kubectl get secret pgbouncer-secret -o yaml

# Check PgBouncer logs for auth errors
kubectl logs -l app=pgbouncer | grep -i auth

# Verify userlist.txt format: "username" "md5<hash>"
```

### High Wait Times

```bash
# Check current pool stats
kubectl port-forward svc/pgbouncer 9127:9127
curl -s http://localhost:9127/metrics | grep -E 'cl_waiting|sv_active|sv_idle'

# Check PostgreSQL for slow queries
kubectl exec postgres-0 -- psql -U clpr -d clpr_db -c \
  "SELECT pid, usename, query, state, wait_event
   FROM pg_stat_activity
   WHERE datname = 'clpr_db' AND state != 'idle';"

# Increase pool size if needed
kubectl edit configmap pgbouncer-config
# Update default_pool_size
kubectl rollout restart deployment/pgbouncer
```

### Metrics Not Appearing

```bash
# Check exporter container logs
kubectl logs -l app=pgbouncer -c pgbouncer-exporter

# Verify metrics endpoint
kubectl port-forward svc/pgbouncer 9127:9127
curl http://localhost:9127/metrics

# Check Prometheus target
# Prometheus UI → Status → Targets → pgbouncer

# Verify Prometheus config includes pgbouncer job
kubectl exec -n clpr-monitoring deployment/prometheus -- \
  cat /etc/prometheus/prometheus.yml | grep pgbouncer
```

## Security Considerations

1. **Secret Management**: Use Vault or external secrets operator in production
2. **Network Policies**: Ensure only backend pods can reach PgBouncer
3. **Authentication**: Use strong passwords with MD5 hashing
4. **Access Control**: Limit admin access to PgBouncer stats interface
5. **Monitoring**: Alert on authentication failures and suspicious connection patterns

## Performance Impact

### Benefits
- Reduced connection overhead (connection setup/teardown eliminated)
- Better resource utilization on PostgreSQL
- Improved scalability (500 clients → 50 server connections)
- Connection reuse in transaction mode

### Trade-offs
- Additional network hop (~1ms latency)
- Session-level features not available (prepared statements in transaction mode)
- Slightly increased complexity in debugging

## Related Issues

- [#852](https://git.subcult.tv/subculture-collective/clpr/issues/852) - Kubernetes Cluster Setup
- [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805) - Related Infrastructure Issue

## References

- [PgBouncer Documentation](https://www.pgbouncer.org/config.html)
- [PgBouncer Best Practices](https://www.pgbouncer.org/faq.html)
- [pgx Connection Pooling](https://github.com/jackc/pgx)
- Load Test Report: `backend/tests/load/reports/load_test_report_20251216_215538.md`
