---
title: "CDN FAILOVER RUNBOOK"
summary: "This document describes CDN failover behavior, testing procedures, and operational runbooks for handling CDN outages in the Clipper application."
tags: ["operations","runbook"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# CDN Failover Testing and Operations Runbook

## Overview

This document describes CDN failover behavior, testing procedures, and operational runbooks for handling CDN outages in the Clipper application.

## Architecture

### CDN Providers

Clipper supports multiple CDN providers:
- **Cloudflare** (primary recommended)
- **Bunny CDN** (alternative)
- **AWS CloudFront** (enterprise)

### Failover Strategy

When the CDN is unavailable (5xx errors, timeouts, DNS failures), requests automatically fall back to the origin server:

```
Client Request → CDN (primary) 
                 ↓ (on failure)
                 Origin Server (fallback)
```

### Assets Covered

1. **Static Assets**
   - Thumbnails (.jpg, .png, .webp)
   - User avatars
   - JavaScript bundles
   - CSS files
   - Images

2. **HLS Streaming**
   - Master playlists (.m3u8)
   - Quality variant playlists
   - Media segments (.ts files)

## Configuration

### Environment Variables

```bash
# Enable CDN
CDN_ENABLED=true

# CDN Provider
CDN_PROVIDER=cloudflare  # or bunny, aws-cloudfront

# Cloudflare
CLOUDFLARE_ZONE_ID=your_zone_id
CLOUDFLARE_API_KEY=your_api_key

# Bunny CDN
BUNNY_API_KEY=your_api_key
BUNNY_STORAGE_ZONE=your_storage_zone

# AWS CloudFront
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_CLOUDFRONT_DISTRIBUTION_ID=your_distribution_id

# Failover Configuration
CDN_FAILOVER_ENABLED=true
CDN_FAILOVER_RETRY_COUNT=3
CDN_FAILOVER_RETRY_DELAY=100  # milliseconds
CDN_FAILOVER_TIMEOUT=5000      # milliseconds
```

### Caddyfile Configuration

To enable CDN failover in production, configure Caddy to try CDN first, then origin:

```caddyfile
handle /api/v1/clips/*/thumbnail {
    @cdn_fail expression {
        {http.error.status_code} >= 500
    }
    
    try_files {
        try cdn.example.com{uri}
        try origin.example.com{uri}
    }
    
    header @cdn_fail X-CDN-Failover "true"
    header @cdn_fail X-CDN-Failover-Reason "error"
}
```

### Nginx Configuration

Alternative configuration for Nginx:

```nginx
location ~ ^/api/v1/clips/.*/thumbnail$ {
    proxy_pass https://cdn.example.com;
    proxy_next_upstream error timeout http_502 http_503 http_504;
    proxy_next_upstream_tries 2;
    
    # Fallback to origin
    error_page 502 503 504 = @origin_fallback;
}

location @origin_fallback {
    proxy_pass https://origin.example.com;
    add_header X-CDN-Failover "true";
    add_header X-CDN-Failover-Reason "error";
}
```

## Testing

### Running Tests

#### Backend Integration Tests

```bash
# Start test infrastructure
make test-setup

# Run CDN failover tests
cd backend
go test -v -tags=integration ./tests/integration/cdn/...

# Cleanup
docker compose -f docker-compose.test.yml down
```

#### Frontend E2E Tests

```bash
cd frontend

# Run CDN failover E2E tests
npm run test:e2e -- cdn-failover.spec.ts

# Run with failover mode enabled
E2E_CDN_FAILOVER_MODE=true npm run test:e2e -- cdn-failover.spec.ts

# Run in headed mode for debugging
npm run test:e2e -- cdn-failover.spec.ts --headed
```

#### Load Tests

```bash
# Set failover mode
export CDN_FAILOVER_MODE=true

# Run CDN failover load test
k6 run backend/tests/load/scenarios/cdn_failover.js

# With custom base URL
k6 run -e BASE_URL=http://staging:8080 -e CDN_FAILOVER_MODE=true \
  backend/tests/load/scenarios/cdn_failover.js
```

### Test Environments

#### Local Testing with Caddy

```bash
# Start test CDN and origin servers
caddy run --config Caddyfile.cdn-test

# In another terminal, run your tests
npm run test:e2e -- cdn-failover.spec.ts
```

#### Staging Environment

```bash
# Deploy to staging with failover testing enabled
export CDN_FAILOVER_MODE=true
docker compose -f docker-compose.staging.yml up -d

# Run tests against staging
E2E_CDN_FAILOVER_MODE=true PLAYWRIGHT_BASE_URL=https://staging.example.com \
  npm run test:e2e -- cdn-failover.spec.ts
```

## Observability

### Metrics

The following Prometheus metrics are emitted:

```promql
# Total CDN failover events
cdn_failover_total{reason="timeout|error|dns_failure"}

# Failover latency
cdn_failover_duration_ms

# CDN request rate
cdn_requests_total{provider="cloudflare|bunny|aws"}

# CDN error rate
cdn_errors_total{provider="cloudflare|bunny|aws", status_code="5xx"}
```

### Logs

CDN failover events are logged with structured logging:

```json
{
  "level": "warn",
  "msg": "CDN failover triggered",
  "cdn_provider": "cloudflare",
  "failover_reason": "timeout",
  "asset_type": "thumbnail",
  "clip_id": "abc123",
  "latency_ms": 250,
  "retry_count": 2
}
```

### Alerts

#### CDN Failover Rate High

Triggers when CDN failover rate exceeds 5 requests/second:

```yaml
alert: CDNFailoverRateHigh
expr: rate(cdn_failover_total[5m]) > 5
for: 5m
labels:
  severity: warning
annotations:
  summary: "CDN failover rate is high"
  description: "CDN is experiencing {{ $value }} failovers/sec for 5 minutes"
```

#### CDN Failover Rate Critical

Triggers when CDN failover rate exceeds 20 requests/second:

```yaml
alert: CDNFailoverRateCritical
expr: rate(cdn_failover_total[5m]) > 20
for: 2m
labels:
  severity: critical
annotations:
  summary: "CDN failover rate is critical"
  description: "CDN is experiencing {{ $value }} failovers/sec - immediate action required"
```

#### CDN Failover Latency High

Triggers when P95 failover latency exceeds 500ms:

```yaml
alert: CDNFailoverLatencyHigh
expr: histogram_quantile(0.95, rate(cdn_failover_duration_ms_bucket[5m])) > 500
for: 10m
labels:
  severity: warning
annotations:
  summary: "CDN failover latency is high"
  description: "P95 failover latency is {{ $value }}ms"
```

## Operational Runbooks

### CDN Outage Detected

#### Symptoms
- High CDN failover rate (> 5/sec)
- Increased origin server traffic
- Slower asset load times
- Alert: CDNFailoverRateHigh or CDNFailoverRateCritical

#### Investigation Steps

1. **Check CDN Status**
   ```bash
   # Check Cloudflare status
   curl -X GET "https://api.cloudflare.com/client/v4/zones/{zone_id}" \
     -H "Authorization: Bearer {api_token}"
   
   # Check Bunny CDN status
   curl -X GET "https://api.bunny.net/storagezone/{storage_zone}" \
     -H "AccessKey: {api_key}"
   ```

2. **Verify Metrics**
   ```promql
   # Current failover rate
   rate(cdn_failover_total[5m])
   
   # Failover reasons breakdown
   sum by (reason) (rate(cdn_failover_total[5m]))
   
   # Origin server load
   rate(http_requests_total{source="origin"}[5m])
   ```

3. **Check Logs**
   ```bash
   # Tail CDN failover logs
   kubectl logs -l app=clpr-backend --tail=100 | grep "cdn_failover"
   
   # Count recent failovers
   kubectl logs -l app=clpr-backend --since=10m | grep "cdn_failover" | wc -l
   ```

#### Resolution Steps

1. **If CDN Provider is Down**
   - Origin is already serving traffic (failover active)
   - Monitor origin server capacity
   - Consider scaling origin if needed
   - Contact CDN provider support
   - Update status page for users

2. **If Configuration Issue**
   - Review recent CDN configuration changes
   - Check DNS settings
   - Verify SSL/TLS certificates
   - Check firewall rules

3. **If Origin Server Overloaded**
   - Scale origin horizontally (add more instances)
   - Enable additional caching layers
   - Consider temporary rate limiting

### Disable CDN (Emergency)

If CDN is causing issues, temporarily disable it:

```bash
# Set environment variable
export CDN_ENABLED=false

# Restart application
kubectl rollout restart deployment/clpr-backend

# Or update ConfigMap
kubectl edit configmap clpr-config
# Set CDN_ENABLED: "false"

# Restart pods to pick up new config
kubectl rollout restart deployment/clpr-backend
```

### Enable CDN After Outage

Once CDN provider is healthy:

```bash
# Re-enable CDN
export CDN_ENABLED=true

# Restart application
kubectl rollout restart deployment/clpr-backend

# Monitor metrics for any issues
watch -n 5 'curl -s http://localhost:9090/metrics | grep cdn_'
```

### Purge CDN Cache

If serving stale content:

```bash
# Via API
curl -X POST http://localhost:8080/api/v1/admin/cdn/purge \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"clip_id": "abc123"}'

# Purge entire cache (use cautiously)
curl -X POST http://localhost:8080/api/v1/admin/cdn/purge-all \
  -H "Authorization: Bearer {token}"
```

## Performance Expectations

### Normal Operation (CDN Serving)
- Asset load time: < 100ms (P95)
- HLS playlist load: < 200ms (P95)
- HLS segment load: < 100ms (P95)
- Cache hit rate: > 95%

### Failover Mode (Origin Serving)
- Asset load time: < 500ms (P95)
- HLS playlist load: < 500ms (P95)
- HLS segment load: < 300ms (P95)
- Error rate: < 5%

### Recovery Expectations
- Automatic failover: < 1 second
- Player stall duration: < 2 seconds
- Full recovery time: < 5 minutes

## Best Practices

1. **Always have origin capacity** to handle 100% of traffic during CDN failures
2. **Monitor CDN health** continuously with synthetic checks
3. **Test failover regularly** (monthly) to ensure it works
4. **Set appropriate cache TTLs**:
   - Static assets: 1 hour - 1 day
   - HLS playlists: 1 minute
   - HLS segments: 1 day
5. **Use versioned URLs** for cache busting when needed
6. **Implement retry logic** with exponential backoff (max 3 retries)
7. **Log all failover events** for post-incident analysis

## Related Documentation

- [Testing Guide](testing/TESTING.md)
- [Search Failover Tests](testing/TESTING.md#search-failover-tests)

## Related Issues

- [#689 Watch Parties Sync](https://git.subcult.tv/subculture-collective/clpr/issues/689)
- [#694 Chat/WebSocket Backend](https://git.subcult.tv/subculture-collective/clpr/issues/694)
