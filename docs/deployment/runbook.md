---
title: "Operations Runbook"
summary: "Operational procedures and commands for managing Clipper in production."
tags: ["deployment", "runbook", "ops"]
area: "deployment"
status: "stable"
owner: "team-ops"
version: "1.0"
last_reviewed: 2025-12-24
aliases: ["runbook", "ops procedures"]
---

# Operations Runbook

> **Note**: The comprehensive operations runbook is maintained at [[../operations/runbook|operations/runbook.md]]. This page provides quick deployment-specific procedures.

For complete operational procedures including monitoring, incident response, deployment testing, and rollback drills, see [[../operations/runbook|Operations Runbook]].

## Quick Deployment Procedures

### Deploy New Version

```bash
# Deploy new version
kubectl set image deployment/backend backend=clpr:v1.2.3 -n clpr

# Check rollout status
kubectl rollout status deployment/backend -n clpr

# Rollback if needed
kubectl rollout undo deployment/backend -n clpr
```

### Pre-Deployment Checklist

1. **Run Deployment Tests**: Verify all deployment scripts pass tests
   ```bash
   cd scripts && ./test-deployment-harness.sh
   ```

2. **Run Rollback Drill**: Validate rollback procedures
   ```bash
   DRY_RUN=true ./rollback-drill.sh
   ```

3. **Review Artifacts**: Check logs and reports for warnings

4. **Staging Rehearsal**: Full end-to-end test
   ```bash
   ./staging-rehearsal.sh
   ```

### Post-Deployment Verification

1. Check service health: `kubectl get pods -n clpr`
2. Check logs: `kubectl logs -f deployment/backend -n clpr`
3. Verify metrics in Grafana
4. Run smoke tests
5. Monitor for 30 minutes

## Common Deployment Issues

### Deployment Stuck

1. Check pod events: `kubectl describe pod <pod-name> -n clpr`
2. Check resource limits and quotas
3. Verify image pull credentials
4. Check for failing health checks

### Rolling Update Failed

1. Check rollout status: `kubectl rollout status deployment/backend -n clpr`
2. Review pod logs for errors
3. Rollback: `kubectl rollout undo deployment/backend -n clpr`
4. Investigate and fix issues before re-deploying

## Related Documentation

- [[../operations/runbook|Operations Runbook]] - Complete operational procedures
- [[docker|Docker Deployment]] - Container-based deployment guide
- [[ci_cd|CI/CD Pipeline]] - Automated deployment workflows
- [[../operations/blue-green-deployment|Blue-Green Deployment]] - Zero-downtime deployments
- [[../operations/preflight|Preflight Checklist]] - Pre-deployment validation

---

**See also:**
[[../operations/runbook|Full Operations Runbook]] ·
[[index|Deployment Index]] ·
[[../index|Documentation Home]]
