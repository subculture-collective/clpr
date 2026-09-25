---
title: "APPLICATION LOGS"
summary: "The backend log collection endpoint (`POST /api/v1/logs`) provides centralized log aggregation for frontend and mobile clients. This enables monitoring, debugging, and analytics across the entire plat"
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Application Log Collection Endpoint

## Overview
The backend log collection endpoint (`POST /api/v1/logs`) provides centralized log aggregation for frontend and mobile clients. This enables monitoring, debugging, and analytics across the entire platform.

## Endpoint

### POST /api/v1/logs
Accepts log entries from frontend (web) and mobile (iOS/Android) clients.

#### Authentication
- Optional authentication via `OptionalAuthMiddleware`
- Both authenticated and anonymous logs are accepted
- User ID is automatically captured if authenticated

#### Rate Limiting
- **100 requests/minute per endpoint per IP address**
- Authenticated and anonymous users behind the same IP share this limit
- Returns `429 Too Many Requests` when limit exceeded

#### Request Format

```json
{
  "level": "error",
  "message": "Failed to load clip",
  "timestamp": "2026-01-02T12:34:56Z",
  "service": "clpr-frontend",
  "platform": "web",
  "context": {
    "clipId": "clip_123",
    "url": "/clips/123"
  },
  "user_agent": "Mozilla/5.0...",
  "session_id": "session_abc123",
  "trace_id": "trace_xyz789",
  "url": "https://clpr.video/clips/123",
  "device_id": "device_456",
  "app_version": "1.0.0",
  "error": "NetworkError: Failed to fetch",
  "stack": "Error: Failed to fetch\n    at fetchClip..."
}
```

#### Required Fields
- `level` (string): Must be one of: `debug`, `info`, `warn`, `error`
- `message` (string): Log message, max 10,000 characters

#### Optional Fields
- `timestamp` (string, ISO 8601): Client-side timestamp (server uses current time if omitted)
- `service` (string): Service identifier (auto-detected from platform if omitted)
- `platform` (string): `web`, `ios`, or `android`
- `context` (object): Additional structured data (JSONB)
- `user_agent` (string): User agent string
- `session_id` (string): Client session identifier
- `trace_id` (string): Request trace identifier for correlation
- `url` (string): Current URL or screen
- `device_id` (string): Hashed device identifier (mobile only)
- `app_version` (string): Application version
- `error` (string): Error message (for error-level logs)
- `stack` (string): Stack trace (for error-level logs)

#### Response
- **Success**: `204 No Content` (no response body)
- **Validation Error**: `400 Bad Request`
- **Too Large**: `413 Request Entity Too Large` (message exceeds 100KB)
- **Rate Limit**: `429 Too Many Requests`
- **Server Error**: `500 Internal Server Error`

## Security

### Sensitive Data Filtering
The backend automatically filters sensitive information from logs:

#### Filtered Content
- Passwords (`password`, `passwd`, `pwd`)
- Secrets (`secret`)
- API Keys (`apikey`, `api_key`)
- Auth Tokens (`token`, `access_token`, `auth_token`, `authorization`)
- Bearer tokens

#### Filtering Behavior
- **Messages**: If a log message contains sensitive keywords, it's replaced with `[REDACTED - contains sensitive data]`
- **Context Fields**: Sensitive keys are replaced with `[REDACTED]` while preserving field names
- **Nested Objects**: Filtering applied recursively

### Size Limits
- Maximum message size: 100KB
- Requests exceeding limit return `413 Request Entity Too Large`

## Database Schema

### Table: `application_logs`

```sql
CREATE TABLE application_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(10) NOT NULL CHECK (level IN ('debug', 'info', 'warn', 'error')),
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    service VARCHAR(50) NOT NULL,
    platform VARCHAR(20),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    trace_id VARCHAR(255),
    url TEXT,
    user_agent TEXT,
    device_id VARCHAR(255),
    app_version VARCHAR(50),
    error TEXT,
    stack TEXT,
    context JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Indexes
- `idx_application_logs_level`: For filtering by log level
- `idx_application_logs_timestamp`: For time-based queries
- `idx_application_logs_user_id`: For user-specific logs
- `idx_application_logs_service`: For service-specific logs
- `idx_application_logs_created_at`: For log retention cleanup
- `idx_application_logs_level_timestamp`: Composite for level + time queries

## Log Retention

### Retention Policy
Logs should be retained based on their level:
- **Debug logs**: 30 days
- **Info/Warn logs**: 30 days
- **Error logs**: 90 days

### Cleanup
The `DeleteOldLogs` repository method deletes logs older than a specified retention period. To implement level-based retention, call this method separately for each level with different retention values:

```go
// Delete debug logs older than 30 days
deletedCount, err := logRepo.DeleteOldLogsByLevel(ctx, "debug", 30)

