---
title: "Operational Runbooks"
summary: "Index of operational runbooks for production support"
tags: ["operations", "runbooks", "index"]
area: "operations"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["runbooks"]
---

# Operational Runbooks

Comprehensive operational runbooks for managing Clipper in production. These runbooks provide step-by-step procedures for common operational tasks, troubleshooting, and incident response.

## Quick Reference

### Emergency Contacts

- **On-Call Engineer**: PagerDuty alert (< 5 min response)
- **Operations Team**: ops@clpr.tv, Slack #ops-team
- **Security Team**: security@clpr.tv (< 30 min response)
- **Engineering Manager**: (555) 123-4567

### Critical Procedures

- [Emergency Ban User](./moderation-operations.md#emergency-ban-user)
- [Emergency Revoke Moderator](./moderation-operations.md#emergency-revoke-moderator)
- [Emergency System Shutdown](./moderation-rollback.md#complete-moderation-system-shutdown)
- [Rollback Deployment](./moderation-rollback.md#application-rollback)

---

## Moderation System Runbooks

Complete operational documentation for the moderation system.

### Core Operations

#### [Moderation Operations](./moderation-operations.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Primary operational runbook for day-to-day moderation tasks.

**Covers:**
- Emergency procedures (ban user, revoke moderator)
- Manual ban/unban operations
- Moderator management (add/remove)
- User ban status checks
- Daily health checks
- Weekly audit reviews

**When to use:**
- Emergency ban needed
- Moderator access issues
- Routine maintenance tasks
- Ban/unban user requests

---

#### [Audit Log Operations](./audit-log-operations.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Procedures for reviewing, exporting, and analyzing audit logs.

**Covers:**
- Viewing recent logs
- Filtering by action, user, time range
- Exporting to CSV/JSON
- Scheduled exports
- Common patterns (unauthorized access, suspicious activity)
- Investigation procedures
- Compliance and retention

**When to use:**
- Compliance audits
- Security investigations
- User action history requests
- Incident analysis
- GDPR data requests

---

#### [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Troubleshooting guide for Twitch ban synchronization issues.

**Covers:**
- OAuth scope issues
- Twitch API rate limits
- Invalid broadcaster ID
- Authentication failures
- Network timeouts
- Manual sync retry procedures
- Rate limit management

**When to use:**
- Ban sync fails
- Twitch integration errors
- Rate limit exceeded
- OAuth permission errors
- Sync performance issues

---

### Access Control & Security

#### [Permission Escalation](./permission-escalation.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Managing user permissions and emergency access grants.

**Covers:**
- Emergency admin access
- Temporary moderator access
- Grant/revoke permissions
- Permission troubleshooting
- Audit and compliance
- Permission matrix

**When to use:**
- Grant emergency access
- Permission denied errors
- Add/remove moderators
- Security audits
- Access reviews

---

### Recovery & Rollback

#### [Moderation Rollback](./moderation-rollback.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Emergency rollback procedures for moderation system issues.

**Covers:**
- Feature flag rollback
- Database rollback (PITR)
- Application rollback
- Emergency disable procedures
- Post-rollback verification

**When to use:**
- Critical bugs deployed
- Data corruption detected
- Security vulnerabilities
- Service unavailable
- Mass false positives

---

### Monitoring & Alerts

#### [Moderation Monitoring](./moderation-monitoring.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Monitoring setup, metrics, and alert configuration.

**Covers:**
- Key metrics to monitor
- Prometheus configuration
- Grafana dashboards
- Alert configuration (critical, warning, info)
- Alert response procedures
- Log monitoring

**When to use:**
- Setting up monitoring
- Configuring alerts
- Dashboard creation
- Performance analysis
- Capacity planning

---

### Incident Response

#### [Moderation Incidents](./moderation-incidents.md)
**Status**: ✅ Complete | **Reviewed**: 2026-02-03

Incident response and common issue resolution.

**Covers:**
- Incident response framework
- Security incidents (unauthorized access, compromised accounts)
- Common issues (user can't submit, moderator can't ban, sync failures)
- Contact procedures
- Post-incident reviews

**When to use:**
- Security incidents
- Service degradation
- User reports issues
- Escalation needed
- Post-mortem creation

---

## Other Infrastructure Runbooks

### Platform Operations

#### [Alert Validation](./alert-validation.md)
Validating alert configurations and testing alerting pipelines.

#### [Background Jobs](./background-jobs.md)
Troubleshooting background job processing and queue management.

#### [HPA Scaling](./hpa-scaling.md)
Horizontal Pod Autoscaling configuration and troubleshooting.

---

## Runbook Structure

All runbooks follow a consistent structure:

```markdown
# Title

## Overview
- Purpose and audience
- Prerequisites

## Table of Contents

## Main Content
- Step-by-step procedures
- Code examples (bash, SQL, API calls)
- Troubleshooting steps
- Verification procedures

## Related Runbooks
- Links to related documentation

## Emergency Contacts
```

---

## Using These Runbooks

### For On-Call Engineers

1. **Check monitoring dashboards** for current system status
2. **Identify the issue** using symptoms described in runbooks
3. **Follow the procedure** step-by-step
4. **Verify the fix** using verification steps
5. **Document** actions taken in incident ticket
6. **Escalate** if unresolved after following runbook

### For Operations Team

1. **Daily**: Review [daily health checks](./moderation-operations.md#daily-health-checks)
2. **Weekly**: Complete [audit reviews](./audit-log-operations.md#compliance-and-retention)
3. **Monthly**: Review runbooks for updates
4. **Quarterly**: Conduct runbook testing exercises

### For Support Team

1. **User issues**: Start with [Common Issues](./moderation-incidents.md#common-issues-and-solutions)
2. **Permission problems**: See [Permission Escalation](./permission-escalation.md#troubleshooting-permission-issues)
3. **Ban questions**: Use [Moderation Operations](./moderation-operations.md#check-user-ban-status)
4. **Escalation**: Follow [contact procedures](./moderation-incidents.md#contact-procedures)

---

## Runbook Maintenance

### Review Frequency

- **Quarterly**: Full review of all runbooks
- **After incidents**: Update based on lessons learned
- **After deployments**: Verify procedures still valid
- **On request**: Ad-hoc updates as needed

### Update Process

1. Identify outdated content or missing procedures
2. Create issue in repository
3. Update runbook with changes
4. Test procedures in staging
5. Review with team
6. Merge and publish
7. Notify team of changes

### Testing Runbooks

Regularly test runbooks to ensure accuracy:

- **Monthly**: Test emergency procedures in staging
- **Quarterly**: Full runbook walkthrough
- **Annually**: Disaster recovery drill

---

## Contributing

Found an issue or have a suggestion?

1. **Create an issue**: [GitHub Issues](https://git.subcult.tv/subculture-collective/clpr/issues)
2. **Submit a PR**: Update the runbook and submit for review
3. **Contact ops team**: ops@clpr.tv or Slack #ops-team

All contributions should:
- Follow existing runbook structure
- Include tested procedures
- Provide code examples
- Add verification steps

---

## Additional Resources

### Documentation

- [Operations Index](../index.md) - Operations hub
- [Moderation API Docs](../../backend/moderation-api.md) - API reference
- [Architecture](../../../ARCHITECTURE.md) - System architecture

### External Resources

- [Twitch API Documentation](https://dev.twitch.tv/docs/api/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)

### Training

- New hire onboarding includes runbook familiarization
- Monthly on-call training sessions
- Quarterly incident response drills

---

## Runbook Index

| Runbook | Status | Purpose | Last Reviewed |
|---------|--------|---------|---------------|
| [Moderation Operations](./moderation-operations.md) | ✅ Complete | Emergency procedures, manual operations | 2026-02-03 |
| [Audit Log Operations](./audit-log-operations.md) | ✅ Complete | Log review, export, compliance | 2026-02-03 |
| [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md) | ✅ Complete | Twitch sync issues | 2026-02-03 |
| [Permission Escalation](./permission-escalation.md) | ✅ Complete | Access control, permissions | 2026-02-03 |
| [Moderation Rollback](./moderation-rollback.md) | ✅ Complete | Rollback procedures | 2026-02-03 |
| [Moderation Monitoring](./moderation-monitoring.md) | ✅ Complete | Metrics, alerts, dashboards | 2026-02-03 |
| [Moderation Incidents](./moderation-incidents.md) | ✅ Complete | Incident response, common issues | 2026-02-03 |
| [Alert Validation](./alert-validation.md) | ✅ Complete | Alert testing | 2025-12-24 |
| [Background Jobs](./background-jobs.md) | ✅ Complete | Job troubleshooting | 2025-12-24 |
| [HPA Scaling](./hpa-scaling.md) | ✅ Complete | Autoscaling | 2025-12-24 |

---

**Document Owner**: Operations Team  
**Last Updated**: 2026-02-03  
**Review Frequency**: Quarterly  
**Contact**: ops@clpr.tv
