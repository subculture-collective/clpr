---
title: "Centralized Logging"
summary: "This document describes the centralized logging infrastructure for Clipper."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Centralized Logging

This document describes the centralized logging infrastructure for Clipper.

## Overview

Clipper uses structured JSON logging with centralized aggregation via Grafana Loki. All services (backend, frontend, mobile) emit structured logs that are collected, indexed, and made searchable through a unified interface.

## Log Levels

All services use consistent log levels:

- **DEBUG**: Detailed information for debugging (development only)
- **INFO**: General informational messages
- **WARN**: Warning messages for potentially harmful situations
- **ERROR**: Error messages for failures that don't stop the application
- **FATAL**: Critical errors that cause application shutdown (backend only)

## Log Structure

All logs follow a consistent JSON structure:

```json
{
  "timestamp": "2024-12-15T19:00:00Z",
  "level": "info",
  "message": "HTTP Request",
  "service": "clpr-backend",
  "trace_id": "abc123-def456",
  "user_id": "hashed_user_id",
  "method": "GET",
  "path": "/api/v1/clips",
  "status_code": 200,
  "latency": "25ms",
  "fields": {
    "custom_field": "value"
  }
}
```

### Standard Fields

- `timestamp`: ISO 8601 timestamp in UTC
- `level`: Log level (debug, info, warn, error, fatal)
- `message`: Human-readable log message
- `service`: Service name (clpr-backend, clpr-frontend, clpr-mobile)
- `trace_id`: Request/trace ID for correlation
- `user_id`: Hashed user ID (for privacy)
- `error`: Error message (if applicable)
- `stack`: Stack trace (for errors)
- `fields`: Additional contextual fields

## Backend Logging (Go)

### Using the Logger

```go
import "git.subcult.tv/subculture-collective/clpr/pkg/utils"

// Initialize logger (done in main.go)
utils.InitLogger(utils.LogLevelInfo)

// Log messages
utils.Info("User logged in", map[string]interface{}{
    "user_id": userID,
})

utils.Error("Failed to fetch clips", err, map[string]interface{}{
    "query": queryParams,
})

utils.Fatal("Database connection failed", err, nil)
```

### HTTP Request Logging

The structured logger includes Gin middleware that automatically logs all HTTP requests:

```go
router.Use(logger.GinLogger())
```

This logs:
- HTTP method, path, and query parameters
- Response status code
- Request latency
- User ID (hashed)
- Trace ID (request ID)
- Client IP and User-Agent
- Errors (if any)

## Frontend Logging (TypeScript)

### Using the Logger

```typescript
import { info, error, warn, debug } from '@/lib/logger';

// Log messages
info('User navigated to clips page', { page: 'clips' });

try {
  // ... code
} catch (err) {
  error('Failed to load clips', err as Error, { 
    filter: activeFilter 
  });
}
```

### Initialization

Initialize the logger in your app entry point:

```typescript
import { initLogger, LogLevel } from '@/lib/logger';

// In development
initLogger(LogLevel.DEBUG);

// In production
initLogger(LogLevel.INFO);
```

## Mobile Logging (React Native)

### Using the Logger

```typescript
import { info, error, warn, debug } from '@/lib/logger';

// Async logging
await info('Screen mounted', { screen: 'Home' });

try {
  // ... code
} catch (err) {
  await error('Failed to sync data', err as Error, {
    syncType: 'clips'
  });
}
```

## PII Redaction

All loggers automatically redact personally identifiable information (PII):

### Redacted Data Types

- **Email addresses**: `user@example.com` → `[REDACTED_EMAIL]`
- **Phone numbers**: `555-123-4567` → `[REDACTED_PHONE]`
- **Credit card numbers**: `4111-1111-1111-1111` → `[REDACTED_CARD]`
- **SSN**: `123-45-6789` → `[REDACTED_SSN]`
- **Passwords**: `password=secret123` → `password="[REDACTED]"`
- **API tokens**: `Bearer eyJ...` → `Bearer [REDACTED_TOKEN]`

### Sensitive Field Names

Fields with these names are automatically redacted:
- password, passwd, pwd
- secret, token, api_key, apikey
- authorization, auth
- access_token, auth_token

## Log Aggregation

### Architecture

```
┌─────────────┐
│   Backend   │────┐
└─────────────┘    │
                   │
┌─────────────┐    │    ┌──────────┐    ┌──────────┐    ┌─────────┐
│  Frontend   │────┼───→│ Promtail │───→│   Loki   │───→│ Grafana │
└─────────────┘    │    └──────────┘    └──────────┘    └─────────┘
                   │
┌─────────────┐    │
│   Mobile    │────┘
└─────────────┘
```

1. **Services** emit structured JSON logs to stdout/files
2. **Promtail** collects logs from Docker containers and log files
3. **Loki** aggregates, indexes, and stores logs
4. **Grafana** provides search, filtering, and visualization

### Starting the Logging Stack

```bash
cd monitoring
docker-compose -f docker-compose.monitoring.yml up -d loki promtail grafana
```

### Accessing Logs

**Grafana Log Explorer**: <http://localhost:3000/explore>

Query examples:
```logql
# All error logs
{level="error"}

# Backend errors in the last hour
{service="clpr-backend", level="error"} [1h]

# Logs for a specific trace ID
{trace_id="abc123"}

# Failed authentication attempts
{message=~".*authentication failed.*"}

# High latency requests
{service="clpr-backend"} | json | latency > 1s
```

## Log Retention

