---
title: "Webhook Monitoring"
summary: "This guide covers the webhook monitoring infrastructure, metrics, alerts, and troubleshooting procedures for the Clipper webhook delivery system."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Webhook Monitoring and Alerting Guide

## Overview

This guide covers the webhook monitoring infrastructure, metrics, alerts, and troubleshooting procedures for the Clipper webhook delivery system.

## Monitoring Architecture

### Metrics Collection

- **Prometheus**: Collects metrics from the backend service every 10 seconds
- **Grafana**: Visualizes metrics through the dedicated Webhook Monitoring Dashboard
- **Alertmanager**: Routes alerts to on-call engineers via configured channels

### Dashboard Access

- **Webhook Monitoring Dashboard**: `https://grafana.clpr.com/d/webhook-monitoring/webhook-monitoring-dashboard`
- **System Health**: Available via Grafana → Dashboards → "System Health Dashboard"

## Key Metrics

### Delivery Metrics

- `webhook_delivery_total` - Total webhook delivery attempts by event type and status
- `webhook_delivery_duration_seconds` - Histogram of delivery latency
- `webhook_http_status_code_total` - HTTP status codes returned by webhook endpoints

### Queue Metrics

- `webhook_retry_queue_size` - Current size of the retry queue
- `webhook_dead_letter_queue_size` - Current size of the dead-letter queue
- `webhook_retry_queue_processing_rate` - Rate of processing items from retry queue

### Subscription Metrics

- `webhook_subscriptions_active` - Number of active webhook subscriptions
- `webhook_subscription_delivery_total` - Deliveries per subscription
- `webhook_consecutive_failures` - Consecutive failures per subscription

### Performance Metrics

- `webhook_time_to_success_seconds` - Time from first attempt to successful delivery
- `webhook_retry_attempts` - Distribution of retry attempts
- `webhook_retry_total` - Total retry attempts by retry number
- `webhook_dlq_movements_total` - Rate of items moved to DLQ by reason

## Alert Severity Levels

### Critical (P1)

**Response Time**: Immediate (within 15 minutes)

Alerts:
- `CriticalWebhookFailureRate` - Failure rate > 50%
- `CriticalWebhookRetryQueue` - Retry queue > 500 items
- `CriticalWebhookDeadLetterQueue` - DLQ > 50 items
- `CriticalDLQMovementRate` - > 5 webhooks/sec moving to DLQ
- `WebhookSubscriptionCriticalFailures` - > 20 consecutive failures
- `WebhookDeliveryStalled` - Processing appears stalled

**Action**: Page on-call engineer immediately

### Warning (P2)

**Response Time**: Within 1 hour during business hours

Alerts:
- `HighWebhookFailureRate` - Failure rate > 10%
- `LargeWebhookRetryQueue` - Retry queue > 100 items
- `WebhookDeadLetterQueueItems` - DLQ > 10 items
- `HighWebhookDeliveryLatency` - P95 latency > 5s
- `WebhookDeliveryLatencySpike` - Latency doubled vs baseline
- `HighDLQMovementRate` - > 1 webhook/sec moving to DLQ
- `WebhookSubscriptionConsecutiveFailures` - > 5 consecutive failures
- `WebhookSubscriptionHealthDegradation` - Per-subscription failure rate > 50%
- `HighRetryExhaustionRate` - > 30% of retries exhausting

**Action**: Investigate during next available window or escalate if worsening

### Info (P3)

**Response Time**: Review during next business day

Alerts:
- `NoActiveWebhookSubscriptions` - No active subscriptions for 1 hour

**Action**: Informational only, no immediate action required

## Troubleshooting Runbooks

### High Webhook Failure Rate

**Symptoms**: `HighWebhookFailureRate` or `CriticalWebhookFailureRate` alert firing

**Investigation Steps**:
1. Check the Webhook Monitoring Dashboard
   - Identify which event types are failing
   - Check HTTP status code distribution
   - Review per-subscription failure rates

2. Query recent failures:
   ```promql
   topk(10, sum(rate(webhook_delivery_total{status="failed"}[5m])) by (event_type))
   ```

3. Check webhook delivery logs:
   ```bash
   kubectl logs -l app=backend --tail=100 | grep WEBHOOK | grep -i error
   ```

4. Identify failing subscriptions:
   ```sql
   SELECT s.id, s.url, s.user_id, COUNT(*) as failures
   FROM webhook_deliveries wd
   JOIN webhook_subscriptions s ON wd.subscription_id = s.id
   WHERE wd.status = 'failed'
     AND wd.created_at > NOW() - INTERVAL '1 hour'
   GROUP BY s.id, s.url, s.user_id
   ORDER BY failures DESC
   LIMIT 10;
   ```

