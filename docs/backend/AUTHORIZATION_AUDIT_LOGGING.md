---
title: "AUTHORIZATION AUDIT LOGGING"
summary: "The authorization middleware implements comprehensive audit logging for all authorization decisions to support security monitoring, compliance, and incident investigation."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Authorization Audit Logging

## Overview

The authorization middleware implements comprehensive audit logging for all authorization decisions to support security monitoring, compliance, and incident investigation.

## Features

- **All Decisions Logged**: Both allowed and denied authorization attempts are logged
- **Structured JSON Format**: Logs use JSON format for easy parsing and indexing
- **Rich Context**: Includes user ID, resource details, action, decision, reason, IP address, and user agent
- **GDPR Compliant**: No PII (Personally Identifiable Information) is logged directly; user IDs are anonymized UUIDs
- **Centralized Logging**: Integrates with the existing structured logging system for forwarding to ELK/Loki
- **Searchable**: All logs are indexed and searchable for audit trail analysis

## Log Format

### Standard Authorization Log Entry

```json
{
  "timestamp": "2026-01-05T12:34:56Z",
  "level": "info",
  "message": "Authorization allowed: user=550e8400-e29b-41d4-a716-446655440000 resource=comment:123e4567-e89b-12d3-a456-426614174000 action=delete reason=user_is_owner",
  "service": "clpr-backend",
  "fields": {
    "audit_type": "authorization",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "resource": "comment",
    "resource_id": "123e4567-e89b-12d3-a456-426614174000",
    "action": "delete",
    "decision": "allowed",
    "reason": "user_is_owner",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
  }
}
```

### Denied Authorization Log Entry

```json
{
  "timestamp": "2026-01-05T12:35:23Z",
  "level": "warn",
  "message": "Authorization denied: user=550e8400-e29b-41d4-a716-446655440001 resource=clip:123e4567-e89b-12d3-a456-426614174000 action=delete reason=insufficient_role",
  "service": "clpr-backend",
  "fields": {
    "audit_type": "authorization",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "resource": "clip",
    "resource_id": "123e4567-e89b-12d3-a456-426614174000",
    "action": "delete",
    "decision": "denied",
    "reason": "insufficient_role",
    "ip_address": "192.168.1.101",
    "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
    "metadata": {
      "user_role": "user",
      "required_roles": ["admin"]
    }
  }
}
```

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string (ISO 8601) | UTC timestamp of the authorization decision |
| `level` | string | Log level: "info" for allowed, "warn" for denied |
| `message` | string | Human-readable summary of the decision |
| `service` | string | Service name (always "clpr-backend") |
| `fields.audit_type` | string | Type of audit log (always "authorization") |
| `fields.user_id` | string (UUID) | User ID making the request |
| `fields.resource` | string | Resource type (clip, comment, user, etc.) |
| `fields.resource_id` | string (UUID) | Specific resource ID being accessed |
| `fields.action` | string | Action being performed (create, read, update, delete) |
| `fields.decision` | string | Authorization decision: "allowed", "denied", or "error" |
| `fields.reason` | string | Detailed reason for the decision |
| `fields.ip_address` | string | Client IP address |
| `fields.user_agent` | string | Client user agent string |
| `fields.metadata` | object (optional) | Additional context-specific information |

## Decision Reasons

### Allowed Reasons

- `user_is_owner`: User owns the resource
- `elevated_role_<role>`: User has elevated role (admin, moderator)
- `role_based_access_<role>`: Access granted based on user role
- `account_type_access_<type>`: Access granted based on account type
- `no_restrictions`: Resource has no access restrictions

### Denied Reasons

- `not_owner_insufficient_role`: User is not the owner and lacks required role
- `insufficient_role`: User does not have required role
- `insufficient_account_type`: User's account type doesn't have permission
- `no_permission_rule`: No permission rule defined for resource/action
- `ownership_check_failed`: Error checking resource ownership
- `authorization_check_failed`: System error during authorization

