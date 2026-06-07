# Moderation API Documentation

Complete documentation for the Clipper moderation API endpoints, including authentication, permissions, rate limiting, and code examples.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Authorization & Permissions](#authorization--permissions)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [Endpoints](#endpoints)
  - [Sync Bans](#sync-bans)
  - [List Bans](#list-bans)
  - [Create Ban](#create-ban)
  - [Get Ban Details](#get-ban-details)
  - [Revoke Ban](#revoke-ban)
  - [List Moderators](#list-moderators)
  - [Add Moderator](#add-moderator)
  - [Remove Moderator](#remove-moderator)
  - [Update Moderator Permissions](#update-moderator-permissions)
  - [List Audit Logs](#list-audit-logs)
  - [Export Audit Logs](#export-audit-logs)
  - [Get Audit Log](#get-audit-log)
- [Code Examples](#code-examples)
- [Permission Matrix](#permission-matrix)
- [Deployment Guide](#deployment-guide)

---

## Overview

The Moderation API provides comprehensive tools for managing user bans, moderator roles, and audit logging within the Clipper platform. All endpoints require authentication and appropriate permissions.

**Base URL**: `https://api.clpr.tv/api/v1/moderation`

**API Version**: v1

**Content-Type**: `application/json`

---

## Authentication

All moderation endpoints require JWT Bearer token authentication:

```http
Authorization: Bearer <your_jwt_token>
```

### Obtaining a Token

1. Authenticate via Twitch OAuth: `GET /api/v1/auth/twitch`
2. Use the returned JWT token in subsequent requests
3. Refresh tokens when expired: `POST /api/v1/auth/refresh`

### Token Format

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### Token Expiration

- Access tokens expire after 1 hour
- Refresh tokens expire after 30 days
- Use the refresh token to obtain a new access token without re-authenticating

---

## Authorization & Permissions

### User Roles

- **User**: Standard user with basic access
- **Moderator**: Can moderate content within assigned communities/channels
- **Admin**: Full access to all moderation features

### Permission Scopes

Moderators can have different scopes:

- **Global**: Can moderate across all communities
- **Community**: Can moderate specific communities/channels only

### Permission Validation

Each endpoint validates permissions based on:

1. User role (admin, moderator, user)
2. Resource ownership (e.g., channel owner)
3. Moderator scope (global vs. community-specific)

---

## Rate Limiting

All endpoints are rate-limited to prevent abuse. Limits are enforced per IP address and user account.

### Rate Limit Headers

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1738540800
```

### Endpoint Rate Limits

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| POST /sync-bans | 5 requests | 1 hour |
| GET /bans | 60 requests | 1 minute |
| POST /ban | 10 requests | 1 hour |
| DELETE /ban/:id | 10 requests | 1 hour |
| GET /ban/:id | 60 requests | 1 minute |
| GET /moderators | 60 requests | 1 minute |
| POST /moderators | 10 requests | 1 hour |
| DELETE /moderators/:id | 10 requests | 1 hour |
| PATCH /moderators/:id | 10 requests | 1 hour |
| GET /audit-logs | 60 requests | 1 minute |
| GET /audit-logs/export | 10 requests | 1 hour |
| GET /audit-logs/:id | 60 requests | 1 minute |

### Rate Limit Exceeded Response

```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 3600
}
```

**Status Code**: `429 Too Many Requests`

---

## Error Handling

### Standard Error Response Format

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | INVALID_REQUEST | Malformed request or invalid parameters |
| 401 | UNAUTHORIZED | Missing or invalid authentication token |
| 403 | FORBIDDEN | Insufficient permissions for the requested action |
| 404 | NOT_FOUND | Requested resource does not exist |
| 409 | CONFLICT | Resource conflict (e.g., user already banned) |
| 429 | RATE_LIMIT_EXCEEDED | Rate limit exceeded |
| 500 | INTERNAL_ERROR | Server error |
| 503 | SERVICE_UNAVAILABLE | Service temporarily unavailable |

### Error Code Examples

#### Invalid UUID Format

```json
{
  "error": "Invalid UUID format",
  "code": "INVALID_REQUEST",
  "details": {
    "field": "user_id",
    "provided": "invalid-uuid"
  }
}
```

#### Permission Denied

```json
{
  "error": "Permission denied: You must be a channel owner or admin to perform this action",
  "code": "FORBIDDEN"
}
```

#### Resource Not Found

```json
{
  "error": "Channel not found",
  "code": "NOT_FOUND",
  "details": {
    "channel_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

---

## Endpoints

### Sync Bans

Synchronizes ban status from Twitch for a specific channel.

**Endpoint**: `POST /api/v1/moderation/sync-bans`

**Authentication**: Required

**Rate Limit**: 5 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Body**:
```json
{
  "channel_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| channel_id | string (UUID) | Yes | Channel ID to sync bans for |

#### Response

**Success (200 OK)**:
```json
{
  "status": "syncing",
  "job_id": "987fcdeb-51a2-43c1-b789-123456789abc",
  "message": "Ban sync started for channel 123e4567-e89b-12d3-a456-426614174000"
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| status | string | Sync status: "syncing" |
| job_id | string (UUID) | Unique identifier for the sync job |
| message | string | Human-readable status message |

#### Errors

**400 Bad Request** - Invalid or missing channel_id:
```json
{
  "error": "channel_id is required and must be a valid UUID",
  "code": "INVALID_REQUEST"
}
```

**401 Unauthorized** - Missing or invalid token:
```json
{
  "error": "Authorization token required",
  "code": "UNAUTHORIZED"
}
```

**503 Service Unavailable** - Twitch sync service unavailable:
```json
{
  "error": "Twitch ban sync service is currently unavailable",
  "code": "SERVICE_UNAVAILABLE"
}
```

#### Notes

- Sync operation runs asynchronously with a 5-minute timeout
- Returns immediately; ban data is updated in the background
- Requires Twitch moderator OAuth scopes for the channel
- Only channel owners, moderators, or admins can sync bans

---

### List Bans

Retrieves a paginated list of bans for a specific channel.

**Endpoint**: `GET /api/v1/moderation/bans`

**Authentication**: Required

**Rate Limit**: 60 requests per minute

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| channelId | string (UUID) | Yes | - | Channel ID to list bans for |
| limit | integer | No | 10 | Number of results per page (max: 100) |
| offset | integer | No | 0 | Number of results to skip (must be multiple of limit) |

**Example Request**:
```http
GET /api/v1/moderation/bans?channelId=123e4567-e89b-12d3-a456-426614174000&limit=20&offset=0
```

#### Response

**Success (200 OK)**:
```json
{
  "bans": [
    {
      "id": "98765432-e89b-12d3-a456-426614174001",
      "channelId": "123e4567-e89b-12d3-a456-426614174000",
      "userId": "user-uuid-here",
      "bannedBy": "moderator-uuid-here",
      "reason": "Violation of community guidelines",
      "bannedAt": "2024-01-15T10:30:00Z",
      "expiresAt": null,
      "isPermanent": true,
      "username": "banneduser123",
      "bannedByUsername": "moderator456"
    }
  ],
  "total": 150,
  "limit": 20,
  "offset": 0
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| bans | array | Array of ban objects |
| bans[].id | string (UUID) | Unique ban identifier |
| bans[].channelId | string (UUID) | Channel where user is banned |
| bans[].userId | string (UUID) | Banned user ID |
| bans[].bannedBy | string (UUID) | Moderator who issued the ban |
| bans[].reason | string | Reason for the ban |
| bans[].bannedAt | string (ISO 8601) | Timestamp when ban was issued |
| bans[].expiresAt | string/null | Ban expiration timestamp (null if permanent) |
| bans[].isPermanent | boolean | Whether ban is permanent |
| bans[].username | string | Banned user's username |
| bans[].bannedByUsername | string | Moderator's username |
| total | integer | Total number of bans |
| limit | integer | Results per page |
| offset | integer | Current offset |

#### Errors

**400 Bad Request** - Invalid parameters:
```json
{
  "error": "channelId is required and must be a valid UUID",
  "code": "INVALID_REQUEST"
}
```

**400 Bad Request** - Invalid offset:
```json
{
  "error": "offset must be a multiple of limit",
  "code": "INVALID_REQUEST",
  "details": {
    "offset": 15,
    "limit": 10
  }
}
```

**403 Forbidden** - Insufficient permissions:
```json
{
  "error": "Permission denied: You do not have access to view bans for this channel",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - Channel not found:
```json
{
  "error": "Channel not found",
  "code": "NOT_FOUND"
}
```

---

### Create Ban

Creates a new ban for a user in a specific channel.

**Endpoint**: `POST /api/v1/moderation/ban`

**Authentication**: Required

**Permissions**: Channel owner/admin or site moderator

**Rate Limit**: 10 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Body**:
```json
{
  "channelId": "123e4567-e89b-12d3-a456-426614174000",
  "userId": "user-uuid-to-ban",
  "reason": "Violation of community guidelines"
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| channelId | string (UUID) | Yes | Channel ID where user will be banned |
| userId | string (UUID) | Yes | User ID to ban |
| reason | string | No | Reason for the ban (max 1000 characters) |

#### Response

**Success (201 Created)**:
```json
{
  "id": "ban-uuid-here",
  "channelId": "123e4567-e89b-12d3-a456-426614174000",
  "userId": "user-uuid-to-ban",
  "bannedBy": "moderator-uuid-here",
  "reason": "Violation of community guidelines",
  "bannedAt": "2024-01-15T10:30:00Z",
  "expiresAt": null,
  "isPermanent": true
}
```

#### Errors

**400 Bad Request** - Invalid UUID format:
```json
{
  "error": "Invalid UUID format for channelId",
  "code": "INVALID_REQUEST"
}
```

**400 Bad Request** - Cannot ban channel owner:
```json
{
  "error": "Cannot ban the channel owner",
  "code": "INVALID_REQUEST"
}
```

**403 Forbidden** - Insufficient permissions:
```json
{
  "error": "Permission denied: You must be a channel owner, admin, or moderator to ban users",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - User or channel not found:
```json
{
  "error": "User not found",
  "code": "NOT_FOUND"
}
```

**409 Conflict** - User already banned:
```json
{
  "error": "User is already banned in this channel",
  "code": "CONFLICT"
}
```

#### Notes

- Permanent bans are created by default (expiresAt is null)
- Creates an audit log entry automatically
- Broadcasts ban event via WebSocket to connected clients

---

### Get Ban Details

Retrieves detailed information about a specific ban.

**Endpoint**: `GET /api/v1/moderation/ban/:id`

**Authentication**: Required

**Rate Limit**: 60 requests per minute

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**URL Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string (UUID) | Yes | Ban ID |

**Example Request**:
```http
GET /api/v1/moderation/ban/98765432-e89b-12d3-a456-426614174001
```

#### Response

**Success (200 OK)**:
```json
{
  "id": "98765432-e89b-12d3-a456-426614174001",
  "channelId": "123e4567-e89b-12d3-a456-426614174000",
  "userId": "user-uuid-here",
  "bannedBy": "moderator-uuid-here",
  "reason": "Violation of community guidelines",
  "bannedAt": "2024-01-15T10:30:00Z",
  "expiresAt": null,
  "isPermanent": true,
  "username": "banneduser123",
  "bannedByUsername": "moderator456",
  "channelName": "My Community"
}
```

#### Errors

**400 Bad Request** - Invalid ban ID:
```json
{
  "error": "Invalid UUID format for ban ID",
  "code": "INVALID_REQUEST"
}
```

**403 Forbidden** - No permission to view:
```json
{
  "error": "Permission denied: You do not have access to view this ban",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - Ban not found:
```json
{
  "error": "Ban not found",
  "code": "NOT_FOUND"
}
```

---

### Revoke Ban

Removes an existing ban, allowing the user to access the channel again.

**Endpoint**: `DELETE /api/v1/moderation/ban/:id`

**Authentication**: Required

**Permissions**: Channel owner/admin or site moderator

**Rate Limit**: 10 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**URL Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string (UUID) | Yes | Ban ID to revoke |

**Example Request**:
```http
DELETE /api/v1/moderation/ban/98765432-e89b-12d3-a456-426614174001
```

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "message": "Ban revoked successfully"
}
```

#### Errors

**400 Bad Request** - Invalid ban ID:
```json
{
  "error": "Invalid UUID format for ban ID",
  "code": "INVALID_REQUEST"
}
```

**403 Forbidden** - Insufficient permissions:
```json
{
  "error": "Permission denied: You must be a channel owner, admin, or moderator to revoke bans",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - Ban not found:
```json
{
  "error": "Ban not found or already revoked",
  "code": "NOT_FOUND"
}
```

#### Notes

- Creates an audit log entry for the revocation
- Broadcasts unban event via WebSocket
- User can immediately access the channel again

---

### List Moderators

Retrieves a paginated list of moderators for a specific channel.

**Endpoint**: `GET /api/v1/moderation/moderators`

**Authentication**: Required

**Rate Limit**: 60 requests per minute

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| channelId | string (UUID) | Yes | - | Channel ID to list moderators for |
| limit | integer | No | 50 | Number of results per page (max: 100) |
| offset | integer | No | 0 | Number of results to skip |

**Example Request**:
```http
GET /api/v1/moderation/moderators?channelId=123e4567-e89b-12d3-a456-426614174000&limit=20
```

#### Response

**Success (200 OK)**:
```json
{
  "moderators": [
    {
      "id": "member-uuid-here",
      "userId": "user-uuid-here",
      "channelId": "123e4567-e89b-12d3-a456-426614174000",
      "role": "moderator",
      "addedAt": "2024-01-10T09:00:00Z",
      "addedBy": "owner-uuid-here",
      "username": "mod_user123",
      "displayName": "Mod User",
      "avatarUrl": "https://example.com/avatar.jpg"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| moderators | array | Array of moderator objects |
| moderators[].id | string (UUID) | Channel member ID |
| moderators[].userId | string (UUID) | User ID of the moderator |
| moderators[].channelId | string (UUID) | Channel ID |
| moderators[].role | string | Role: "moderator" or "admin" |
| moderators[].addedAt | string (ISO 8601) | When moderator was added |
| moderators[].addedBy | string (UUID) | User who added the moderator |
| moderators[].username | string | Moderator's username |
| moderators[].displayName | string | Display name |
| moderators[].avatarUrl | string | Avatar URL |
| total | integer | Total number of moderators |
| limit | integer | Results per page |
| offset | integer | Current offset |

#### Errors

**400 Bad Request** - Missing or invalid channelId:
```json
{
  "error": "channelId query parameter is required",
  "code": "INVALID_REQUEST"
}
```

**403 Forbidden** - Insufficient permissions:
```json
{
  "error": "Permission denied: You do not have access to view moderators for this channel",
  "code": "FORBIDDEN"
}
```

**500 Internal Server Error**:
```json
{
  "error": "Failed to retrieve moderators",
  "code": "INTERNAL_ERROR"
}
```

---

### Add Moderator

Adds a new moderator to a specific channel.

**Endpoint**: `POST /api/v1/moderation/moderators`

**Authentication**: Required

**Permissions**: Channel owner/admin only

**Rate Limit**: 10 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Body**:
```json
{
  "userId": "user-uuid-to-make-moderator",
  "channelId": "123e4567-e89b-12d3-a456-426614174000",
  "reason": "Trusted community member"
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| userId | string (UUID) | Yes | User ID to add as moderator |
| channelId | string (UUID) | Yes | Channel ID |
| reason | string | No | Reason for adding moderator |

#### Response

**Success (201 Created)**:
```json
{
  "success": true,
  "moderator": {
    "id": "member-uuid-here",
    "userId": "user-uuid-to-make-moderator",
    "channelId": "123e4567-e89b-12d3-a456-426614174000",
    "role": "moderator",
    "addedAt": "2024-01-15T10:30:00Z",
    "addedBy": "owner-uuid-here"
  },
  "message": "Moderator added successfully"
}
```

#### Errors

**400 Bad Request** - Invalid UUID:
```json
{
  "error": "Invalid UUID format for userId",
  "code": "INVALID_REQUEST"
}
```

**400 Bad Request** - User already admin:
```json
{
  "error": "User is already an admin in this channel",
  "code": "INVALID_REQUEST"
}
```

**403 Forbidden** - Not channel owner/admin:
```json
{
  "error": "Permission denied: Only channel owners and admins can add moderators",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - User or channel not found:
```json
{
  "error": "User not found",
  "code": "NOT_FOUND"
}
```

#### Notes

- Creates an audit log entry
- If user is already a member, upgrades them to moderator
- If user is not a member, adds them as a new moderator
- Validates moderator scope for community-specific permissions

---

### Remove Moderator

Removes moderator privileges from a user in a specific channel.

**Endpoint**: `DELETE /api/v1/moderation/moderators/:id`

**Authentication**: Required

**Permissions**: Channel owner/admin only

**Rate Limit**: 10 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**URL Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string (UUID) | Yes | Channel member ID (not user ID) |

**Example Request**:
```http
DELETE /api/v1/moderation/moderators/member-uuid-here
```

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "message": "Moderator removed successfully"
}
```

#### Errors

**400 Bad Request** - Cannot remove channel owner:
```json
{
  "error": "Cannot remove the channel owner",
  "code": "INVALID_REQUEST"
}
```

**403 Forbidden** - Not channel owner/admin:
```json
{
  "error": "Permission denied: Only channel owners and admins can remove moderators",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - Moderator not found:
```json
{
  "error": "Moderator not found",
  "code": "NOT_FOUND"
}
```

#### Notes

- Downgrades moderator role to "member" instead of removing them entirely
- Creates an audit log entry
- Cannot remove the channel owner

---

### Update Moderator Permissions

Updates the permissions for an existing moderator.

**Endpoint**: `PATCH /api/v1/moderation/moderators/:id`

**Authentication**: Required

**Permissions**: Channel owner/admin only

**Rate Limit**: 10 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**URL Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string (UUID) | Yes | Channel member ID |

**Body**:
```json
{
  "permissions": ["manage_bans", "manage_content", "view_reports"]
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| permissions | array of strings | Yes | List of permission names |

**Available Permissions**:
- `manage_bans`: Can ban/unban users
- `manage_content`: Can approve/reject content
- `view_reports`: Can view user reports
- `manage_moderators`: Can add/remove other moderators
- `view_analytics`: Can view moderation analytics

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "moderator": {
    "id": "member-uuid-here",
    "userId": "user-uuid-here",
    "permissions": ["manage_bans", "manage_content", "view_reports"],
    "updatedAt": "2024-01-15T10:30:00Z"
  },
  "message": "Permissions updated successfully"
}
```

#### Errors

**403 Forbidden** - Not channel owner/admin:
```json
{
  "error": "Permission denied: Only channel owners and admins can update moderator permissions",
  "code": "FORBIDDEN"
}
```

**404 Not Found** - Moderator not found:
```json
{
  "error": "Moderator not found",
  "code": "NOT_FOUND"
}
```

#### Notes

- Creates an audit log entry
- Permissions are stored as a JSON array
- Invalid permissions are silently ignored

---

### List Audit Logs

Retrieves moderation audit logs with filtering and pagination.

**Endpoint**: `GET /api/v1/moderation/audit-logs`

**Authentication**: Required

**Permissions**: Admin or moderator only

**Rate Limit**: 60 requests per minute

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| moderator_id | string (UUID) | No | - | Filter by moderator user ID |
| action | string | No | - | Filter by action type |
| start_date | string (YYYY-MM-DD) | No | - | Filter logs from this date |
| end_date | string (YYYY-MM-DD) | No | - | Filter logs until this date |
| limit | integer | No | 100 | Results per page (max: 1000) |
| offset | integer | No | 0 | Results to skip |

**Valid Action Types**:
- `approve`
- `reject`
- `escalate`
- `ban_user`
- `unban_user`
- `add_moderator`
- `remove_moderator`

**Example Request**:
```http
GET /api/v1/moderation/audit-logs?action=ban_user&start_date=2024-01-01&limit=50
```

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "data": [
    {
      "id": "log-uuid-here",
      "moderatorId": "moderator-uuid",
      "moderatorUsername": "mod_user123",
      "action": "ban_user",
      "contentType": "user",
      "contentId": "user-uuid-banned",
      "reason": "Violation of community guidelines",
      "metadata": {
        "channelId": "channel-uuid",
        "duration": "permanent"
      },
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "total": 1250,
    "limit": 50,
    "offset": 0
  }
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| success | boolean | Request success status |
| data | array | Array of audit log entries |
| data[].id | string (UUID) | Audit log entry ID |
| data[].moderatorId | string (UUID) | Moderator who performed the action |
| data[].moderatorUsername | string | Moderator's username |
| data[].action | string | Action performed |
| data[].contentType | string | Type of content affected |
| data[].contentId | string (UUID) | ID of affected content |
| data[].reason | string | Reason for the action |
| data[].metadata | object | Additional action metadata |
| data[].createdAt | string (ISO 8601) | When action was performed |
| meta.total | integer | Total matching logs |
| meta.limit | integer | Results per page |
| meta.offset | integer | Current offset |

#### Errors

**400 Bad Request** - Invalid parameters:
```json
{
  "error": "Invalid action type. Must be one of: approve, reject, escalate, ban_user",
  "code": "INVALID_REQUEST"
}
```

**401 Unauthorized**:
```json
{
  "error": "Authorization token required",
  "code": "UNAUTHORIZED"
}
```

**403 Forbidden** - Not admin/moderator:
```json
{
  "error": "Permission denied: You must be an admin or moderator to view audit logs",
  "code": "FORBIDDEN"
}
```

---

### Export Audit Logs

Exports audit logs to CSV format for a specified date range.

**Endpoint**: `GET /api/v1/moderation/audit-logs/export`

**Authentication**: Required

**Permissions**: Admin or moderator only

**Rate Limit**: 10 requests per hour

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| start_date | string (YYYY-MM-DD) | No | Start date for export |
| end_date | string (YYYY-MM-DD) | No | End date for export |

**Example Request**:
```http
GET /api/v1/moderation/audit-logs/export?start_date=2024-01-01&end_date=2024-01-31
```

#### Response

**Success (200 OK)**:

```csv
id,moderator_id,moderator_username,action,content_type,content_id,reason,created_at
log-uuid-1,mod-uuid,mod_user123,ban_user,user,user-uuid,Violation of rules,2024-01-15T10:30:00Z
log-uuid-2,mod-uuid,mod_user123,approve,submission,sub-uuid,Good content,2024-01-15T11:00:00Z
```

**Content-Type**: `text/csv`

**Content-Disposition**: `attachment; filename="audit-logs-2024-01-01-to-2024-01-31.csv"`

#### Errors

**401 Unauthorized**:
```json
{
  "error": "Authorization token required",
  "code": "UNAUTHORIZED"
}
```

**403 Forbidden**:
```json
{
  "error": "Permission denied: You must be an admin or moderator to export audit logs",
  "code": "FORBIDDEN"
}
```

#### Notes

- Exports all matching logs without pagination limits
- Large exports may take longer to process
- CSV encoding uses UTF-8

---

### Get Audit Log

Retrieves a specific audit log entry by ID.

**Endpoint**: `GET /api/v1/moderation/audit-logs/:id`

**Authentication**: Required

**Permissions**: Admin or moderator only

**Rate Limit**: 60 requests per minute

#### Request

**Headers**:
```http
Authorization: Bearer <token>
```

**URL Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string (UUID) | Yes | Audit log entry ID |

**Example Request**:
```http
GET /api/v1/moderation/audit-logs/log-uuid-here
```

#### Response

**Success (200 OK)**:
```json
{
  "id": "log-uuid-here",
  "moderatorId": "moderator-uuid",
  "moderatorUsername": "mod_user123",
  "action": "ban_user",
  "contentType": "user",
  "contentId": "user-uuid-banned",
  "reason": "Violation of community guidelines",
  "metadata": {
    "channelId": "channel-uuid",
    "duration": "permanent",
    "userAgent": "Mozilla/5.0...",
    "ipAddress": "192.168.1.1"
  },
  "createdAt": "2024-01-15T10:30:00Z"
}
```

#### Errors

**401 Unauthorized**:
```json
{
  "error": "Authorization token required",
  "code": "UNAUTHORIZED"
}
```

**403 Forbidden**:
```json
{
  "error": "Permission denied: You must be an admin or moderator to view audit logs",
  "code": "FORBIDDEN"
}
```

**404 Not Found**:
```json
{
  "error": "Audit log not found",
  "code": "NOT_FOUND"
}
```

---

## Code Examples

### cURL Examples

#### Sync Bans

```bash
curl -X POST https://api.clpr.tv/api/v1/moderation/sync-bans \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

#### List Bans

```bash
curl -X GET "https://api.clpr.tv/api/v1/moderation/bans?channelId=123e4567-e89b-12d3-a456-426614174000&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Create Ban

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

#### Revoke Ban

```bash
curl -X DELETE https://api.clpr.tv/api/v1/moderation/ban/ban-uuid-here \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Add Moderator

```bash
curl -X POST https://api.clpr.tv/api/v1/moderation/moderators \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-uuid-to-make-moderator",
    "channelId": "123e4567-e89b-12d3-a456-426614174000",
    "reason": "Trusted community member"
  }'
```

#### List Audit Logs

```bash
curl -X GET "https://api.clpr.tv/api/v1/moderation/audit-logs?action=ban_user&limit=50" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### JavaScript Examples

#### Using Fetch API

```javascript
// Configuration
const API_BASE = 'https://api.clpr.tv/api/v1/moderation';
const AUTH_TOKEN = 'YOUR_TOKEN';

// Helper function for API calls
async function apiCall(endpoint, options = {}) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'API request failed');
  }

  return response.json();
}

// Sync bans
async function syncBans(channelId) {
  return apiCall('/sync-bans', {
    method: 'POST',
    body: JSON.stringify({ channel_id: channelId }),
  });
}

// List bans
async function listBans(channelId, limit = 20, offset = 0) {
  const params = new URLSearchParams({ channelId, limit, offset });
  return apiCall(`/bans?${params}`);
}

// Create ban
async function createBan(channelId, userId, reason) {
  return apiCall('/ban', {
    method: 'POST',
    body: JSON.stringify({ channelId, userId, reason }),
  });
}

// Revoke ban
async function revokeBan(banId) {
  return apiCall(`/ban/${banId}`, {
    method: 'DELETE',
  });
}

// Add moderator
async function addModerator(userId, channelId, reason) {
  return apiCall('/moderators', {
    method: 'POST',
    body: JSON.stringify({ userId, channelId, reason }),
  });
}

// List audit logs
async function listAuditLogs(filters = {}) {
  const params = new URLSearchParams(filters);
  return apiCall(`/audit-logs?${params}`);
}

// Usage examples
try {
  // Sync bans
  const syncResult = await syncBans('123e4567-e89b-12d3-a456-426614174000');
  console.log('Sync started:', syncResult.job_id);

  // List bans
  const bans = await listBans('123e4567-e89b-12d3-a456-426614174000', 20, 0);
  console.log('Total bans:', bans.total);

  // Create ban
  const ban = await createBan(
    '123e4567-e89b-12d3-a456-426614174000',
    'user-uuid-to-ban',
    'Violation of rules'
  );
  console.log('Ban created:', ban.id);

} catch (error) {
  console.error('API Error:', error.message);
}
```

#### Using Axios

```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'https://api.clpr.tv/api/v1/moderation',
  headers: {
    'Authorization': `Bearer ${process.env.API_TOKEN}`,
    'Content-Type': 'application/json',
  },
});

// Sync bans
async function syncBans(channelId) {
  const { data } = await api.post('/sync-bans', {
    channel_id: channelId,
  });
  return data;
}

// List bans with pagination
async function listBans(channelId, limit = 20, offset = 0) {
  const { data } = await api.get('/bans', {
    params: { channelId, limit, offset },
  });
  return data;
}

// Create ban
async function createBan(channelId, userId, reason) {
  const { data } = await api.post('/ban', {
    channelId,
    userId,
    reason,
  });
  return data;
}

// Remove moderator
async function removeModerator(memberId) {
  const { data } = await api.delete(`/moderators/${memberId}`);
  return data;
}

// Export audit logs
async function exportAuditLogs(startDate, endDate) {
  const { data } = await api.get('/audit-logs/export', {
    params: { start_date: startDate, end_date: endDate },
    responseType: 'blob',
  });
  return data;
}
```

---

### Go Examples

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	APIBase = "https://api.clpr.tv/api/v1/moderation"
)

// API Client
type ModerationClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewClient creates a new moderation API client
func NewClient(token string) *ModerationClient {
	return &ModerationClient{
		BaseURL: APIBase,
		Token:   token,
		Client:  &http.Client{},
	}
}

// Helper method for making requests
func (c *ModerationClient) do(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.Client.Do(req)
}

// SyncBansRequest represents the sync bans request
type SyncBansRequest struct {
	ChannelID string `json:"channel_id"`
}

// SyncBansResponse represents the sync bans response
type SyncBansResponse struct {
	Status  string `json:"status"`
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

// SyncBans synchronizes bans from Twitch
func (c *ModerationClient) SyncBans(channelID string) (*SyncBansResponse, error) {
	req := SyncBansRequest{ChannelID: channelID}
	resp, err := c.do("POST", "/sync-bans", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sync failed: status %d", resp.StatusCode)
	}

	var result SyncBansResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Ban represents a ban
type Ban struct {
	ID          string  `json:"id"`
	ChannelID   string  `json:"channelId"`
	UserID      string  `json:"userId"`
	BannedBy    string  `json:"bannedBy"`
	Reason      string  `json:"reason"`
	BannedAt    string  `json:"bannedAt"`
	ExpiresAt   *string `json:"expiresAt"`
	IsPermanent bool    `json:"isPermanent"`
}

// ListBansResponse represents the list bans response
type ListBansResponse struct {
	Bans   []Ban `json:"bans"`
	Total  int   `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// ListBans retrieves a paginated list of bans
func (c *ModerationClient) ListBans(channelID string, limit, offset int) (*ListBansResponse, error) {
	params := url.Values{}
	params.Add("channelId", channelID)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))

	resp, err := c.do("GET", "/bans?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list bans failed: status %d", resp.StatusCode)
	}

	var result ListBansResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateBanRequest represents the create ban request
type CreateBanRequest struct {
	ChannelID string `json:"channelId"`
	UserID    string `json:"userId"`
	Reason    string `json:"reason,omitempty"`
}

// CreateBan creates a new ban
func (c *ModerationClient) CreateBan(channelID, userID, reason string) (*Ban, error) {
	req := CreateBanRequest{
		ChannelID: channelID,
		UserID:    userID,
		Reason:    reason,
	}

	resp, err := c.do("POST", "/ban", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create ban failed: status %d", resp.StatusCode)
	}

	var ban Ban
	if err := json.NewDecoder(resp.Body).Decode(&ban); err != nil {
		return nil, err
	}

	return &ban, nil
}

// RevokeBan removes an existing ban
func (c *ModerationClient) RevokeBan(banID string) error {
	resp, err := c.do("DELETE", "/ban/"+banID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke ban failed: status %d", resp.StatusCode)
	}

	return nil
}

// Moderator represents a moderator
type Moderator struct {
	ID        string   `json:"id"`
	UserID    string   `json:"userId"`
	ChannelID string   `json:"channelId"`
	Role      string   `json:"role"`
	AddedAt   string   `json:"addedAt"`
	AddedBy   string   `json:"addedBy"`
	Username  string   `json:"username"`
}

// AddModeratorRequest represents the add moderator request
type AddModeratorRequest struct {
	UserID    string `json:"userId"`
	ChannelID string `json:"channelId"`
	Reason    string `json:"reason,omitempty"`
}

// AddModeratorResponse represents the add moderator response
type AddModeratorResponse struct {
	Success   bool      `json:"success"`
	Moderator Moderator `json:"moderator"`
	Message   string    `json:"message"`
}

// AddModerator adds a new moderator
func (c *ModerationClient) AddModerator(userID, channelID, reason string) (*AddModeratorResponse, error) {
	req := AddModeratorRequest{
		UserID:    userID,
		ChannelID: channelID,
		Reason:    reason,
	}

	resp, err := c.do("POST", "/moderators", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("add moderator failed: status %d", resp.StatusCode)
	}

	var result AddModeratorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Usage example
func main() {
	client := NewClient("YOUR_TOKEN")

	// Sync bans
	syncResult, err := client.SyncBans("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		fmt.Printf("Error syncing bans: %v\n", err)
		return
	}
	fmt.Printf("Sync started with job ID: %s\n", syncResult.JobID)

	// List bans
	bans, err := client.ListBans("123e4567-e89b-12d3-a456-426614174000", 20, 0)
	if err != nil {
		fmt.Printf("Error listing bans: %v\n", err)
		return
	}
	fmt.Printf("Total bans: %d\n", bans.Total)

	// Create ban
	ban, err := client.CreateBan(
		"123e4567-e89b-12d3-a456-426614174000",
		"user-uuid-to-ban",
		"Violation of community guidelines",
	)
	if err != nil {
		fmt.Printf("Error creating ban: %v\n", err)
		return
	}
	fmt.Printf("Ban created with ID: %s\n", ban.ID)
}
```

---

## Permission Matrix

### Actions by Role

| Action | User | Moderator | Channel Owner/Admin | Site Admin |
|--------|------|-----------|---------------------|------------|
| **Bans** |
| View channel bans | ❌ | ✅ (own channels) | ✅ | ✅ |
| Create ban | ❌ | ✅ (own channels) | ✅ | ✅ |
| Revoke ban | ❌ | ✅ (own channels) | ✅ | ✅ |
| Sync bans from Twitch | ❌ | ✅ (own channels) | ✅ | ✅ |
| View ban details | ❌ | ✅ (own channels) | ✅ | ✅ |
| **Moderators** |
| View moderators | ✅ | ✅ | ✅ | ✅ |
| Add moderator | ❌ | ❌ | ✅ | ✅ |
| Remove moderator | ❌ | ❌ | ✅ | ✅ |
| Update moderator permissions | ❌ | ❌ | ✅ | ✅ |
| **Audit Logs** |
| View audit logs | ❌ | ✅ | ✅ | ✅ |
| Export audit logs | ❌ | ✅ | ✅ | ✅ |
| View specific audit log | ❌ | ✅ | ✅ | ✅ |

### Moderator Scopes

**Global Moderators**:
- Can moderate any channel/community
- Assigned by site admins only
- Full moderation permissions across platform

**Community Moderators**:
- Can only moderate assigned channels/communities
- Assigned by channel owners or admins
- Permissions scoped to specific channels

---

## Deployment Guide

### Prerequisites

1. **Environment Variables**

Required environment variables for moderation features:

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
RATE_LIMIT_WHITELIST_IPS=127.0.0.1,::1

# Audit Logging
AUDIT_LOG_ENABLED=true
AUDIT_LOG_RETENTION_DAYS=90
```

2. **Database Migrations**

Run moderation-related migrations:

```bash
# Navigate to backend directory
cd backend

# Run all migrations
make migrate-up

# Or run specific moderation migrations
psql $DATABASE_URL -f migrations/XXX_create_bans_table.sql
psql $DATABASE_URL -f migrations/XXX_create_audit_logs_table.sql
psql $DATABASE_URL -f migrations/XXX_create_channel_members_table.sql
```

3. **Dependencies**

Install required dependencies:

```bash
# Backend (Go)
cd backend
go mod download

# Redis
docker run -d -p 6379:6379 redis:7-alpine

# PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_DB=clpr \
  -e POSTGRES_USER=clpr \
  -e POSTGRES_PASSWORD=clpr \
  postgres:15-alpine
```

### Configuration

1. **Rate Limiting Configuration**

Update rate limits in `backend/config/config.go`:

```go
type RateLimitConfig struct {
    Enabled       bool
    WhitelistIPs  string
    DefaultLimit  int
    BurstLimit    int
}
```

2. **Twitch OAuth Scopes**

Required Twitch OAuth scopes for ban sync:

- `moderator:read:banned_users`
- `moderator:manage:banned_users`
- `channel:moderate`

Configure in Twitch Developer Console.

3. **Audit Log Retention**

Configure audit log retention policy:

```sql
-- Create cleanup job (run daily)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs()
RETURNS void AS $$
BEGIN
  DELETE FROM audit_logs
  WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Schedule cleanup (using pg_cron)
SELECT cron.schedule('cleanup-audit-logs', '0 2 * * *', 'SELECT cleanup_old_audit_logs()');
```

### Building

```bash
# Build backend
cd backend
make build

# Or using Docker
docker build -t clpr-backend .
```

### Running

#### Development

```bash
# Start backend
cd backend
make run

# Or with hot reload
make dev
```

#### Production

```bash
# Using Docker Compose
docker-compose -f docker-compose.prod.yml up -d

# Or using binary
./backend/bin/api
```

### Health Checks

Verify moderation services are running:

```bash
# Health check endpoint
curl http://localhost:8080/health

# Specific checks
curl http://localhost:8080/health/db
curl http://localhost:8080/health/redis
```

Expected response:

```json
{
  "status": "healthy",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "twitch_api": "healthy"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Monitoring

1. **Prometheus Metrics**

Moderation metrics are exposed at `/metrics`:

```
# Ban operations
clpr_moderation_bans_created_total
clpr_moderation_bans_revoked_total
clpr_moderation_ban_sync_duration_seconds

# Moderator operations
clpr_moderation_moderators_added_total
clpr_moderation_moderators_removed_total

# Audit logs
clpr_moderation_audit_logs_created_total
```

2. **Logging**

Moderation actions are logged with structured logging:

```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "action": "ban_created",
  "moderator_id": "uuid",
  "user_id": "uuid",
  "channel_id": "uuid",
  "reason": "Violation of rules"
}
```

3. **Alerting**

Set up alerts for critical events:

- High rate of ban failures
- Twitch API sync failures
- Unauthorized access attempts
- Rate limit violations

### Security

1. **HTTPS/TLS**

Always use HTTPS in production:

```nginx
server {
    listen 443 ssl http2;
    server_name api.clpr.tv;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location /api/v1/moderation {
        proxy_pass http://backend:8080;
        proxy_set_header Authorization $http_authorization;
    }
}
```

2. **API Key Rotation**

Rotate Twitch API credentials regularly:

```bash
# Update environment variables
export TWITCH_CLIENT_ID=new_client_id
export TWITCH_CLIENT_SECRET=new_secret

# Restart service
docker-compose restart backend
```

3. **Access Control**

Implement IP whitelisting for sensitive operations:

```env
RATE_LIMIT_WHITELIST_IPS=10.0.0.0/8,172.16.0.0/12
```

### Troubleshooting

#### Ban Sync Not Working

1. Verify Twitch OAuth scopes
2. Check Twitch API credentials
3. Verify moderator permissions on Twitch
4. Check audit logs for errors

```bash
# Check logs
docker logs clpr-backend | grep "ban_sync"

# Test Twitch API connection
curl -X POST http://localhost:8080/api/v1/moderation/sync-bans \
  -H "Authorization: Bearer TOKEN" \
  -d '{"channel_id": "uuid"}'
```

#### Rate Limiting Issues

1. Check Redis connection
2. Verify rate limit configuration
3. Check IP whitelist

```bash
# Test Redis connection
redis-cli ping

# Check rate limit status
curl http://localhost:8080/api/v1/health/redis
```

#### Permission Errors

1. Verify JWT token is valid
2. Check user role and permissions
3. Verify channel membership

```bash
# Decode JWT token
echo "YOUR_TOKEN" | base64 -d | jq

# Check user permissions
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer TOKEN"
```

---

## Additional Resources

- [OpenAPI Specification](/docs/openapi/openapi.yaml)
- [Authentication Guide](/docs/backend/authentication.md)
- [Authorization Framework](/docs/backend/authorization-framework.md)
- [Rate Limiting](/docs/backend/rate-limiting.md)
- [GitHub Issues](https://git.subcult.tv/subculture-collective/clpr/issues)
- [API Status Page](https://status.clpr.tv)

---

**Last Updated**: 2024-01-15  
**Version**: 1.0.0  
**Maintainer**: team-core