- **Retention Period**: 90 days
- **Compaction**: Runs every 10 minutes
- **Delete Delay**: 2 hours (for safety)

Configure in `monitoring/loki-config.yml`:

```yaml
limits_config:
  retention_period: 2160h  # 90 days

compactor:
  retention_enabled: true
  retention_delete_delay: 2h
```

## Log-Based Alerts

Alertmanager monitors logs for specific patterns and triggers alerts.

### Configured Alerts

1. **HighErrorLogRate**: > 10 errors/sec
2. **CriticalErrorSpike**: > 50 errors/sec
3. **FailedAuthenticationSpike**: > 5 failed auth/sec
4. **SQLInjectionAttempt**: SQL injection patterns detected
5. **SuspiciousSecurityEvent**: Security-related warnings
6. **ApplicationPanic**: Panic/crash events
7. **NoLogsReceived**: No logs for 10 minutes
8. **DatabaseConnectionErrors**: DB connection failures
9. **RedisConnectionErrors**: Redis connection failures

View alerts in Alertmanager: <http://localhost:9093>

## Best Practices

### DO

✅ Use appropriate log levels
✅ Include contextual information in fields
✅ Log at decision points and error boundaries
✅ Use trace IDs to correlate related logs
✅ Log security-relevant events
✅ Test logging in development before production

### DON'T

❌ Log sensitive data (passwords, tokens, PII)
❌ Log in tight loops (high cardinality)
❌ Use string concatenation for structured fields
❌ Log at DEBUG level in production
❌ Include full stack traces for expected errors
❌ Log the same event multiple times

### Example: Good Logging

```go
// Good: Structured with context
utils.Info("Clip created", map[string]interface{}{
    "clip_id": clipID,
    "user_id": hashUserID(userID),
    "duration": duration,
})

// Good: Error with context
utils.Error("Failed to update clip", err, map[string]interface{}{
    "clip_id": clipID,
    "operation": "update",
})
```

### Example: Bad Logging

```go
// Bad: Not structured
log.Printf("Clip %s created by user %s", clipID, userEmail)

// Bad: PII exposure
utils.Info("User logged in", map[string]interface{}{
    "email": user.Email,  // PII!
    "password": password,  // Never log passwords!
})

// Bad: In a loop
for _, clip := range clips {
    utils.Debug("Processing clip", map[string]interface{}{"clip": clip})
}
```

## Searching Logs

### By Time Range

```logql
{service="clpr-backend"} [5m]  # Last 5 minutes
{service="clpr-backend"} [1h]  # Last hour
{service="clpr-backend"} [24h] # Last day
```

### By Log Level

```logql
{level="error"}                    # All errors
{level=~"error|fatal"}            # Errors and fatal
{level!="debug"}                  # Everything except debug
```

### By Service

```logql
{service="clpr-backend"}
{service="clpr-frontend"}
{service="clpr-mobile"}
{service=~"clpr-.*"}            # All services
```

### By Message Content

```logql
{message=~".*failed.*"}            # Contains "failed"
{message=~"(?i)error"}             # Contains "error" (case-insensitive)
{message!~".*health.*"}            # Doesn't contain "health"
```

### Using JSON Fields

```logql
# Parse JSON and filter
{service="clpr-backend"} | json | status_code >= 500

# Filter by user
{service="clpr-backend"} | json | user_id="abc123"

# Filter by latency
{service="clpr-backend"} | json | latency > 100ms
```

### Aggregations

```logql
# Count logs by level
sum(count_over_time({service="clpr-backend"}[5m])) by (level)

# Average latency
avg_over_time({service="clpr-backend"} | json | unwrap latency [5m])

# Error rate
sum(rate({level="error"}[5m])) by (service)
```

## Troubleshooting

### No Logs Appearing

1. Check Promtail is running:
   ```bash
   docker-compose -f docker-compose.monitoring.yml ps promtail
   ```

2. Check Promtail logs:
   ```bash
   docker-compose -f docker-compose.monitoring.yml logs promtail
   ```

3. Verify Loki is accessible:
   ```bash
   curl http://localhost:3100/ready
   ```

### High Log Volume

If generating too many logs:

1. Increase log level to INFO or WARN in production
2. Use sampling for high-frequency events
3. Reduce retention period if storage is limited
4. Add filters in Promtail to exclude noisy logs

### Loki Performance

If Loki is slow:

1. Check index cache size in `loki-config.yml`
2. Increase compactor frequency
3. Use label filters (faster than regex on message content)
4. Consider sharding for high volume

## Monitoring Log Health

Dashboard: **Centralized Logging Dashboard** in Grafana

Key metrics:
- Log volume by service
- Log level distribution
- Error rate trends
- Top error messages
- Security events

## Security Considerations

1. **PII Redaction**: All loggers automatically redact PII
2. **Access Control**: Restrict Grafana access to authorized personnel
3. **Audit Trail**: Log all security-relevant events
4. **Retention**: 90 days for compliance and debugging
5. **Secure Transport**: Use TLS for log shipping in production

## Integration with Other Observability Tools

- **Metrics**: Prometheus for quantitative data
- **Traces**: (Future) Jaeger/Tempo for distributed tracing
- **Alerts**: Alertmanager for log-based alerts
- **Dashboards**: Grafana for unified visualization

## References

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Promtail Configuration](https://grafana.com/docs/loki/latest/clients/promtail/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
- [Grafana Explore](https://grafana.com/docs/grafana/latest/explore/)
