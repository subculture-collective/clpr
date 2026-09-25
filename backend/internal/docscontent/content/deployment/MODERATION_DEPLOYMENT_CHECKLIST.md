# Moderation System Production Deployment Checklist

Use this checklist to ensure all steps are completed for a safe production deployment.

## Pre-Deployment (1 Week Before)

- [ ] Staging deployment completed successfully
- [ ] Staging validation tests passed
- [ ] Smoke tests verified on staging
- [ ] Performance testing completed on staging
- [ ] Security review completed
- [ ] Documentation reviewed and updated
- [ ] Team training on moderation features completed
- [ ] Rollback procedure tested on staging
- [ ] Deployment window scheduled (date/time: ____________)
- [ ] Change management ticket created
- [ ] Stakeholders notified

## Pre-Deployment (24 Hours Before)

- [ ] Announce maintenance window to users
- [ ] Verify backup system is operational
- [ ] Check database disk space (>20% free)
- [ ] Verify monitoring dashboards accessible
- [ ] Alert on-call engineers
- [ ] Review recent production incidents
- [ ] Confirm no conflicting deployments scheduled

## Pre-Deployment (1 Hour Before)

- [ ] Run pre-flight checks: `./scripts/preflight-moderation.sh --env production`
- [ ] Verify all checks passed
- [ ] Create full database backup: `./scripts/backup.sh`
- [ ] Verify backup completed successfully
- [ ] Test backup restore on staging (if time permits)
- [ ] Check system resources (CPU, memory, disk)
- [ ] Verify database connections stable
- [ ] Review application logs for existing errors

## Deployment

### Migration Execution

- [ ] Start deployment at: ____________ (time)
- [ ] Enable maintenance mode (if required)
- [ ] Run migration: `./scripts/migrate-moderation.sh --env production`
- [ ] Confirm production deployment (type 'yes')
- [ ] Migration completed successfully
- [ ] Note completion time: ____________
- [ ] Migration duration: ____________ minutes

### Validation

- [ ] Run validation: `./scripts/validate-moderation.sh --env production --report validation.txt`
- [ ] All validation checks passed
- [ ] Review validation report
- [ ] No errors in validation output

### Health Checks

- [ ] Disable maintenance mode (if enabled)
- [ ] Run health check: `./scripts/health-check.sh`
- [ ] Backend API responding (200 OK)
- [ ] Database connections stable
- [ ] All services healthy

### Smoke Tests

- [ ] Test moderation queue endpoint
- [ ] Test moderation actions endpoint
- [ ] Test moderation appeals endpoint
- [ ] Create test moderation queue item
- [ ] Resolve test moderation queue item
- [ ] Verify audit logs created
- [ ] Delete test data

## Post-Deployment Monitoring (1 Hour)

- [ ] Monitor application logs (no critical errors)
- [ ] Check error rates (normal levels)
- [ ] Verify response times (no degradation)
- [ ] Check database performance (queries executing normally)
- [ ] Monitor database connections (stable)
- [ ] Verify moderation features accessible to moderators
- [ ] Check CPU usage (normal levels)
- [ ] Check memory usage (normal levels)
- [ ] Review monitoring dashboards

## Post-Deployment Communication

- [ ] Notify team of successful deployment
- [ ] Update deployment status in change management system
- [ ] Post announcement to team chat
- [ ] Document any issues encountered
- [ ] Update deployment notes
- [ ] Schedule retrospective (if needed)

## Post-Deployment Tasks (24 Hours)

- [ ] Review logs for any warnings
- [ ] Check moderation queue usage
- [ ] Verify no data integrity issues
- [ ] Monitor user reports/feedback
- [ ] Verify backup retention policy
- [ ] Archive deployment artifacts
- [ ] Update documentation with lessons learned

## Rollback Trigger Conditions

Roll back immediately if any of these occur:

- [ ] Critical errors in application logs
- [ ] Validation checks fail
- [ ] Database performance degraded significantly
- [ ] Moderation features not accessible
- [ ] Data integrity issues detected
- [ ] High error rate (>5% increase)
- [ ] Service downtime >5 minutes

## Rollback Procedure (If Needed)

- [ ] Announce rollback to team
- [ ] Run rollback: `./scripts/rollback-moderation.sh --env production --target 10`
- [ ] Verify rollback completed
- [ ] Run health checks
- [ ] Verify application functionality
- [ ] Document rollback reason
- [ ] Schedule post-mortem

## Sign-Off

**Deployment Lead:** _______________________ Date: _______

**Database Admin:** _______________________ Date: _______

**QA Engineer:** _______________________ Date: _______

**Notes:**
```
_________________________________________________________________

_________________________________________________________________

_________________________________________________________________
```

## Emergency Contacts

- **On-Call Engineer:** ________________________
- **Database Admin:** ________________________
- **DevOps Lead:** ________________________
- **Product Owner:** ________________________

## Rollback Decision Matrix

| Severity | Impact | Action |
|----------|--------|--------|
| Critical | Service down | Rollback immediately |
| High | Major features broken | Rollback within 15 min |
| Medium | Minor features affected | Monitor and fix forward |
| Low | Cosmetic issues | Fix in next release |

---

**Deployment Date:** _______________

**Deployment Status:** [ ] Success [ ] Rolled Back [ ] Partial

**Completion Time:** _______________

**Notes:** _________________________________________________________
