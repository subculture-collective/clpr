---
title: "Alert Testing Staging"
summary: "This document describes procedures for testing alert rules in the staging environment to ensure they fire correctly before deploying to production."
tags: ["operations","testing"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Alert Testing in Staging

This document describes procedures for testing alert rules in the staging environment to ensure they fire correctly before deploying to production.

**Related Issues (Roadmap 5.0 - Phase 5.3):**
- [#805 - Observability Infrastructure](https://git.subcult.tv/subculture-collective/clpr/issues/805)
- [#858 - Grafana Dashboards](https://git.subcult.tv/subculture-collective/clpr/issues/858)
- [#860 - Alerting Configuration](https://git.subcult.tv/subculture-collective/clpr/issues/860)

## Overview

Testing alerts in staging ensures:
- Alert rules are syntactically correct
- Thresholds trigger at expected values
- Routing to Slack/PagerDuty works correctly
- Runbook links are accessible
- Inhibition rules function properly

## Prerequisites

### Access Required

- [ ] Staging Prometheus: `https://prometheus-staging.clpr.app` or `http://localhost:9090` (port-forward)
- [ ] Staging Alertmanager: `https://alertmanager-staging.clpr.app` or `http://localhost:9093`
- [ ] Staging Grafana: `https://grafana-staging.clpr.app`
- [ ] Staging Slack workspace (or test channels in main workspace)
- [ ] PagerDuty staging integration keys

### Tools Required

- `kubectl` access to staging cluster
- `curl` and `jq` for API calls
- `promtool` for rule validation (optional but recommended)

### Staging Setup

Ensure staging has monitoring stack deployed:
```bash
# Check Prometheus is running
kubectl get pods -n monitoring | grep prometheus

# Check Alertmanager is running
kubectl get pods -n monitoring | grep alertmanager

# Port-forward if needed
kubectl port-forward -n monitoring svc/prometheus 9090:9090 &
kubectl port-forward -n monitoring svc/alertmanager 9093:9093 &
```

## Test Procedures

### 1. Validate Alert Rules Syntax

**Using promtool (recommended):**
```bash
cd monitoring

# Validate alert rules
promtool check rules alerts.yml

# Output should show:
# Checking alerts.yml
#   SUCCESS: 153 rules found
```

**Using test script:**
```bash
cd monitoring
./test-alerts.sh validate
```

**Expected output:**
```
[INFO] Checking Prometheus connectivity...
[SUCCESS] Prometheus is accessible
[INFO] Checking Alertmanager connectivity...
[SUCCESS] Alertmanager is accessible
[INFO] Validating alert rules syntax...
[SUCCESS] Alert rules are valid
[INFO] Checking alert coverage...
[SUCCESS] ✓ ServiceDown
[SUCCESS] ✓ DatabaseDown
...
[SUCCESS] All required alerts are configured
```

### 2. Test Critical Alert Routing

Critical alerts should route to PagerDuty and Slack #incidents.

**Test procedure:**
```bash
cd monitoring
./test-alerts.sh test-critical
```

**Manual test:**
```bash
# Send test critical alert
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestCriticalAlert",
      "severity": "critical",
      "service": "test"
    },
    "annotations": {
      "summary": "Test critical alert",
      "description": "Testing alert routing to PagerDuty and Slack"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify:**
- [ ] Alert appears in Alertmanager UI
- [ ] PagerDuty incident created (check "Clipper Critical" service)
- [ ] Slack message posted to #incidents channel
- [ ] Alert includes summary, description, and timestamp
- [ ] Alert acknowledges and resolves properly

**Cleanup:**
```bash
# Resolve the test alert
curl -X DELETE "http://localhost:9093/api/v1/alerts?filter=alertname=TestCriticalAlert"
```

### 3. Test Warning Alert Routing

Warning alerts should route to Slack #alerts only (not PagerDuty).

**Test procedure:**
```bash
cd monitoring
./test-alerts.sh test-warning
```

**Manual test:**
```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestWarningAlert",
      "severity": "warning",
      "service": "test"
    },
    "annotations": {
      "summary": "Test warning alert",
      "description": "Testing warning level routing"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify:**
- [ ] Alert appears in Alertmanager UI
- [ ] Slack message posted to #alerts channel
- [ ] NO PagerDuty incident created
- [ ] Alert acknowledges and resolves properly

### 4. Test Security Alert Routing

Security alerts should route to dedicated PagerDuty security service and Slack #security.

**Test procedure:**
```bash
cd monitoring
./test-alerts.sh test-security
```

**Manual test:**
```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestSecurityAlert",
      "severity": "critical",
      "security": "true",
      "service": "test"
    },
    "annotations": {
      "summary": "Test security alert",
      "description": "Testing security alert routing"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify:**
- [ ] Alert appears in Alertmanager UI
- [ ] PagerDuty incident created in "Clipper Security" service
- [ ] Slack message posted to #security channel
- [ ] Alert includes security context
- [ ] Response time is faster (5s group_wait vs 10s)

### 5. Test SLO Breach Alert

SLO breach alerts should route to dedicated PagerDuty SLO service and Slack #incidents.

**Test procedure:**
```bash
cd monitoring
./test-alerts.sh test-slo
```

**Manual test:**
```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "SLOAvailabilityBreach",
      "severity": "critical",
      "slo": "availability"
    },
    "annotations": {
      "summary": "Test SLO availability breach",
      "description": "Service availability is 99.3%, target is 99.5%",
      "runbook": "docs/operations/playbooks/slo-breach-response.md"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify:**
- [ ] Alert appears in Alertmanager UI
- [ ] PagerDuty incident created in "Clipper SLO" service
- [ ] Slack message posted to #incidents channel
- [ ] Runbook link included and accessible
- [ ] Alert shows SLO type and threshold

### 6. Test Alert Inhibition

Inhibition rules should suppress lower-priority alerts when higher-priority alerts fire.

**Test procedure:**
```bash
cd monitoring
./test-alerts.sh test-inhibition
```

**Manual test:**
```bash
# Step 1: Send ServiceDown alert (critical)
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "ServiceDown",
      "severity": "critical",
      "job": "test-service"
    },
    "annotations": {
      "summary": "Test service is down",
      "description": "Testing inhibition rules"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'

# Step 2: Wait 2 seconds
sleep 2

# Step 3: Send HighErrorRate alert (warning) - should be inhibited
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "HighErrorRate",
      "severity": "warning",
      "job": "test-service"
    },
    "annotations": {
      "summary": "Test high error rate",
      "description": "Testing inhibition - should be suppressed"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify in Alertmanager UI:**
- [ ] ServiceDown alert is active
- [ ] HighErrorRate alert is shown but marked as "inhibited"
- [ ] Only ServiceDown notification sent to Slack/PagerDuty
- [ ] Inhibition reason shown in Alertmanager

**Check inhibition rules:**
```bash
# View configured inhibition rules
kubectl get configmap -n monitoring alertmanager-config -o yaml | grep -A 20 "inhibit_rules"
```

### 7. Test Real Alert Firing

Trigger real alerts by simulating actual conditions.

#### Test ServiceDown Alert

**Trigger:**
```bash
# Scale deployment to 0 (service will be down)
kubectl scale deployment backend --replicas=0 -n staging

# Wait 2 minutes for alert to fire
sleep 120

# Check alert status
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="ServiceDown")'
```

**Verify:**
- [ ] ServiceDown alert fires after ~1 minute
- [ ] Alert routes to PagerDuty and Slack
- [ ] Alert includes correct service labels
- [ ] Runbook reference present

**Cleanup:**
```bash
# Restore service
kubectl scale deployment backend --replicas=3 -n staging
```

#### Test High Error Rate Alert

**Trigger:**
```bash
# Generate errors by calling invalid endpoint repeatedly
for i in {1..100}; do
  curl -s https://staging.clpr.app/api/v1/invalid-endpoint > /dev/null &
done

# Wait for error rate to accumulate
sleep 60

# Check alert status
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="HighErrorRate")'
```

**Verify:**
- [ ] HighErrorRate or CriticalErrorRate alert fires
- [ ] Alert routes correctly based on severity
- [ ] Error rate shown in alert description

#### Test High Memory Usage Alert

**Trigger:**
```bash
# Stress test memory (use load testing tool)
# OR simulate by reducing memory limits temporarily
kubectl set resources deployment backend \
  --limits=memory=256Mi -n staging

# Generate load
# Wait for memory alert to fire
```

**Verify:**
- [ ] HighMemoryUsage alert fires when > 80%
- [ ] Alert routes to Slack #alerts
- [ ] Instance/pod information included

**Cleanup:**
```bash
# Restore original memory limits
kubectl set resources deployment backend \
  --limits=memory=2Gi -n staging
```

#### Test Security Alerts

**Trigger FailedAuthenticationSpike:**
```bash
# Attempt multiple failed logins
for i in {1..50}; do
  curl -X POST https://staging.clpr.app/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"invalid","password":"invalid"}' &
done

# Wait for alert
sleep 180
```

**Verify:**
- [ ] FailedAuthenticationSpike alert fires
- [ ] Routes to PagerDuty Security and #security channel
- [ ] Alert includes rate information

### 8. Test Alert Silencing

Test that silences work correctly.

**Create silence:**
```bash
# Using amtool
amtool silence add alertname=TestAlert \
  --duration=10m \
  --comment="Testing silence functionality" \
  --author="test@clpr.app"

# Using Alertmanager API
curl -X POST http://localhost:9093/api/v1/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {
        "name": "alertname",
        "value": "TestAlert",
        "isRegex": false
      }
    ],
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
    "endsAt": "'$(date -u +%Y-%m-%dT%H:%M:%S -d @$(($(date +%s) + 600)))Z'",
    "comment": "Testing silence",
    "createdBy": "test@clpr.app"
  }'
