---
title: "HPA Scaling"
summary: "This runbook covers Horizontal Pod Autoscaler (HPA) operations, troubleshooting, and best practices for the Clipper backend and frontend services."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# HPA Scaling Operations Runbook

## Overview

This runbook covers Horizontal Pod Autoscaler (HPA) operations, troubleshooting, and best practices for the Clipper backend and frontend services.

## HPA Configuration

### Backend Service
- **Min Replicas**: 2 (staging), 3 (production)
- **Max Replicas**: 10 (default), 20 (production)
- **Metrics**:
  - CPU: 70% target utilization
  - Memory: 80% target utilization
  - Custom: ~1000 requests/second per pod
- **Scale-down stabilization**: 300 seconds (5 minutes)

### Frontend Service
- **Min Replicas**: 2 (staging), 3 (production)
- **Max Replicas**: 8 (default), 8 (production)
- **Metrics**:
  - CPU: 70% target utilization
  - Memory: 80% target utilization
  - Custom: ~1000 requests/second per pod
- **Scale-down stabilization**: 300 seconds (5 minutes)

## Prerequisites

- Metrics Server installed and running
- Prometheus Adapter configured for custom metrics
- Prometheus scraping application metrics
- kubectl access to the cluster

## Common Operations

### Check HPA Status

```bash
# List all HPAs
kubectl get hpa -n clpr-production

# Detailed HPA status
kubectl describe hpa clpr-backend -n clpr-production
kubectl describe hpa clpr-frontend -n clpr-production

# Watch HPA in real-time
kubectl get hpa -n clpr-production -w
```

### Check Current Metrics

```bash
# Resource metrics (CPU/Memory)
kubectl top pods -n clpr-production

# Custom metrics via Prometheus Adapter
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/clpr-production/pods/*/http_requests_per_second" | jq .

# Check metrics-server health
kubectl get apiservice v1beta1.metrics.k8s.io

# Check prometheus-adapter health
kubectl get apiservice v1beta1.custom.metrics.k8s.io
```

### Manual Scaling

```bash
# Temporarily override HPA (not recommended for production)
kubectl scale deployment clpr-backend -n clpr-production --replicas=5

# Re-enable HPA
kubectl patch hpa clpr-backend -n clpr-production -p '{"spec":{"minReplicas":3}}'
```

## Troubleshooting

### HPA Maxed Out

**Alert**: `HPAMaxedOut`  
**Symptom**: HPA has been at maximum replicas for 15+ minutes.

**Investigation**:
```bash
# Check current load
kubectl top pods -n clpr-production -l app.kubernetes.io/name=clpr-backend

# Check custom metrics
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/clpr-production/pods/*/http_requests_per_second" | jq .

# Review recent scaling events
kubectl describe hpa clpr-backend -n clpr-production | grep -A 20 "Events:"
```

**Resolution**:
1. If sustained high load, increase `maxReplicas` in values file
2. Review application performance - may indicate need for optimization
3. Check for traffic anomalies (DDoS, bot traffic)
4. Consider vertical scaling (increase pod resources)

**Update maxReplicas**:
```bash
# Edit HPA directly (temporary)
kubectl edit hpa clpr-backend -n clpr-production

# Or update values and redeploy (permanent)
# Edit helm/charts/backend/values.yaml
# Then: helm upgrade clpr-backend ./helm/charts/backend
```

### HPA Unable to Scale

**Alert**: `HPAUnableToScale`  
**Symptom**: HPA is unable to scale due to constraints.

**Investigation**:
```bash
# Check HPA conditions
kubectl describe hpa clpr-backend -n clpr-production

# Check node capacity
kubectl top nodes

# Check resource quotas
kubectl describe resourcequota -n clpr-production

# Check pod scheduling
kubectl get pods -n clpr-production -o wide
kubectl describe nodes
```

**Common Causes**:
1. **Insufficient node capacity**: Nodes don't have enough CPU/memory
2. **Resource quotas**: Namespace quota exceeded
3. **Pod affinity/anti-affinity**: Can't find suitable nodes
4. **PodDisruptionBudget**: Preventing scale-down

