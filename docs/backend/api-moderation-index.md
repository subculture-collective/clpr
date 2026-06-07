# Moderation API - Quick Start Guide

Complete API documentation for managing moderation, bans, moderators, and audit logging in the Clipper platform.

## 📚 Documentation

### Main Documentation
- **[Complete Moderation API Reference](./moderation-api.md)** - Full API documentation with all endpoints, examples, and deployment guide

### OpenAPI Specification
- **[OpenAPI Spec](../openapi/openapi.yaml)** - Machine-readable API specification
- **[Interactive API Docs](../openapi/index.html)** - Browse and test the API

## 🚀 Quick Start

### 1. Authentication

All moderation endpoints require JWT Bearer authentication:

```bash
# Obtain token via Twitch OAuth
curl -X GET https://api.clpr.tv/api/v1/auth/twitch

# Use token in requests
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.clpr.tv/api/v1/moderation/bans?channelId=CHANNEL_UUID
```

### 2. Common Operations

#### Ban a User

```bash
curl -X POST https://api.clpr.tv/api/v1/moderation/ban \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channelId": "123e4567-e89b-12d3-a456-426614174000",
    "userId": "user-uuid-to-ban",
    "reason": "Violation of community guidelines"
  }'
```

#### List Moderators

```bash
curl -X GET "https://api.clpr.tv/api/v1/moderation/moderators?channelId=123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### View Audit Logs

```bash
curl -X GET "https://api.clpr.tv/api/v1/moderation/audit-logs?action=ban_user&limit=50" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📖 Endpoints Overview

### Ban Management

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| POST | `/api/v1/moderation/sync-bans` | Sync bans from Twitch | 5/hour |
| GET | `/api/v1/moderation/bans` | List channel bans | 60/min |
| POST | `/api/v1/moderation/ban` | Create a new ban | 10/hour |
| GET | `/api/v1/moderation/ban/:id` | Get ban details | 60/min |
| DELETE | `/api/v1/moderation/ban/:id` | Revoke a ban | 10/hour |

### Moderator Management

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| GET | `/api/v1/moderation/moderators` | List moderators | 60/min |
| POST | `/api/v1/moderation/moderators` | Add moderator | 10/hour |
| DELETE | `/api/v1/moderation/moderators/:id` | Remove moderator | 10/hour |
| PATCH | `/api/v1/moderation/moderators/:id` | Update permissions | 10/hour |

### Audit Logging

| Method | Endpoint | Description | Rate Limit |
|--------|----------|-------------|------------|
| GET | `/api/v1/moderation/audit-logs` | List audit logs | 60/min |
| GET | `/api/v1/moderation/audit-logs/export` | Export logs to CSV | 10/hour |
| GET | `/api/v1/moderation/audit-logs/:id` | Get specific log | 60/min |

## 🔐 Permissions

### Required Roles

- **Channel Owner/Admin**: Full moderation access for their channels
- **Moderator**: Can moderate assigned channels only
- **Site Admin**: Global moderation access

### Permission Matrix

| Action | User | Moderator | Channel Owner/Admin | Site Admin |
|--------|------|-----------|---------------------|------------|
| View bans | ❌ | ✅ | ✅ | ✅ |
| Create ban | ❌ | ✅ | ✅ | ✅ |
| Revoke ban | ❌ | ✅ | ✅ | ✅ |
| Add moderator | ❌ | ❌ | ✅ | ✅ |
| Remove moderator | ❌ | ❌ | ✅ | ✅ |
| View audit logs | ❌ | ✅ | ✅ | ✅ |

See [Complete Permission Matrix](./moderation-api.md#permission-matrix) for details.

## 💻 Code Examples

### JavaScript (Fetch)

```javascript
const API_BASE = 'https://api.clpr.tv/api/v1/moderation';
const AUTH_TOKEN = 'YOUR_TOKEN';

// List bans
async function listBans(channelId) {
  const response = await fetch(`${API_BASE}/bans?channelId=${channelId}`, {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`
    }
  });
  return response.json();
}

// Create ban
async function createBan(channelId, userId, reason) {
  const response = await fetch(`${API_BASE}/ban`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ channelId, userId, reason })
  });
  return response.json();
}
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

const (
    APIBase = "https://api.clpr.tv/api/v1/moderation"
)

type Client struct {
    Token  string
    Client *http.Client
}

func (c *Client) CreateBan(channelID, userID, reason string) error {
    body := map[string]string{
        "channelId": channelID,
        "userId":    userID,
        "reason":    reason,
    }
    
    jsonData, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", APIBase+"/ban", bytes.NewBuffer(jsonData))
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.Client.Do(req)
    defer resp.Body.Close()
    return err
}
```

See [Complete Code Examples](./moderation-api.md#code-examples) for more languages.

## 🚨 Error Handling

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Malformed request or invalid parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication token |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource does not exist |
| 409 | `CONFLICT` | Resource conflict (e.g., already banned) |
| 429 | `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

### Error Response Format

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

See [Error Handling](./moderation-api.md#error-handling) for complete documentation.

## 📊 Rate Limiting

Rate limits are enforced per IP address and user account:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1738540800
```

| Endpoint Type | Limit | Window |
|---------------|-------|--------|
| Read operations (GET) | 60 requests | 1 minute |
| Write operations (POST/DELETE) | 10 requests | 1 hour |
| Sync operations | 5 requests | 1 hour |

See [Rate Limiting](./moderation-api.md#rate-limiting) for details.

## 🛠️ Testing

### Using cURL

```bash
# Test authentication
export TOKEN="your_jwt_token"

# List bans
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.clpr.tv/api/v1/moderation/bans?channelId=CHANNEL_UUID"

# Create ban
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"channelId":"CHANNEL_UUID","userId":"USER_UUID","reason":"Test"}' \
  https://api.clpr.tv/api/v1/moderation/ban
```

### Using Postman

1. Import the OpenAPI spec: `docs/openapi/openapi.yaml`
2. Set up environment variable for `Bearer Token`
3. Test endpoints directly in Postman

### Using Browser (Swagger UI)

```bash
# Serve the API docs locally
npm run openapi:serve

# Open in browser
open http://localhost:8081
```

## 🔧 Deployment

### Environment Variables

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/clpr

# Redis (for rate limiting)
REDIS_URL=redis://localhost:6379

# Twitch API (for ban sync)
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret

# JWT Authentication
JWT_SECRET=your_jwt_secret
JWT_EXPIRATION=3600

# Rate Limiting
RATE_LIMIT_ENABLED=true

# Audit Logging
AUDIT_LOG_ENABLED=true
AUDIT_LOG_RETENTION_DAYS=90
```

See [Deployment Guide](./moderation-api.md#deployment-guide) for complete setup.

## 📝 Additional Resources

- [Complete API Reference](./moderation-api.md)
- [OpenAPI Specification](../openapi/openapi.yaml)
- [Authentication Guide](./authentication.md)
- [Authorization Framework](./authorization-framework.md)
- [Audit Log Service](./AUDIT_LOG_SERVICE.md)
- [GitHub Issues](https://git.subcult.tv/subculture-collective/clpr/issues)

## 💬 Support

- **Documentation Issues**: [Open an issue](https://git.subcult.tv/subculture-collective/clpr/issues/new)
- **API Questions**: [GitHub Discussions](https://git.subcult.tv/subculture-collective/clpr/discussions)
- **Email**: support@clpr.tv

---

**Last Updated**: 2024-01-15  
**Version**: 1.0.0  
**Maintainer**: team-core
