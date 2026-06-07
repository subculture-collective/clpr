---
title: "README"
summary: "This directory contains incident response playbooks and operational procedures for managing Clipper in production."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Operations Playbooks Index

This directory contains incident response playbooks and operational procedures for managing Clipper in production.

## Quick Links

- 🚨 **[On-Call Quick Reference](../ON_CALL_QUICK_REFERENCE.md)** - Essential info for on-call engineers
- 📊 **[SLO Definitions](../slos.md)** - Service level objectives and targets
- 🔧 **[Operations Runbook](../runbook.md)** - Common tasks and procedures

## Incident Response Playbooks

### Primary Playbooks

#### [SLO Breach Response](./slo-breach-response.md) 🆕

**When to use:** Alert for SLO breach (availability, latency, error rate)

**Covers:**
- Availability SLO breach response
- Latency SLO breach response
- Error rate SLO breach response
- Search performance issues
- Webhook delivery problems
- Communication templates
- Post-incident procedures

**Response Time:** < 15 minutes for critical alerts

---

#### [Search Incidents](./search-incidents.md)

**When to use:** Search functionality degraded or failing

**Covers:**
- High search latency
- Embedding generation failures
- Low embedding coverage
- Cache issues
- Zero results troubleshooting
- Fallback issues
- Indexing failures
- Vector search performance
- BM25 search performance

**Response Time:** < 15 minutes for critical, < 1 hour for warnings

---

## Alert Severity Guide

| Severity | Response Time | Escalation | Playbook |
|----------|--------------|------------|----------|
| **Critical (P1)** | < 15 minutes | Yes, after 15 min | [SLO Breach Response](./slo-breach-response.md) |
| **Warning (P2)** | < 1 hour | Yes, after 1 hour | [SLO Breach Response](./slo-breach-response.md) |
| **Info (P3)** | < 4 hours | No | [Operations Runbook](../runbook.md) |

## Common Scenarios

### Service Degradation

- **High Error Rate** → [SLO Breach Response - Error Rate](./slo-breach-response.md#error-rate-slo-breach)
- **Slow Response Times** → [SLO Breach Response - Latency](./slo-breach-response.md#latency-slo-breach)
- **Service Unavailable** → [SLO Breach Response - Availability](./slo-breach-response.md#availability-slo-breach)

### Search Issues

- **Search is Slow** → [Search Incidents - High Latency](./search-incidents.md#high-latency)
- **No Search Results** → [Search Incidents - Zero Results](./search-incidents.md#zero-results)
- **Embedding Failures** → [Search Incidents - Embedding Failures](./search-incidents.md#embedding-failures)

### Infrastructure

- **Database Problems** → [Operations Runbook - Database](../runbook.md#database-operations)
- **Redis Issues** → [Operations Runbook - Cache](../runbook.md#cache-operations)
- **Deployment Issues** → [Operations Runbook - Deployments](../runbook.md#deployments)

### Security

- **Authentication Failures** → [SLO Breach Response - Security](./slo-breach-response.md)
- **SQL Injection Attempts** → [SLO Breach Response - Security](./slo-breach-response.md)
- **Suspicious Activity** → [SLO Breach Response - Security](./slo-breach-response.md)

## Monitoring & Alerting

### Dashboards

- **SLO Compliance:** <http://localhost:3000/d/slo-dashboard>
- **Application Overview:** <http://localhost:3000/d/app-overview>
- **API Performance:** <http://localhost:3000/d/api-performance>
- **System Health:** <http://localhost:3000/d/system-health>
- **Search Quality:** <http://localhost:3000/d/search-quality>
- **Webhook Monitoring:** <http://localhost:3000/d/webhook-monitoring>

### Alert Management

- **Prometheus Alerts:** <http://localhost:9090/alerts>
- **Alertmanager:** <http://localhost:9093>
- **PagerDuty:** <https://clpr.pagerduty.com>

### Configuration

- **Alert Rules:** `../../monitoring/alerts.yml`
- **Alert Routing:** `../../monitoring/alertmanager.yml`
- **Alertmanager Setup:** `../../monitoring/ALERTMANAGER_SETUP.md`

## Escalation Policy

```
Level 1: On-call Engineer (0-15 min)
   ↓ (no acknowledgment after 15 min)
Level 2: On-call Lead (15-30 min)
   ↓ (no resolution after 30 min)
Level 3: Engineering Manager (30-45 min)
   ↓ (ongoing incident after 1 hour)
Level 4: VP Engineering + CTO (1+ hour)
```

## Communication Channels

- **#incidents** - Critical alerts and active incident coordination
- **#alerts** - Warning level alerts and potential issues
- **#monitoring** - Informational alerts and metrics trends
- **#security** - Security events and authentication failures

## Quick Command Reference

### Service Health

```bash
# Kubernetes
kubectl get pods -n clpr
kubectl describe pod <pod-name> -n clpr
kubectl logs -f deployment/backend -n clpr

# Docker Compose
docker-compose ps
docker-compose logs -f backend
```

### Database

```bash
# Connection test
psql $POSTGRES_URL -c "SELECT 1"

# Active connections
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Slow queries
psql -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

### Cache

```bash
# Connection test
redis-cli PING

# Memory usage
redis-cli INFO memory

# Cache stats
redis-cli INFO stats
```

### Deployments

```bash
# Rollback
kubectl rollout undo deployment/backend -n clpr

# Scale
kubectl scale deployment backend --replicas=5 -n clpr

# Restart
kubectl rollout restart deployment/backend -n clpr
```

## Creating New Playbooks

When creating new playbooks, follow this structure:

1. **Overview** - What this playbook covers
2. **Detection** - How you know there's an issue
3. **Immediate Actions** - What to do right now (< 5 min)
4. **Investigation Steps** - How to diagnose the problem
5. **Mitigation Strategies** - How to fix common causes
6. **Recovery Verification** - How to confirm it's fixed
7. **Post-Incident** - What to do after resolution

Use existing playbooks as templates.

## Contributing

When updating playbooks:

1. Test procedures before documenting
2. Include actual commands (not placeholders)
3. Add examples and screenshots where helpful
4. Link to relevant documentation
5. Update this index with new playbooks
6. Get review from team members

## Post-Incident Review

After every incident:

1. **Update playbooks** with learnings
2. **Add new scenarios** if encountered new issue
3. **Improve commands** based on what worked
4. **Update timings** if estimates were wrong
5. **Add screenshots** of useful dashboards
6. **Link to incident reports** for context

## Resources

### Internal Documentation

- [SLO Documentation](../slos.md)
- [Monitoring Setup](../../monitoring/README.md)
- [Alert Configuration](../../monitoring/alerts.yml)
- [Deployment Guide](../deployment.md)
- [Security Procedures](../break-glass-procedures.md)

### External Resources

- [Google SRE Book](https://sre.google/sre-book/table-of-contents/)
- [Incident Response Guide](https://response.pagerduty.com/)
- [Kubernetes Debugging](https://kubernetes.io/docs/tasks/debug/)
- [PostgreSQL Performance](https://www.postgresql.org/docs/current/performance-tips.html)

---

**Last Updated:** 2025-12-21  
**Maintained By:** Platform Engineering Team

For questions about playbooks or to request new playbooks, contact the Platform Engineering team or open an issue.