**Resolution**:
```bash
# Scale cluster (GKE example)
gcloud container clusters resize clpr-prod --num-nodes=5

# Or enable cluster autoscaler
gcloud container clusters update clpr-prod --enable-autoscaling --min-nodes=3 --max-nodes=10

# Check and adjust resource quotas if needed
kubectl get resourcequota -n clpr-production
```

### HPA Metrics Unavailable

**Alert**: `HPAMetricsUnavailable`  
**Symptom**: HPA cannot obtain metrics for scaling decisions.

**Investigation**:
```bash
# Check metrics-server
kubectl get deployment metrics-server -n kube-system
kubectl logs -n kube-system deployment/metrics-server

# Check prometheus-adapter
kubectl get deployment prometheus-adapter -n custom-metrics
kubectl logs -n custom-metrics deployment/prometheus-adapter

# Test metrics API
kubectl top nodes
kubectl top pods -n clpr-production
```

**Resolution**:

1. **Metrics Server Down**:
```bash
# Restart metrics-server
kubectl rollout restart deployment/metrics-server -n kube-system

# Verify it's running
kubectl get pods -n kube-system -l k8s-app=metrics-server
```

2. **Prometheus Adapter Issues**:
```bash
# Check Prometheus connectivity
kubectl exec -it -n custom-metrics deployment/prometheus-adapter -- wget -O- http://clpr-monitoring-prometheus-server.clpr-monitoring.svc:9090/-/healthy

# Restart prometheus-adapter
kubectl rollout restart deployment/prometheus-adapter -n custom-metrics

# Check logs for configuration errors
kubectl logs -n custom-metrics deployment/prometheus-adapter --tail=100
```

3. **Application not exposing metrics**:
```bash
# Check if app exposes metrics
kubectl exec -it -n clpr-production deployment/clpr-backend -- wget -O- http://localhost:8080/debug/metrics

# Verify service monitor/scrape config
kubectl get servicemonitor -n clpr-production
```

### HPA Frequent Scaling

**Alert**: `HPAFrequentScaling`  
**Symptom**: HPA is changing desired replicas frequently (flapping).

**Causes**:
- Oscillating load patterns
- Thresholds set too low
- Insufficient scale-down stabilization
- Bursty traffic

**Investigation**:
```bash
# Monitor HPA decisions over time
kubectl get hpa clpr-backend -n clpr-production -w

# Check metrics trend
# (Use Grafana dashboard or Prometheus)

# Review HPA behavior configuration
kubectl get hpa clpr-backend -n clpr-production -o yaml | grep -A 20 behavior
```

**Resolution**:
1. **Adjust thresholds**: Increase CPU/memory targets or RPS target
2. **Increase stabilization window**:
```yaml
behavior:
  scaleDown:
    stabilizationWindowSeconds: 600  # Increase from 300
```
3. **Add scale-up delay**:
```yaml
behavior:
  scaleUp:
    stabilizationWindowSeconds: 60  # Add delay
```

### Replica Mismatch

**Alert**: `HPADesiredReplicasMismatch`  
**Symptom**: Desired replicas don't match current replicas for 15+ minutes.

**Investigation**:
```bash
# Check pod status
kubectl get pods -n clpr-production -l app.kubernetes.io/name=clpr-backend

# Check deployment rollout
kubectl rollout status deployment/clpr-backend -n clpr-production

# Look for scheduling issues
kubectl get events -n clpr-production --sort-by='.lastTimestamp' | head -20

# Check for failing pods
kubectl get pods -n clpr-production | grep -v Running
```

**Common Causes**:
1. Pods stuck in Pending (resource constraints)
2. Image pull failures
3. Crashing pods (CrashLoopBackOff)
4. Node pressure (disk, memory, PID)

**Resolution**: Address underlying pod scheduling/startup issues.

### Custom Metrics Not Available

**Alert**: `HPACustomMetricsNotAvailable`  
**Symptom**: Prometheus Adapter is down; HPA falls back to CPU/memory only.

