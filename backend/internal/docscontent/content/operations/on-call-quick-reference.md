---
title: "On Call Quick Reference"
summary: "> **Print this card and keep it handy during your on-call rotation**"
tags: ["operations","quick-reference"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# On-Call Quick Reference Card

> **Print this card and keep it handy during your on-call rotation**

## Emergency Contacts

- **On-Call Lead:** [Phone number]
- **Engineering Manager:** [Phone number]
- **VP Engineering:** [Phone number]
- **Security Team:** <security@clpr.app>

## Alert Response Times

| Severity | Response Time | Escalation |
|----------|--------------|------------|
| **P1 Critical** | < 15 minutes | Auto-escalate after 15 min |
| **P2 Warning** | < 1 hour | Auto-escalate after 1 hour |
| **P3 Info** | < 4 hours | No escalation |

## Quick Links

### Dashboards

- **SLO Dashboard:** <http://localhost:3000/d/slo-dashboard>
- **Prometheus Alerts:** <http://localhost:9090/alerts>
- **Alertmanager:** <http://localhost:9093>
- **Application Overview:** <http://localhost:3000/d/app-overview>

### Documentation

- **SLO Definitions:** `docs/operations/slos.md`
- **Response Playbook:** `docs/operations/playbooks/slo-breach-response.md`
- **Operations Runbook:** `docs/operations/runbook.md`

### Communication

- **Slack Incidents:** #incidents
- **Slack Alerts:** #alerts
- **PagerDuty:** <https://clpr.pagerduty.com>

## Common Alert Response Checklist

### When You Receive an Alert

- [ ] **Acknowledge** in PagerDuty (within 15 min for P1)
- [ ] **Post in #incidents:** "I'm responding to [Alert Name]"
- [ ] **Open dashboards:** Check SLO dashboard and metrics
- [ ] **Assess impact:** How many users? What functionality?
- [ ] **Follow runbook:** Use playbook for specific alert type

### Investigation Quick Commands

```bash
# Check service health
kubectl get pods -n clpr
docker-compose ps

# View recent logs
kubectl logs -l app=backend --tail=100 -n clpr
{service="clpr-backend", level="error"} [15m]  # Loki

# Check database
psql $POSTGRES_URL -c "SELECT 1"
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check Redis
redis-cli PING
redis-cli INFO stats

# Recent deployments
kubectl rollout history deployment/backend -n clpr
```

### Common Mitigation Actions

**Rollback deployment:**
```bash
kubectl rollout undo deployment/backend -n clpr
```

**Scale up:**
```bash
kubectl scale deployment backend --replicas=5 -n clpr
```

**Restart service:**
```bash
kubectl rollout restart deployment/backend -n clpr
```

**Clear cache:**
```bash
redis-cli FLUSHDB
```

## SLO Targets (Know These!)

| SLO | Target | Alert Threshold |
|-----|--------|----------------|
| **Availability** | 99.5% | < 99.5% for 5 min |
| **Error Rate** | < 0.5% | > 0.5% for 5 min |
| **Latency (List)** | P95 < 100ms | > 100ms for 5 min |
| **Latency (Detail)** | P95 < 50ms | > 50ms for 5 min |
| **Search** | P95 < 200ms | > 200ms for 5 min |
| **Webhooks** | > 90% success | < 90% for 10 min |

## Alert Type Quick Guide

### 🚨 SLOAvailabilityBreach

- **Meaning:** Less than 99.5% of requests succeeding
- **Check:** Service health, error logs, recent deployments
- **Actions:** Rollback, scale up, restart unhealthy pods
- **Runbook:** [Availability SLO Breach](../../docs/operations/playbooks/slo-breach-response.md#availability-slo-breach)

### 🚨 SLOErrorRateBreach

- **Meaning:** More than 0.5% of requests returning 5xx errors
- **Check:** Error logs, stack traces, database connectivity
- **Actions:** Rollback, fix database connections, hotfix
- **Runbook:** [Error Rate SLO Breach](../../docs/operations/playbooks/slo-breach-response.md#error-rate-slo-breach)

### 🚨 SLOLatencyBreach

- **Meaning:** P95 response time exceeds target
- **Check:** Slow endpoints, database queries, cache hit rate
- **Actions:** Add indexes, scale database, optimize queries
- **Runbook:** [Latency SLO Breach](../../docs/operations/playbooks/slo-breach-response.md#latency-slo-breach)

### 🚨 ServiceDown

- **Meaning:** Service not responding to health checks
- **Check:** Pod status, container logs, crashes
- **Actions:** Check logs for panic, restart pods, rollback
- **Runbook:** [Operations Runbook](../../docs/operations/runbook.md#high-error-rate)

### ⚠️ ErrorBudgetFastBurn

- **Meaning:** Consuming > 10% error budget in 1 hour
- **Check:** Recent changes, error patterns, traffic spikes
- **Actions:** Immediate mitigation to stop budget burn
- **Runbook:** [SLO Breach Response](../../docs/operations/playbooks/slo-breach-response.md)

### ⚠️ HighErrorRate

- **Meaning:** Error rate elevated but below SLO breach
- **Check:** Error logs, specific endpoints, error types
- **Actions:** Monitor, investigate, prepare mitigation
- **Runbook:** [High Error Rate](../../docs/operations/runbook.md#high-error-rate)

### ⚠️ HighResponseTime

- **Meaning:** Response times elevated
- **Check:** Slow queries, high load, resource usage
- **Actions:** Optimize queries, scale if needed
- **Runbook:** [High Latency](../../docs/operations/runbook.md#high-latency)

### 🛡️ Security Alerts

- **Meaning:** Suspicious activity detected
- **Check:** Security logs, IP addresses, attack patterns
- **Actions:** Block IPs, notify security team, verify input validation
- **Runbook:** [Security Incidents](../../docs/operations/playbooks/slo-breach-response.md)

## Communication Templates

### Initial Response (Post Immediately)

```
🚨 Incident: [Alert Name]
Status: Investigating
Started: [Time]
Impact: [Brief description]
Responder: @[your-name]

Dashboard: [Link]
Updates every 15 minutes
```

### Status Update (Every 15 min for P1)

```
📊 Update [Time]
Status: [Investigating/Mitigating/Resolving]
Actions: [What you've done]
Next: [What you're doing next]
Impact: [Current state]
```

### Resolution

```
✅ Resolved [Time]
Duration: [How long]
Root Cause: [Brief description]
Resolution: [What fixed it]
Next: Post-mortem scheduled
```

## Post-Incident Checklist

- [ ] Update error budget tracking
- [ ] File incident report (within 24 hours)
- [ ] Schedule post-mortem (within 2-3 days)
- [ ] Create action items
- [ ] Update runbooks with learnings

## Don't Panic

Remember:
- ✅ Follow the runbooks - they're tested
- ✅ Ask for help if stuck - escalate!
- ✅ Communicate often - keep team informed
- ✅ Document everything - helps with post-mortem
- ✅ Safety first - don't make it worse
- ✅ Learn from incidents - they make us better

## Pro Tips

1. **Keep this card visible** during your rotation
2. **Pre-load dashboards** in browser tabs before rotation starts
3. **Test PagerDuty** at start of rotation to verify it works
4. **Keep phone charged** and volume on
5. **Know your limits** - escalate if unsure
6. **Stay calm** - panic doesn't help
7. **Coffee is your friend** ☕

---

**Last Updated:** 2025-12-21  
**Version:** 1.0  
**Owner:** Platform Engineering Team

For questions or updates to this card, contact the Platform Engineering team.
