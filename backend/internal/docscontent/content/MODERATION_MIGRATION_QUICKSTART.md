# Moderation Migration Scripts - Quick Start Guide

Quick reference guide for deploying the moderation system using the migration scripts.

## Prerequisites

- [ ] golang-migrate installed
- [ ] Database credentials configured
- [ ] Backup system operational
- [ ] Access to staging/production servers

## Quick Commands

### Test Everything First

```bash
# Run test suite
./scripts/test-migration-scripts.sh
```

### Staging Deployment

```bash
# 1. Pre-flight checks
./scripts/preflight-moderation.sh --env staging

# 2. Dry run
./scripts/migrate-moderation.sh --env staging --dry-run

# 3. Execute migration
./scripts/migrate-moderation.sh --env staging

# 4. Validate
./scripts/validate-moderation.sh --env staging
```

### Production Deployment

```bash
# 1. Pre-flight checks with report
./scripts/preflight-moderation.sh --env production --report preflight-prod.txt

# 2. Create backup
./scripts/backup.sh

# 3. Execute migration (requires 'yes' confirmation)
./scripts/migrate-moderation.sh --env production

# 4. Validate with report
./scripts/validate-moderation.sh --env production --report validation-prod.txt
```

### Rollback (If Needed)

```bash
# Dry run first
./scripts/rollback-moderation.sh --env production --dry-run --target 10

# Execute rollback (requires 'ROLLBACK' confirmation)
./scripts/rollback-moderation.sh --env production --target 10
```

## Script Options Reference

### preflight-moderation.sh

```bash
Options:
  -e, --env ENV          Environment (staging|production) [default: staging]
  -r, --report FILE      Generate report to file
  -h, --help             Show help
```

### migrate-moderation.sh

```bash
Options:
  -e, --env ENV          Environment (staging|production) [default: staging]
  --dry-run              Test mode without making changes
  --skip-backup          Skip database backup (not recommended)
  --skip-validation      Skip post-migration validation
  -h, --help             Show help
```

### validate-moderation.sh

```bash
Options:
  -e, --env ENV          Environment (staging|production) [default: staging]
  -r, --report FILE      Generate report to file
  -h, --help             Show help
```

### rollback-moderation.sh

```bash
Options:
  -e, --env ENV          Environment (staging|production) [default: staging]
  --target VERSION       Target migration version to rollback to
  --dry-run              Test mode without making changes
  --skip-backup          Skip data backup (not recommended)
  -h, --help             Show help
```

## Common Workflows

### First-Time Production Deployment

```bash
# Week before: Test on staging
./scripts/preflight-moderation.sh --env staging
./scripts/migrate-moderation.sh --env staging
./scripts/validate-moderation.sh --env staging --report validation-staging.txt

# Day before: Pre-flight production
./scripts/preflight-moderation.sh --env production --report preflight-prod.txt

# Deployment day: Execute
./scripts/backup.sh
./scripts/migrate-moderation.sh --env production
./scripts/validate-moderation.sh --env production --report validation-prod.txt

# Monitor for 1 hour
docker compose logs -f backend | grep -i "moderation"
```

### Emergency Rollback

```bash
# Immediate rollback to before moderation features
./scripts/rollback-moderation.sh --env production --target 10

# Verify rollback
./scripts/health-check.sh
docker compose logs -f backend
```

### Partial Rollback

```bash
# Rollback to keep basic moderation but remove appeals
./scripts/rollback-moderation.sh --env production --target 49
```

## Troubleshooting

### Issue: Pre-flight checks fail

```bash
# Review specific failures
./scripts/preflight-moderation.sh --env production --report preflight.txt
cat preflight.txt

# Fix issues and re-run
```

### Issue: Migration fails (dirty state)

```bash
# Check current version
migrate -path backend/migrations -database "$DB_URL" version

# Force clean (carefully!)
migrate -path backend/migrations -database "$DB_URL" force VERSION

# Retry
./scripts/migrate-moderation.sh --env production
```

### Issue: Validation fails

```bash
# Get detailed report
./scripts/validate-moderation.sh --env production --report validation.txt
cat validation.txt

# Check specific issues and fix
```

## Environment Variables

Required in `.env` file:

```bash
# Database
DB_HOST=localhost
DB_PORT=5436
DB_USER=clpr
DB_PASSWORD=secure_password
DB_NAME=clpr_db
DB_SSLMODE=require  # For production/staging; use 'disable' only for local development

# Optional
BACKUP_DIR=/var/backups/clpr
```

## Exit Codes

All scripts return:
- `0` - Success
- `1` - Failure

## Getting Help

### Script Help

```bash
./scripts/preflight-moderation.sh --help
./scripts/migrate-moderation.sh --help
./scripts/validate-moderation.sh --help
./scripts/rollback-moderation.sh --help
```

### Documentation

- [Complete Deployment Guide](../docs/deployment/moderation-deployment.md)
- [Production Deployment Checklist](../docs/deployment/MODERATION_DEPLOYMENT_CHECKLIST.md)
- [Scripts README](./README.md)

## Safety Tips

1. ✅ **Always test on staging first**
2. ✅ **Always create backups before migration**
3. ✅ **Use dry-run mode to preview changes**
4. ✅ **Monitor logs during and after deployment**
5. ✅ **Know how to rollback quickly**
6. ✅ **Schedule deployments during low-traffic periods**
7. ✅ **Keep team informed during deployment**

## Migrations Included

| Version | Description | Tables Added |
|---------|-------------|--------------|
| 000011 | Moderation Audit Logs | `moderation_audit_logs` |
| 000049 | Moderation Queue System | `moderation_queue`, `moderation_decisions` |
| 000050 | Moderation Appeals | `moderation_appeals` |
| 000069 | Forum Moderation | Forum-specific tables |
| 000097 | Updated Audit Logs | Enhanced audit logs |

## Support

For issues:
1. Check troubleshooting section above
2. Review deployment guide
3. Check application logs
4. Contact DevOps team

---

**Last Updated:** 2026-02-04
**Version:** 1.0
