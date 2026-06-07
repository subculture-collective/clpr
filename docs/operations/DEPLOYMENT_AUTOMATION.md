---
title: "DEPLOYMENT AUTOMATION"
summary: "This document describes the completed deployment automation features."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Deployment Automation Documentation

This document describes the completed deployment automation features.

## Overview

Three key deployment automation TODOs have been implemented:

1. **Automated Migration Execution** in blue-green deployments
2. **k6 Metrics Extraction** for benchmark reports
3. **Password Security Validation** for environment files

## 1. Automated Migration Execution

### Location
`scripts/blue-green-deploy.sh` - `run_migrations()` function

### Implementation Details

The migration execution now includes:

- **Pre-flight validation**: Checks database connectivity before running migrations
- **Migration execution**: Uses `golang-migrate` Docker container to run migrations
- **Post-migration verification**: Verifies current migration version
- **Error handling**: Returns non-zero exit code on failure for rollback

### Usage

Migrations run automatically during blue-green deployment:

```bash
./scripts/blue-green-deploy.sh
```

The script will:
1. Check if migrations directory exists
2. Validate database connectivity using `pg_isready`
3. Run migrations using the `migrate/migrate:v4.17.0` Docker image
4. Verify migration status
5. Continue with deployment if successful, or rollback if failed

### Configuration

Set these environment variables:
- `POSTGRES_USER` (default: clpr)
- `POSTGRES_PASSWORD` (required)
- `POSTGRES_DB` (default: clpr_db)
- `DEPLOY_DIR` (default: /opt/clpr)

### Example

```bash
# Load secrets securely from a secrets file (recommended)
export POSTGRES_PASSWORD="$(cat /run/secrets/postgres_password)"
export DEPLOY_DIR="/opt/clpr"
./scripts/blue-green-deploy.sh
```

## 2. k6 Metrics Extraction

### Location
`backend/tests/load/run_all_benchmarks.sh` - `extract_k6_metrics()` function

### Implementation Details

The metrics extraction includes:

- **JSON parsing**: Uses `jq` to parse k6 JSON output
- **Log fallback**: Falls back to parsing console log output if JSON unavailable
- **Metrics extracted**:
  - p50, p95, p99 response times
  - Error rate percentage
  - Requests per second (RPS)
  - Cache hit rate percentage
- **Report generation**: Populates benchmark report table with actual metrics

### Usage

Run all benchmarks:

```bash
cd backend/tests/load
./run_all_benchmarks.sh [output_dir]
```

The script will:
1. Run all k6 benchmark scripts
2. Capture JSON and log output
3. Extract metrics from output
4. Generate a consolidated markdown report with metrics table
5. Exit with code 1 if any benchmarks fail (for CI integration)

### Metrics Table Format

```markdown
| Endpoint | Status | p50 | p95 | p99 | Error Rate | RPS | Cache Hit % |
|----------|--------|-----|-----|-----|------------|-----|-------------|
| feed_list | PASS | 18.23ms | 67.45ms | 142.89ms | 0.32% | 52.34 | 73.21% |
```

### CI Integration

The load tests workflow already integrates with CI and will fail builds on benchmark regression:

```yaml
# .github/workflows/load-tests.yml
- name: Run load tests
  run: |
    cd backend/tests/load
    ./run_all_benchmarks.sh ./reports/benchmarks_${TIMESTAMP}
```

## 3. Password Security Validation

### Location
`scripts/validate-hardening.sh` - `validate_env_placeholders()` function

### Implementation Details

The password validation includes:

- **Placeholder validation**: Ensures example files use `CHANGEME` placeholders
- **Secret detection**: Identifies variables containing passwords, secrets, keys, tokens
- **False positive filtering**: Excludes timing/config variables (EXPIRY, TTL, TIMEOUT, etc.)
- **Production validation**: Checks actual production files for remaining placeholders
- **Multi-file support**: Validates both production and staging examples

### Usage

Run hardening validation:

```bash
./scripts/validate-hardening.sh
```

