---
title: "Websocket Configuration"
summary: "This document describes the WebSocket CORS (Cross-Origin Resource Sharing) configuration for the clpr backend."
tags: ["backend"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# WebSocket Configuration

This document describes the WebSocket CORS (Cross-Origin Resource Sharing) configuration for the clpr backend.

## Overview

The WebSocket server provides real-time chat functionality for channels. To prevent unauthorized access and cross-site attacks, WebSocket connections are restricted to allowed origins configured via environment variables.

## Configuration

### Environment Variable

The WebSocket server reads its allowed origins from the `WEBSOCKET_ALLOWED_ORIGINS` environment variable.

```bash
# Development
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# Staging
WEBSOCKET_ALLOWED_ORIGINS=https://staging.clpr.tv

# Production
WEBSOCKET_ALLOWED_ORIGINS=https://clpr.tv,https://www.clpr.tv
```

### Format

- **Comma-separated list**: Multiple origins can be specified, separated by commas
- **No spaces**: Do not add spaces around commas
- **Include protocol**: Always include `http://` or `https://`
- **Include port**: For development URLs with non-standard ports (e.g., `http://localhost:5173`)

### Wildcard Patterns

The configuration supports wildcard patterns for subdomain matching:

```bash
# Allow all subdomains of clpr.tv
WEBSOCKET_ALLOWED_ORIGINS=*.clpr.tv

# This would allow:
# - https://clpr.tv (base domain)
# - https://staging.clpr.tv
# - https://api.staging.clpr.tv
# - etc.
```

**Important Notes:**
- Wildcard patterns must start with `*.`
- Only subdomain wildcards are supported (not path or protocol wildcards)
- The pattern `*.clpr.tv` matches both the base domain `clpr.tv` and any subdomain
- Wildcards work with both HTTP and HTTPS protocols

### Examples

#### Development Environment

```bash
# .env.development
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

This allows WebSocket connections from:
- The main frontend dev server at `http://localhost:5173`
- Alternative dev server at `http://localhost:3000`

#### Staging Environment

```bash
# .env.staging
WEBSOCKET_ALLOWED_ORIGINS=https://staging.clpr.tv,https://staging.clpr.app
```

This allows WebSocket connections from the staging frontend servers.

#### Production Environment

```bash
# .env.production
WEBSOCKET_ALLOWED_ORIGINS=https://clpr.tv,https://www.clpr.tv
```

This allows WebSocket connections from:
- The main production domain `https://clpr.tv`
- The www subdomain `https://www.clpr.tv`

#### Production with Wildcard

If you have multiple subdomains in production:

```bash
# .env.production
WEBSOCKET_ALLOWED_ORIGINS=*.clpr.tv
```

This allows WebSocket connections from any subdomain of `clpr.tv`, including:
- `https://clpr.tv`
- `https://www.clpr.tv`
- `https://beta.clpr.tv`
- `https://api.clpr.tv`
- etc.

## Security Considerations

### Never Use `*` (Wildcard All)

**NEVER** use `*` as an allowed origin in production:

```bash
# ❌ INSECURE - DO NOT USE
WEBSOCKET_ALLOWED_ORIGINS=*
```

This allows WebSocket connections from any origin, which is a major security vulnerability that can lead to:
- Cross-site WebSocket hijacking (CSWSH) attacks
- Unauthorized access to chat channels
- Data exfiltration
- Session hijacking

### Default Behavior

If `WEBSOCKET_ALLOWED_ORIGINS` is not set or is empty:
- Default origins are used: `http://localhost:5173,http://localhost:3000`
- This is suitable for local development only
- **Always configure this explicitly for staging and production**

### Validation on Startup

The WebSocket server validates the allowed origins configuration on startup and logs warnings for:
- Empty configuration
- Use of `*` (allows all origins)
- Overly broad wildcards like `*.*`

### Origin Rejection Logging

When a WebSocket connection is rejected due to an invalid origin, the server logs:
```
WebSocket connection rejected: origin https://evil.com not in allowed list
```

This helps with debugging legitimate connection issues while also alerting to potential attacks.

## Testing

### Manual Testing

To test WebSocket CORS configuration:

1. Start the backend server with your configuration:
   ```bash
   WEBSOCKET_ALLOWED_ORIGINS=https://clpr.tv go run ./cmd/api
   ```

2. Try to connect from different origins and verify:
   - Allowed origins connect successfully
   - Disallowed origins are rejected

### Automated Tests

Run the WebSocket origin validation tests:

```bash
cd backend
go test ./internal/websocket -run TestIsOriginAllowed -v
go test ./internal/websocket -run TestServerCheckOrigin -v
```

## Implementation Details

### Code Structure

The WebSocket CORS configuration is implemented across several files:

- `backend/config/config.go` - Configuration loading from environment variables
- `backend/internal/websocket/server.go` - WebSocket server with CheckOrigin middleware
- `backend/internal/websocket/origin.go` - Origin validation and pattern matching logic
- `backend/cmd/api/main.go` - Server initialization with configuration

### Origin Validation Flow

1. Client attempts to establish WebSocket connection
2. Browser sends `Origin` header with the request
3. `CheckOrigin` function in the Upgrader is called
4. Origin is validated against configured patterns using `isOriginAllowed()`
5. If allowed, connection is upgraded to WebSocket
6. If rejected, connection is closed with appropriate logging

### Pattern Matching Algorithm

For wildcard patterns like `*.clpr.tv`:

1. Extract domain from the origin (remove protocol and port)
2. Check if domain ends with the pattern suffix (e.g., `clpr.tv`)
3. Ensure proper subdomain boundary (must end with `.clpr.tv` or be exactly `clpr.tv`)
4. Reject partial matches (e.g., `fakeclpr.tv` does not match `*.clpr.tv`)

## Troubleshooting

### WebSocket Connections Failing

If WebSocket connections are failing:

1. **Check server logs** for rejection messages:
   ```
   WebSocket connection rejected: origin <origin> not in allowed list
   ```

2. **Verify environment variable** is set correctly:
   ```bash
   echo $WEBSOCKET_ALLOWED_ORIGINS
   ```

3. **Check frontend origin** matches exactly (including protocol and port):
   - Browser DevTools → Network → WebSocket connection → Headers → Origin

4. **Restart server** after changing environment variables

### Common Issues

**Issue**: Frontend on `http://localhost:5173` cannot connect

**Solution**: Ensure `WEBSOCKET_ALLOWED_ORIGINS` includes `http://localhost:5173` with the correct port

---

**Issue**: Production connections failing with wildcard pattern

**Solution**: Verify the wildcard pattern matches the frontend domain structure. For `www.clpr.tv`, use either:
- Exact: `https://www.clpr.tv`
- Wildcard: `*.clpr.tv` (matches any subdomain including www)

---

**Issue**: Staging environment cannot connect

**Solution**: Check that:
1. `WEBSOCKET_ALLOWED_ORIGINS` is set in staging environment
2. Value matches the staging frontend URL exactly
3. HTTPS is used if frontend is served over HTTPS

## Migration Guide

If you're migrating from hardcoded origins to environment-based configuration:

### Before (Hardcoded)
```go
allowedOrigins := []string{
    "http://localhost:3000",
    "http://localhost:5173",
    "https://clpr.subculture.gg",
}
```

### After (Environment-based)
```bash
# .env file
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,https://clpr.tv
```

### Steps

1. Add `WEBSOCKET_ALLOWED_ORIGINS` to all environment files (`.env`, `.env.staging`, `.env.production`)
2. Update values to match your current frontend URLs
3. Deploy backend with new configuration
4. Verify WebSocket connections work
5. Monitor server logs for any rejected connections

## Related Documentation

- [Backend API Documentation](./api.md)
- [Security Guide](./security.md)