**Investigation**:
```bash
# Check prometheus-adapter status
kubectl get pods -n custom-metrics
kubectl logs -n custom-metrics deployment/prometheus-adapter

# Verify Prometheus connectivity
kubectl port-forward -n clpr-monitoring svc/clpr-monitoring-prometheus-server 9090:9090
# Then visit http://localhost:9090 and check targets

# Test custom metrics API
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq .
```

**Resolution**:
```bash
# Restart prometheus-adapter
kubectl rollout restart deployment/prometheus-adapter -n custom-metrics

# If configuration issue, update ConfigMap
kubectl edit configmap prometheus-adapter-config -n custom-metrics
kubectl rollout restart deployment/prometheus-adapter -n custom-metrics

# Verify metrics are available
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/clpr-production/pods/*/http_requests_per_second" | jq .
```

### Metrics Server Down

**Alert**: `MetricsServerDown`  
**Symptom**: Metrics Server unavailable; HPA cannot function.

**Investigation**:
```bash
# Check metrics-server
kubectl get pods -n kube-system -l k8s-app=metrics-server
kubectl logs -n kube-system deployment/metrics-server

# Check API service
kubectl get apiservice v1beta1.metrics.k8s.io
```

**Resolution**:
```bash
# Restart metrics-server
kubectl rollout restart deployment/metrics-server -n kube-system

# If still failing, check node connectivity
kubectl describe nodes

# Re-deploy if necessary
kubectl apply -f infrastructure/k8s/base/metrics-server.yaml
```

### Slow Scale-Up

**Alert**: `BackendHighUtilizationBeforeScaling` or `FrontendHighUtilizationBeforeScaling`  
**Symptom**: Pods exceeding target threshold but HPA not scaling.

**Investigation**:
```bash
# Check HPA evaluation interval (default: 15s)
kubectl describe hpa clpr-backend -n clpr-production

# Check metric collection lag
kubectl top pods -n clpr-production

# Review HPA events
kubectl get events -n clpr-production --field-selector involvedObject.name=clpr-backend
```

**Causes**:
- Metrics lag (stale data)
- HPA stabilization window preventing scale-up
- Multiple metrics: HPA uses highest recommendation

**Resolution**:
1. Reduce stabilization window for scale-up (already at 0s by default)
2. Lower thresholds to trigger scaling earlier
3. Pre-scale before known high-load periods

## Best Practices

### 1. Set Appropriate Resource Requests

HPA uses resource requests as the baseline for percentage-based metrics.

```yaml
resources:
  requests:
    cpu: 250m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

### 2. Use Multiple Metrics

Combine CPU, memory, and custom metrics for better scaling decisions:

```yaml
autoscaling:
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
  targetRequestsPerSecond: "1000"
```

### 3. Configure Scale-Down Stabilization

Prevent flapping by adding stabilization:

```yaml
behavior:
  scaleDown:
    stabilizationWindowSeconds: 300  # Wait 5 minutes before scaling down
```

### 4. Monitor HPA Metrics

- Set up Grafana dashboards for HPA metrics
- Alert on HPA at max replicas
- Track scaling events over time

### 5. Load Testing

Before production:
```bash
# Example with k6 or hey
hey -z 60s -c 100 https://clpr.tv/api/v1/clips

# Monitor HPA response
kubectl get hpa -n clpr-production -w
```

### 6. Plan for Peak Traffic

- Set minReplicas higher during known peak periods
- Use scheduled scaling (external tool like kube-schedule-scaler)
- Pre-warm cache before traffic spikes

## Metrics Reference

### Resource Metrics (via Metrics Server)

- `container_cpu_usage_seconds_total`: CPU usage
- `container_memory_working_set_bytes`: Memory usage

### Custom Metrics (via Prometheus Adapter)

- `http_requests_per_second`: Rate of HTTP requests per pod
- Configure additional metrics in `prometheus-adapter-config` ConfigMap

## Related Documentation

- [Kubernetes HPA Documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server)
- [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter)
- [Clipper Monitoring Setup](./monitoring.md)

## Support

For HPA-related issues:
1. Check this runbook
2. Review HPA status and events
3. Check monitoring dashboards
4. Escalate to infrastructure team if unresolved
