---
title: "AUDIT LOG SERVICE"
summary: "The AuditLogService provides comprehensive audit logging capabilities for all moderation actions in the system. It tracks who did what, when, where, and why with full context."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Audit Log Service Implementation

## Overview
The AuditLogService provides comprehensive audit logging capabilities for all moderation actions in the system. It tracks who did what, when, where, and why with full context.

## Features

### Core Functionality
- **Comprehensive Action Logging**: Records all moderation actions with full context
- **Advanced Filtering**: Filter by action, actor, target, channel, entity type, and date range
- **Pagination**: Efficient handling of large audit log datasets
- **Metadata Support**: Store additional action-specific data as JSON
- **Context Capture**: IP address and user agent for security and compliance
- **CSV Export**: Export filtered audit logs for compliance and reporting

### Database Schema
The `moderation_audit_logs` table includes:
- `id`: Unique identifier
- `action`: Type of action performed (e.g., ban, approve, reject)
- `entity_type`: Type of entity acted upon (e.g., user, clip, comment, channel)
- `entity_id`: ID of the affected entity
- `moderator_id`: User who performed the action
- `reason`: Optional explanation for the action
- `metadata`: JSON field for additional context
- `ip_address`: IP address from which the action was performed
- `user_agent`: User agent string of the client
- `channel_id`: Optional channel context for chat/moderation actions
- `created_at`: Timestamp of the action

### Performance Optimizations
Database indexes on:
- `moderator_id` - Fast lookups by actor
- `entity_type` and `entity_id` - Fast lookups by target
- `created_at DESC` - Efficient time-based queries
- `action` - Fast filtering by action type
- `channel_id` (partial, where not null) - Efficient channel queries
- `ip_address` (partial, where not null) - Security investigations

## API Usage

### Logging an Action

```go
import (
    "context"
    "github.com/google/uuid"
    "git.subcult.tv/subculture-collective/clpr/internal/services"
)

// Initialize service (typically done in dependency injection)
auditLogService := services.NewAuditLogService(auditLogRepo)

// Log a moderation action with full context
ctx := context.Background()
action := "ban"
actor := userID                    // UUID of moderator
target := bannedUserID             // UUID of banned user
entityType := "user"
channelID := chatChannelID         // Optional: UUID

// In a Gin handler, extract IP and user agent from request:
ipAddress := c.ClientIP()          // Extracted from request context (e.g., Gin *gin.Context)
userAgent := c.Request.UserAgent() // Extracted from request headers

reason := "Repeated spam violations"
metadata := map[string]interface{}{
    "duration": "7d",
    "severity": "high",
    "auto_expire": true,
}

err := auditLogService.LogAction(
    ctx, action, actor, target, entityType,
    services.AuditLogOptions{
        Channel:   &channelID,
        Reason:    &reason,
        Metadata:  metadata,
        IPAddress: &ipAddress,
        UserAgent: &userAgent,
    },
)
```

### Querying Audit Logs

```go
import (
    "context"
    "time"
    "github.com/google/uuid"
    "git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// Build filters
moderatorID := uuid.MustParse("...")
channelID := uuid.MustParse("...")
startDate := time.Now().Add(-7 * 24 * time.Hour)
endDate := time.Now()

filters := repository.AuditLogFilters{
    ModeratorID: &moderatorID,        // Optional: filter by who performed actions
    Action:      "ban",               // Optional: filter by action type
    EntityType:  "user",              // Optional: filter by entity type
    EntityID:    &targetUserID,       // Optional: filter by specific target
    ChannelID:   &channelID,          // Optional: filter by channel
    StartDate:   &startDate,          // Optional: filter by date range
    EndDate:     &endDate,            // Optional: filter by date range
}

// Retrieve logs with pagination
page := 1
limit := 50
logs, total, err := auditLogService.GetAuditLogs(ctx, filters, page, limit)

// Process results
for _, log := range logs {
    fmt.Printf("Action: %s by %s on %s\n", 
        log.Action, 
        log.Moderator.Username,
        log.CreatedAt)
}
```

### Filter Parsing from HTTP Query Parameters

```go
// In HTTP handler
filters, err := services.ParseAuditLogFilters(
    c.Query("moderator_id"),
    c.Query("action"),
    c.Query("entity_type"),
    c.Query("entity_id"),
    c.Query("channel_id"),
    c.Query("start_date"),  // RFC3339 format
    c.Query("end_date"),    // RFC3339 format
)
```

