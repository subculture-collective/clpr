---
title: "Moderation Monitoring and Alerting"
summary: "Monitoring setup, key metrics, and alert configuration for moderation system"
tags: ["operations", "runbook", "monitoring", "alerts", "metrics"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["monitoring", "alerts", "metrics"]
---

# Moderation Monitoring and Alerting

## Overview

This runbook provides guidance on monitoring the moderation system, configuring alerts, and responding to metric anomalies. Proper monitoring ensures early detection of issues and maintains system health.

**Audience**: Operations team, SRE team, on-call engineers

**Prerequisites**:
- Access to monitoring dashboards (Grafana, Prometheus)
- Alert manager access
- Understanding of moderation system architecture

## Table of Contents

- [Key Metrics to Monitor](#key-metrics-to-monitor)
- [Monitoring Setup](#monitoring-setup)
  - [Prometheus Configuration](#prometheus-configuration)
  - [Grafana Dashboards](#grafana-dashboards)
- [Alert Configuration](#alert-configuration)
  - [Critical Alerts](#critical-alerts)
  - [Warning Alerts](#warning-alerts)
  - [Info Alerts](#info-alerts)
- [Alert Response](#alert-response)
- [Dashboard Reference](#dashboard-reference)
- [Log Monitoring](#log-monitoring)
- [Related Runbooks](#related-runbooks)

---

## Key Metrics to Monitor

### System Health Metrics

| Metric | Description | Healthy Range | Alert Threshold |
|--------|-------------|---------------|-----------------|
| `moderation_api_uptime` | API availability percentage | > 99.9% | < 99.5% |
| `moderation_api_latency_p95` | 95th percentile response time | < 200ms | > 500ms |
| `moderation_api_latency_p99` | 99th percentile response time | < 500ms | > 1000ms |
| `moderation_api_error_rate` | Error rate percentage | < 0.1% | > 1% |
| `moderation_db_connections` | Active database connections | < 80 | > 90 |
| `moderation_cache_hit_rate` | Redis cache hit rate | > 90% | < 70% |

### Operational Metrics

| Metric | Description | Healthy Range | Alert Threshold |
|--------|-------------|---------------|-----------------|
| `moderation_bans_total` | Total active bans | N/A | N/A (trending) |
| `moderation_bans_created_rate` | Bans created per minute | < 10/min | > 100/min (spike) |
| `moderation_unbans_rate` | Unbans per minute | < 5/min | > 50/min (spike) |
| `moderation_sync_success_rate` | Ban sync success rate | > 95% | < 80% |
| `moderation_sync_duration` | Ban sync average duration | < 5s | > 30s |
| `moderation_moderator_count` | Total active moderators | N/A | N/A (trending) |

### Audit and Compliance Metrics

| Metric | Description | Healthy Range | Alert Threshold |
|--------|-------------|---------------|-----------------|
| `audit_log_write_rate` | Audit logs written per second | > 0 | = 0 (critical) |
| `audit_log_write_failures` | Failed audit log writes | 0 | > 0 |
| `audit_log_lag` | Time delay in log writes | < 1s | > 5s |
| `audit_log_export_count` | Exports per day | N/A | N/A (compliance) |

### Twitch Integration Metrics

| Metric | Description | Healthy Range | Alert Threshold |
|--------|-------------|---------------|-----------------|
| `twitch_api_latency` | Twitch API response time | < 500ms | > 2000ms |
| `twitch_api_error_rate` | Twitch API errors | < 1% | > 5% |
| `twitch_rate_limit_remaining` | Remaining Twitch API calls | > 100 | < 50 |
| `twitch_auth_refresh_success_rate` | Token refresh success | > 99% | < 95% |

---

## Monitoring Setup

### Prometheus Configuration

#### Metrics Endpoints

```yaml
# /etc/prometheus/prometheus.yml

scrape_configs:
  - job_name: 'clpr-moderation'
    scrape_interval: 15s
    metrics_path: '/api/v1/metrics'
    static_configs:
      - targets:
          - 'api.clpr.tv:443'
    scheme: https
    bearer_token: 'YOUR_METRICS_TOKEN'
    
  - job_name: 'clpr-moderation-db'
    scrape_interval: 30s
    static_configs:
      - targets:
          - 'db-exporter.clpr.tv:9187'
    
  - job_name: 'clpr-redis'
    scrape_interval: 15s
    static_configs:
      - targets:
          - 'redis-exporter.clpr.tv:9121'
```

#### Custom Metrics Queries

```promql
# API Error Rate (percentage)
rate(moderation_api_errors_total[5m]) / rate(moderation_api_requests_total[5m]) * 100

# p95 Latency
histogram_quantile(0.95, rate(moderation_api_latency_seconds_bucket[5m]))

# Ban Creation Rate (per minute)
rate(moderation_bans_created_total[1m]) * 60

# Sync Success Rate
rate(moderation_sync_success_total[5m]) / rate(moderation_sync_attempts_total[5m]) * 100

# Twitch Rate Limit Usage
(twitch_rate_limit_total - twitch_rate_limit_remaining) / twitch_rate_limit_total * 100
```

---

### Grafana Dashboards

#### Create Moderation Dashboard

```json
{
  "dashboard": {
    "title": "Moderation System Dashboard",
    "tags": ["moderation", "operations"],
    "timezone": "utc",
    "panels": [
      {
        "title": "API Requests per Minute",
        "targets": [
          {
            "expr": "rate(moderation_api_requests_total[1m]) * 60",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(moderation_api_errors_total[5m]) / rate(moderation_api_requests_total[5m]) * 100",
            "legendFormat": "Error %"
          }
        ],
        "type": "graph",
        "alert": {
          "name": "High Error Rate",
          "conditions": [
            {
              "evaluator": {
                "params": [1],
                "type": "gt"
              },
              "operator": {
                "type": "and"
              },
              "query": {
                "params": ["A", "5m", "now"]
              },
              "reducer": {
                "params": [],
                "type": "avg"
              },
              "type": "query"
            }
          ]
        }
      },
      {
        "title": "Active Bans",
        "targets": [
          {
            "expr": "moderation_bans_active_total",
            "legendFormat": "Active Bans"
          }
        ],
        "type": "stat"
      },
      {
        "title": "Ban Sync Success Rate",
        "targets": [
          {
            "expr": "rate(moderation_sync_success_total[5m]) / rate(moderation_sync_attempts_total[5m]) * 100",
            "legendFormat": "Success Rate %"
          }
        ],
        "type": "gauge",
        "thresholds": [
          {"value": 80, "color": "red"},
          {"value": 95, "color": "yellow"},
          {"value": 99, "color": "green"}
        ]
      },
      {
        "title": "Audit Log Write Rate",
        "targets": [
          {
            "expr": "rate(audit_log_writes_total[1m]) * 60",
            "legendFormat": "Logs per minute"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Twitch API Rate Limit",
        "targets": [
          {
            "expr": "twitch_rate_limit_remaining",
            "legendFormat": "Remaining"
          },
          {
            "expr": "twitch_rate_limit_total",
            "legendFormat": "Total"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

#### Import Dashboard

```bash
# Export dashboard ID from Grafana community
DASHBOARD_ID="12345"

# Import via API
curl -X POST https://grafana.clpr.tv/api/dashboards/import \
  -H "Authorization: Bearer $GRAFANA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "dashboard": {
      "id": null,
      "uid": "moderation-ops",
      "title": "Moderation Operations"
    },
    "inputs": [{
      "name": "DS_PROMETHEUS",
      "type": "datasource",
      "pluginId": "prometheus",
      "value": "Prometheus"
    }],
    "overwrite": true
  }'
```

---

## Alert Configuration

### Critical Alerts

#### High Error Rate

```yaml
# /etc/alertmanager/rules/moderation-critical.yml

groups:
  - name: moderation_critical
    interval: 30s
    rules:
      - alert: ModerationAPIHighErrorRate
        expr: |
          rate(moderation_api_errors_total[5m]) / 
          rate(moderation_api_requests_total[5m]) * 100 > 1
        for: 2m
        labels:
          severity: critical
          component: moderation-api
        annotations:
          summary: "Moderation API error rate is {{ $value }}%"
          description: "Error rate above 1% for 2 minutes"
          runbook: "https://docs.clpr.tv/operations/runbooks/moderation-incidents"
```

#### API Down

```yaml
      - alert: ModerationAPIDown
        expr: up{job="clpr-moderation"} == 0
        for: 1m
        labels:
          severity: critical
          component: moderation-api
        annotations:
          summary: "Moderation API is down"
          description: "API has been unavailable for 1 minute"
          action: "Check service status and restart if needed"
```

#### Audit Log Failure

```yaml
      - alert: AuditLogWriteFailure
        expr: rate(audit_log_write_failures_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
          component: audit-logging
          compliance: "true"
        annotations:
          summary: "Audit log writes are failing"
          description: "{{ $value }} failed writes in last 5 minutes"
          action: "Check database connectivity and disk space"
```

---

### Warning Alerts

#### High Latency

```yaml
  - name: moderation_warnings
    interval: 1m
    rules:
      - alert: ModerationAPIHighLatency
        expr: |
          histogram_quantile(0.95, 
            rate(moderation_api_latency_seconds_bucket[5m])
          ) > 0.5
        for: 5m
        labels:
          severity: warning
          component: moderation-api
        annotations:
          summary: "API p95 latency is {{ $value }}s"
          description: "Latency above 500ms threshold"
```

#### Low Cache Hit Rate

```yaml
      - alert: ModerationCacheLowHitRate
        expr: |
          rate(moderation_cache_hits[5m]) / 
          (rate(moderation_cache_hits[5m]) + rate(moderation_cache_misses[5m])) 
          * 100 < 70
        for: 10m
        labels:
          severity: warning
          component: redis-cache
        annotations:
          summary: "Cache hit rate is {{ $value }}%"
          description: "Hit rate below 70% threshold"
          action: "Check Redis memory usage and eviction policy"
```

#### Ban Sync Failures

```yaml
      - alert: BanSyncHighFailureRate
        expr: |
          rate(moderation_sync_failures_total[10m]) / 
          rate(moderation_sync_attempts_total[10m]) * 100 > 20
        for: 15m
        labels:
          severity: warning
          component: ban-sync
        annotations:
          summary: "Ban sync failure rate is {{ $value }}%"
          description: "More than 20% of syncs failing"
          runbook: "https://docs.clpr.tv/operations/runbooks/ban-sync-troubleshooting"
```

---

### Info Alerts

#### Unusual Activity

```yaml
  - name: moderation_info
    interval: 5m
    rules:
      - alert: UnusualBanActivity
        expr: |
          rate(moderation_bans_created_total[5m]) > 
          avg_over_time(rate(moderation_bans_created_total[5m])[1h:5m]) * 3
        for: 10m
        labels:
          severity: info
          component: moderation
        annotations:
          summary: "Ban creation rate {{ $value }}x above average"
          description: "Possible spam attack or mass moderation event"
          action: "Review recent bans in audit logs"
```

#### Rate Limit Warning

```yaml
      - alert: TwitchRateLimitWarning
        expr: twitch_rate_limit_remaining < 100
        for: 2m
        labels:
          severity: info
          component: twitch-integration
        annotations:
          summary: "Twitch rate limit at {{ $value }} remaining"
          description: "Approaching rate limit threshold"
          action: "Throttle sync operations if needed"
```

---

## Alert Response

### Alert Severity Levels

| Severity | Response Time | Escalation | Examples |
|----------|--------------|------------|----------|
| **Critical** | Immediate | PagerDuty | API down, audit log failure |
| **Warning** | < 30 min | Slack notification | High latency, sync failures |
| **Info** | < 2 hours | Email | Unusual activity, rate limits |

### Response Procedures

#### Critical Alert Response

1. **Acknowledge alert** in PagerDuty
2. **Assess impact** 
   - Check dashboard for affected services
   - Review recent deployments
3. **Mitigate immediately**
   - Follow relevant runbook
   - Consider rollback if needed
4. **Escalate if unresolved** after 15 minutes
5. **Document** in incident ticket

#### Warning Alert Response

1. **Check dashboard** for trends
2. **Review logs** for errors
3. **Take corrective action** if needed
4. **Monitor** for 30 minutes
5. **Escalate** if becomes critical

---

## Dashboard Reference

### Quick Links

- **Main Dashboard**: https://grafana.clpr.tv/d/moderation-ops
- **Error Dashboard**: https://grafana.clpr.tv/d/moderation-errors
- **Sync Dashboard**: https://grafana.clpr.tv/d/ban-sync
- **Audit Log Dashboard**: https://grafana.clpr.tv/d/audit-logs

### Key Panels

1. **API Health Overview**
   - Request rate
   - Error rate
   - Latency percentiles
   - Uptime

2. **Ban Operations**
   - Active bans
   - Ban creation rate
   - Unban rate
   - Ban duration distribution

3. **Sync Operations**
   - Sync success/failure rate
   - Sync duration
   - Twitch API latency
   - Rate limit status

4. **Audit Logs**
   - Write rate
   - Write failures
   - Export count
   - Storage usage

---

## Log Monitoring

### Important Log Patterns

#### Error Patterns to Watch

```bash
# High error rate in logs
tail -f /var/log/clpr/moderation.log | grep -E "ERROR|FATAL"

# Authentication failures
tail -f /var/log/clpr/moderation.log | grep "authentication failed"

# Database connection issues
tail -f /var/log/clpr/moderation.log | grep "database connection"

# Twitch API errors
tail -f /var/log/clpr/moderation.log | grep "twitch.*error"
```

#### Success Patterns

```bash
# Successful bans
tail -f /var/log/clpr/moderation.log | grep "ban_user.*success"

# Successful syncs
tail -f /var/log/clpr/moderation.log | grep "sync_bans.*completed"
```

### Centralized Logging

#### Elasticsearch Query Examples

```json
{
  "query": {
    "bool": {
      "must": [
        {"match": {"component": "moderation"}},
        {"match": {"level": "ERROR"}},
        {"range": {"@timestamp": {"gte": "now-1h"}}}
      ]
    }
  },
  "sort": [{"@timestamp": "desc"}],
  "size": 100
}
```

#### Kibana Dashboard

```bash
# Create index pattern
curl -X POST https://kibana.clpr.tv/api/saved_objects/index-pattern/moderation-logs \
  -H "kbn-xsrf: true" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "title": "moderation-*",
      "timeFieldName": "@timestamp"
    }
  }'
```

---

## Monitoring Best Practices

### Proactive Monitoring

- [ ] Review dashboards daily
- [ ] Check alert history weekly
- [ ] Analyze trends monthly
- [ ] Update thresholds quarterly

### Alert Hygiene

- [ ] Acknowledge all alerts
- [ ] Document resolutions
- [ ] Tune noisy alerts
- [ ] Remove obsolete alerts
- [ ] Test alerts monthly

### On-Call Checklist

Daily:
- [ ] Check dashboards for anomalies
- [ ] Review overnight alerts
- [ ] Verify backup jobs ran
- [ ] Check audit log exports

Weekly:
- [ ] Review error trends
- [ ] Update runbooks
- [ ] Test alert escalation
- [ ] Capacity planning review

---

## Related Runbooks

- [Moderation Operations](./moderation-operations.md) - Operational procedures
- [Moderation Incidents](./moderation-incidents.md) - Incident response
- [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md) - Sync issues
- [Moderation Rollback](./moderation-rollback.md) - Rollback procedures

---

**Last Updated**: 2026-02-03  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
