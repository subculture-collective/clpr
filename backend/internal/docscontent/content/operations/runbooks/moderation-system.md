# Moderation System Monitoring Runbook

**Related Issue:** [#1021 - Set Up Monitoring & Alerts](https://git.subcult.tv/subculture-collective/clpr/issues/1021)

## Overview

This runbook provides operational guidance for monitoring and troubleshooting the moderation system, including ban operations, sync operations, permission checks, and audit logs.

## Quick Links

- **Dashboard:** [Moderation System Monitoring](http://localhost:3000/d/moderation-system)
- **Prometheus:** [Moderation Metrics](http://localhost:9090/graph?g0.expr=moderation_ban_operations_total)
- **Related Runbooks:**
  - [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md)
  - [Moderation Operations](./moderation-operations.md)
  - [Moderation Incidents](./moderation-incidents.md)
  - [Permission Escalation](./permission-escalation.md)
  - [Audit Log Operations](./audit-log-operations.md)

## Metrics Reference

### Ban Operations

| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| `moderation_ban_operations_total` | Total ban operations | Counter | `operation`, `status`, `error_type` |
| `moderation_ban_operation_duration_seconds` | Ban operation latency | Histogram | `operation` |
| `moderation_active_bans` | Current active bans | Gauge | `community_type` |

**Key Thresholds:**
- **Success Rate:** > 90% (warning < 90%, critical < 50%)
- **P95 Latency:** < 2s (warning > 2s, critical > 5s)
- **P99 Latency:** < 5s

### Sync Operations

| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| `moderation_sync_operations_total` | Total sync operations | Counter | `status`, `error_type` |
| `moderation_sync_operation_duration_seconds` | Sync operation latency | Histogram | `sync_type` |
| `moderation_sync_bans_processed_total` | Bans processed during sync | Counter | `status` |

**Key Thresholds:**
- **Success Rate:** > 90% (warning < 90%, critical < 50%)
- **P95 Latency:** < 30s (warning > 30s)
- **Sync Interval:** Every 15 minutes (configurable)

### Permission Checks

| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| `moderation_permission_checks_total` | Total permission checks | Counter | `permission_type`, `result` |
| `moderation_permission_check_duration_seconds` | Permission check latency | Histogram | `permission_type` |
| `moderation_permission_denials_total` | Permission denials | Counter | `permission_type`, `reason` |

**Key Thresholds:**
- **P95 Latency:** < 100ms (warning > 100ms, critical > 250ms)
- **Denial Rate:** < 50% under normal conditions
- **Denial Spike:** < 10/sec (warning > 10/sec)

### Audit Logs

| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| `moderation_audit_log_operations_total` | Audit log operations | Counter | `action`, `status` |
| `moderation_audit_log_operation_duration_seconds` | Audit log operation latency | Histogram | `action` |
| `moderation_audit_log_volume` | Audit log volume | Gauge | `period` |

**Key Thresholds:**
- **Success Rate:** > 99%
- **P95 Latency:** < 500ms (warning > 500ms)

### Database Performance

| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| `moderation_database_query_duration_seconds` | Database query latency | Histogram | `query_type` |
| `moderation_slow_queries_total` | Slow queries (>1s) | Counter | `query_type` |

**Key Thresholds:**
- **P95 Latency:** < 100ms for most queries
- **Slow Query Rate:** < 1/sec (warning > 1/sec, critical > 5/sec)

### API Errors

| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| `moderation_api_errors_total` | API errors | Counter | `endpoint`, `error_code` |

**Key Thresholds:**
- **Error Rate:** < 5% (warning > 5%, critical > 10%)
- **5xx Error Rate:** < 1% (critical > 1%)

---

## Alert Response Procedures

### Ban Operation Failures

#### ModerationBanHighFailureRate

**Severity:** Warning  
**Threshold:** > 10% failure rate for 10 minutes

**Symptoms:**
- Users unable to ban/unban
- Error messages in application logs
- Increased support tickets

**Investigation Steps:**

1. Check current failure rate:
   ```promql
   sum(rate(moderation_ban_operations_total{status="failed"}[5m])) 
   / sum(rate(moderation_ban_operations_total[5m]))
   ```

2. Identify error types:
   ```promql
   sum(rate(moderation_ban_operations_total{status="failed"}[5m])) by (error_type)
   ```

3. Check recent errors in logs:
   ```bash
   kubectl logs -l app=clpr-backend --tail=100 | grep -i "ban.*error"
   ```

4. Verify database connectivity:
   ```bash
   kubectl exec -it deployment/clpr-backend -- psql -U clpr -c "SELECT 1"
   ```

**Common Causes & Solutions:**

| Cause | Solution |
|-------|----------|
| Database connection issues | Restart backend pods, check DB health |
| Permission configuration errors | Review role assignments, verify ACLs |
| Network timeouts | Check network latency, increase timeouts |
| Rate limiting | Implement exponential backoff |

**Mitigation:**
- Enable circuit breaker for ban operations
- Implement retry logic with backoff
- Queue failed operations for later retry

**Escalation:** Escalate to on-call engineer if failure rate > 50% or persists > 30 minutes

---

#### ModerationBanCriticalFailureRate

**Severity:** Critical  
**Threshold:** > 50% failure rate for 5 minutes

**Immediate Actions:**

1. **Alert stakeholders** in #incidents channel
2. **Check service health:**
   ```bash
   kubectl get pods -l app=clpr-backend
   kubectl top pods -l app=clpr-backend
   ```

3. **Review recent deployments:**
   ```bash
   kubectl rollout history deployment/clpr-backend
   ```

4. **Consider rollback** if related to recent deployment:
   ```bash
   kubectl rollout undo deployment/clpr-backend
   ```

5. **Enable maintenance mode** if issue persists:
   ```bash
   kubectl patch deployment clpr-backend -p '{"spec":{"replicas":0}}'
   ```

**Recovery Steps:**
- Restore from last known good configuration
- Scale backend pods if resource exhaustion
- Clear any stuck database locks
- Verify external dependencies (Twitch API, etc.)

---

### Sync Operation Failures

#### ModerationSyncFailures

**Severity:** Warning  
**Threshold:** > 0.1 failures/sec for 10 minutes

**Investigation Steps:**

1. Check sync failure reasons:
   ```promql
   sum(rate(moderation_sync_operations_total{status="failed"}[5m])) by (error_type)
   ```

2. Verify Twitch API connectivity:
   ```bash
   curl -H "Client-ID: $TWITCH_CLIENT_ID" https://api.twitch.tv/helix/users
   ```

3. Check authentication tokens:
   ```bash
   kubectl exec -it deployment/clpr-backend -- \
     psql -U clpr -d clpr -c \
     "SELECT user_id, expires_at FROM twitch_auth WHERE expires_at < NOW() LIMIT 5"
   ```

4. Review sync job logs:
   ```bash
   kubectl logs -l app=clpr-backend --tail=200 | grep -i "sync"
   ```

**Common Issues:**

| Error Type | Resolution |
|------------|------------|
| `api_error` | Check Twitch API status page |
| `auth_error` | Refresh OAuth tokens |
| `rate_limit` | Implement backoff, reduce frequency |
| `timeout` | Increase timeout values |

**Related:** See [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md) for detailed guidance

---

#### ModerationSyncCriticalFailureRate

**Severity:** Critical  
**Threshold:** > 50% failure rate

**Impact:** Ban data severely out of sync with Twitch

**Immediate Actions:**

1. Disable automatic sync temporarily:
   ```bash
   kubectl set env deployment/clpr-backend SYNC_ENABLED=false
   ```

2. Check Twitch API status:
   - Visit: https://devstatus.twitch.tv/

3. Verify OAuth configuration:
   ```bash
   kubectl get secret twitch-oauth -o jsonpath='{.data.client_id}' | base64 -d
   ```

4. Manual sync attempt:
   ```bash
   kubectl exec -it deployment/clpr-backend -- \
     curl -X POST http://localhost:8080/api/v1/moderation/sync
   ```

**Recovery:**
- Wait for Twitch API recovery if external issue
- Re-authenticate users if auth issues
- Consider manual import as last resort

---

### Permission Denial Spikes

#### ModerationPermissionDenialSpike

**Severity:** Warning  
**Threshold:** > 10 denials/sec for 5 minutes

**Investigation:**

1. Identify affected permission types:
   ```promql
   topk(5, sum(rate(moderation_permission_denials_total[5m])) by (permission_type, reason))
   ```

2. Check for configuration changes:
   ```bash
   git log --since="1 hour ago" -- backend/internal/models/roles.go
   ```

3. Review recent user activity:
   ```bash
   kubectl exec -it deployment/clpr-backend -- \
     psql -U clpr -d clpr -c \
     "SELECT user_id, action, COUNT(*) FROM moderation_audit_logs 
      WHERE created_at > NOW() - INTERVAL '1 hour' 
      GROUP BY user_id, action ORDER BY COUNT DESC LIMIT 10"
   ```

**Common Scenarios:**

| Scenario | Action |
|----------|--------|
| Role configuration change | Review and revert if incorrect |
| Coordinated attack | Enable rate limiting |
| Bug in permission check | Deploy hotfix |
| Mass user role changes | Verify intentional |

---

#### ModerationPermissionDenialCritical

**Severity:** Critical  
**Threshold:** > 50% of checks denied

**Symptoms:** Most/all users unable to perform moderation actions

**Emergency Response:**

1. **Identify the issue:**
   ```promql
   sum(rate(moderation_permission_checks_total{result="denied"}[5m])) 
   / sum(rate(moderation_permission_checks_total[5m]))
   ```

2. **Check role configuration:**
   ```bash
   kubectl exec -it deployment/clpr-backend -- \
     psql -U clpr -d clpr -c \
     "SELECT role, COUNT(*) FROM users GROUP BY role"
   ```

3. **Emergency permission bypass** (use with caution):
   ```sql
   -- Temporarily grant moderator role to specific users
   UPDATE users SET role = 'moderator' WHERE id = '<user_id>';
   ```

4. **Rollback recent changes:**
   ```bash
   kubectl rollout undo deployment/clpr-backend
   ```

---

### Slow Queries

#### ModerationSlowQueries

**Severity:** Warning  
**Threshold:** > 1 slow query/sec

**Investigation:**

1. Identify slow query types:
   ```promql
   topk(5, sum(rate(moderation_slow_queries_total[5m])) by (query_type))
   ```

2. Check database performance:
   ```bash
   kubectl exec -it postgres-0 -- \
     psql -U clpr -d clpr -c \
     "SELECT query, calls, mean_exec_time FROM pg_stat_statements 
      WHERE query LIKE '%moderation%' 
      ORDER BY mean_exec_time DESC LIMIT 10"
   ```

3. Review indexes:
   ```sql
   SELECT schemaname, tablename, indexname, idx_scan 
   FROM pg_stat_user_indexes 
   WHERE schemaname = 'public' AND tablename LIKE '%ban%' 
   ORDER BY idx_scan;
   ```

**Optimization Steps:**

1. **Add missing indexes:**
   ```sql
   CREATE INDEX CONCURRENTLY idx_bans_community_user 
   ON community_bans(community_id, user_id);
   ```

2. **Update statistics:**
   ```sql
   ANALYZE community_bans;
   ANALYZE moderation_audit_logs;
   ```

3. **Optimize query:**
   - Review query plan: `EXPLAIN ANALYZE <query>`
   - Reduce joins
   - Add appropriate WHERE clauses

---

#### ModerationSlowQueriesCritical

**Severity:** Critical  
**Threshold:** > 5 slow queries/sec

**Immediate Actions:**

1. **Check for table locks:**
   ```sql
   SELECT * FROM pg_locks WHERE NOT granted;
   ```

2. **Identify blocking queries:**
   ```sql
   SELECT pid, usename, query, state, wait_event_type 
   FROM pg_stat_activity 
   WHERE state != 'idle' AND query LIKE '%moderation%';
   ```

3. **Kill long-running queries if needed:**
   ```sql
   SELECT pg_terminate_backend(<pid>);
   ```

4. **Enable connection pooling:**
   ```bash
   kubectl scale deployment pgbouncer --replicas=3
   ```

**Long-term Solutions:**
- Partition large tables
- Implement read replicas
- Cache frequent queries
- Optimize database schema

---

### API Errors

#### ModerationAPIHighErrorRate

**Severity:** Warning  
**Threshold:** > 5% error rate

**Investigation:**

1. Check error distribution:
   ```promql
   sum(rate(moderation_api_errors_total[5m])) by (endpoint, error_code)
   ```

2. Review application logs:
   ```bash
   kubectl logs -l app=clpr-backend --tail=500 | grep -E "ERROR|FATAL"
   ```

3. Test endpoints manually:
   ```bash
   curl -X POST http://clpr-backend:8080/api/v1/moderation/ban \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"user_id": "test", "reason": "test"}'
   ```

**Common Error Codes:**

| Code | Meaning | Solution |
|------|---------|----------|
| 400 | Bad Request | Validate input data |
| 403 | Forbidden | Check permissions |
| 429 | Rate Limited | Implement backoff |
| 500 | Internal Error | Check logs, restart pods |
| 503 | Service Unavailable | Scale up, check dependencies |

---

#### ModerationAPICriticalErrorRate

**Severity:** Critical  
**Threshold:** > 10% 5xx error rate

**Immediate Actions:**

1. Check pod health:
   ```bash
   kubectl get pods -l app=clpr-backend
   kubectl describe pod <pod-name>
   ```

2. Review recent deployments:
   ```bash
   kubectl rollout status deployment/clpr-backend
   ```

3. Check resource usage:
   ```bash
   kubectl top pods -l app=clpr-backend
   ```

4. Scale up if needed:
   ```bash
   kubectl scale deployment clpr-backend --replicas=6
   ```

5. Consider rollback:
   ```bash
   kubectl rollout undo deployment/clpr-backend
   ```

---

### Audit Log Failures

#### ModerationAuditLogFailures

**Severity:** Warning  
**Threshold:** > 1 failure/sec

**Impact:** Audit trail may be incomplete (compliance risk)

**Investigation:**

1. Check failure types:
   ```promql
   sum(rate(moderation_audit_log_operations_total{status="failed"}[5m])) by (action)
   ```

2. Verify database writes:
   ```sql
   SELECT COUNT(*) FROM moderation_audit_logs 
   WHERE created_at > NOW() - INTERVAL '5 minutes';
   ```

3. Check disk space:
   ```bash
   kubectl exec -it postgres-0 -- df -h /var/lib/postgresql/data
   ```

**Resolution:**
- Ensure database has sufficient storage
- Check for write permission issues
- Verify audit log table integrity
- Consider async audit logging

**Compliance Note:** Document any audit log gaps for compliance purposes

---

#### ModerationAuditLogHighLatency

**Severity:** Warning  
**Threshold:** P95 > 500ms

**Impact:** Moderation operations may be slowed

**Investigation:**

1. Check audit log volume:
   ```promql
   rate(moderation_audit_log_operations_total[5m])
   ```

2. Review table size:
   ```sql
   SELECT pg_size_pretty(pg_total_relation_size('moderation_audit_logs'));
   ```

3. Check for missing indexes:
   ```sql
   SELECT * FROM pg_stat_user_tables WHERE relname = 'moderation_audit_logs';
   ```

**Optimization:**
- Implement async audit logging
- Archive old audit logs (> 90 days)
- Add indexes on frequently queried columns
- Consider separate audit database

---

## Dashboard Usage

### System Health Overview

**URL:** http://localhost:3000/d/moderation-system

**Key Panels:**
1. **Ban Operations Rate** - Monitor ban/unban throughput
2. **Ban Operation Failure Rate** - Gauge of system health
3. **Sync Failure Rate** - Sync operation success
4. **Ban Operation Latency** - P50/P95/P99 latencies
5. **Sync Operation Latency** - Sync performance
6. **Bans Processed During Sync** - Sync effectiveness
7. **Permission Checks Rate** - Permission system load
8. **Permission Denials by Reason** - Permission issues
9. **Permission Check Latency** - Permission check performance
10. **Audit Log Operations Rate** - Audit system health
11. **Audit Log Operation Latency** - Audit performance
12. **API Error Rate by Endpoint** - Endpoint health
13. **Slow Queries Rate** - Database performance
14. **Database Query Latency** - Query performance
15. **Active Bans by Type** - Current ban counts

**Recommended Views:**
- **Last 1 hour:** Real-time monitoring
- **Last 6 hours:** Trend analysis
- **Last 24 hours:** Daily patterns
- **Last 7 days:** Weekly trends

---

## Testing & Validation

### Metrics Export Validation

Verify metrics are being exported:

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="clpr-backend")'

# Verify specific metrics exist
curl 'http://localhost:9090/api/v1/query?query=moderation_ban_operations_total' | jq '.data.result'
```

### Alert Testing

Test alerts fire correctly:

```bash
# Generate test ban operations
for i in {1..100}; do
  curl -X POST http://clpr-backend:8080/api/v1/moderation/ban \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"user_id\": \"test_$i\", \"reason\": \"test\"}"
done

# Check if alert fired
curl http://localhost:9093/api/v1/alerts | jq '.data[] | select(.labels.alertname=="ModerationBanHighFailureRate")'
```

### Dashboard Data Validation

Verify dashboard panels show data:

1. Open dashboard: http://localhost:3000/d/moderation-system
2. Select time range: Last 1 hour
3. Verify all panels display data
4. Check for "No data" messages
5. Validate calculations match Prometheus queries

---

## Maintenance

### Regular Tasks

**Daily:**
- Review dashboard for anomalies
- Check alert history in Alertmanager
- Verify audit log completeness

**Weekly:**
- Review slow query logs
- Analyze permission denial patterns
- Audit active ban counts

**Monthly:**
- Review and tune alert thresholds
- Archive old audit logs
- Update runbooks with learnings

### Capacity Planning

Monitor these metrics for capacity planning:

```promql
# Ban operations trend
avg_over_time(rate(moderation_ban_operations_total[1h])[30d:1h])

# Audit log growth rate
avg_over_time(rate(moderation_audit_log_operations_total[1h])[30d:1h])

# Database query load
avg_over_time(rate(moderation_database_query_duration_seconds_count[1h])[30d:1h])
```

---

## References

- [Moderation Operations Runbook](./moderation-operations.md)
- [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md)
- [Permission Escalation](./permission-escalation.md)
- [Audit Log Operations](./audit-log-operations.md)
- [Moderation Incidents](./moderation-incidents.md)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboard Guide](https://grafana.com/docs/grafana/latest/dashboards/)

---

## Escalation

**P1 (Critical - < 15 min response):**
- > 50% failure rate for ban operations
- > 50% permission denial rate
- Complete moderation system outage

**P2 (High - < 1 hour response):**
- 10-50% failure rate
- Sync completely failing
- High API error rates

**P3 (Medium - < 4 hours response):**
- Elevated slow queries
- Permission denial spikes
- Audit log latency

**Contact:**
- On-call Engineer: PagerDuty
- #incidents Slack channel
- Engineering Manager (escalation)

---

**Last Updated:** 2026-02-06  
**Maintained By:** Platform Engineering Team  
**Related Issue:** #1021