**Common Causes**:
- **External Service Down**: Target webhook endpoints are unavailable
  - *Solution*: Contact subscription owners, consider temporarily disabling unhealthy subscriptions
  
- **Network Issues**: Connectivity problems to external services
  - *Solution*: Check network connectivity, verify DNS resolution, check firewall rules
  
- **Rate Limiting**: Webhook endpoints rate limiting our requests
  - *Solution*: Review rate limiting, coordinate with subscription owners
  
- **Invalid Endpoints**: Subscriptions with invalid URLs or SSL issues
  - *Solution*: Identify and fix/disable problematic subscriptions

**Resolution**:
- If external service is down: Monitor and wait for recovery
- If configuration issue: Fix configuration and retry failed deliveries
- If persistent issue with specific subscription: Disable subscription and notify owner

### Large Retry Queue

**Symptoms**: `LargeWebhookRetryQueue` or `CriticalWebhookRetryQueue` alert firing

**Investigation Steps**:
1. Check retry queue size trend in Grafana
2. Query retry queue statistics:
   ```sql
   SELECT event_type, COUNT(*) as count, MAX(retry_count) as max_retries
   FROM webhook_retry_queue
   GROUP BY event_type
   ORDER BY count DESC;
   ```

3. Check oldest items in queue:
   ```sql
   SELECT * FROM webhook_retry_queue
   ORDER BY created_at ASC
   LIMIT 10;
   ```

**Common Causes**:
- **Processing Backlog**: Delivery processing can't keep up with incoming events
  - *Solution*: Scale up webhook delivery workers
  
- **Stuck Items**: Items repeatedly failing and being retried
  - *Solution*: Identify and move to DLQ manually
  
- **Rate Limiting**: Target services are rate limiting
  - *Solution*: Adjust retry backoff strategy

**Resolution**:
```bash
# Scale up webhook workers
kubectl scale deployment backend --replicas=5

# Monitor queue drain rate
watch -n 5 'curl -s http://backend:8080/debug/metrics | grep webhook_retry_queue_size'
```

### Dead Letter Queue Items

**Symptoms**: `WebhookDeadLetterQueueItems` or `CriticalWebhookDeadLetterQueue` alert firing

**Investigation Steps**:
1. Query DLQ items:
   ```sql
   SELECT 
     wdlq.event_type,
     wdlq.reason,
     COUNT(*) as count,
     MIN(wdlq.created_at) as oldest_item,
     MAX(wdlq.created_at) as newest_item
   FROM webhook_dead_letter_queue wdlq
   GROUP BY wdlq.event_type, wdlq.reason
   ORDER BY count DESC;
   ```

2. Check DLQ movement rate by reason:
   ```promql
   sum(rate(webhook_dlq_movements_total[1h])) by (reason)
   ```

3. Sample failed deliveries:
   ```sql
   SELECT * FROM webhook_dead_letter_queue
   ORDER BY created_at DESC
   LIMIT 5;
   ```

**Common Causes**:
- **Invalid Subscriptions**: Subscriptions with permanently broken endpoints
  - *Solution*: Disable problematic subscriptions
  
- **Repeated Client Errors (4xx)**: Invalid payloads or authentication issues
  - *Solution*: Review payload format, fix if needed, disable subscription if client-side issue
  
- **Extended Outages**: Target services down for extended period
  - *Solution*: Contact subscription owners, consider reprocessing after recovery

**Resolution**:
```sql
-- Disable problematic subscription
UPDATE webhook_subscriptions
SET is_active = false
WHERE id = 'problematic-subscription-id';

-- Clear old DLQ items (after investigation)
DELETE FROM webhook_dead_letter_queue
WHERE created_at < NOW() - INTERVAL '7 days';
```

### High Delivery Latency

**Symptoms**: `HighWebhookDeliveryLatency` or `WebhookDeliveryLatencySpike` alert firing

**Investigation Steps**:
1. Check latency percentiles in Grafana dashboard
2. Query slowest deliveries:
   ```promql
   topk(10, histogram_quantile(0.95,
     sum(rate(webhook_delivery_duration_seconds_bucket[5m])) by (le, event_type)
   ))
   ```

3. Check if specific subscriptions are slow:
   ```sql
   SELECT 
     s.id, 
     s.url,
     AVG(EXTRACT(EPOCH FROM (wd.updated_at - wd.created_at))) as avg_duration_seconds
   FROM webhook_deliveries wd
   JOIN webhook_subscriptions s ON wd.subscription_id = s.id
   WHERE wd.status = 'delivered'
     AND wd.created_at > NOW() - INTERVAL '1 hour'
   GROUP BY s.id, s.url
   ORDER BY avg_duration_seconds DESC
   LIMIT 10;
   ```