// Delete info logs older than 30 days
deletedCount, err := logRepo.DeleteOldLogsByLevel(ctx, "info", 30)

// Delete warn logs older than 30 days
deletedCount, err := logRepo.DeleteOldLogsByLevel(ctx, "warn", 30)

// Delete error logs older than 90 days
deletedCount, err := logRepo.DeleteOldLogsByLevel(ctx, "error", 90)
```

Alternatively, for uniform retention across all levels:
```go
// Delete all logs older than 30 days
deletedCount, err := logRepo.DeleteOldLogs(ctx, 30)
```

**Note**: The current `DeleteOldLogs` implementation applies uniform retention across all log levels. To implement the level-based retention policy described above, you would need to add a `DeleteOldLogsByLevel` method to the repository or use separate cleanup jobs for each level.

## Client Integration

### Frontend (TypeScript)

```typescript
// logger.ts automatically sends logs in production
const logger = getLogger();
logger.error('Failed to load clip', new Error('Network error'), {
  clipId: 'clip_123',
  action: 'load'
});
```

The frontend logger:
- Sends logs to `/api/v1/logs` in production mode
- Includes PII redaction before sending
- Handles errors gracefully (no infinite loops)
- Enriches logs with browser context (URL, user agent)

### Mobile (TypeScript/React Native)

```typescript
// Mobile logger automatically sends logs in production
const logger = getLogger();
await logger.error('Failed to sync data', new Error('Sync failed'), {
  lastSync: '2026-01-02T10:00:00Z'
});
```

The mobile logger:
- Sends logs to `{API_BASE_URL}/api/v1/logs` in production
- Includes platform-specific context (iOS/Android, app version, device ID)
- Handles errors gracefully
- Uses `EXPO_PUBLIC_API_URL` environment variable for API base URL

## Statistics Endpoint

### GET /api/v1/logs/stats
Returns aggregated log statistics (admin only).

#### Authentication
- Requires authentication
- Requires admin role

#### Response

```json
{
  "success": true,
  "stats": {
    "total_logs": 1000000,
    "unique_users": 5000,
    "error_count": 20000,
    "warn_count": 50000,
    "info_count": 800000,
    "debug_count": 130000,
    "logs_last_hour": 50000,
    "logs_last_24h": 1000000
  }
}
```

## Best Practices

### Client-Side
1. **Filter PII before sending**: Both frontend and mobile loggers filter PII
2. **Include context**: Add relevant contextual information to `context` field
3. **Use appropriate log levels**: 
   - `debug`: Detailed diagnostic information
   - `info`: General informational messages
   - `warn`: Warning messages (recoverable issues)
   - `error`: Error events (require attention)
4. **Avoid logging in hot paths**: Don't log in high-frequency loops
5. **Include trace IDs**: For request correlation across services

### Server-Side
1. **Monitor rate limits**: Track 429 responses
2. **Set up log retention**: Configure automated cleanup
3. **Review error logs regularly**: High error rates may indicate issues
4. **Monitor storage**: Logs can consume significant database space

## Monitoring

### Metrics to Track
- Logs per minute (by level)
- Unique users reporting errors
- Error rate trends
- Rate limit violations
- Storage usage

### Alerts
- Error rate spike (>10% increase)
- Storage threshold (>80% capacity)
- High rate limit violations

## Troubleshooting

### Common Issues

#### Logs not appearing
1. Check frontend/mobile is in production mode
2. Verify API endpoint URL is correct
3. Check network connectivity
4. Review rate limiting (may be blocked)

#### Rate limit exceeded
1. Review client code for excessive logging
2. Check for infinite loops
3. Verify rate limit thresholds are appropriate
4. Consider implementing client-side batching

#### Missing context
1. Ensure context object is properly serialized
2. Check for circular references in context
3. Verify context size is within limits

## Examples

### Error Log
```json
{
  "level": "error",
  "message": "Failed to authenticate user",
  "error": "Invalid credentials",
  "context": {
    "attemptCount": 3,
    "lastAttempt": "2026-01-02T12:34:56Z"
  }
}
```

### Info Log
```json
{
  "level": "info",
  "message": "User logged in successfully",
  "context": {
    "loginMethod": "oauth",
    "provider": "twitch"
  }
}
```

### Debug Log
```json
{
  "level": "debug",
  "message": "Cache hit for clip metadata",
  "context": {
    "clipId": "clip_123",
    "cacheKey": "metadata:clip_123",
    "ttl": 3600
  }
}
```