## Resource Types

- `clip`: Video clips
- `comment`: User comments
- `user`: User profiles
- `favorite`: User favorites
- `subscription`: User subscriptions
- `submission`: Clip submissions

## Actions

- `create`: Creating new resources
- `read`: Reading/viewing resources
- `update`: Modifying existing resources
- `delete`: Deleting resources

## Privacy & GDPR Compliance

### No Direct PII

The audit logs do not contain any direct PII such as:
- Email addresses
- Real names
- Phone numbers
- Payment information

### Anonymized Identifiers

- **User IDs**: UUIDs that don't reveal personal information
- **IP Addresses**: Logged for security purposes but can be anonymized/hashed based on retention policy
- **User Agents**: Standard browser identification strings

### Data Retention

Log retention policies should be configured based on compliance requirements:
- **Security Logs**: Typically 90-365 days
- **Audit Logs**: May require longer retention (1-7 years) depending on regulations
- **Access Logs**: Can be purged after shorter periods (30-90 days)

Configure retention in your centralized logging system (ELK/Loki).

## Integration with Centralized Logging

The authorization logs integrate with the existing `StructuredLogger` from `pkg/utils/logger.go`, which:

1. Outputs JSON to stdout
2. Can be forwarded to centralized logging systems:
   - **Elasticsearch/Kibana**: For search and visualization
   - **Loki/Grafana**: For lightweight log aggregation
   - **Splunk**: For enterprise log management
3. Automatically redacts PII patterns from log messages
4. Supports log levels for filtering (INFO for allowed, WARN for denied)

### Example ELK Query

Search for denied access attempts to clips:
```
service:"clpr-backend" AND audit_type:"authorization" AND decision:"denied" AND resource:"clip"
```

Search for authorization errors:
```
service:"clpr-backend" AND audit_type:"authorization" AND decision:"error"
```

### Example Loki Query

```logql
{service="clpr-backend"} |= "audit_type" |= "authorization" | json | decision="denied"
```

## Usage in Code

### Automatic Logging in Middleware

The `RequireResourceOwnership` middleware automatically logs all authorization decisions:

```go
router.DELETE("/clips/:id", 
    middleware.AuthMiddleware(),
    middleware.RequireResourceOwnership(
        middleware.ResourceTypeClip,
        middleware.ActionDelete,
        clipOwnershipChecker,
    ),
    handler.DeleteClip,
)
```

### Manual Logging

For custom authorization logic, use `LogAuthorizationDecision`:

```go
import "git.subcult.tv/subculture-collective/clpr/internal/middleware"

middleware.LogAuthorizationDecision(
    userID,
    middleware.ResourceTypeClip,
    clipID,
    middleware.ActionDelete,
    "allowed",
    "user_is_owner",
    c.ClientIP(),
    c.Request.UserAgent(),
    map[string]interface{}{
        "custom_field": "value",
    },
)
```

## Monitoring & Alerts

### Recommended Alerts

1. **Excessive Denied Attempts**: Alert if a user has >10 denied attempts in 5 minutes
2. **Authorization Errors**: Alert on any authorization system errors
3. **Privilege Escalation**: Alert on denied attempts to access admin-only resources
4. **Geographic Anomalies**: Alert on access from unusual IP addresses for a user

### Dashboard Metrics

- Authorization decisions by user
- Authorization decisions by resource type
- Authorization decisions by action
- Denied attempts over time
- Most denied users/resources

## Testing

Run the authorization audit logging tests:

```bash
cd backend
go test -v ./internal/middleware -run TestLogAuthorizationDecision
go test -v ./internal/middleware -run TestCanAccessResource
```

## References

- [Authorization Middleware](../internal/middleware/authorization.go)
- [Structured Logger](../pkg/utils/logger.go)
- [GDPR Compliance Documentation](https://gdpr.eu/)
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
