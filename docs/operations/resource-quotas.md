---
title: "Resource Quotas"
summary: "**Status**: Implemented (Roadmap 5.0 Phase 5.2)"
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Resource Quotas & Limits

**Status**: Implemented (Roadmap 5.0 Phase 5.2)  
**Related Issues**: [#853](https://git.subcult.tv/subculture-collective/clpr/issues/853), [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805)

## Overview

This document describes the resource quota and limit policies enforced in the Clipper Kubernetes cluster to prevent resource exhaustion and ensure fair allocation across namespaces.

## Architecture

### ResourceQuotas

ResourceQuotas enforce hard limits on aggregate resource consumption per namespace. They prevent any single namespace from consuming excessive cluster resources.

**Namespaces with ResourceQuotas:**
- `clpr-production` - Production workloads
- `clpr-staging` - Staging/testing workloads
- `clpr-monitoring` - Monitoring infrastructure (Prometheus, Grafana, Loki)

### LimitRanges

LimitRanges define default, minimum, and maximum resource constraints for individual pods and containers within a namespace. They ensure that every container has appropriate resource requests and limits.

## Resource Allocations

### Production Environment (`clpr-production`)

#### ResourceQuota
```yaml
requests.cpu: "20"           # Total CPU requests
requests.memory: "40Gi"      # Total memory requests
limits.cpu: "40"             # Total CPU limits
limits.memory: "80Gi"        # Total memory limits
requests.storage: "200Gi"    # Total storage
pods: "100"                  # Max pods
persistentvolumeclaims: "20" # Max PVCs
```

#### LimitRange (Container)
```yaml
default:
  cpu: "1000m"
  memory: "1Gi"
defaultRequest:
  cpu: "100m"
  memory: "256Mi"
max:
  cpu: "4000m"
  memory: "8Gi"
min:
  cpu: "10m"
  memory: "32Mi"
maxLimitRequestRatio:
  cpu: 10
  memory: 8
```

### Staging Environment (`clpr-staging`)

#### ResourceQuota
```yaml
requests.cpu: "10"           # Total CPU requests
requests.memory: "20Gi"      # Total memory requests
limits.cpu: "20"             # Total CPU limits
limits.memory: "40Gi"        # Total memory limits
requests.storage: "100Gi"    # Total storage
pods: "50"                   # Max pods
persistentvolumeclaims: "15" # Max PVCs
```

#### LimitRange (Container)
```yaml
default:
  cpu: "500m"
  memory: "512Mi"
defaultRequest:
  cpu: "50m"
  memory: "128Mi"
max:
  cpu: "2000m"
  memory: "4Gi"
min:
  cpu: "10m"
  memory: "32Mi"
```

### Monitoring Environment (`clpr-monitoring`)

#### ResourceQuota
```yaml
requests.cpu: "8"            # Total CPU requests
requests.memory: "16Gi"      # Total memory requests
limits.cpu: "16"             # Total CPU limits
limits.memory: "32Gi"        # Total memory limits
requests.storage: "200Gi"    # Total storage (metrics/logs)
pods: "30"                   # Max pods
persistentvolumeclaims: "10" # Max PVCs
```

#### LimitRange (Container)
```yaml
default:
  cpu: "1000m"
  memory: "2Gi"
defaultRequest:
  cpu: "100m"
  memory: "512Mi"
max:
  cpu: "4000m"
  memory: "16Gi"
min:
  cpu: "10m"
  memory: "64Mi"
```

## Service-Specific Resource Configurations

### Backend Service

**Production:**
```yaml
requests:
  cpu: 250m
  memory: 512Mi
limits:
  cpu: 1000m
  memory: 1Gi
```

**Staging:**
```yaml
requests:
  cpu: 100m
  memory: 256Mi
limits:
  cpu: 500m
  memory: 512Mi
```

### Frontend Service

```yaml
requests:
  cpu: 50m
  memory: 64Mi
limits:
  cpu: 200m
  memory: 256Mi
```

### PostgreSQL

```yaml
requests:
  cpu: 250m
  memory: 512Mi
limits:
  cpu: 1000m
  memory: 2Gi
```

### Redis

```yaml
requests:
  cpu: 100m
  memory: 256Mi
limits:
  cpu: 500m
  memory: 512Mi
```

## Deployment

### Apply ResourceQuotas and LimitRanges

```bash
# Apply to all namespaces
kubectl apply -f infrastructure/k8s/base/resource-quotas.yaml
kubectl apply -f infrastructure/k8s/base/limit-ranges.yaml

# Verify application
kubectl get resourcequota -A
kubectl get limitrange -A
```

### Check Quota Usage

```bash
# View quota status for a namespace
kubectl describe resourcequota -n clpr-production

# View limit ranges
kubectl describe limitrange -n clpr-production

# Check quota usage across all namespaces
kubectl get resourcequota -A -o json | \
  jq -r '.items[] | "\(.metadata.namespace): CPU=\(.status.used["requests.cpu"])/\(.status.hard["requests.cpu"]) Memory=\(.status.used["requests.memory"])/\(.status.hard["requests.memory"])"'
```

## Out-of-Memory (OOM) Behavior

### OOM Kill Process

When a container exceeds its memory limit:

1. **Detection**: Kernel OOM killer detects memory exhaustion
2. **Termination**: Container process is killed (exit code 137)
3. **Restart**: Kubernetes restarts the container based on restart policy
4. **Status**: Pod status shows `OOMKilled` reason
5. **Alert**: Prometheus alert `ContainerOOMKilled` fires

### Testing OOM Behavior

#### Create a Test Pod

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: oom-test
  namespace: clpr-staging
spec:
  containers:
  - name: memory-hog
    image: polinux/stress:1.0.4
    resources:
      limits:
        memory: "128Mi"
      requests:
        memory: "64Mi"
    command: ["stress"]
    args: ["--vm", "1", "--vm-bytes", "256M", "--vm-hang", "1"]
  restartPolicy: Never
EOF
```

#### Monitor OOM Events

```bash
# Watch pod status
kubectl get pod oom-test -n clpr-staging -w

# Check termination reason
kubectl describe pod oom-test -n clpr-staging | grep -A 5 "Last State"

# View OOM events
kubectl get events -n clpr-staging --field-selector involvedObject.name=oom-test
```

#### Expected Behavior

1. Pod starts successfully
2. Container attempts to allocate 256MB (exceeds 128Mi limit)
3. OOM killer terminates the container
4. Pod status shows `OOMKilled`
5. Alert fires in Prometheus/Alertmanager

#### Cleanup

```bash
kubectl delete pod oom-test -n clpr-staging
```

### OOM Prevention

1. **Set appropriate limits**: Base limits on actual memory usage patterns
2. **Monitor trends**: Use Grafana dashboards to track memory usage over time
3. **Configure alerts**: Set alerts at 90% of memory limit
4. **Implement graceful degradation**: Handle low-memory conditions in application code
5. **Use HPA**: Scale horizontally before hitting memory limits

## Monitoring and Alerts

### Grafana Dashboard

Import the Resource Quotas & Limits dashboard:
- **Location**: `monitoring/dashboards/resource-quotas.json`
- **URL**: `http://grafana.clpr.tv/d/resource-quotas`

**Dashboard Panels:**
- CPU/Memory quota usage by namespace
- Pod count vs quota
- Container memory usage vs limits
- CPU throttling rates
- OOM kill events
- Quota violation alerts

### Prometheus Alerts

The following alerts are configured in `monitoring/alerts.yml`:

#### Critical Alerts
- `ContainerOOMKilled` - Container terminated due to OOM
- `ResourceQuotaCPUNearLimit` - CPU quota >80% used
- `ResourceQuotaMemoryNearLimit` - Memory quota >80% used

#### Warning Alerts
- `PodMemoryNearLimit` - Container using >90% of memory limit
- `PodCPUThrottling` - Significant CPU throttling detected
- `ResourceQuotaPodCountNearLimit` - Pod count >90% of quota
- `StorageQuotaNearLimit` - Storage quota >85% used
- `LimitRangeViolation` - Pod violates LimitRange constraints

### Alert Runbooks

When quota alerts fire:

1. **Identify the issue**:
   ```bash
   kubectl describe resourcequota -n <namespace>
   kubectl top pods -n <namespace>
   ```

2. **Check current usage**:
   ```bash
   # View resource requests/limits per pod
   kubectl get pods -n <namespace> -o json | \
     jq '.items[] | {name: .metadata.name, requests: .spec.containers[].resources.requests, limits: .spec.containers[].resources.limits}'
   ```

3. **Take action**:
   - Scale down non-critical services
   - Optimize resource requests/limits
   - Increase namespace quota (if justified)
   - Clean up unused resources

4. **Prevent recurrence**:
   - Review and adjust resource allocations
   - Implement auto-scaling where appropriate
   - Set up capacity planning alerts

## Troubleshooting

### Pod Creation Fails with Quota Exceeded

**Error**: `forbidden: exceeded quota`

**Solution**:
```bash
# Check current quota usage
kubectl describe resourcequota -n <namespace>

# View all pod resource requests
kubectl get pods -n <namespace> -o json | jq '.items[].spec.containers[].resources'

# Options:
# 1. Scale down other deployments
kubectl scale deployment <name> --replicas=<lower-number> -n <namespace>

# 2. Adjust resource requests in deployment
# 3. Request quota increase (with justification)
```

### Container Constantly OOMKilled

**Symptoms**: Pod restarts repeatedly with `OOMKilled` status

**Diagnosis**:
```bash
# Check restart count
kubectl get pods -n <namespace>

# View memory usage before crash
kubectl top pod <pod-name> -n <namespace> --containers

# Check events
kubectl describe pod <pod-name> -n <namespace>
```

**Solution**:
```bash
# Increase memory limit in deployment
kubectl edit deployment <name> -n <namespace>

# Or update Helm values
helm upgrade <release> ./chart -n <namespace> \
  --set resources.limits.memory=2Gi
```

### CPU Throttling Impacting Performance

**Symptoms**: Slow response times, high throttling rate

**Diagnosis**:
```bash
# Check CPU throttling in Grafana
# Or query Prometheus:
sum by (pod, container) (rate(container_cpu_cfs_throttled_seconds_total[5m]))
```

**Solution**:
1. Increase CPU limits (if headroom available)
2. Optimize application CPU usage
3. Scale horizontally (more pods, lower per-pod CPU)

### LimitRange Blocks Pod Creation

**Error**: `Pod exceeds LimitRange constraints`

**Diagnosis**:
```bash
# View LimitRange constraints
kubectl describe limitrange -n <namespace>

# Check pod resource specifications
kubectl get pod <pod-name> -n <namespace> -o yaml
```

**Solution**:
1. Adjust pod resources to fit within LimitRange
2. Modify LimitRange constraints (if justified)
3. Use namespace with appropriate LimitRange for workload

## Best Practices

### Resource Requests and Limits

1. **Always set both requests and limits**
   - Requests: Resources guaranteed to the pod
   - Limits: Maximum resources pod can consume

2. **Base on actual usage**
   - Monitor actual usage in production
   - Set requests at P90 usage
   - Set limits at P99 usage + 20% buffer

3. **Maintain reasonable ratios**
   - CPU limit/request ratio: ≤10x
   - Memory limit/request ratio: ≤8x

4. **Account for startup overhead**
   - Include memory for initialization
   - Consider JVM heap + non-heap memory

### Quota Management

1. **Regular review**
   - Review quota usage monthly
   - Adjust based on trends
   - Plan for growth

2. **Buffer allocation**
   - Keep 20-30% quota headroom
   - Reserve capacity for spikes
   - Plan for failover scenarios

3. **Documentation**
   - Document quota increases
   - Track historical usage
   - Justify resource requests

## Related Documentation

- [Kubernetes Scaling](./kubernetes-scaling.md)
- [Performance Tuning](./performance.md)
- [Monitoring Guide](./monitoring.md)
- [Helm Charts README](../../helm/README.md)

## References

- [Kubernetes ResourceQuotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Kubernetes LimitRanges](https://kubernetes.io/docs/concepts/policy/limit-range/)
- [Managing Resources for Containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- Issue [#853 - Application Helm Charts](https://git.subcult.tv/subculture-collective/clpr/issues/853)
- Issue [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805)
