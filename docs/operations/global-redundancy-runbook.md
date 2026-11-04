---
title: "Global Redundancy Runbook"
summary: "This runbook provides operational procedures for managing global redundancy, failover scenarios, and multi-region infrastructure."
tags: ["operations","runbook"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Global Redundancy & Failover Runbook

This runbook provides operational procedures for managing global redundancy, failover scenarios, and multi-region infrastructure.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Monitoring](#monitoring)
4. [Common Scenarios](#common-scenarios)
5. [Incident Response](#incident-response)
6. [Maintenance Procedures](#maintenance-procedures)
7. [Troubleshooting](#troubleshooting)

## Overview

### Components

- **Primary Database**: PostgreSQL (read/write)
- **Read Replicas**: Regional PostgreSQL instances (read-only)
- **Health Checks**: Regional health monitoring

### Regions

- **us-east-1**: Primary (US East Coast)
- **eu-west-1**: Secondary (Europe)
- **ap-southeast-1**: Tertiary (Asia-Pacific)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Global Load Balancer                        │
│          (Geo-routing + Health-based failover)               │
└────────────┬──────────────────────────────┬─────────────────┘
             │                              │
    ┌────────▼────────┐          ┌─────────▼────────┐
    │   US-EAST-1     │          │    EU-WEST-1     │
    │   (Primary)     │          │   (Secondary)    │
    │                 │          │                  │
    │ - API Server    │          │ - API Server     │
    │ - Primary DB    │◄────────►│ - Read Replica   │
    └─────────────────┘          └──────────────────┘
```

## Monitoring

### Health Check Endpoints

```bash
# Check primary region health
curl https://api.clpr.gg/health

# Check specific region
curl https://us-east-1.api.clpr.gg/health
curl https://eu-west-1.api.clpr.gg/health
curl https://ap-southeast-1.api.clpr.gg/health
```

### Key Metrics

1. **Database Replication Lag**
   ```promql
   # Alert if lag > 5 seconds
   pg_replication_lag_seconds > 5
   ```

2. **Region Health**
   ```promql
   # Alert if region unhealthy
   region_health_status{status="unhealthy"} == 1
   ```

### Dashboards

- **Global Overview**: [Grafana Dashboard](http://grafana.clpr.gg/d/global-overview)
- **Region Health**: [Grafana Dashboard](http://grafana.clpr.gg/d/region-health)

## Common Scenarios

### Scenario 1: Primary Region Degraded

**Symptoms:**
- Increased latency in us-east-1
- Health checks showing degraded status
- No complete outage

**Actions:**
1. Verify the issue via monitoring
2. Check if automatic failover is working
3. Monitor user impact
4. Investigate root cause

**Commands:**
```bash
# Check region health
curl https://api.clpr.gg/v1/health/regions

# Check replication status
psql -h replica.eu-west-1 -c "SELECT * FROM pg_stat_replication;"

```

### Scenario 2: Primary Database Failure

**Symptoms:**
- Database connection errors
- Write operations failing
- Read replicas still working

**Actions:**
1. **Immediate (0-5 minutes)**
   - Verify primary is down
   - Check automatic failover triggered
   - Monitor error rates

2. **Short-term (5-15 minutes)**
   - If auto-failover didn't trigger, initiate manual failover
   - Promote read replica to primary
   - Update connection strings

3. **Recovery (15+ minutes)**
   - Investigate cause of failure
   - Restore original primary
   - Re-establish replication

**Failover Commands:**
```bash
# Promote replica to primary (PostgreSQL)
pg_ctl promote -D /var/lib/postgresql/data

# Update application config
export DB_HOST=eu-west-1.db.clpr.gg
systemctl restart clpr-api

# Verify new primary
psql -h eu-west-1.db.clpr.gg -c "SELECT pg_is_in_recovery();"
# Should return: f (false = not in recovery = primary)
```

## Incident Response

### Severity Levels

**P0 - Critical**
- Complete service outage
- Data loss risk
- Security breach

**P1 - High**
- Major functionality broken
- Single region failure
- Performance degradation >50%

**P2 - Medium**
- Minor functionality issues
- Performance degradation <50%
- Non-critical errors

**P3 - Low**
- Cosmetic issues
- Minor bugs
- Enhancement requests

### Response Procedures

#### P0 - Critical Incident

1. **Alert** (0-2 minutes)
   - PagerDuty alerts on-call engineer
   - Create incident channel in Slack
   - Page backup if no response in 5 min

2. **Assess** (2-5 minutes)
   - Determine scope and impact
   - Identify affected components
   - Estimate users impacted

3. **Communicate** (5-10 minutes)
   - Update status page
   - Notify stakeholders
   - Post in incident channel

4. **Mitigate** (10-30 minutes)
   - Execute failover if needed
   - Roll back if recent deploy
   - Implement workaround

5. **Resolve** (30+ minutes)
   - Fix root cause
   - Verify resolution
   - Monitor for stability

6. **Post-Mortem** (24-48 hours)
   - Write incident report
   - Identify action items
   - Schedule follow-up

#### P1 - High Severity

Similar to P0 but with relaxed timelines:
- Alert: 0-5 minutes
- Assess: 5-10 minutes
- Communicate: 10-15 minutes
- Mitigate: 15-60 minutes

### Communication Templates

**Initial Alert:**
```
🚨 INCIDENT: [Brief Description]
Severity: P0/P1/P2
Status: Investigating
Impact: [User impact description]
ETA: [Estimated resolution time]
Incident Lead: @engineer
```

**Update:**
```
📊 UPDATE: [Progress description]
Status: Mitigating/Monitoring
Actions Taken: [List actions]
Next Steps: [What's next]
ETA: [Updated estimate]
```

**Resolution:**
```
✅ RESOLVED: [Description]
Duration: [Total time]
Root Cause: [Brief explanation]
Follow-up: [Link to post-mortem]
```

## Maintenance Procedures

### Database Maintenance

#### Read Replica Setup

```bash
# 1. Create base backup from primary
pg_basebackup -h primary.db.clpr.gg -D /var/lib/postgresql/replica -U replication -v -P

# 2. Configure recovery.conf
cat > /var/lib/postgresql/replica/recovery.conf <<EOF
standby_mode = 'on'
primary_conninfo = 'host=primary.db.clpr.gg port=5432 user=replication'
trigger_file = '/tmp/postgresql.trigger'
EOF

# 3. Start replica
pg_ctl -D /var/lib/postgresql/replica start

# 4. Verify replication
psql -h replica -c "SELECT pg_is_in_recovery();"
psql -h primary -c "SELECT * FROM pg_stat_replication;"
```

#### Failover Test

```bash
# Schedule monthly failover tests

# 1. Announce maintenance window
# 2. Promote replica
pg_ctl promote -D /var/lib/postgresql/replica

# 3. Update application
export DB_HOST=replica.db.clpr.gg
systemctl restart clpr-api

# 4. Monitor for issues (30 minutes)

# 5. Fail back to primary
# ... reverse the process
```

## Troubleshooting

### Database Replication Issues

**Problem:** Replication lag increasing

```bash
# Check lag on replica
psql -h replica -c "SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) AS lag_seconds;"

# Check WAL sender status
psql -h primary -c "SELECT * FROM pg_stat_replication;"

# Common causes:
# - High write volume on primary
# - Network issues between primary and replica
# - Replica under-resourced

# Solutions:
# - Increase replica resources
# - Check network connectivity
# - Optimize queries on primary
```

**Problem:** Replication stopped

```bash
# Check replication slots
psql -h primary -c "SELECT * FROM pg_replication_slots;"

# If slot inactive, recreate
psql -h replica -c "SELECT pg_drop_replication_slot('replica_slot');"
psql -h primary -c "SELECT pg_create_physical_replication_slot('replica_slot');"

# Restart replica
pg_ctl -D /var/lib/postgresql/replica restart
```

## Contacts

### On-Call Schedule

- **Primary**: Check PagerDuty
- **Backup**: Check PagerDuty
- **Manager**: [Email/Phone]

### Escalation

1. On-call engineer (immediate)
2. Tech lead (15 min)
3. Engineering manager (30 min)
4. CTO (1 hour)

### External Contacts

- **Database Support**: [Contact info]
- **Infrastructure Provider**: [Contact info]

## Related Documentation

- Database Operations
- Deployment Guide