**Common Causes**:
- **Slow Endpoints**: Target webhook endpoints responding slowly
  - *Solution*: Contact subscription owners about performance
  
- **Network Latency**: High network latency to target endpoints
  - *Solution*: Monitor network performance, consider geographic distribution
  
- **Resource Constraints**: Backend service under resource pressure
  - *Solution*: Check CPU/memory usage, scale up if needed

**Resolution**:
- Monitor slow subscriptions and notify owners
- Consider implementing timeout adjustments
- Scale backend service if under resource pressure

### Subscription Consecutive Failures

**Symptoms**: `WebhookSubscriptionConsecutiveFailures` or `WebhookSubscriptionCriticalFailures` alert firing

**Investigation Steps**:
1. Identify failing subscription:
   ```promql
   webhook_consecutive_failures > 5
   ```

2. Query subscription details:
   ```sql
   SELECT 
     s.*,
     COUNT(wd.id) as total_deliveries,
     SUM(CASE WHEN wd.status = 'failed' THEN 1 ELSE 0 END) as failures
   FROM webhook_subscriptions s
   LEFT JOIN webhook_deliveries wd ON s.id = wd.subscription_id
   WHERE s.id = 'subscription-id'
     AND wd.created_at > NOW() - INTERVAL '1 hour'
   GROUP BY s.id;
   ```

3. Check recent delivery attempts:
   ```sql
   SELECT * FROM webhook_deliveries
   WHERE subscription_id = 'subscription-id'
   ORDER BY created_at DESC
   LIMIT 20;
   ```

**Common Causes**:
- **Endpoint Down**: Target endpoint is unavailable
- **Authentication Issues**: Signature verification failing
- **Configuration Changes**: Recent changes to subscription or endpoint

**Resolution**:
```sql
-- Temporarily disable failing subscription
UPDATE webhook_subscriptions
SET is_active = false
WHERE id = 'subscription-id';

-- Notify subscription owner
INSERT INTO notifications (user_id, type, message)
VALUES (
  (SELECT user_id FROM webhook_subscriptions WHERE id = 'subscription-id'),
  'webhook_subscription_health',
  'Your webhook subscription has been temporarily disabled due to repeated failures.'
);
```

### Webhook Delivery Stalled

**Symptoms**: `WebhookDeliveryStalled` alert firing

**Investigation Steps**:
1. Check webhook scheduler status:
   ```bash
   kubectl logs -l app=backend | grep "webhook.*scheduler" | tail -50
   ```

2. Verify background workers are running:
   ```bash
   kubectl get pods -l app=backend
   kubectl describe pod <backend-pod-name>
   ```

3. Check for deadlocks or stuck transactions:
   ```sql
   SELECT * FROM pg_stat_activity
   WHERE query LIKE '%webhook%'
     AND state = 'active'
     AND query_start < NOW() - INTERVAL '5 minutes';
   ```

**Common Causes**:
- **Scheduler Not Running**: Background job scheduler crashed or stopped
- **Database Issues**: Connection pool exhausted or deadlocks
- **Infinite Loop**: Bug causing processing to hang

**Resolution**:
```bash
# Restart backend pods
kubectl rollout restart deployment backend

# Monitor recovery
watch -n 2 'kubectl logs -l app=backend --tail=20 | grep WEBHOOK'

# Verify queue is draining
curl http://backend:8080/health/webhooks
```

## Alert Configuration

### Alertmanager Routes

Webhook alerts are routed based on severity:

```yaml
routes:
  - match:
      severity: critical
    receiver: pagerduty-critical
    group_wait: 10s
    group_interval: 5m
    repeat_interval: 4h
    
  - match:
      severity: warning
    receiver: slack-alerts
    group_wait: 30s
    group_interval: 15m
    repeat_interval: 12h
    
  - match:
      severity: info
    receiver: email-notifications
    group_wait: 1m
    group_interval: 1h
    repeat_interval: 24h
```

### On-Call Rotation

- **Primary On-Call**: Receives PagerDuty alerts for critical issues
- **Secondary On-Call**: Backup for primary
- **Escalation**: After 15 minutes, escalates to engineering manager

## Maintenance Procedures

### Regular Maintenance Tasks

**Daily**:
- Review DLQ items from previous day
- Check for subscriptions with high failure rates
- Monitor queue sizes and latency trends

**Weekly**:
- Review and clean up old DLQ items (> 7 days)
- Audit inactive subscriptions
- Check for performance degradation trends

