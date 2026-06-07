---
title: "Background Jobs"
summary: "This runbook provides troubleshooting guidance for background job monitoring and alerting."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Background Jobs Runbook

This runbook provides troubleshooting guidance for background job monitoring and alerting.

## Overview

Clipper uses several background jobs (schedulers) to perform periodic maintenance tasks:

- **hot_score_refresh**: Updates hot scores for trending clips (every 5 minutes)
- **trending_score_refresh**: Recalculates trending scores (every 60 minutes)
- **clip_sync**: Syncs clips from Twitch API (every 15 minutes)
- **reputation_tasks**: Awards badges and updates user stats (every 6 hours)
- **webhook_retry**: Retries failed webhook deliveries (every 1 minute)
- **embedding_generation**: Generates embeddings for new clips (configurable interval)

## Metrics

All background jobs expose the following Prometheus metrics:

- `job_execution_total{job_name, status}`: Total job executions (status: success, failed)
- `job_execution_duration_seconds{job_name}`: Job execution duration histogram
- `job_last_success_timestamp_seconds{job_name}`: Unix timestamp of last successful run
- `job_items_processed_total{job_name, status}`: Items processed (status: success, failed, skipped)
- `job_queue_size{job_name}`: Current queue size (for jobs with queues)

## Alert Responses

### Job Failures

**Alert**: `BackgroundJobFailing`  
**Severity**: Warning  
**Threshold**: > 0.1 failures/sec for 5 minutes

#### Investigation Steps

1. Check job logs for error messages:
   ```bash
   kubectl logs -f deployment/backend -n clpr | grep "job_name"
   ```

2. Identify error patterns in recent logs:
   ```bash
   kubectl logs deployment/backend -n clpr --tail=1000 | grep -i "error\|failed"
   ```

3. Check dependent services:
   - Database connectivity: `kubectl get pods -n clpr | grep postgres`
   - Redis connectivity: `kubectl get pods -n clpr | grep redis`
   - External APIs (Twitch, OpenSearch): Check network and API status

4. Review recent changes:
   ```bash
   kubectl rollout history deployment/backend -n clpr
   ```

#### Resolution

- **Database issues**: Check connection pool settings, verify credentials
- **External API failures**: Check rate limits, API keys, network connectivity
- **Code errors**: Review stack traces, consider rollback if recent deployment
- **Resource constraints**: Check CPU/memory usage, scale if needed

### Critical Failure Rate

**Alert**: `BackgroundJobCriticalFailureRate`  
**Severity**: Critical  
**Threshold**: > 50% failure rate for 5 minutes

#### Investigation Steps

1. Immediate: Check if service is healthy:
   ```bash
   kubectl get pods -n clpr -l app=backend
   kubectl describe pod <pod-name> -n clpr
   ```

2. Check for cascading failures:
   ```bash
   kubectl logs deployment/backend -n clpr --tail=500 | grep -E "panic|fatal|critical"
   ```

3. Verify database and cache health:
   ```bash
   kubectl exec -it postgres-pod -n clpr -- psql -U clpr -c "SELECT 1;"
   kubectl exec -it redis-pod -n clpr -- redis-cli PING
   ```

#### Resolution

- **Immediate**: If recent deployment, rollback:
  ```bash
  kubectl rollout undo deployment/backend -n clpr
  ```
- **Database down**: Restart database pod or check cloud provider status
- **Code panic**: Review panic stack traces, deploy hotfix
- **Resource exhaustion**: Scale deployment immediately:
  ```bash
  kubectl scale deployment backend --replicas=5 -n clpr
  ```

### Job Not Running (Stale)

**Alert**: `BackgroundJobNotRunning`  
**Severity**: Warning  
**Threshold**: No successful run for > 2 hours

#### Investigation Steps

1. Check if job is scheduled to run:
   ```bash
   kubectl logs deployment/backend -n clpr | grep "Starting.*scheduler"
   ```

2. Verify job didn't get stuck:
   ```bash
   kubectl top pods -n clpr
   # Look for high CPU usage that might indicate stuck job
   ```

3. Check for deadlocks or long-running operations:
   ```bash
   kubectl exec -it postgres-pod -n clpr -- psql -U clpr -c \
     "SELECT pid, now() - pg_stat_activity.query_start AS duration, query 
      FROM pg_stat_activity 
      WHERE state = 'active' 
      ORDER BY duration DESC;"
   ```

#### Resolution

- **Job stuck**: Restart backend pod:
  ```bash
  kubectl rollout restart deployment/backend -n clpr
  ```
- **Database lock**: Kill long-running query:
  ```bash
  kubectl exec -it postgres-pod -n clpr -- psql -U clpr -c \
    "SELECT pg_terminate_backend(<pid>);"
  ```
- **Configuration issue**: Check job interval settings in environment variables

### Job Critically Stale

**Alert**: `BackgroundJobCriticallyStale`  
**Severity**: Critical  
**Threshold**: No successful run for > 24 hours