The script will:
1. Check example files have CHANGEME placeholders
2. Validate no hardcoded secrets in examples
3. Check actual production files (if present) for CHANGEME markers
4. Exit with code 1 if critical issues found

### CI Integration

The hardening validation runs automatically in CI:

```yaml
# .github/workflows/ci.yml
hardening-validation:
  name: Production Hardening Validation
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    
    - name: Run hardening validation
      run: |
        chmod +x scripts/validate-hardening.sh
        ./scripts/validate-hardening.sh
    
    - name: Check for TODO comments in deployment scripts
      run: |
        if grep -r "TODO" scripts/blue-green-deploy.sh scripts/validate-hardening.sh backend/tests/load/run_all_benchmarks.sh 2>/dev/null; then
          echo "ERROR: Found TODO comments in deployment automation scripts"
          exit 1
        fi
```

### Validation Rules

1. **Example files** (.env.*.example):
   - Must contain CHANGEME for sensitive variables
   - Should not have hardcoded secrets
   - Timing/configuration values are allowed

2. **Production files** (.env.production, .env):
   - Must NOT contain CHANGEME placeholders
   - Will fail validation if CHANGEME found
   - Only checked when file exists or ENVIRONMENT=production

3. **Excluded variables**:
   - Variables ending in: EXPIRY, EXPIRES, TTL, TIMEOUT, DURATION, INTERVAL
   - Variables ending in: MAX, MIN, LIMIT, PORT, HOST, URL
   - Boolean flags: ENABLE, DISABLE

## Testing

Run the comprehensive test suite:

```bash
./scripts/test-deployment-automation.sh
```

Tests include:
- Shell script syntax validation
- TODO comment detection
- Migration implementation verification
- k6 metrics extraction functionality
- Password validation logic

## Benefits

### Zero-Downtime Deployments
- Migrations run automatically before traffic switch
- Database validated before migration execution
- Automatic rollback on failure

### Performance Monitoring
- Detailed metrics in benchmark reports
- Easy identification of performance regressions
- CI integration prevents bad deployments

### Security Hardening
- Prevents accidental deployment with placeholder secrets
- Validates environment configuration
- CI gates protect production

## Troubleshooting

### Migration Failures

If migrations fail:

```bash
# Check database connectivity
docker exec clpr-postgres pg_isready -U clpr

# Check migration status
# Note: Use environment variable for DB URL to avoid exposing password in process list
export CLIPPER_DB_URL="postgresql://clpr:YOUR_PASSWORD@postgres:5432/clpr_db?sslmode=disable"
docker run --rm --network clpr-network \
  -v /opt/clpr/backend/migrations:/migrations:ro \
  -e DATABASE_URL="$CLIPPER_DB_URL" \
  migrate/migrate:v4.17.0 \
  -path /migrations \
  -database "$DATABASE_URL" \
  version
```

### k6 Metrics Missing

If metrics aren't extracted:

1. Ensure `jq` is installed: `apt-get install jq`
2. Check k6 output files exist in report directory
3. Verify log file contains summary output
4. Check function logs for errors

### Validation False Positives

If validation incorrectly flags configuration:

1. Check if variable name matches exclude patterns
2. Update exclude patterns in `validate_env_placeholders()` if needed
3. Numeric values and booleans are automatically excluded

## Related Files

- `scripts/blue-green-deploy.sh` - Main deployment script
- `scripts/validate-hardening.sh` - Security validation
- `backend/tests/load/run_all_benchmarks.sh` - Benchmark runner
- `scripts/test-deployment-automation.sh` - Test suite
- `.github/workflows/ci.yml` - CI integration
- `.github/workflows/load-tests.yml` - Load test workflow

## See Also

- [Blue-Green Deployment Guide](../docs/BLUE_GREEN_DEPLOYMENT.md)
- [Load Testing Guide](../backend/tests/load/README.md)
- [Production Hardening Checklist](../docs/PRODUCTION_HARDENING.md)
