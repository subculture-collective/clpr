---
title: "Moderation Rollback Procedures"
summary: "Emergency rollback procedures for moderation system"
tags: ["operations", "runbook", "rollback", "emergency", "disaster-recovery"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["rollback", "disaster-recovery"]
---

# Moderation Rollback Procedures

## Overview

This runbook provides emergency rollback procedures for the moderation system when critical issues arise. Use these procedures to quickly disable or revert moderation features to maintain system stability.

**Audience**: Operations team, on-call engineers, incident commanders

**Prerequisites**:
- Admin or super-admin access
- Access to production environment
- Understanding of feature flag system
- Database backup access (for data rollback)

## Table of Contents

- [When to Rollback](#when-to-rollback)
- [Rollback Types](#rollback-types)
- [Feature Flag Rollback](#feature-flag-rollback)
  - [Disable Twitch Moderation](#disable-twitch-moderation)
  - [Disable Ban Sync](#disable-ban-sync)
  - [Disable Audit Logging](#disable-audit-logging)
- [Database Rollback](#database-rollback)
  - [Revert Recent Bans](#revert-recent-bans)
  - [Restore from Backup](#restore-from-backup)
- [Application Rollback](#application-rollback)
  - [Rollback Backend Deployment](#rollback-backend-deployment)
  - [Rollback Frontend Deployment](#rollback-frontend-deployment)
- [Emergency Disable Procedures](#emergency-disable-procedures)
- [Post-Rollback Verification](#post-rollback-verification)
- [Related Runbooks](#related-runbooks)

---

## When to Rollback

### Rollback Triggers

Initiate rollback if:

- [ ] **Critical bug** causing data corruption
- [ ] **Security vulnerability** actively being exploited
- [ ] **Service unavailability** (>5 minute outage)
- [ ] **Mass false positives** (incorrect bans/unbans)
- [ ] **Performance degradation** (>10x normal latency)
- [ ] **Database integrity issues** detected

### Decision Matrix

| Severity | Issue Type | Action | Rollback Type |
|----------|-----------|--------|---------------|
| **P0** | Data corruption, security breach | Immediate rollback | Full rollback |
| **P1** | Service down, major bug | Quick disable | Feature flag only |
| **P2** | Performance issues | Gradual rollback | Phased feature disable |
| **P3** | Minor bugs | Monitor, fix forward | No rollback |

---

## Rollback Types

### 1. Feature Flag Rollback (Fastest)

**Time**: < 2 minutes  
**Impact**: Disables feature, no data changes  
**Reversible**: Yes, immediately

### 2. Database Rollback (Medium)

**Time**: 5-30 minutes  
**Impact**: Reverts data changes  
**Reversible**: Partial (depending on backup)

### 3. Application Rollback (Slowest)

**Time**: 15-60 minutes  
**Impact**: Reverts code changes  
**Reversible**: Yes, but requires redeployment

---

## Feature Flag Rollback

### Disable Twitch Moderation

**Use case**: Twitch ban/unban feature causing issues

#### Via Environment Variable

```bash
# SSH to production server
ssh production-server

# Update environment variable
sudo sed -i 's/FEATURE_TWITCH_MODERATION=true/FEATURE_TWITCH_MODERATION=false/' /etc/clpr/.env

# Restart services
sudo systemctl restart clpr-backend
sudo systemctl restart clpr-frontend

# Verify
curl -s https://api.clpr.tv/api/v1/features | jq '.twitch_moderation'
# Should return: false
```

#### Via Database Feature Flag

```bash
# Connect to database
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod

# Disable feature
UPDATE feature_flags 
SET enabled = false, updated_at = NOW()
WHERE name = 'twitch_moderation';

-- Verify
SELECT name, enabled, updated_at FROM feature_flags WHERE name = 'twitch_moderation';

\q
```

#### Via API

```bash
API_TOKEN="${API_TOKEN}"

# Disable feature flag
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/twitch_moderation \
  -d '{"enabled": false, "reason": "Emergency rollback - INC-2026-001"}'

# Verify
curl -s -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/features | jq '.twitch_moderation'
```

#### Verification

```bash
# 1. Check feature flag status
curl -s https://api.clpr.tv/api/v1/features | jq '.twitch_moderation'

# 2. Test Twitch ban endpoint (should return feature disabled)
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/moderation/twitch/ban \
  -d '{"broadcaster_id":"123","user_id":"456"}' | jq

# Expected: {"error": "FEATURE_DISABLED", "message": "Twitch moderation is currently disabled"}

# 3. Check frontend UI
# Ban buttons should be hidden or disabled
```

---

### Disable Ban Sync

**Use case**: Ban sync causing rate limit issues or data inconsistencies

```bash
# Database method
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod << EOF
UPDATE feature_flags 
SET enabled = false 
WHERE name = 'ban_sync';

-- Also pause any scheduled sync jobs
UPDATE background_jobs 
SET enabled = false 
WHERE job_type = 'sync_bans';
EOF

# Verify
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/moderation/sync-bans \
  -d '{"broadcaster_id":"123"}' | jq
# Should return feature disabled error
```

---

### Disable Audit Logging

**Use case**: Audit logging causing database performance issues

**⚠️ WARNING**: Only disable as last resort. This affects compliance.

```bash
# Temporarily disable audit log writes
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/audit_logging \
  -d '{"enabled": false, "reason": "Emergency - database performance"}'

# Alternative: Switch to async logging
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/settings/audit_log_mode \
  -d '{"mode": "async", "buffer_size": 1000}'
```

**Post-Disable**:
- [ ] Notify compliance team immediately
- [ ] Document time period without logs
- [ ] Re-enable within 1 hour maximum
- [ ] Investigate performance issue

---

## Database Rollback

### Revert Recent Bans

**Use case**: Mass ban error, incorrect bans applied

#### Revert Bans from Last Hour

```bash
# Connect to database
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod

-- Check bans in last hour
SELECT id, user_id, channel_id, reason, created_at 
FROM bans 
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- If confirmed to revert
BEGIN;

-- Log the rollback
INSERT INTO audit_logs (action, actor_id, details, created_at)
VALUES (
  'rollback_bans',
  'system',
  '{"reason": "Emergency rollback", "incident": "INC-2026-001"}',
  NOW()
);

-- Delete recent bans
DELETE FROM bans 
WHERE created_at > NOW() - INTERVAL '1 hour';

-- Record count
SELECT 'Deleted ' || COUNT(*) || ' bans' AS result
FROM bans 
WHERE created_at > NOW() - INTERVAL '1 hour';

COMMIT;

\q
```

#### Revert Specific Ban Batch

```bash
# If you have ban IDs from incident
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod << 'EOF'
\set ban_ids '(''ban-abc123'', ''ban-def456'', ''ban-ghi789'')'

BEGIN;

-- Verify bans to be deleted
SELECT id, user_id, reason FROM bans WHERE id IN :ban_ids;

-- Delete if correct
DELETE FROM bans WHERE id IN :ban_ids;

COMMIT;
EOF
```

---

### Restore from Backup

**Use case**: Catastrophic data corruption, need point-in-time recovery

#### Point-in-Time Recovery (PITR)

```bash
#!/bin/bash
# restore-moderation-tables.sh

set -euo pipefail

BACKUP_TIMESTAMP="${1:-}"  # Format: 2026-02-03T10:00:00Z
INCIDENT_ID="${2:-INC-unknown}"

if [ -z "$BACKUP_TIMESTAMP" ]; then
  echo "Usage: $0 <backup_timestamp> [incident_id]"
  echo "Example: $0 2026-02-03T10:00:00Z INC-2026-001"
  exit 1
fi

echo "=== Moderation Database Restore ==="
echo "Restore Point: $BACKUP_TIMESTAMP"
echo "Incident: $INCIDENT_ID"
echo "==================================="
echo

# 1. Stop writes to affected tables
echo "Step 1: Disabling moderation features..."
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/all_moderation \
  -d '{"enabled": false}'

# 2. Create current state backup
echo "Step 2: Creating pre-restore backup..."
pg_dump -h production-db.clpr.tv -U clpr_admin -d clpr_prod \
  -t bans -t moderators -t audit_logs \
  > "/tmp/pre-restore-backup-$(date +%Y%m%d-%H%M%S).sql"

# 3. Restore from backup
echo "Step 3: Restoring from backup..."
# Using continuous archiving / PITR
pg_basebackup -h backup-db.clpr.tv -U replication_user \
  -D /var/lib/postgresql/restore \
  --recovery-target-time="$BACKUP_TIMESTAMP"

# 4. Import restored tables to production
echo "Step 4: Importing restored data..."
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod << SQL
-- Backup current state to history table
CREATE TABLE IF NOT EXISTS bans_history_${INCIDENT_ID} AS 
SELECT * FROM bans;

-- Restore from backup
TRUNCATE bans CASCADE;
TRUNCATE moderators CASCADE;

-- Import from restored backup
\copy bans FROM '/var/lib/postgresql/restore/bans.csv' CSV HEADER
\copy moderators FROM '/var/lib/postgresql/restore/moderators.csv' CSV HEADER

SQL

# 5. Verify restore
echo "Step 5: Verifying restored data..."
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod << SQL
SELECT 
  'bans' AS table_name,
  COUNT(*) AS row_count,
  MAX(created_at) AS latest_record
FROM bans

UNION ALL

SELECT 
  'moderators' AS table_name,
  COUNT(*) AS row_count,
  MAX(created_at) AS latest_record
FROM moderators;
SQL

# 6. Re-enable features
echo "Step 6: Re-enabling moderation features..."
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/all_moderation \
  -d '{"enabled": true}'

echo
echo "=== Restore Complete ==="
echo "Backup file: /tmp/pre-restore-backup-*.sql"
echo "History table: bans_history_${INCIDENT_ID}"
echo "Next steps:"
echo "  1. Verify functionality"
echo "  2. Update incident ticket"
echo "  3. Notify stakeholders"
```

---

## Application Rollback

### Rollback Backend Deployment

**Use case**: New backend version causing critical issues

#### Kubernetes Rollback

```bash
# View deployment history
kubectl rollout history deployment/clpr-backend -n production

# Rollback to previous version
kubectl rollout undo deployment/clpr-backend -n production

# Rollback to specific revision
kubectl rollout undo deployment/clpr-backend -n production --to-revision=3

# Monitor rollback progress
kubectl rollout status deployment/clpr-backend -n production

# Verify pods are running
kubectl get pods -n production -l app=clpr-backend
```

#### Docker Compose Rollback

```bash
# SSH to server
ssh production-server

# Stop current containers
cd /opt/clpr
sudo docker-compose down

# Pull previous version
PREV_VERSION="v1.2.3"  # Get from git tags
git checkout $PREV_VERSION

# Rebuild and start
sudo docker-compose build clpr-backend
sudo docker-compose up -d clpr-backend

# Check logs
sudo docker-compose logs -f clpr-backend
```

---

### Rollback Frontend Deployment

**Use case**: Frontend bug affecting moderation UI

#### CDN Rollback

```bash
# If using Cloudflare or similar CDN
# Purge cache to force old version

# Purge specific URLs
curl -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CF_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"files": ["https://clpr.tv/js/moderation.js", "https://clpr.tv/css/moderation.css"]}'

# Or purge everything
curl -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CF_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"purge_everything": true}'
```

#### Static File Rollback

```bash
# Replace with previous version
cd /var/www/clpr/frontend
sudo tar -xzf /backups/frontend-v1.2.3.tar.gz -C /var/www/clpr/frontend

# Restart web server
sudo systemctl restart nginx
```

---

## Emergency Disable Procedures

### Complete Moderation System Shutdown

**Use case**: Critical security breach, system-wide failure

**⚠️ EXTREME MEASURE**: Only use in catastrophic scenarios

```bash
#!/bin/bash
# emergency-shutdown.sh

INCIDENT_ID="${1:-INC-unknown}"

echo "=== EMERGENCY MODERATION SHUTDOWN ==="
echo "Incident: $INCIDENT_ID"
echo "Time: $(date -u)"
echo "======================================"
echo

# 1. Disable all moderation features
echo "Disabling all features..."
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/emergency/disable-moderation \
  -d "{\"incident_id\": \"$INCIDENT_ID\", \"reason\": \"Emergency shutdown\"}"

# 2. Close moderation API endpoints at firewall
echo "Blocking API endpoints..."
sudo iptables -A OUTPUT -p tcp --dport 443 -d api.clpr.tv -m string \
  --string "/api/v1/moderation" --algo bm -j DROP

# 3. Stop background jobs
echo "Stopping background jobs..."
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod << SQL
UPDATE background_jobs 
SET enabled = false 
WHERE job_type LIKE '%moderation%' OR job_type LIKE '%ban%';
SQL

# 4. Create incident log
echo "Creating incident log..."
cat > "/var/log/clpr/emergency-shutdown-${INCIDENT_ID}.log" << LOG
Emergency Shutdown - $INCIDENT_ID
Time: $(date -u)
Initiated by: $(whoami)
Actions taken:
  - All moderation features disabled
  - API endpoints blocked
  - Background jobs stopped
LOG

echo
echo "=== SHUTDOWN COMPLETE ==="
echo "Log file: /var/log/clpr/emergency-shutdown-${INCIDENT_ID}.log"
echo
echo "To restore:"
echo "  1. Fix underlying issue"
echo "  2. Run restore-moderation.sh"
echo "  3. Verify functionality"
```

### Restore After Emergency Shutdown

```bash
#!/bin/bash
# restore-moderation.sh

echo "=== Restoring Moderation System ==="

# 1. Remove firewall blocks
sudo iptables -D OUTPUT -p tcp --dport 443 -d api.clpr.tv -m string \
  --string "/api/v1/moderation" --algo bm -j DROP

# 2. Re-enable features gradually
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/ban_users \
  -d '{"enabled": true}'

sleep 5

curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/moderate_content \
  -d '{"enabled": true}'

sleep 5

curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/admin/features/twitch_moderation \
  -d '{"enabled": true}'

# 3. Re-enable background jobs
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod << SQL
UPDATE background_jobs 
SET enabled = true 
WHERE job_type LIKE '%moderation%' OR job_type LIKE '%ban%';
SQL

echo "=== Restore Complete ==="
echo "Run verification tests"
```

---

## Post-Rollback Verification

### Verification Checklist

After any rollback, verify:

- [ ] **Feature flags** reflect intended state
- [ ] **API endpoints** return expected responses
- [ ] **UI elements** display correctly
- [ ] **Background jobs** running normally
- [ ] **Audit logs** being created
- [ ] **Database integrity** verified
- [ ] **Performance metrics** within normal range

### Automated Verification Script

```bash
#!/bin/bash
# verify-rollback.sh

echo "=== Post-Rollback Verification ==="
echo

# 1. Check feature flags
echo "1. Feature Flags:"
curl -s https://api.clpr.tv/api/v1/features | jq '{
  twitch_moderation,
  ban_sync,
  audit_logging
}'

# 2. Test API health
echo
echo "2. API Health:"
curl -s https://api.clpr.tv/api/v1/moderation/health | jq

# 3. Verify database connectivity
echo
echo "3. Database:"
psql -h production-db.clpr.tv -U clpr_admin -d clpr_prod -c \
  "SELECT COUNT(*) as ban_count FROM bans WHERE created_at > NOW() - INTERVAL '1 hour';"

# 4. Check recent audit logs
echo
echo "4. Audit Logs (last 5):"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  https://api.clpr.tv/api/v1/moderation/audit-logs?limit=5 | \
  jq '.logs[] | {time: .created_at, action: .action}'

# 5. Performance check
echo
echo "5. Performance:"
START=$(date +%s%N)
curl -s https://api.clpr.tv/api/v1/moderation/bans?limit=10 > /dev/null
END=$(date +%s%N)
LATENCY=$(( (END - START) / 1000000 ))
echo "API latency: ${LATENCY}ms"

if [ $LATENCY -lt 500 ]; then
  echo "✓ Performance OK"
else
  echo "⚠ High latency detected"
fi

echo
echo "=== Verification Complete ==="
```

---

## Related Runbooks

- [Moderation Operations](./moderation-operations.md) - Normal operations
- [Moderation Incidents](./moderation-incidents.md) - Incident response
- [Moderation Monitoring](./moderation-monitoring.md) - Metrics and alerts

---

## Emergency Contacts

- **Incident Commander**: ops-oncall@clpr.tv
- **Engineering Lead**: eng-lead@clpr.tv
- **Database Team**: dba@clpr.tv
- **Security Team**: security@clpr.tv

---

**Last Updated**: 2026-02-03  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
