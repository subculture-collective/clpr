---
title: "CDN FAILOVER TESTING"
summary: "This guide provides quick commands to run the CDN failover tests."
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-30
---

# CDN Failover Testing Quick Start

This guide provides quick commands to run the CDN failover tests.

## Status

✅ **Configuration Complete** - CDN failover tests are now fully configured and ready to run.

- Backend CDN service with failover support: ✅ Implemented
- Frontend E2E tests: ✅ Ready (5 test suites, 13 tests)
- Test environment configuration: ✅ Complete
- Documentation: ✅ Updated

The CDN failover tests validate that assets (images, videos, thumbnails) gracefully fall back to origin servers when CDN endpoints fail or timeout.

## Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Node.js 18+
- k6 load testing tool
- PostgreSQL and Redis (via docker-compose.test.yml)

## Setup Test Environment

```bash
# Start test infrastructure (PostgreSQL, Redis, OpenSearch)
make test-setup

# This will:
# - Start Docker containers for test databases
# - Run database migrations
# - Seed test data
```

## Backend Integration Tests

Test CDN failover behavior at the API level:

```bash
# Navigate to backend directory
cd backend

# Run all CDN failover tests
go test -v -tags=integration ./tests/integration/cdn/...

# Run specific test
go test -v -tags=integration ./tests/integration/cdn/... -run TestCDNFailover_StaticAssets

# Run with coverage
go test -v -tags=integration ./tests/integration/cdn/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**What's tested:**
- Static asset failover (images, thumbnails)
- HLS playlist failover
- HLS segment failover
- Retry/backoff behavior
- Cache header validation
- Failover metrics and headers

## Load Tests (k6)

Test system behavior under sustained CDN failures:

```bash
# Enable failover mode
export CDN_FAILOVER_MODE=true

# Run CDN failover load test
k6 run backend/tests/load/scenarios/cdn_failover.js

# Run with custom configuration
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e CDN_FAILOVER_MODE=true \
  backend/tests/load/scenarios/cdn_failover.js

# Run with verbose output
k6 run --verbose backend/tests/load/scenarios/cdn_failover.js
```

**What's tested:**
- 40% static asset requests
- 30% HLS playlist requests
- 25% HLS segment requests
- 5% mixed asset requests
- Request storm prevention
- Alert threshold validation
- Throughput under failover

## Frontend E2E Tests (Playwright)

Test user-facing behavior during CDN failures:

```bash
# Navigate to frontend directory
cd frontend

# Run all CDN failover E2E tests
npm run test:e2e -- cdn-failover.spec.ts

# Run with failover mode enabled
# This enables CDN failover simulation in the test environment
E2E_CDN_FAILOVER_MODE=true npm run test:e2e -- cdn-failover.spec.ts

# Run in headed mode (see browser)
npm run test:e2e -- cdn-failover.spec.ts --headed

# Run in UI mode (debugging)
npm run test:e2e:ui -- cdn-failover.spec.ts

# Run specific test
npm run test:e2e -- cdn-failover.spec.ts -g "should load thumbnails from origin"
```

**Note**: CDN failover E2E tests run using the standard E2E test commands, but CDN failover simulation is **opt-in**. To enable failover mode, you must explicitly set the `E2E_CDN_FAILOVER_MODE` environment variable to `true` (see the example above); only then does the Playwright configuration pass the corresponding flag to the dev server.

**What's tested:**
- Thumbnail loading from origin
- User avatar loading from origin
- Broken image handling
- HLS video playback from origin
- Video stall and resume
- Loading states
- UI responsiveness
- Page navigation during failure

## Local Testing with Caddy Simulation

Simulate CDN failures locally using Caddy:

```bash
# Terminal 1: Start Caddy with test configuration
caddy run --config Caddyfile.cdn-test

# Terminal 2: Run your application
make dev

# Terminal 3: Run E2E tests
cd frontend
npm run test:e2e -- cdn-failover.spec.ts
```

## Staging Environment Testing

```bash
# Deploy to staging with failover testing enabled
export CDN_FAILOVER_MODE=true
docker compose -f docker-compose.staging.yml up -d

# Run tests against staging
PLAYWRIGHT_BASE_URL=https://staging.example.com \
  E2E_CDN_FAILOVER_MODE=true \
  npm run test:e2e -- cdn-failover.spec.ts

# Run load tests against staging
k6 run \
  -e BASE_URL=https://staging.example.com \
  -e CDN_FAILOVER_MODE=true \
  backend/tests/load/scenarios/cdn_failover.js
```

## Cleanup

```bash
# Stop test infrastructure
docker compose -f docker-compose.test.yml down

# Remove test volumes (clean slate)
docker compose -f docker-compose.test.yml down -v
```

## Verifying Test Results

### Backend Tests
Look for:
- ✅ All tests pass
- ✅ Failover headers present (`X-CDN-Failover: true`)
- ✅ Response status codes are 200 (successful fallback)
- ✅ No panics or crashes

### Load Tests
Look for:
- ✅ Fallback rate > 70% (when CDN_FAILOVER_MODE=true)
- ✅ Error rate < 5%
- ✅ Request storm rate < 1%
- ✅ P95 latency < 1000ms for assets
- ✅ P95 latency < 500ms for segments

### E2E Tests
Look for:
- ✅ All tests pass in Chromium, Firefox, and WebKit
- ✅ Images load successfully
- ✅ Videos play without errors
- ✅ UI remains interactive
- ✅ No console errors

## Monitoring Failover in Production

```bash
# Check failover metrics in Prometheus
curl 'http://prometheus:9090/api/v1/query?query=rate(cdn_failover_total[5m])'

# View failover by reason
curl 'http://prometheus:9090/api/v1/query?query=sum by (reason) (rate(cdn_failover_total[5m]))'

# Check fallback latency
curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, sum(rate(cdn_failover_duration_ms_bucket[5m])) by (le))'

# Check active alerts
curl 'http://prometheus:9090/api/v1/alerts' | jq '.data.alerts[] | select(.labels.alertname | contains("CDN"))'
```

## Troubleshooting

### Tests Fail to Connect to Database
```bash
# Check database is running
docker ps | grep postgres

# Check database is ready
pg_isready -h localhost -p 5437 -U clpr

# Restart database
docker compose -f docker-compose.test.yml restart postgres
```

### Frontend Tests Timeout
```bash
# Increase timeout in playwright.config.ts
timeout: 60 * 1000,  # 60 seconds instead of 30

# Or run with specific timeout
npm run test:e2e -- cdn-failover.spec.ts --timeout=60000
```

### Load Test Errors
```bash
# Check base URL is accessible
curl http://localhost:8080/health

# Reduce concurrent users
# Edit cdn_failover.js stages to use lower targets
{ duration: '2m', target: 5 },   # Reduced from 30
```

## Related Documentation

- [CDN Failover Runbook](docs/operations/CDN_FAILOVER_RUNBOOK.md) - Operational procedures
- [Testing Guide](docs/testing/TESTING.md) - Full testing documentation
- [Backend HLS Implementation](docs/archive/BACKEND_HLS_IMPLEMENTATION.md) - HLS streaming details

## Support

If you encounter issues:
1. Check the [CDN Failover Runbook](docs/operations/CDN_FAILOVER_RUNBOOK.md)
2. Review test logs for error details
3. Verify test infrastructure is running
4. Check that environment variables are set correctly