#### Investigation Steps

1. Verify job is enabled and configured:
   ```bash
   kubectl get configmap backend-config -n clpr -o yaml
   ```

2. Check for panic recovery or restart loops:
   ```bash
   kubectl describe pod <backend-pod> -n clpr
   # Look at restart count and events
   ```

3. Review error logs for the specific job:
   ```bash
   kubectl logs deployment/backend -n clpr --since=24h | grep "<job_name>"
   ```

#### Resolution

- **Job disabled**: Re-enable in configuration and redeploy
- **Crash loop**: Fix underlying issue causing panics
- **Resource starvation**: Increase resource limits

### High Job Duration

**Alert**: `BackgroundJobHighDuration`  
**Severity**: Warning  
**Threshold**: P95 duration > 300 seconds for 15 minutes

#### Investigation Steps

1. Check database query performance:
   ```bash
   kubectl exec -it postgres-pod -n clpr -- psql -U clpr -c \
     "SELECT query, calls, mean_exec_time, max_exec_time 
      FROM pg_stat_statements 
      ORDER BY mean_exec_time DESC 
      LIMIT 10;"
   ```

2. Analyze slow queries:
   ```bash
   kubectl logs deployment/backend -n clpr | grep "slow query\|took.*ms"
   ```

3. Check for lock contention:
   ```bash
   kubectl exec -it postgres-pod -n clpr -- psql -U clpr -c \
     "SELECT * FROM pg_locks WHERE NOT granted;"
   ```

#### Resolution

- **Slow queries**: Add indexes, optimize queries
- **Lock contention**: Review transaction isolation levels
- **Large dataset**: Implement pagination or batch processing
- **External API slow**: Add timeouts, implement circuit breakers

### Critical Job Duration

**Alert**: `BackgroundJobCriticalDuration`  
**Severity**: Critical  
**Threshold**: P95 duration > 600 seconds for 10 minutes

#### Resolution

- **Immediate**: Consider temporarily disabling job if causing system issues
- **Optimize**: Reduce batch size, add pagination, parallelize work
- **Scale**: Add more worker goroutines or pods
- **Circuit break**: Add circuit breakers for external dependencies

### Queue Growing

**Alert**: `BackgroundJobQueueGrowing`  
**Severity**: Warning  
**Threshold**: Queue > 100 and growing by > 50% over 10 minutes

#### Resolution

- **Increase parallelism**: Adjust worker count in job configuration
- **Scale horizontally**: Add more backend replicas
- **Optimize processing**: Reduce per-item processing time
- **Temporary**: Increase batch size to drain queue faster

### Critical Queue Size

**Alert**: `BackgroundJobCriticalQueueSize`  
**Severity**: Critical  
**Threshold**: Queue size > 1000 for 10 minutes

#### Resolution

- **Emergency**: Scale backend replicas immediately:
  ```bash
  kubectl scale deployment backend --replicas=10 -n clpr
  ```
- **Clear queue**: If items are stale, consider manual cleanup
- **Root cause**: Fix underlying processing issue before scaling down

### High Item Failure Rate

**Alert**: `BackgroundJobHighItemFailureRate`  
**Severity**: Warning  
**Threshold**: > 20% item failure rate for 15 minutes

#### Resolution

- **Data issues**: Fix data validation, add error handling
- **External API**: Implement retries with exponential backoff
- **Schema changes**: Run necessary migrations
- **Configuration**: Update job settings or thresholds

## Monitoring Dashboard

Access the Background Jobs dashboard in Grafana:
- URL: `<your-grafana-url>/d/background-jobs`
- Dashboard: "Background Jobs Monitoring"

Key panels:
1. **Job Execution Status Overview**: Overall success/failure rates
2. **Job Success Rate**: Percentage gauge for quick health check
3. **Job Queue Sizes**: Track queue growth
4. **Job Duration (P95)**: Performance tracking
5. **Time Since Last Success**: Staleness detection
6. **Execution Details Table**: Comprehensive job status

## Common Maintenance Tasks

### Manually Trigger Job

```bash
# Connect to backend pod
kubectl exec -it <backend-pod> -n clpr -- /bin/sh

# Trigger job via API (if available)
curl -X POST http://localhost:8080/admin/jobs/trigger \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"job_name": "hot_score_refresh"}'
```

### Adjust Job Intervals

Update environment variables in deployment and redeploy:
```bash
kubectl rollout restart deployment/backend -n clpr
```

### Check Job Configuration

```bash
kubectl exec -it <backend-pod> -n clpr -- env | grep -i "interval\|job"
```

## Escalation

If unable to resolve within 30 minutes:
1. Page on-call engineer: Use PagerDuty incident escalation
2. Notify in Slack: Post in #incidents channel
3. Consider rollback if related to recent deployment
4. Document findings in incident report

## Related Documentation

- [Operations: Monitoring](../monitoring.md)
- [Deployment: Runbook](../runbook.md)
- [Webhook Monitoring](../webhook-monitoring.md)
