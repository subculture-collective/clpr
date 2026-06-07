---
title: "On Call Rotation"
summary: "This document describes the on-call rotation procedures, responsibilities, and escalation policies for the Clipper platform."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# On-Call Rotation Guide

This document describes the on-call rotation procedures, responsibilities, and escalation policies for the Clipper platform.

## Overview

The on-call rotation ensures 24/7 coverage for critical incidents and system failures. Engineers on call are responsible for responding to alerts, mitigating incidents, and coordinating with the team for resolution.

## On-Call Schedule

### Rotation Period

- **Duration:** 1 week (Monday 9:00 AM to following Monday 9:00 AM)
- **Handoff:** Monday morning standups
- **Timezone:** All times in UTC unless specified

### Schedule Template

| Week | Primary | Secondary | Manager |
|------|---------|-----------|---------|
| Week 1 (Jan 1-7) | Engineer A | Engineer B | Manager X |
| Week 2 (Jan 8-14) | Engineer B | Engineer C | Manager X |
| Week 3 (Jan 15-21) | Engineer C | Engineer D | Manager Y |
| Week 4 (Jan 22-28) | Engineer D | Engineer A | Manager Y |

**Note:** Update this table with actual names and maintain in PagerDuty schedule.

### Role Definitions

**Primary On-Call:**
- First responder to all alerts
- Acknowledges alerts within response time SLA
- Leads incident response
- Updates incident channels
- Files incident reports

**Secondary On-Call (Backup):**
- Covers if primary is unavailable
- Auto-escalated after 15 minutes (P1) or 1 hour (P2)
- Assists with complex incidents
- Takes over if primary needs break

**On-Call Manager:**
- Escalation point for critical issues
- Auto-escalated after 30 minutes
- Provides technical leadership
- Makes rollback/architecture decisions
- Coordinates with product/business

## Response Times and SLAs

### Alert Severity Levels

| Severity | Response Time | Acknowledgment | Resolution Target | Escalation |
|----------|--------------|----------------|-------------------|------------|
| **P1 (Critical)** | < 15 minutes | Required | 1 hour | Auto after 15 min |
| **P2 (Warning)** | < 1 hour | Required | 4 hours | Auto after 1 hour |
| **P3 (Info)** | < 4 hours | Optional | 24 hours | No auto-escalation |

### Escalation Path

```
Level 1: Primary On-Call Engineer (0-15 min)
         ↓ (if no ack or needs help)
Level 2: Secondary On-Call Engineer (15-30 min)
         ↓ (if not resolved)
Level 3: On-Call Manager (30-60 min)
         ↓ (if critical/widespread)
Level 4: VP Engineering + CTO (1+ hour)
```

## On-Call Responsibilities

### Before Your Shift