### CSV Export

```go
import "os"

// Export to file
file, err := os.Create("audit_logs.csv")
if err != nil {
    return err
}
defer file.Close()

err = auditLogService.ExportAuditLogsCSV(ctx, filters, file)
```

## HTTP Endpoints

### GET /admin/audit-logs
Retrieve audit logs with filtering and pagination.

**Query Parameters:**
- `page` (int, default: 1): Page number
- `limit` (int, default: 50, max: 100): Items per page
- `moderator_id` (UUID): Filter by moderator
- `action` (string): Filter by action type
- `entity_type` (string): Filter by entity type
- `entity_id` (UUID): Filter by specific entity
- `channel_id` (UUID): Filter by channel
- `start_date` (RFC3339): Start of date range
- `end_date` (RFC3339): End of date range

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "action": "ban",
      "entity_type": "user",
      "entity_id": "uuid",
      "moderator_id": "uuid",
      "reason": "Spam",
      "metadata": {
        "duration": "7d",
        "severity": "high"
      },
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "channel_id": "uuid",
      "created_at": "2026-01-08T20:00:00Z",
      "moderator": {
        "id": "uuid",
        "username": "admin",
        "display_name": "Admin User"
      }
    }
  ],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 150,
    "total_pages": 3
  }
}
```

### GET /admin/audit-logs/export
Export audit logs to CSV with the same filtering options as the list endpoint.

**Query Parameters:** Same as list endpoint

**Response:** CSV file download

## Testing

### Unit Tests
```bash
# Run service tests
go test ./internal/services -run TestAuditLog -v

# Run repository tests  
go test ./internal/repository -run TestAuditLog -v

# Run all tests with filter parsing
go test ./internal/services -run TestParse -v
```

### Test Coverage
- ✅ Filter parsing with validation
- ✅ Filter structure and optional fields
- ✅ Empty filters handling
- ✅ Date range validation
- ✅ UUID validation for all ID fields
- ✅ Repository filter structure

## Migration

The database schema changes for audit log context fields were added in migration `000097_update_moderation_audit_logs`:
- `ip_address` (INET)
- `user_agent` (TEXT)
- `channel_id` (UUID)
- `actor_id` (UUID) - new name for moderator
- `target_user_id` (UUID) - for user-focused actions
- Indexes for performance on all filterable fields

## Security Considerations

1. **IP Address Storage**: IP addresses are stored for security auditing and compliance. **Important**: IP addresses are considered personal data / Personally Identifiable Information (PII) under GDPR and many similar privacy regulations and MUST be handled accordingly.
2. **User Agent Tracking**: Helps identify automated vs. manual actions
3. **No PII in Metadata**: Avoid storing sensitive personal information in the metadata field
4. **Access Control**: Only authorized administrators should access audit logs
5. **Retention Policies**: Implement clear data retention and deletion policies for audit log entries (including IP addresses and user agents), ensuring that data is not kept longer than necessary for the stated purpose and that retention is based on a documented legal basis.
6. **Anonymization / Pseudonymization**: Where full IP addresses are not strictly required, consider applying anonymization or pseudonymization techniques (e.g., IP truncation, hashing, or tokenization) to reduce privacy risk while preserving the utility of audit data.
7. **Regulatory Compliance**: Ensure that the collection and processing of audit log data complies with all applicable privacy and data protection laws (e.g., GDPR, CCPA), including providing appropriate notices, honoring user rights where applicable, and conducting legal/security reviews as needed.

## Performance Considerations

1. **Indexes**: All common filter fields are indexed
2. **Pagination**: Always use pagination for large datasets
3. **Date Range Queries**: Indexed for efficient time-based lookups
4. **Join Optimization**: User data is joined efficiently with proper indexes
5. **Partial Indexes**: IP and channel indexes only include non-null values to save space

## Common Action Types

- `approve` - Approving content submissions
- `reject` - Rejecting content submissions
- `ban` - Banning a user
- `unban` - Unbanning a user
- `timeout` - Temporary restriction
- `delete_message` - Deleting a chat message
- `bulk_approve` - Bulk content approval
- `bulk_reject` - Bulk content rejection
- Custom actions as needed

## Common Entity Types

- `user` - User accounts
- `clip` - Video clips
- `comment` - User comments
- `clip_submission` - Pending clip submissions
- `channel` - Chat channels
- `message` - Chat messages
- Custom entity types as needed