**Monthly**:
- Review alert thresholds and tune as needed
- Update runbooks based on recent incidents
- Capacity planning review

### Disabling Problematic Subscriptions

```sql
-- Find unhealthy subscriptions
SELECT 
  s.id, 
  s.url,
  s.user_id,
  COUNT(wd.id) as total_attempts,
  SUM(CASE WHEN wd.status = 'failed' THEN 1 ELSE 0 END) as failures,
  (SUM(CASE WHEN wd.status = 'failed' THEN 1 ELSE 0 END)::float / COUNT(wd.id)) as failure_rate
FROM webhook_subscriptions s
JOIN webhook_deliveries wd ON s.id = wd.subscription_id
WHERE wd.created_at > NOW() - INTERVAL '24 hours'
  AND s.is_active = true
GROUP BY s.id, s.url, s.user_id
HAVING (SUM(CASE WHEN wd.status = 'failed' THEN 1 ELSE 0 END)::float / COUNT(wd.id)) > 0.8
  AND COUNT(wd.id) > 10
ORDER BY failure_rate DESC;

-- Disable subscription
UPDATE webhook_subscriptions
SET is_active = false
WHERE id = 'subscription-id';
```

### Manual DLQ Reprocessing

For outbound webhooks (generic webhook deliveries), items in the DLQ can be manually reprocessed if the underlying issue has been resolved:

```sql
-- Review items in the outbound webhook DLQ
SELECT 
  subscription_id,
  event_type,
  error_message,
  http_status_code,
  attempt_count,
  COUNT(*) as count
FROM outbound_webhook_dead_letter_queue
WHERE moved_to_dlq_at > NOW() - INTERVAL '24 hours'
GROUP BY subscription_id, event_type, error_message, http_status_code, attempt_count
ORDER BY count DESC;

-- Items can be replayed by updating the replayed_at timestamp
-- This marks them for potential reprocessing by background jobs
UPDATE outbound_webhook_dead_letter_queue
SET replayed_at = NOW()
WHERE reason IN (
  'max_retries_network_error',
  'max_retries_client_error',
  'max_retries_server_error'
)
  AND moved_to_dlq_at > NOW() - INTERVAL '1 hour'
  AND replayed_at IS NULL;
```

For Stripe webhook retries (if applicable):

```sql
-- Move Stripe webhook items from DLQ back to retry queue for reprocessing
INSERT INTO webhook_retry_queue (
  stripe_event_id,
  event_type,
  payload,
  retry_count,
  max_retries,
  next_retry_at,
  created_at
)
SELECT 
  stripe_event_id,
  event_type,
  payload,
  0 as retry_count,
  3 as max_retries,
  NOW() as next_retry_at,
  NOW() as created_at
FROM webhook_dead_letter_queue
WHERE created_at > NOW() - INTERVAL '1 hour'
  AND stripe_event_id NOT IN (SELECT stripe_event_id FROM webhook_retry_queue);

-- Remove from DLQ after moving to retry queue
DELETE FROM webhook_dead_letter_queue
WHERE stripe_event_id IN (
  SELECT stripe_event_id FROM webhook_retry_queue
  WHERE created_at > NOW() - INTERVAL '1 hour'
);
```

## Dashboard Panels

The Webhook Monitoring Dashboard includes:

1. **Summary Stats**: Success rate, active subscriptions, queue sizes
2. **Delivery Rate**: Success, failed, and retry rates over time
3. **Latency**: P50, P95, P99 delivery latency
4. **HTTP Status Codes**: Distribution of response codes
5. **Event Type Distribution**: Deliveries by event type
6. **Retry Analysis**: Retry attempts distribution
7. **Queue Trends**: Retry and DLQ size over time
8. **DLQ Movements**: Rate of items moving to DLQ by reason
9. **Time to Success**: P95 time from first attempt to success
10. **Subscription Health**: Top subscriptions by volume and failure rate
11. **Consecutive Failures**: Subscriptions with ongoing issues

## Contact and Escalation

- **Slack Channel**: `#alerts-webhooks`
- **On-Call**: PagerDuty rotation "Webhook Delivery"
- **Documentation**: [Webhook Integration Guide](../backend/webhooks.md)
- **Escalation**: Engineering Manager (after 15 minutes for P1 alerts)

## References

- [Webhook Integration Guide](../backend/webhooks.md)
- [Webhook Retry System](../backend/webhook-retry.md)
- [Prometheus Configuration](../../monitoring/prometheus.yml)
- [Alert Rules](../../monitoring/alerts.yml)
- [Grafana Dashboard](../../monitoring/dashboards/webhook-monitoring.json)