**Preparation (1-2 days before):**
- [ ] Verify you're in the PagerDuty schedule
- [ ] Test PagerDuty notifications (SMS, push, phone)
- [ ] Update emergency contact information
- [ ] Review recent incidents and post-mortems
- [ ] Read this guide and runbooks
- [ ] Review SLO targets and current status
- [ ] Check monitoring dashboards
- [ ] Ensure you have access to:
  - PagerDuty
  - Slack (#incidents, #alerts, #security)
  - Grafana dashboards
  - Kubernetes clusters (if applicable)
  - Database access (read-only minimum)
  - AWS/Cloud console (if applicable)

**Equipment Checklist:**
- [ ] Laptop charged and accessible
- [ ] Mobile phone charged
- [ ] PagerDuty app installed and working
- [ ] VPN configured (if required)
- [ ] Backup power bank
- [ ] Reliable internet connection

### During Your Shift

**Daily Activities:**
- Monitor #alerts channel for trends
- Review error budgets and SLO compliance
- Check for any degraded services
- Respond to alerts per SLA
- Document all incidents
- Update runbooks with learnings

**Communication Requirements:**
- Post status updates every 15 minutes for P1 incidents
- Keep #incidents channel updated
- Notify manager for critical issues
- Coordinate with team for complex problems
- Clear communication with stakeholders

**Availability:**
- Be reachable 24/7 during your rotation
- Respond to pages within SLA time
- Have reliable internet access
- Keep phone volume on and nearby
- Notify secondary if unavailable (planned break)

### After Your Shift

**Handoff (Monday standup):**
- [ ] Review any open incidents
- [ ] Share key learnings from the week
- [ ] Transfer any in-progress work
- [ ] Document any system changes
- [ ] Update runbooks if needed
- [ ] File incident reports for major issues

**Post-Rotation:**
- [ ] Complete any pending post-mortems
- [ ] Update documentation based on incidents
- [ ] Provide feedback on alerting/runbooks
- [ ] Share improvement suggestions

## Incident Response Workflow

### 1. Alert Received

```
RECEIVE ALERT
    ↓
ACKNOWLEDGE in PagerDuty (< 15 min for P1)
    ↓
POST in #incidents: "I'm responding to [Alert Name]"
    ↓
ASSESS IMPACT: Check dashboards, metrics, logs
    ↓
FOLLOW RUNBOOK for specific alert type
    ↓
MITIGATE or ESCALATE
    ↓
RESOLVE and DOCUMENT
```

### 2. Assessment Questions

When you receive an alert, answer these questions:
- What is alerting? (Service, component, metric)
- What is the severity? (P1, P2, P3)
- What is the impact? (Users affected, functionality)
- Is this a known issue? (Check recent incidents)
- What changed recently? (Deployments, config)

### 3. Initial Response Template

Post immediately in #incidents:
```
🚨 Incident: [Alert Name]
Status: Investigating
Started: [Time]
Severity: [P1/P2/P3]
Impact: [Brief description - e.g., "API endpoints returning 5xx"]
Responder: @[your-name]
Dashboard: [Link to relevant Grafana dashboard]

Updates every 15 minutes
```

### 4. Investigation Checklist

- [ ] Check SLO dashboard for current status
- [ ] Review error logs in Loki/Grafana
- [ ] Check recent deployments/changes
- [ ] Verify service health (pods, containers)
- [ ] Check database and Redis status
- [ ] Review metrics for anomalies
- [ ] Follow alert-specific runbook

### 5. Common Mitigation Actions

**Service Issues:**
```bash
# Restart service
kubectl rollout restart deployment/backend -n clpr

# Scale up
kubectl scale deployment backend --replicas=5 -n clpr

# Rollback deployment
kubectl rollout undo deployment/backend -n clpr
```

**Database Issues:**
```bash
# Check connections
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check slow queries
psql -c "SELECT query, now() - query_start as duration FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC LIMIT 10;"
```

**Cache Issues:**
```bash
# Check Redis
redis-cli PING
redis-cli INFO stats

# Clear cache (use with caution)
redis-cli FLUSHDB
```

### 6. Status Updates

Post updates every **15 minutes for P1**, **30 minutes for P2**:
```
📊 Update [Time]
Status: [Investigating/Mitigating/Resolving/Resolved]
Actions taken: [List what you've done]
Next steps: [What you're doing next]
Current impact: [Updated impact status]
ETA: [If known]
```

### 7. Resolution and Close

When resolved:
```
✅ Resolved [Time]
Duration: [HH:MM]
Root cause: [Brief description]
Resolution: [What fixed it]
Follow-up: Post-mortem scheduled for [Date/Time]

Service metrics back to normal. Monitoring for 30 minutes before closing.
```

## Alerting Channels and Tools

### PagerDuty

**Services:**
- **Clipper Critical:** General P1 alerts (ServiceDown, DatabaseDown, etc.)
- **Clipper SLO:** SLO breach alerts
- **Clipper Security:** Security events and authentication failures

**Access:** https://clpr.pagerduty.com

### Slack Channels

- **#incidents:** Critical incidents (P1), active coordination
- **#alerts:** Warning alerts (P2), monitoring
- **#monitoring:** Informational alerts (P3), metrics
- **#security:** Security events, authentication failures

### Monitoring Dashboards

- **SLO Dashboard:** http://localhost:3000/d/slo-dashboard
- **Application Overview:** http://localhost:3000/d/app-overview
- **System Health:** http://localhost:3000/d/system-health
- **Prometheus Alerts:** http://localhost:9090/alerts
- **Alertmanager:** http://localhost:9093

**Production URLs:** Replace localhost with production Grafana URL.

### Runbooks and Documentation

- **SLO Breach Response:** [docs/operations/playbooks/slo-breach-response.md](playbooks/slo-breach-response.md)
- **Search Incidents:** [docs/operations/playbooks/search-incidents.md](playbooks/search-incidents.md)
- **Background Jobs:** [docs/operations/runbooks/background-jobs.md](runbooks/background-jobs.md)
- **HPA Scaling:** [docs/operations/runbooks/hpa-scaling.md](runbooks/hpa-scaling.md)
- **CDN Failover:** [docs/operations/CDN_FAILOVER_RUNBOOK.md](CDN_FAILOVER_RUNBOOK.md)
- **Quick Reference:** [docs/operations/on-call-quick-reference.md](on-call-quick-reference.md)

## Emergency Contacts

### Primary Contacts

```
Primary On-Call: [Check PagerDuty schedule]
Secondary On-Call: [Check PagerDuty schedule]
On-Call Manager: [Check PagerDuty schedule]
```

### Escalation Contacts

```
VP Engineering: [Name] - [Phone] - [Email]
CTO: [Name] - [Phone] - [Email]
Security Team: security@clpr.app
Platform Team: platform@clpr.app
```

**Note:** Keep these contacts up-to-date in PagerDuty.

## Alert Types and Response

### Critical Alerts (P1)

**SLOAvailabilityBreach:**
- **Meaning:** < 99.5% availability
- **Response:** Check service health, rollback if needed
- **Runbook:** [Availability SLO Breach](playbooks/slo-breach-response.md#availability-slo-breach)

**SLOErrorRateBreach:**
- **Meaning:** > 0.5% error rate
- **Response:** Check error logs, database connectivity
- **Runbook:** [Error Rate SLO Breach](playbooks/slo-breach-response.md#error-rate-slo-breach)

**ServiceDown:**
- **Meaning:** Service not responding
- **Response:** Check pod status, restart or rollback
- **Runbook:** [Service Down](runbook.md#service-down)

**DatabaseDown:**
- **Meaning:** Database not responding
- **Response:** Check DB status, connections, restart if needed
- **Runbook:** [Database Down](runbook.md#database-down)

**RedisDown:**
- **Meaning:** Redis cache not responding
- **Response:** Check Redis status, restart if needed
- **Runbook:** [Redis Down](runbook.md#redis-down)

### Warning Alerts (P2)

**HighErrorRate:**
- **Meaning:** Elevated error rate (not yet SLO breach)
- **Response:** Monitor, investigate error patterns
- **Runbook:** [High Error Rate](runbook.md#high-error-rate)

**HighResponseTime:**
- **Meaning:** Elevated latency
- **Response:** Check slow queries, optimize or scale
- **Runbook:** [High Latency](runbook.md#high-latency)

**HighMemoryUsage:**
- **Meaning:** Memory usage > 80%
- **Response:** Check for memory leaks, scale if needed
- **Runbook:** [High Memory](runbook.md#high-memory-usage)

**LowDiskSpace:**
- **Meaning:** Disk space < 20%
- **Response:** Clean up logs, expand volume if needed
- **Runbook:** [Low Disk Space](runbook.md#low-disk-space)

### Security Alerts

**FailedAuthenticationSpike:**
- **Meaning:** Unusual failed login attempts
- **Response:** Check for brute force, block IPs
- **Runbook:** [Security Events](playbooks/slo-breach-response.md#security-alerts)

**SQLInjectionAttempt:**
- **Meaning:** Possible SQL injection detected
- **Response:** Block IPs, verify input validation
- **Runbook:** [Security Events](playbooks/slo-breach-response.md#security-alerts)

**SuspiciousSecurityEvent:**
- **Meaning:** Security-related warnings detected
- **Response:** Investigate immediately, notify security team
- **Runbook:** [Security Events](playbooks/slo-breach-response.md#security-alerts)

## Silencing Alerts

### When to Silence

**Valid reasons:**
- Scheduled maintenance
- Known issue being actively worked
- Deployment in progress
- Planned downtime

**Invalid reasons:**
- Alert is annoying
- Too many alerts
- Don't know how to fix

### How to Silence

**Via Alertmanager UI:**
1. Go to Alertmanager: http://localhost:9093
2. Click "Silences" tab
3. Click "New Silence"
4. Set matchers (alertname, service, etc.)
5. Set duration (be conservative)
6. Add comment explaining why
7. Add your name/email as creator

**Via CLI:**
```bash
# Silence specific service (RECOMMENDED)
amtool silence add service=backend \
  --duration=2h \
  --comment="Backend deployment in progress" \
  --author="ops@clpr.app"

# Silence specific alert
amtool silence add alertname=HighMemoryUsage \
  --duration=1h \
  --comment="Investigating memory leak" \
  --author="you@clpr.app"

# List active silences
amtool silence query

# Remove silence
amtool silence expire <silence-id>
```

**⚠️ Important:**
- Always add a comment explaining why
- Use shortest duration possible
- Never silence security alerts without manager approval
- Remove silence when issue is resolved

### Silence Guidelines

| Alert Type | Max Silence Duration | Approval Required |
|------------|---------------------|-------------------|
| Security | Not recommended | Manager + Security team |
| P1 Critical | 1 hour | Manager |
| P2 Warning | 4 hours | No |
| P3 Info | 24 hours | No |
| Maintenance | Duration of maintenance | Manager for >4h |

## Best Practices

### Do's ✅

- **Acknowledge alerts promptly** (within SLA)
- **Post status updates** regularly
- **Follow runbooks** - they're tested
- **Ask for help** when stuck - escalate!
- **Document everything** - helps with post-mortems
- **Keep team informed** - communication is key
- **Test before your shift** - verify PagerDuty works
- **Learn from incidents** - update runbooks
- **Stay calm** - panic doesn't help

### Don'ts ❌

- **Don't ignore alerts** - even if they seem minor
- **Don't silence without reason** - especially security
- **Don't make changes without understanding** - ask first
- **Don't skip post-mortems** - we learn from failures
- **Don't work alone on P1s** - get help early
- **Don't forget to eat/sleep** - escalate if tired
- **Don't hesitate to rollback** - safety first
- **Don't leave incidents undocumented** - always file reports

## On-Call Compensation

### Time Off

- 1 day off for every week on call (taken within 30 days)
- Additional time off for major incidents (>4h continuous work)
- Flex time for off-hours responses

### Financial

- On-call stipend: [Amount] per week
- Incident response bonus for major incidents
- Meal/transportation reimbursement during incidents

**Note:** Check with HR for current compensation policy.

## Health and Wellness

### Taking Breaks

- Short breaks (1-2h): Notify secondary, keep phone accessible
- Longer breaks: Swap with secondary, update PagerDuty
- Personal emergencies: Contact manager immediately

### Avoiding Burnout

- Use your time off after rotation
- Speak up if rotation is too frequent
- Share feedback about alert fatigue
- Escalate if overwhelmed - it's not a weakness
- Maintain work-life balance

### Support Resources

- Slack: #on-call-support
- Manager: Schedule 1:1 to discuss concerns
- HR: Mental health resources available
- Team: We're here to help!

## Training and Resources

### Required Reading

- [ ] This document (on-call-rotation.md)
- [ ] [On-Call Quick Reference](on-call-quick-reference.md)
- [ ] [SLO Documentation](slos.md)
- [ ] [SLO Breach Response Playbook](playbooks/slo-breach-response.md)
- [ ] [Operations Runbook](runbook.md)

### Recommended Reading

- [ ] [Alertmanager Setup Guide](../../monitoring/ALERTMANAGER_SETUP.md)
- [ ] [Centralized Logging](centralized-logging.md)
- [ ] [Kubernetes Runbook](kubernetes-runbook.md)
- [ ] [Background Jobs Runbook](runbooks/background-jobs.md)

### Training Sessions

- New engineer on-call training (before first shift)
- Shadow an experienced on-call engineer (recommended)
- Incident response simulation (quarterly)
- Runbook review sessions (monthly)

## Incident Post-Mortems

### Required for:

- All P1 incidents (no matter how brief)
- P2 incidents lasting > 2 hours
- Any incident affecting users
- Security incidents
- SLO breaches

### Timeline

- **Within 24h:** Initial incident report filed
- **Within 48h:** Draft post-mortem shared
- **Within 7 days:** Post-mortem meeting scheduled
- **Within 14 days:** Action items assigned and tracked

### Post-Mortem Template

Use the template in `docs/operations/incident-post-mortem-template.md` (create if missing).

**Required sections:**
- Timeline of events
- Root cause analysis
- Impact assessment
- What went well / What didn't
- Action items with owners

## Continuous Improvement

### Feedback Loop

- Weekly: Share learnings in team sync
- Monthly: Review alert metrics and thresholds
- Quarterly: Update runbooks and documentation
- Annually: Review on-call process and compensation

### Metrics to Track

- Mean time to acknowledge (MTTA)
- Mean time to resolve (MTTR)
- Alert frequency by severity
- False positive rate
- Escalation frequency
- On-call satisfaction scores

### Making Changes

- **Runbooks:** Update immediately after incidents
- **Alerts:** Create ticket for threshold changes
- **Process:** Discuss in team meetings
- **Documentation:** PRs always welcome

## FAQs

**Q: What if I miss an alert?**
A: It will auto-escalate to secondary. Acknowledge ASAP and post in #incidents explaining the delay.

**Q: Should I page someone at 3 AM?**
A: For P1 incidents, yes. That's what we're here for. For P2, assess if it can wait until morning.

**Q: What if I'm not sure what to do?**
A: Follow the runbook. If still unsure, escalate to secondary or manager. Never guess on P1 incidents.

**Q: Can I silence an alert that's noisy?**
A: Temporarily (< 1 hour) while investigating. But file a ticket to fix the root cause or adjust the threshold.

**Q: What if I need to cancel my on-call shift?**
A: Notify manager at least 48 hours in advance. Find someone to swap with if possible.

**Q: Should I deploy a fix during my shift?**
A: For hotfixes to resolve P1, yes (with manager approval). For non-urgent fixes, wait for business hours.

**Q: How do I handle alerts during a deployment?**
A: Consider silencing expected alerts. But watch dashboards closely and rollback if issues arise.

## Contact and Support

**Questions about this guide:**
- Slack: #platform-team
- Email: platform@clpr.app

**On-call support:**
- Slack: #on-call-support
- PagerDuty: Escalate per policy
- Manager: Direct message in Slack or call

**Updates to this document:**
- File PR with changes
- Discuss in team meeting
- Owner: Platform Engineering Team

---

**Document Version:** 1.0  
**Last Updated:** 2026-01-02  
**Owner:** Platform Engineering Team  
**Related Issues:** #860 (Roadmap 5.0 Phase 5.3)

## Related Documentation

- [On-Call Quick Reference Card](on-call-quick-reference.md) - Print and keep handy
- [Alertmanager Setup Guide](../../monitoring/ALERTMANAGER_SETUP.md)
- [SLO Documentation](slos.md)
- [Runbooks](runbooks/) and [Playbooks](playbooks/)
- [Monitoring README](../../monitoring/README.md)
