---
title: "Alert Validation"
summary: "This runbook provides step-by-step procedures for validating monitoring alert rules, ensuring they fire correctly, contain proper labels, and clear on recovery."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Alert Validation Runbook

## Overview

This runbook provides step-by-step procedures for validating monitoring alert rules, ensuring they fire correctly, contain proper labels, and clear on recovery.

## Table of Contents

- [Quick Reference](#quick-reference)
- [Prerequisites](#prerequisites)
- [Validation Procedures](#validation-procedures)
- [Synthetic Signal Generation](#synthetic-signal-generation)
- [Alert Testing](#alert-testing)
- [Dashboard Validation](#dashboard-validation)
- [Troubleshooting](#troubleshooting)
- [Continuous Validation](#continuous-validation)

## Quick Reference

### Common Commands

```bash
# Validate alert rules syntax
cd monitoring
./test-alerts.sh validate

# Run all alert validation tests
cd monitoring/tests
./alert-validation-test.sh all

# Run specific alert test
./alert-validation-test.sh latency-alert

# Validate all dashboards
./dashboard-validation-test.sh all

# Generate synthetic signals
cd monitoring/tools
./synthetic-signal-generator.sh all
```

### Service URLs

- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093
- **Grafana**: http://localhost:3000
- **Pushgateway**: http://localhost:9091

## Prerequisites

### Required Services

1. **Prometheus** - Running and scraping metrics
2. **Alertmanager** - Running and configured
3. **Prometheus Pushgateway** - Running for synthetic metrics
4. **Grafana** (optional) - For dashboard validation

### Starting Monitoring Stack

```bash
cd monitoring
docker-compose -f docker-compose.monitoring.yml up -d

# Verify services are running
docker-compose -f docker-compose.monitoring.yml ps

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check Alertmanager status
curl http://localhost:9093/api/v1/status
```

### Required Tools

- `curl` - For API requests
- `jq` - For JSON processing
- `bc` - For calculations
- `bash` - Version 4.0+

Install on Ubuntu/Debian:
```bash
sudo apt-get update
sudo apt-get install -y curl jq bc
```

## Validation Procedures

### Step 1: Validate Alert Rule Syntax

Before running tests, ensure alert rules are syntactically correct:

```bash
cd monitoring
./test-alerts.sh validate
```

This checks:
- Alert rule YAML syntax
- Required labels (severity, etc.)
- Alert coverage for critical services

### Step 2: Run Alert Validation Tests

Run the comprehensive alert validation suite:

```bash
cd monitoring/tests
./alert-validation-test.sh all
```

**What This Tests:**
- Alerts fire when thresholds are exceeded
- Alerts contain correct labels (service, severity, runbook)
- Alerts clear on recovery
- No excessive flapping (< 2 state changes/60s)

**Expected Duration**: ~30 minutes for all tests

### Step 3: Run Dashboard Validation Tests

Validate that dashboard panels accurately reflect metrics:

```bash
cd monitoring/tests
./dashboard-validation-test.sh all
```

**What This Tests:**
- Dashboard queries return expected values
- Metrics match synthetic signals within tolerance (5%)
- All dashboard panels are accessible

**Expected Duration**: ~15 minutes

## Synthetic Signal Generation

Synthetic signals simulate production conditions to trigger alerts for testing.

### Generate Individual Signals

#### High Latency Signal

```bash
cd monitoring/tools
# Generate 150ms latency for 60 seconds
./synthetic-signal-generator.sh latency 60 150
```

#### High Error Rate Signal

```bash
# Generate 1% error rate for 120 seconds
./synthetic-signal-generator.sh error-rate 120 0.01
```

#### Webhook Failure Signal

```bash
# Generate 15% webhook failure rate for 60 seconds
./synthetic-signal-generator.sh webhook-failure 60 0.15
```

#### Queue Depth Signal

```bash
# Generate queue with 500 items for 90 seconds
./synthetic-signal-generator.sh queue-depth 90 500
```

#### Search Failover Signal

```bash
# Generate 10 failovers/min for 60 seconds
./synthetic-signal-generator.sh search-failover 60 10
```

#### CDN Failover Signal

```bash
# Generate 10 failovers/sec for 60 seconds
./synthetic-signal-generator.sh cdn-failover 60 10
```

### Generate All Signals

Run all signal generators simultaneously:

```bash
./synthetic-signal-generator.sh all
```

### Send Recovery Signals

After testing, send recovery signals to clear alerts:

```bash
# Clear all alerts
./synthetic-signal-generator.sh recovery all

# Clear specific alert
./synthetic-signal-generator.sh recovery latency
```

## Alert Testing

### Manual Alert Testing Workflow

1. **Start Monitoring Stack**
   ```bash
   cd monitoring
   docker-compose -f docker-compose.monitoring.yml up -d
   ```

2. **Generate Synthetic Signal**
   ```bash
   cd monitoring/tools
   ./synthetic-signal-generator.sh latency 120 150
   ```

3. **Wait for Alert to Fire** (typically 5-10 minutes)
   - Check Prometheus alerts: http://localhost:9090/alerts
   - Check Alertmanager: http://localhost:9093

4. **Verify Alert Labels**
   ```bash
   curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="SLOLatencyBreach")'
   ```

5. **Send Recovery Signal**
   ```bash
   ./synthetic-signal-generator.sh recovery latency
   ```

6. **Verify Alert Cleared** (wait 5-10 minutes)
   ```bash
   curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="SLOLatencyBreach")'
   ```

### Expected Alert Labels

All alerts should include:

- `alertname` - Name of the alert
- `severity` - critical, warning, or info
- `service` or `job` - Affected service
- Runbook annotation with link to troubleshooting docs

Example:
```json
{
  "labels": {
    "alertname": "SLOLatencyBreach",
    "severity": "warning",
    "slo": "latency"
  },
  "annotations": {
    "summary": "Latency SLO breach for list endpoints",
    "description": "P95 latency is 0.15s, target is < 0.1s.",
    "runbook": "Check database query performance..."
  }
}
```

## Dashboard Validation

### Validate Dashboard Metrics

1. **Generate Known Metric Values**
   ```bash
   cd monitoring/tools
   ./synthetic-signal-generator.sh latency 30 150
   ```

2. **Query Prometheus for Actual Values**
   ```bash
   # Wait 35 seconds for metrics to be scraped
   sleep 35
   
   # Query P95 latency
   curl -s 'http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))' | jq '.data.result[0].value[1]'
   ```

3. **Compare Expected vs. Actual**
   - Expected: ~0.15 (150ms)
   - Tolerance: ±5%
   - Result should be between 0.1425 and 0.1575

4. **Verify in Grafana Dashboard**
   - Open http://localhost:3000
   - Navigate to SLO Dashboard
   - Check P95 latency panel shows ~150ms

### Dashboard Panel Checklist

For each dashboard panel:
- [ ] Panel loads without errors
- [ ] Query returns data
- [ ] Metric values match expected ranges
- [ ] Units are correct (ms, %, count, etc.)
- [ ] Time ranges work correctly
- [ ] Thresholds are visible and accurate

## Troubleshooting

### Alerts Not Firing

**Symptom**: Synthetic signals generated but alerts don't fire

**Diagnosis**:
1. Check Prometheus is scraping metrics:
   ```bash
   curl http://localhost:9090/api/v1/targets
   ```

2. Verify metrics exist:
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=http_request_duration_seconds_bucket' | jq '.data.result'
   ```

3. Check alert rule evaluation:
   ```bash
   curl http://localhost:9090/api/v1/alerts
   ```

**Common Causes**:
- Pushgateway not running or not scraped by Prometheus
- Alert rule syntax error (check Prometheus logs)
- `for` duration not elapsed (wait full duration)
- Metric labels don't match alert rule selectors

**Resolution**:
```bash
# Restart Prometheus to reload config
docker-compose -f docker-compose.monitoring.yml restart prometheus

# Check Prometheus logs
docker-compose -f docker-compose.monitoring.yml logs prometheus
```

### Alerts Not Clearing

**Symptom**: Alerts remain active after recovery signal sent

**Diagnosis**:
1. Check if recovery metrics were pushed:
   ```bash
   curl http://localhost:9091/metrics | grep http_request_duration_seconds
   ```

2. Verify Prometheus scraped updated metrics:
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=http_request_duration_seconds_count' | jq
   ```

**Resolution**:
- Wait for full alert `for` duration + scrape interval (typically 5-6 minutes)
- Manually delete metrics from Pushgateway and resend:
  ```bash
  curl -X DELETE http://localhost:9091/metrics/job/clpr-backend/instance/test
  ./synthetic-signal-generator.sh recovery latency
  ```

### False Positives During Testing

**Symptom**: Unexpected alerts fire during validation

**Causes**:
- Previous test signals still active
- Background services generating real metrics
- Incorrect threshold in synthetic signal generator

**Resolution**:
1. Clear all metrics from Pushgateway:
   ```bash
   curl -X PUT http://localhost:9091/api/v1/admin/wipe
   ```

2. Send recovery signals:
   ```bash
   ./synthetic-signal-generator.sh recovery all
   ```

3. Wait for alerts to clear before starting new tests

### Dashboard Metrics Not Matching

**Symptom**: Dashboard shows different values than expected

**Diagnosis**:
1. Check Prometheus query directly:
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=YOUR_QUERY' | jq
   ```

2. Compare Grafana query to Prometheus query
3. Check time range alignment

**Common Issues**:
- Grafana caching (disable in panel settings)
- Time range not aligned with synthetic signal
- Query aggregation differences

## Continuous Validation

### Automated CI Validation

Alert validation runs automatically via GitHub Actions:

- **Trigger**: Daily at 02:00 UTC
- **Also runs**: On changes to monitoring configuration
- **Duration**: ~30 minutes
- **Artifacts**: Validation report published

**View Results**:
1. Go to GitHub Actions tab
2. Select "Alert Validation" workflow
3. Download validation report artifact

### Setting Up Cron Validation

For production environments, set up periodic validation:

```bash
# Add to crontab
0 2 * * * cd /opt/clpr/monitoring/tests && ./alert-validation-test.sh all > /var/log/alert-validation.log 2>&1
```

### Validation Schedule Recommendations

- **Syntax Validation**: On every config change (via CI)
- **Alert Testing**: Daily or weekly
- **Dashboard Validation**: Weekly
- **Full Validation**: Before major releases
- **Manual Testing**: After threshold changes

## Best Practices

1. **Always Test in Staging First**: Never test alerts in production
2. **Use Dedicated Test Labels**: Tag synthetic metrics with `instance=test`
3. **Clean Up After Testing**: Send recovery signals and clear Pushgateway
4. **Document Threshold Changes**: Update alert validation report
5. **Review Validation Reports**: Check for trends in false positives
6. **Rotate On-Call Testing**: Different team members run manual tests
7. **Keep Runbooks Updated**: Document new alert types and procedures

## Related Documentation

- [Alert Validation Report](../../monitoring/docs/ALERT_VALIDATION_REPORT.md)
- [Monitoring README](../../monitoring/README.md)
- [Alert Testing Guide](alert-testing-staging.md)
- [Alertmanager Setup](../../monitoring/ALERTMANAGER_SETUP.md)
- [SLO Documentation](slos.md)

## Changelog

- **2026-01-29**: Initial alert validation runbook created
  - Complete validation procedures documented
  - Synthetic signal generation guide
  - Troubleshooting section added
  - CI integration documented
