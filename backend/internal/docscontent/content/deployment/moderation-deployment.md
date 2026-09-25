---
title: "Moderation System Deployment"
summary: "Deployment guide for moderation system features"
tags: ["deployment", "moderation", "migrations"]
area: "deployment"
status: "stable"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-04
aliases: ["moderation deployment", "moderation migration"]
---

# Moderation System Deployment Guide

Comprehensive guide for deploying the moderation system to staging and production environments.

## Overview

The moderation system includes:

- **Moderation Audit Logs** - Track all moderation actions
- **Moderation Queue System** - Queue for reviewing flagged content
- **Moderation Appeals** - Allow users to appeal moderation decisions
- **Forum Moderation** - Moderation tools for forums and communities
- **Updated Audit Logs** - Enhanced audit logging

### Migrations Included

| Migration | Description | Tables Added |
|-----------|-------------|--------------|
| 000011 | Moderation Audit Logs | `moderation_audit_logs` |
| 000049 | Moderation Queue System | `moderation_queue`, `moderation_decisions` |
| 000050 | Moderation Appeals | `moderation_appeals` |
| 000069 | Forum Moderation | Forum-specific moderation tables |
| 000097 | Updated Moderation Audit Logs | Enhanced `moderation_audit_logs` |

## Prerequisites

### System Requirements

- PostgreSQL 12+ (for JSON support)
- golang-migrate tool installed
- Database backup capability
- Moderator/admin users configured

### Dependencies

The following tables must exist before migration:
- `users` (with role support)
- `clips` (for content moderation)
- `comments` (for comment moderation)

## Staging Deployment

### Step 1: Pre-flight Checks

Run pre-flight checks to validate the environment:

```bash
cd /opt/clpr
./scripts/preflight-moderation.sh --env staging
```

Expected output:
```
✓ All pre-flight checks passed!
Moderation system migration may proceed.
```

If checks fail, fix all issues before proceeding.

### Step 2: Dry Run

Test the migration in dry-run mode:

```bash
./scripts/migrate-moderation.sh --env staging --dry-run
```

Review the output to understand what will be executed.

### Step 3: Create Backup

Create a database backup (if not using `--skip-backup`):

```bash
./scripts/backup.sh
```

Verify backup was created:

```bash
ls -lh /var/backups/clpr/
```

### Step 4: Run Migration

Execute the migration:

```bash
./scripts/migrate-moderation.sh --env staging
```

**Expected Duration:** 2-5 minutes depending on database size

### Step 5: Validate Migration

Run post-migration validation:

```bash
./scripts/validate-moderation.sh --env staging --report validation-staging.txt
```

Expected output:
```
✓ All validation checks passed!
Moderation system is ready for use.
```

### Step 6: Smoke Tests

Test basic moderation functionality:

```bash
# Test moderation queue API endpoint
curl -X GET https://staging.clpr.app/api/moderation/queue \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Expected: 200 OK with empty or populated queue
```

### Step 7: Monitor

Monitor application logs for errors:

```bash
docker compose logs -f backend | grep -i "moderation"
```

## Production Deployment

### Pre-Deployment Checklist

- [ ] Staging deployment completed successfully
- [ ] Staging validation passed
- [ ] Smoke tests completed on staging
- [ ] Database backup verified and tested
- [ ] Rollback procedure reviewed
- [ ] Deployment window scheduled (low-traffic period)
- [ ] Team notified of deployment
- [ ] Monitoring dashboards ready

### Production Deployment Steps

#### 1. Announce Maintenance Window

Notify users of upcoming maintenance (if downtime expected):

```bash
# Post announcement via admin panel or API
```

#### 2. Enable Maintenance Mode (Optional)

If downtime is required:

```bash
# Enable maintenance mode
docker compose exec backend /app/maintenance on
```

#### 3. Pre-flight Checks

```bash
cd /opt/clpr
./scripts/preflight-moderation.sh --env production --report preflight-prod.txt
```

Review the report and ensure all checks pass.

#### 4. Create Production Backup

```bash
# Full database backup
./scripts/backup.sh

# Verify backup
LATEST_BACKUP=$(ls -t /var/backups/clpr/db-*.sql.gz | head -1)
echo "Latest backup: $LATEST_BACKUP"
du -h "$LATEST_BACKUP"
```

**Critical:** Test backup restore on staging before proceeding!

#### 5. Run Migration with Confirmation

```bash
./scripts/migrate-moderation.sh --env production
```

You will be prompted to type 'yes' to confirm production deployment.