```

**Send test alert:**
```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning"
    },
    "annotations": {
      "summary": "Test alert - should be silenced"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify:**
- [ ] Alert appears in Alertmanager UI as "silenced"
- [ ] NO notification sent to Slack or PagerDuty
- [ ] Silence listed in Silences tab
- [ ] Silence auto-expires after 10 minutes

**List and remove silences:**
```bash
# List active silences
amtool silence query

# Remove silence
amtool silence expire <silence-id>
```

### 9. Test Alert Resolution

**Send and resolve test alert:**
```bash
# Send alert
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestResolutionAlert",
      "severity": "warning"
    },
    "annotations": {
      "summary": "Testing alert resolution"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'

# Wait 30 seconds
sleep 30

# Resolve alert by sending with endsAt
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestResolutionAlert",
      "severity": "warning"
    },
    "annotations": {
      "summary": "Testing alert resolution"
    },
    "endsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'
```

**Verify:**
- [ ] Initial alert notification sent
- [ ] Resolution notification sent (if send_resolved: true)
- [ ] Alert removed from active alerts in Alertmanager
- [ ] Resolution appears in Slack with green indicator

### 10. End-to-End Integration Test

Complete test of entire alert pipeline.

**Test scenario: Database connection pool exhaustion**

```bash
# 1. Trigger condition (requires access to database)
# Simulate by creating many connections or adjusting PgBouncer limits

# 2. Wait for alert to fire (5 minutes)
watch 'curl -s http://localhost:9090/api/v1/alerts | jq ".data.alerts[] | select(.labels.alertname==\"PgBouncerPoolExhaustion\")"'

# 3. Verify alert routing
# - Check PagerDuty
# - Check Slack #incidents
# - Check Grafana dashboard

# 4. Verify runbook link works
# Click runbook URL from alert

# 5. Follow runbook to mitigate
# (Simulation only in staging)

# 6. Verify alert resolves
# Restore normal conditions and wait for resolution
```

**Verify end-to-end:**
- [ ] Alert fires when threshold exceeded
- [ ] Alert routes to correct channels
- [ ] Runbook link is accessible and helpful
- [ ] Alert resolves when condition clears
- [ ] Resolution notification sent
- [ ] No false positives or missed alerts

## Test Checklist

Use this checklist when testing alerts in staging:

### Pre-Deployment Validation
- [ ] Alert rules syntax validated with promtool
- [ ] All required alerts present (run coverage check)
- [ ] Runbook links verified and accessible
- [ ] Alert thresholds reviewed and reasonable

### Routing Tests
- [ ] Critical alerts route to PagerDuty + Slack #incidents
- [ ] Warning alerts route to Slack #alerts only
- [ ] Info alerts route to Slack #monitoring only
- [ ] Security alerts route to PagerDuty Security + Slack #security
- [ ] SLO alerts route to PagerDuty SLO + Slack #incidents

### Alert Behavior Tests
- [ ] Inhibition rules work (critical suppresses warnings)
- [ ] Grouping works (related alerts grouped)
- [ ] Silencing works correctly
- [ ] Alert resolution sends notification
- [ ] Repeat interval respected

### Integration Tests
- [ ] At least one real alert triggered and verified
- [ ] PagerDuty incident created correctly
- [ ] Slack notifications formatted properly
- [ ] Grafana dashboard links work
- [ ] Runbooks accessible from alerts

### Documentation
- [ ] All alerts have runbook references
- [ ] Runbooks are up-to-date and accurate
- [ ] On-call procedures documented
- [ ] Escalation paths clear

## Troubleshooting Test Failures

### Alert Not Firing

**Symptoms:** Alert doesn't fire when condition met

**Check:**
1. Verify Prometheus is scraping metrics: `http://localhost:9090/targets`
2. Check alert rule expression in Prometheus UI: `http://localhost:9090/alerts`
3. Verify threshold is actually exceeded (query metrics directly)
4. Check `for` duration - alert may still be pending
5. Review Prometheus logs for errors

**Fix:**
- Adjust threshold if too aggressive
- Fix PromQL expression if incorrect
- Ensure metrics are being exported
- Check for typos in alert name or labels

### Alert Not Routing

**Symptoms:** Alert fires but doesn't reach Slack/PagerDuty

**Check:**
1. Verify alert appears in Alertmanager: `http://localhost:9093`
2. Check Alertmanager routing configuration
3. Verify webhook URLs are correct
4. Check Alertmanager logs for delivery errors
5. Test webhook URLs directly with curl

**Fix:**
- Update webhook URLs in alertmanager.yml
- Check network connectivity to external services
- Verify secrets/API keys are correct
- Review routing rules and matchers

### Alert Routing to Wrong Channel

**Symptoms:** Alert goes to incorrect Slack channel or PagerDuty service

**Check:**
1. Review alert labels (especially severity)
2. Check routing rules in alertmanager.yml
3. Verify matchers are correct (regex vs exact)
4. Check for conflicting routing rules

**Fix:**
- Ensure alert has correct labels
- Adjust routing rules if needed
- Use more specific matchers
- Test with `amtool config routes test`

### Inhibition Not Working

**Symptoms:** Lower-priority alerts not suppressed

**Check:**
1. Verify both alerts are firing
2. Check inhibition rules in alertmanager.yml
3. Ensure label matchers are correct
4. Verify inhibition rule equality constraints

**Fix:**
- Adjust inhibition rule matchers
- Ensure alerts have required labels
- Check label values match exactly
- Review inhibition rules order

### Too Many Notifications

**Symptoms:** Getting spammed with alert notifications

**Check:**
1. Review repeat_interval settings
2. Check if alerts are flapping (firing/resolving quickly)
3. Verify grouping is working correctly
4. Check for duplicate alert rules

**Fix:**
- Increase repeat_interval
- Add hysteresis to alert rules (increase `for` duration)
- Improve grouping configuration
- Remove duplicate alerts
- Adjust thresholds to reduce noise

## Metrics for Alert Quality

Track these metrics to assess alert quality:

### Alert Volume
- Total alerts fired per day/week
- Alerts by severity (P1, P2, P3)
- Alerts by service/component
- Trend over time

**Target:** Decreasing over time as issues are fixed

### Alert Accuracy
- True positives vs false positives
- Actionable vs noise
- Alerts with clear root cause

**Target:** >90% true positive rate

### Response Metrics
- Mean time to acknowledge (MTTA)
- Mean time to resolve (MTTR)
- Escalation frequency

**Target:** MTTA <15min for P1, MTTR <1h for P1

### Coverage
- Services without alerts
- Critical components without monitoring
- Gaps in alert rules

**Target:** 100% coverage of critical services

## Continuous Improvement

### After Each Staging Test

1. **Document findings** in test results file
2. **Update runbooks** with new learnings
3. **Adjust thresholds** if needed
4. **Fix false positives** immediately
5. **Share results** with team

### Monthly Review

- Review alert metrics and trends
- Identify noisy or redundant alerts
- Update alert rules based on learnings
- Refine thresholds based on data

### Quarterly Review

- Complete audit of all alert rules
- Review SLOs and alert thresholds
- Update escalation procedures
- Team training on new alerts

## Related Documentation

- [Alertmanager Setup Guide](../../monitoring/ALERTMANAGER_SETUP.md)
- [On-Call Rotation Guide](on-call-rotation.md)
- [On-Call Quick Reference](on-call-quick-reference.md)
- [SLO Documentation](slos.md)
- [Monitoring README](../../monitoring/README.md)

## Contact

Questions about alert testing:
- Slack: #platform-team
- Email: platform@clpr.app

---

**Document Version:** 1.0  
**Last Updated:** 2026-01-02  
**Owner:** Platform Engineering Team  
**Related Issues:** 
- [#860 - Alerting Configuration (Roadmap 5.0)](https://git.subcult.tv/subculture-collective/clpr/issues/860)
- [#858 - Grafana Dashboards](https://git.subcult.tv/subculture-collective/clpr/issues/858)
- [#805 - Observability Infrastructure](https://git.subcult.tv/subculture-collective/clpr/issues/805)