**Expected Duration:** 5-15 minutes for production databases

#### 6. Validate Migration

```bash
./scripts/validate-moderation.sh --env production --report validation-prod.txt
```

Review the validation report carefully.

#### 7. Disable Maintenance Mode

```bash
docker compose exec backend /app/maintenance off
```

#### 8. Health Checks

```bash
./scripts/health-check.sh
```

Verify all services are healthy:
- Backend API responding
- Database connections stable
- Moderation endpoints accessible

#### 9. Smoke Tests

Test moderation functionality:

```bash
# Test moderation queue
curl -X GET https://clpr.app/api/moderation/queue \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Test creating a moderation action
curl -X POST https://clpr.app/api/moderation/queue \
  -H "Authorization: Bearer $MODERATOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content_type": "comment", "content_id": "test-id", "reason": "spam"}'
```

#### 10. Monitor Production

Monitor for at least 1 hour post-deployment:

```bash
# Watch logs
docker compose logs -f backend

# Check error rates
# (Use your monitoring dashboard: Grafana, Datadog, etc.)

# Check database connections
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM pg_stat_activity WHERE datname = 'clpr_db';"
```

### Post-Deployment Checklist

- [ ] All validation checks passed
- [ ] Smoke tests completed successfully
- [ ] No errors in application logs
- [ ] Database performance normal
- [ ] Moderation features accessible
- [ ] Team notified of successful deployment
- [ ] Documentation updated

## Rollback Procedures

If issues are detected, rollback immediately.

### When to Rollback

Rollback if:
- Validation checks fail
- Critical errors in application logs
- Moderation features not working
- Database performance degraded
- Data integrity issues detected

### Rollback Steps

#### 1. Quick Rollback (Recommended)

Use the rollback script:

```bash
# Dry run first
./scripts/rollback-moderation.sh --env production --dry-run

# Execute rollback
./scripts/rollback-moderation.sh --env production --target 10
```

This will:
1. Backup current moderation data
2. Rollback migrations to version 10 (before moderation)
3. Verify rollback completed successfully

**Expected Duration:** 2-5 minutes

#### 2. Manual Rollback

If the script fails, use golang-migrate directly:

```bash
cd /opt/clpr

# Check current version
migrate -path backend/migrations -database "$DB_URL" version

# Rollback to specific version
migrate -path backend/migrations -database "$DB_URL" goto 10
```

#### 3. Restore from Backup (Last Resort)

If rollback fails or data is corrupted:

```bash
# Stop application
docker compose down

# Restore from backup
BACKUP_FILE="/var/backups/clpr/db-YYYYMMDD-HHMMSS.sql.gz"
gunzip -c "$BACKUP_FILE" | psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Restart application
docker compose up -d

# Verify
./scripts/health-check.sh
```

#### 4. Verify Rollback

```bash
# Check migration version
migrate -path backend/migrations -database "$DB_URL" version

# Verify moderation tables removed
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt moderation*"

# Test application
./scripts/health-check.sh
```

### Post-Rollback Actions

1. **Document the issue** - Record what went wrong
2. **Notify team** - Alert team of rollback
3. **Investigate root cause** - Debug the issue
4. **Plan retry** - Fix issues and reschedule deployment
5. **Update runbook** - Document lessons learned

## Verification Procedures

### Database Verification

```bash
# Connect to database
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Check moderation tables
\dt moderation*

# Check table row counts
SELECT 'moderation_audit_logs' as table_name, COUNT(*) FROM moderation_audit_logs
UNION ALL
SELECT 'moderation_queue', COUNT(*) FROM moderation_queue
UNION ALL
SELECT 'moderation_decisions', COUNT(*) FROM moderation_decisions
UNION ALL
SELECT 'moderation_appeals', COUNT(*) FROM moderation_appeals;

# Check indexes
SELECT tablename, indexname 
FROM pg_indexes 
WHERE tablename LIKE 'moderation%'
ORDER BY tablename, indexname;

# Check constraints
SELECT conname, contype, conrelid::regclass 
FROM pg_constraint 
WHERE conrelid::regclass::text LIKE 'moderation%'
ORDER BY conrelid::regclass::text;
```

### Application Verification

```bash
# Test moderation queue endpoint
curl -X GET https://clpr.app/api/moderation/queue \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq

# Test moderation actions
curl -X GET https://clpr.app/api/moderation/decisions \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq

# Test moderation appeals
curl -X GET https://clpr.app/api/moderation/appeals \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq
```

### Performance Verification

```sql
-- Check query performance on moderation_queue
EXPLAIN ANALYZE 
SELECT * FROM moderation_queue 
WHERE status = 'pending' 
ORDER BY priority DESC, created_at 
LIMIT 10;

-- Should use index: idx_modqueue_status_priority

-- Check foreign key performance
EXPLAIN ANALYZE
SELECT mq.*, md.action, md.reason
FROM moderation_queue mq
LEFT JOIN moderation_decisions md ON md.queue_item_id = mq.id
WHERE mq.status = 'approved'
LIMIT 10;
```

## Troubleshooting

### Common Issues

#### Issue: Pre-flight checks fail

**Symptom:** `preflight-moderation.sh` reports failed checks

**Solution:**
1. Review the specific failed checks
2. Fix configuration issues (database connection, environment variables)
3. Ensure prerequisite tables exist
4. Run pre-flight checks again

#### Issue: Migration fails with "dirty" state

**Symptom:** `Error: Dirty database version`

**Solution:**
```bash
# Check current version and dirty state
migrate -path backend/migrations -database "$DB_URL" version

# Force clean state (use carefully!)
# First, manually verify what failed and fix if needed
migrate -path backend/migrations -database "$DB_URL" force VERSION

# Then retry migration
./scripts/migrate-moderation.sh --env production
```

#### Issue: Validation fails after migration

**Symptom:** `validate-moderation.sh` reports missing tables or indexes

**Solution:**
1. Check specific validation failures
2. Verify migrations actually ran:
   ```bash
   migrate -path backend/migrations -database "$DB_URL" version
   ```
3. Check for errors in migration logs
4. If tables are missing, re-run specific migration:
   ```bash
   migrate -path backend/migrations -database "$DB_URL" goto VERSION
   ```

#### Issue: Performance degradation after migration

**Symptom:** Slow queries, high database load

**Solution:**
1. Check if indexes are being used:
   ```sql
   EXPLAIN ANALYZE SELECT * FROM moderation_queue WHERE status = 'pending';
   ```
2. Verify indexes exist:
   ```sql
   SELECT * FROM pg_indexes WHERE tablename = 'moderation_queue';
   ```
3. Analyze tables:
   ```sql
   ANALYZE moderation_queue;
   ANALYZE moderation_decisions;
   ANALYZE moderation_appeals;
   ```

## Monitoring

### Key Metrics to Monitor

- **Migration duration** - Should complete in 5-15 minutes
- **Database connections** - Should remain stable
- **Error rate** - Should not increase
- **Response times** - Should not degrade
- **Moderation queue size** - Should be accessible

### Alerts

Set up alerts for:
- Migration failures
- Validation failures
- High error rates post-deployment
- Database connection issues
- Slow query performance

## Scripts Reference

### preflight-moderation.sh

Pre-flight checks for moderation system deployment.

```bash
./scripts/preflight-moderation.sh --env production --report preflight.txt
```

**Options:**
- `--env` - Environment (staging|production)
- `--report` - Generate report file

### migrate-moderation.sh

Migration runner for moderation system.

```bash
./scripts/migrate-moderation.sh --env production
```

**Options:**
- `--env` - Environment (staging|production)
- `--dry-run` - Show what would be done
- `--skip-backup` - Skip backup (not recommended)
- `--skip-validation` - Skip post-migration validation

### validate-moderation.sh

Post-migration validation script.

```bash
./scripts/validate-moderation.sh --env production --report validation.txt
```

**Options:**
- `--env` - Environment (staging|production)
- `--report` - Generate report file

### rollback-moderation.sh

Rollback script for moderation system.

```bash
./scripts/rollback-moderation.sh --env production --target 10
```

**Options:**
- `--env` - Environment (staging|production)
- `--target` - Target migration version
- `--dry-run` - Show what would be done
- `--skip-backup` - Skip data backup (not recommended)

## Best Practices

1. **Always test on staging first** - Deploy to staging before production
2. **Create backups** - Always backup before migration
3. **Monitor closely** - Watch logs and metrics during and after deployment
4. **Schedule wisely** - Deploy during low-traffic periods
5. **Have rollback ready** - Know how to rollback quickly
6. **Document everything** - Record what was done and any issues
7. **Communicate** - Keep team informed throughout deployment

## Related Documentation

- [[migration|Database Migrations]]
- [[deployment|Deployment Procedures]]
- [[../operations/runbook|Operations Runbook]]
- [[../backend/moderation|Moderation System Architecture]]

---

[[index|← Back to Deployment Index]]
