---
title: "Permission Model"
summary: "Comprehensive guide to the Clipper permission system for operators and developers."
tags: ["backend", "permissions", "security", "moderation", "authorization"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["permission-model", "permissions", "moderation-permissions"]
---

# Permission Model

This document provides a comprehensive guide to Clipper's permission system for operators and developers.

## Table of Contents

- [Overview](#overview)
- [Permission Hierarchy](#permission-hierarchy)
- [Role Definitions](#role-definitions)
- [Permission Matrix](#permission-matrix)
- [Scope System](#scope-system)
- [Role Examples](#role-examples)
- [Permission Check Code Flow](#permission-check-code-flow)
- [Implementation Guide](#implementation-guide)
- [Troubleshooting](#troubleshooting)
- [Related Documentation](#related-documentation)

## Overview

Clipper uses a dual-layer permission system combining **Roles** and **Account Types** to provide flexible, granular access control:

- **Roles** (Legacy): Basic access control (user, moderator, admin)
- **Account Types** (Current): Granular permission system (member, broadcaster, moderator, community_moderator, admin)

### Design Principles

1. **Hierarchical Permissions**: Higher account types inherit permissions from lower types
2. **Scope-Based Access**: Community moderators are limited to specific channels
3. **Backend Enforcement**: All authorization checks enforced server-side
4. **Audit Trail**: All moderation actions are logged
5. **Defense in Depth**: Multiple layers of permission checks

## Permission Hierarchy

The permission system follows a strict hierarchy where higher account types inherit all permissions from lower types:

```
Regular User (member)
    ↓ inherits all + adds broadcaster features
Broadcaster
    ↓ inherits all + adds sitewide moderation
Site Moderator (moderator)
    ↓ inherits all + adds admin capabilities
Admin
```

**Community Moderator** is a special lateral role with limited, channel-scoped permissions (does not inherit from broadcaster).

### Account Types

| Account Type | Description | Scope |
|-------------|-------------|-------|
| **member** | Regular user with basic permissions | N/A |
| **broadcaster** | Content creator with analytics access | N/A |
| **moderator** | Sitewide moderator with cross-channel access | Sitewide |
| **community_moderator** | Channel-specific moderator | Channel-scoped |
| **admin** | Full administrative access | Sitewide |

### Legacy Roles

For backward compatibility, the system maintains legacy roles:

| Role | Maps To |
|------|---------|
| user | member or broadcaster |
| moderator | moderator or community_moderator |
| admin | admin |

> **Note**: New implementations should use Account Types. Roles are maintained for legacy route guards and backward compatibility.

## Role Definitions

### Regular User (Account Type: member)

Standard account for all registered users.

**Capabilities:**
- Create and submit clips for moderation
- Comment on clips
- Vote on clips and comments
- Follow broadcasters and channels
- Add clips to favorites
- View leaderboards and public analytics
- Manage own profile and settings

**Restrictions:**
- No moderation capabilities
- No access to admin panel
- No access to broadcaster analytics
- Cannot manage other users

**Typical Use Cases:**
- Casual viewers watching and voting on clips
- Community members participating in discussions
- Users curating personal clip collections

---

### Broadcaster (Account Type: broadcaster)

Content creators who can claim and manage their broadcaster profiles.

**Inherits:** All member permissions

**Additional Permissions:**
- `view:broadcaster_analytics` - Access to broadcaster-specific analytics
- `claim:broadcaster_profile` - Ability to claim broadcaster profile

**Capabilities:**
- All member capabilities
- View analytics for their claimed broadcaster profile
- Claim broadcaster profile ownership

**Restrictions:**
- No moderation capabilities
- No access to admin panel
- Cannot moderate other users' content

**Typical Use Cases:**
- Twitch streamers tracking their clip performance
- Content creators analyzing engagement metrics
- Broadcasters managing their profile information

---

### Community Moderator (Account Type: community_moderator)

Channel-specific moderators with limited, scoped permissions.

> **Note**: Community moderators are dedicated moderation accounts that **only** have moderation permissions. They do not inherit basic user permissions (submit, comment, vote, follow). This is intentional to maintain clear separation between moderation and participation roles. If a user needs both moderation and participation capabilities, they should use a separate user account for participation.

**Permissions (4 total):**
- `community:moderate` - Core community moderation capability
- `moderate:users` - Moderate user actions within scope
- `view:channel_analytics` - Access channel-specific analytics
- `manage:moderators` - Manage other moderators in their channel

**Capabilities:**
- Moderate content **only** in assigned channel(s)
- Review and action reports for assigned channels
- Ban/unban users from assigned channels
- View analytics for assigned channels
- Manage other moderators for assigned channels

**Restrictions:**
- **Cannot** submit clips, comment, vote, or follow (use separate user account)
- **Cannot** access content outside assigned channels
- **Cannot** view sitewide analytics
- **Cannot** access admin panel
- **Cannot** manage system configuration
- **Cannot** moderate across all channels

**Channel Scope Validation:**
- Must have `moderator_scope = 'community'`
- Must have at least one channel in `moderation_channels` array
- Permission checks automatically validate channel access
- 403 error returned if accessing out-of-scope channel

**Typical Use Cases:**
- Channel-specific moderators for large communities
- Broadcasters moderating their own channel's content
- Community leaders managing specific gaming communities

---

### Site Moderator (Account Type: moderator)

Trusted community members with sitewide moderation access.

**Inherits:** All broadcaster permissions

**Additional Permissions:**
- `moderate:content` - Moderate clips, comments, submissions
- `moderate:users` - Ban/unban users, manage user accounts
- `create:discovery_lists` - Create and manage discovery lists
- `manage:users` - User management capabilities

**Capabilities:**
- All broadcaster capabilities
- Moderate content across **all channels**
- Access moderation queue for all submissions
- Approve/reject clip submissions sitewide
- Remove inappropriate content (clips and comments)
- Review and action reports from any channel
- Ban/unban users sitewide
- Manage tags
- Create and manage discovery lists
- View moderation audit logs

**Restrictions:**
- **Cannot** delete clips (admin only)
- **Cannot** change user account types (admin only)
- **Note**: Has `manage:users` permission but with limited scope - can manage user bans/unbans and basic user actions, but cannot promote users to admin or change account types
- **Cannot** access full analytics dashboard
- **Cannot** trigger system-level operations
- **Cannot** manage system configuration

**Scope:**
- `moderator_scope = 'site'` (or empty for legacy accounts)
- `moderation_channels` must be empty
- No channel scope restrictions apply

**Typical Use Cases:**
- Experienced community moderators
- Trusted members handling sitewide moderation queue
- Content curators managing discovery lists

---

### Admin (Account Type: admin)

Full administrative access to the entire platform.

**Inherits:** All moderator and community moderator permissions

**Additional Permissions:**
- `manage:users` - Full user management
- `manage:system` - System configuration and management
- `view:analytics_dashboard` - Full platform analytics
- `moderate:override` - Override any moderation decision

**Capabilities:**
- **All permissions** from all other account types
- Full user management (ban, unban, change roles, change account types)
- Delete clips and content
- Trigger clip sync operations
- Access full platform analytics dashboard
- Manage system configuration
- Override any moderation decision
- Manage all tags and discovery lists
- Full audit log access and export

**Restrictions:**
- None (full administrative access)

**Typical Use Cases:**
- Platform operators
- Senior administrators
- System maintenance and configuration
- Emergency intervention and override

## Permission Matrix

### All Permissions by Account Type

| Permission | Member | Broadcaster | Community Mod | Site Mod | Admin |
|-----------|:------:|:-----------:|:-------------:|:--------:|:-----:|
| **Basic Permissions** |
| `create:submission` | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| `create:comment` | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| `create:vote` | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| `create:follow` | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| **Broadcaster Permissions** |
| `view:broadcaster_analytics` | ❌ | ✅ | ❌ | ✅ | ✅ |
| `claim:broadcaster_profile` | ❌ | ✅ | ❌ | ✅ | ✅ |
| **Community Moderator Permissions** |
| `community:moderate` | ❌ | ❌ | ✅ | ❌ | ✅ |
| `moderate:users` | ❌ | ❌ | ✅ | ✅ | ✅ |
| `view:channel_analytics` | ❌ | ❌ | ✅ | ❌ | ✅ |
| `manage:moderators` | ❌ | ❌ | ✅ | ❌ | ✅ |
| **Site Moderator Permissions** |
| `moderate:content` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `create:discovery_lists` | ❌ | ❌ | ❌ | ✅ | ✅ |
| **Admin Permissions** |
| `manage:users` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `manage:system` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `view:analytics_dashboard` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `moderate:override` | ❌ | ❌ | ❌ | ❌ | ✅ |

### Feature Access Matrix

| Feature | Member | Broadcaster | Community Mod | Site Mod | Admin |
|---------|:------:|:-----------:|:-------------:|:--------:|:-----:|
| **Content Actions** |
| View clips | ✅ | ✅ | ✅ | ✅ | ✅ |
| Submit clips | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| Comment on clips | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| Vote on content | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| Add to favorites | ✅ | ✅ | ❌¹ | ✅ | ✅ |
| **Broadcaster Features** |
| View broadcaster analytics | ❌ | ✅ (own) | ❌ | ✅ | ✅ |
| Claim broadcaster profile | ❌ | ✅ | ❌ | ✅ | ✅ |
| **Moderation** |
| Access moderation panel | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| View moderation queue | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| Approve/reject submissions | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| Remove content | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| View reports | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| Action reports | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| Ban users | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| Unban users | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| Manage tags | ❌ | ❌ | ❌ | ✅ | ✅ |
| Create discovery lists | ❌ | ❌ | ❌ | ✅ | ✅ |
| View audit logs | ❌ | ❌ | ✅ (scoped) | ✅ | ✅ |
| **Administration** |
| Delete clips | ❌ | ❌ | ❌ | ❌ | ✅ |
| Change user roles | ❌ | ❌ | ❌ | ❌ | ✅ |
| Change account types | ❌ | ❌ | ❌ | ❌ | ✅ |
| Trigger clip sync | ❌ | ❌ | ❌ | ❌ | ✅ |
| View analytics dashboard | ❌ | ❌ | ❌ | ❌ | ✅ |
| Manage system config | ❌ | ❌ | ❌ | ❌ | ✅ |

**Legend:**
- ✅ = Full access
- ✅ (own) = Only for own resources
- ✅ (scoped) = Limited to assigned channels
- ❌ = No access
- ❌¹ = No access (Community moderators are moderation-only accounts; use separate user account for participation)

## Scope System

### Moderator Scopes

The system supports two moderator scopes:

#### Site Scope (`moderator_scope = 'site'`)

**Characteristics:**
- Cross-channel visibility and access
- No channel restrictions
- `moderation_channels` array must be empty
- Used by Site Moderators and Admins

**Access:**
- Can moderate any channel
- Can view all reports
- Can access all moderation queues
- No scope validation on channel_id

**Example:**
```sql
UPDATE users SET 
  account_type = 'moderator',
  moderator_scope = 'site',
  moderation_channels = '{}'
WHERE username = 'site_moderator';
```

---

#### Community Scope (`moderator_scope = 'community'`)

**Characteristics:**
- Channel-specific access only
- Restricted to assigned channels
- `moderation_channels` array must contain at least one channel UUID
- Used by Community Moderators

**Access:**
- Can only moderate assigned channels
- Permission checks validate channel_id against `moderation_channels`
- 403 Forbidden if accessing out-of-scope channel
- Channel scope applies to all moderation actions

**Example:**
```sql
UPDATE users SET 
  account_type = 'community_moderator',
  moderator_scope = 'community',
  moderation_channels = ARRAY['550e8400-e29b-41d4-a716-446655440000'::uuid]
WHERE username = 'community_mod_user';
```

---

### Channel Scope Validation

For community moderators, the system automatically validates channel access:

**How it works:**
1. User makes request to a moderation endpoint
2. Middleware checks if user is a community moderator
3. If yes, extracts `channel_id` from request (path param, query param, or context)
4. Validates `channel_id` against user's `moderation_channels` array
5. Returns 403 if channel is not in the allowed list

**Channel ID Resolution Priority:**
1. Path parameter: `/api/v1/channels/:channel_id/moderate`
2. Query parameter: `/api/v1/moderate?channel_id=xxx`
3. Context value: Set by handler before middleware

**Example Request Flow:**
```
POST /api/v1/channels/550e8400-e29b-41d4-a716-446655440000/reports/action
Authorization: Bearer <token_for_community_mod>

1. AuthMiddleware validates token, sets user in context
2. RequirePermission('moderate:users') checks permission
3. If community mod, extracts channel_id from path: 550e8400-e29b-41d4-a716-446655440000
4. Validates channel_id is in user.moderation_channels
5. If yes → Continue to handler
6. If no → Return 403 Forbidden
```

---

### Scope Validation Rules

| Account Type | Scope Required | Channel Validation | Access Level |
|-------------|----------------|-------------------|--------------|
| member | None | No | N/A |
| broadcaster | None | No | N/A |
| community_moderator | `community` | **Yes** | Channel-specific |
| moderator | `site` or empty | No | Sitewide |
| admin | `site` or empty | No | Sitewide |

**Important:**
- Site moderators should have `moderator_scope = 'site'` and empty `moderation_channels`
- Community moderators **must** have `moderator_scope = 'community'` and non-empty `moderation_channels`
- Invalid configuration (e.g., site scope with channels) will fail `IsValidModerator()` check

## Role Examples

### Example 1: Regular User Journey

**Scenario:** New user watching and engaging with clips

```
User: alice
Account Type: member
Role: user
Moderator Scope: (empty)
```

**Can Do:**
- ✅ Browse and search clips
- ✅ Vote on clips: `user.Can('create:vote')` → true
- ✅ Comment on clips: `user.Can('create:comment')` → true
- ✅ Submit new clips: `user.Can('create:submission')` → true
- ✅ Add clips to favorites
- ✅ Follow broadcasters

**Cannot Do:**
- ❌ Access moderation panel: `user.Can('moderate:content')` → false
- ❌ Ban users: `user.Can('moderate:users')` → false
- ❌ View broadcaster analytics: `user.Can('view:broadcaster_analytics')` → false
- ❌ Approve submissions

---

### Example 2: Broadcaster Managing Analytics

**Scenario:** Twitch streamer tracking their clips

```
User: streamer_bob
Account Type: broadcaster
Role: user
Moderator Scope: (empty)
```

**Can Do:**
- ✅ All member capabilities
- ✅ Claim broadcaster profile: `user.Can('claim:broadcaster_profile')` → true
- ✅ View own broadcaster analytics: `user.Can('view:broadcaster_analytics')` → true
- ✅ Track clip performance metrics
- ✅ See engagement statistics

**Cannot Do:**
- ❌ Moderate other users' content
- ❌ Access moderation queue
- ❌ View channel analytics: `user.Can('view:channel_analytics')` → false

---

### Example 3: Community Moderator for Single Channel

**Scenario:** Moderator managing "Fortnite" channel only

```
User: mod_carol
Account Type: community_moderator
Role: moderator
Moderator Scope: community
Moderation Channels: [uuid-fortnite-channel]
```

**Can Do (within Fortnite channel only):**
- ✅ Review submissions: `user.Can('community:moderate')` → true
- ✅ Approve/reject clips (channel scope validated)
- ✅ Ban users from Fortnite channel: `user.Can('moderate:users')` → true
- ✅ View Fortnite channel analytics: `user.Can('view:channel_analytics')` → true
- ✅ Manage moderators for Fortnite channel

**Cannot Do:**
- ❌ Moderate Valorant channel content (403 Forbidden - channel scope violation)
- ❌ Access sitewide moderation queue
- ❌ View sitewide analytics
- ❌ Create discovery lists: `user.Can('create:discovery_lists')` → false

**Request Examples:**
```bash
# ✅ ALLOWED - Fortnite channel
POST /api/v1/channels/uuid-fortnite-channel/reports/123/action
# Middleware validates: uuid-fortnite-channel in moderation_channels → Pass

# ❌ FORBIDDEN - Valorant channel
POST /api/v1/channels/uuid-valorant-channel/reports/456/action
# Middleware validates: uuid-valorant-channel NOT in moderation_channels → 403 Forbidden
```

---

### Example 4: Site Moderator

**Scenario:** Trusted moderator managing sitewide content

```
User: mod_dave
Account Type: moderator
Role: moderator
Moderator Scope: site
Moderation Channels: []
```

**Can Do:**
- ✅ All broadcaster capabilities
- ✅ Moderate **any** channel: `user.Can('moderate:content')` → true
- ✅ Access full moderation queue
- ✅ Ban/unban users sitewide: `user.Can('moderate:users')` → true
- ✅ Approve/reject submissions from all channels
- ✅ Remove inappropriate content anywhere
- ✅ Create discovery lists: `user.Can('create:discovery_lists')` → true
- ✅ Manage tags
- ✅ View audit logs

**Cannot Do:**
- ❌ Delete clips: Only admin can delete
- ❌ Change user roles: `user.Can('manage:users')` → true (can manage, but not change account types - admin only)
- ❌ Access analytics dashboard: `user.Can('view:analytics_dashboard')` → false
- ❌ Trigger system operations: `user.Can('manage:system')` → false

**No Channel Restrictions:**
```bash
# ✅ All allowed - no scope validation for site moderators
POST /api/v1/channels/uuid-fortnite-channel/moderate
POST /api/v1/channels/uuid-valorant-channel/moderate
POST /api/v1/channels/uuid-minecraft-channel/moderate
```

---

### Example 5: Platform Admin

**Scenario:** System administrator with full access

```
User: admin_eve
Account Type: admin
Role: admin
Moderator Scope: site
Moderation Channels: []
```

**Can Do:**
- ✅ **Everything** - all permissions: `user.IsAdmin()` → true
- ✅ Delete clips permanently
- ✅ Change user roles: `UPDATE users SET role = 'moderator'`
- ✅ Change account types: `UPDATE users SET account_type = 'community_moderator'`
- ✅ Trigger clip sync operations
- ✅ Access analytics dashboard: `user.Can('view:analytics_dashboard')` → true
- ✅ Manage system configuration: `user.Can('manage:system')` → true
- ✅ Override any moderation decision: `user.Can('moderate:override')` → true
- ✅ Full audit log access and export

**No Restrictions:**
- Admin account type grants all permissions automatically
- `user.Can(any_permission)` always returns true for admins

---

### Example 6: Community Moderator for Multiple Channels

**Scenario:** Moderator managing both "Fortnite" and "Valorant" channels

```
User: mod_frank
Account Type: community_moderator
Role: moderator
Moderator Scope: community
Moderation Channels: [uuid-fortnite-channel, uuid-valorant-channel]
```

**Can Do:**
- ✅ Moderate Fortnite channel
- ✅ Moderate Valorant channel
- ✅ View analytics for both channels
- ✅ Manage moderators for both channels

**Cannot Do:**
- ❌ Moderate Minecraft channel (not in scope)
- ❌ Access sitewide features

**Scope Validation:**
```bash
# ✅ ALLOWED
POST /api/v1/channels/uuid-fortnite-channel/moderate
POST /api/v1/channels/uuid-valorant-channel/moderate

# ❌ FORBIDDEN
POST /api/v1/channels/uuid-minecraft-channel/moderate
```

## Permission Check Code Flow

### Overview

The permission system uses multiple layers of defense with checks at the route, middleware, and service layers.

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Client Request                                            │
│    POST /api/v1/channels/:channel_id/reports/123/action     │
│    Authorization: Bearer <JWT_TOKEN>                         │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. AuthMiddleware                                            │
│    - Validates JWT token                                     │
│    - Extracts user_id and role from claims                   │
│    - Fetches full User object from database                  │
│    - Sets user in Gin context: c.Set("user", userObj)       │
└──────────────────────┬──────────────────────────────────────┘
                       │ If invalid → 401 Unauthorized
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. RequirePermission Middleware                              │
│    - Retrieves user from context                             │
│    - Calls user.Can(permission)                              │
│       → Checks if admin: return true                         │
│       → Checks accountTypePermissions map                    │
│    - If has permission AND is community moderator:           │
│       → Extract channel_id from request                      │
│       → Validate channel_id in user.ModerationChannels       │
│       → If not in list → 403 Forbidden                       │
└──────────────────────┬──────────────────────────────────────┘
                       │ If no permission → 403 Forbidden
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Handler Function                                          │
│    - Business logic execution                                │
│    - Additional permission checks if needed                  │
│    - Resource ownership verification                         │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Service Layer                                             │
│    - Optional: Additional permission validation              │
│    - Optional: Use PermissionCheckService for caching        │
│    - Execute business logic                                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Audit Logging                                             │
│    - Log moderation actions to audit log                     │
│    - Include user_id, action, entity_type, entity_id        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Response                                                  │
│    - Return success/error response                           │
│    - Include appropriate status code                         │
└─────────────────────────────────────────────────────────────┘
```

### Code Flow Details

#### Step 1: Authentication

```go
// JWT Claims structure
type Claims struct {
    UserID   uuid.UUID `json:"user_id"`
    Username string    `json:"username"`
    Role     string    `json:"role"`
    jwt.RegisteredClaims
}

// AuthMiddleware extracts and validates token
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        token := extractToken(c)
        
        // Validate token and extract claims
        claims, err := authService.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // Fetch full user object from database
        user, err := userRepo.GetByID(claims.UserID)
        if err != nil {
            c.JSON(401, gin.H{"error": "User not found"})
            c.Abort()
            return
        }
        
        // Set user in context for downstream middleware/handlers
        c.Set("user", user)
        c.Set("user_id", user.ID)
        c.Next()
    }
}
```

#### Step 2: Permission Check

```go
// RequirePermission middleware
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user from context
        user := c.MustGet("user").(*models.User)
        
        // Check if user has permission
        if !user.Can(permission) {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
        
        // For community moderators, validate channel scope
        if user.ModeratorScope == models.ModeratorScopeCommunity {
            channelID := getChannelIDFromRequest(c)
            if channelID != uuid.Nil {
                hasAccess := false
                for _, ch := range user.ModerationChannels {
                    if ch == channelID {
                        hasAccess = true
                        break
                    }
                }
                if !hasAccess {
                    c.JSON(403, gin.H{
                        "error": "Channel not in moderation scope",
                    })
                    c.Abort()
                    return
                }
            }
        }
        
        c.Next()
    }
}

// User.Can checks permission
func (u *User) Can(permission string) bool {
    // Admins have all permissions
    if u.IsAdmin() || u.AccountType == AccountTypeAdmin {
        return true
    }
    
    // Check account type permissions
    permissions := GetAccountTypePermissions(u.AccountType)
    for _, p := range permissions {
        if p == permission {
            return true
        }
    }
    return false
}
```

#### Step 3: Route Setup

```go
// Example route configuration
func SetupRoutes(router *gin.Engine, deps *Dependencies) {
    v1 := router.Group("/api/v1")
    
    // Public routes (no auth)
    v1.GET("/clips", clipHandler.ListClips)
    
    // Authenticated routes (any logged-in user)
    auth := v1.Group("")
    auth.Use(middleware.AuthMiddleware(authService))
    {
        auth.POST("/clips/:id/vote", clipHandler.Vote)
        auth.POST("/comments", commentHandler.Create)
    }
    
    // Moderation routes (requires moderate:content permission)
    moderation := v1.Group("/moderation")
    moderation.Use(middleware.AuthMiddleware(authService))
    moderation.Use(middleware.RequirePermission(models.PermissionModerateContent))
    {
        moderation.GET("/queue", moderationHandler.GetQueue)
        moderation.POST("/submissions/:id/approve", moderationHandler.Approve)
    }
    
    // Channel-scoped moderation routes
    channels := v1.Group("/channels/:channel_id")
    channels.Use(middleware.AuthMiddleware(authService))
    channels.Use(middleware.RequirePermission(models.PermissionCommunityModerate))
    {
        channels.POST("/reports/:id/action", reportHandler.ActionReport)
        channels.POST("/bans", banHandler.BanUser)
    }
    
    // Admin routes (requires admin account type)
    admin := v1.Group("/admin")
    admin.Use(middleware.AuthMiddleware(authService))
    admin.Use(middleware.RequirePermission(models.PermissionManageSystem))
    {
        admin.POST("/users/:id/role", userHandler.UpdateRole)
        admin.DELETE("/clips/:id", clipHandler.Delete)
    }
}
```

### Performance Optimization

The system uses caching to optimize permission checks:

```go
// PermissionCheckService with Redis caching
type PermissionCheckService struct {
    cache      *redis.Client
    userRepo   repositories.UserRepository
    channelRepo repositories.ChannelRepository
}

// CanModerate checks if user can moderate a channel (with caching)
func (s *PermissionCheckService) CanModerate(
    ctx context.Context,
    user *models.User,
    channelID uuid.UUID,
) (bool, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("perm:mod:%s:%s", user.ID, channelID)
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        return cached == "1", nil
    }
    
    // Not in cache, perform full check
    canModerate := false
    
    // Admins and site moderators can moderate any channel
    if user.IsAdmin() || user.ModeratorScope == models.ModeratorScopeSite {
        canModerate = true
    } else if user.ModeratorScope == models.ModeratorScopeCommunity {
        // Check if channel is in user's moderation list
        for _, ch := range user.ModerationChannels {
            if ch == channelID {
                canModerate = true
                break
            }
        }
    }
    
    // Cache result for 5 minutes
    s.cache.Set(ctx, cacheKey, map[bool]string{true: "1", false: "0"}[canModerate], 5*time.Minute)
    
    return canModerate, nil
}
```

### Error Responses

The system returns consistent error responses:

**401 Unauthorized** - Not authenticated:
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

**403 Forbidden** - Insufficient permissions:
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions",
    "details": {
      "required_permission": "moderate:content",
      "account_type": "member"
    }
  }
}
```

**403 Forbidden** - Channel scope violation:
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Access denied: channel not in moderation scope",
    "details": {
      "required_permission": "community:moderate",
      "account_type": "community_moderator",
      "channel_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

## Implementation Guide

### For Developers

#### Adding a New Protected Endpoint

1. **Define the required permission:**
   ```go
   // In internal/models/roles.go
   const PermissionNewFeature = "feature:new_feature"
   
   // Add to account type permissions
   AccountTypeModerator: {
       // ... existing permissions
       PermissionNewFeature,
   }
   ```

2. **Create route with middleware:**
   ```go
   // In route setup
   router.POST("/new-feature",
       middleware.AuthMiddleware(authService),
       middleware.RequirePermission(models.PermissionNewFeature),
       handler.NewFeature,
   )
   ```

3. **Implement handler:**
   ```go
   func (h *Handler) NewFeature(c *gin.Context) {
       user := c.MustGet("user").(*models.User)
       
       // Optional: Additional permission checks
       if !user.Can(models.PermissionNewFeature) {
           c.JSON(403, gin.H{"error": "Forbidden"})
           return
       }
       
       // Business logic...
   }
   ```

4. **Add tests:**
   ```go
   func TestNewFeature_RequiresPermission(t *testing.T) {
       // Test with member (should fail)
       // Test with moderator (should succeed)
       // Test with admin (should succeed)
   }
   ```

#### Adding Channel-Scoped Moderation

For features that should be channel-scoped for community moderators:

```go
// Ensure channel_id is in path or query params
router.POST("/channels/:channel_id/moderate",
    middleware.AuthMiddleware(authService),
    middleware.RequirePermission(models.PermissionCommunityModerate),
    handler.ModerateChannel,
)

// The middleware will automatically:
// 1. Extract channel_id from path
// 2. Validate it's in user.ModerationChannels (for community mods)
// 3. Allow access if user is site mod or admin
```

### For Operators

#### Promoting a User to Broadcaster

```sql
-- Give broadcaster permissions
UPDATE users 
SET account_type = 'broadcaster'
WHERE username = 'streamer_username';

-- Verify
SELECT username, account_type, role 
FROM users 
WHERE username = 'streamer_username';
```

#### Creating a Community Moderator

```sql
-- Set account type and scope
UPDATE users SET
  account_type = 'community_moderator',
  moderator_scope = 'community',
  moderation_channels = ARRAY['550e8400-e29b-41d4-a716-446655440000'::uuid]
WHERE username = 'community_mod_username';

-- Verify configuration
SELECT 
  username, 
  account_type, 
  moderator_scope,
  array_length(moderation_channels, 1) as channel_count
FROM users 
WHERE username = 'community_mod_username';
```

#### Creating a Site Moderator

```sql
-- Set account type and scope
UPDATE users SET
  account_type = 'moderator',
  moderator_scope = 'site',
  moderation_channels = ARRAY[]::uuid[]  -- Empty array
WHERE username = 'site_mod_username';

-- Also update legacy role for backward compatibility
UPDATE users SET role = 'moderator'
WHERE username = 'site_mod_username';
```

#### Promoting to Admin

```sql
-- Set both account type and role
UPDATE users SET
  account_type = 'admin',
  role = 'admin',
  moderator_scope = 'site',
  moderation_channels = ARRAY[]::uuid[]
WHERE username = 'admin_username';

-- Verify
SELECT username, account_type, role, moderator_scope
FROM users 
WHERE username = 'admin_username';
```

#### Adding Channels to Community Moderator

```sql
-- Add a new channel to existing moderator
UPDATE users SET
  moderation_channels = array_append(
    moderation_channels, 
    '660e8400-e29b-41d4-a716-446655440001'::uuid
  )
WHERE username = 'community_mod_username';

-- Remove a channel
UPDATE users SET
  moderation_channels = array_remove(
    moderation_channels, 
    '660e8400-e29b-41d4-a716-446655440001'::uuid
  )
WHERE username = 'community_mod_username';
```

#### Listing All Moderators

```sql
-- List all moderators with their scopes
SELECT 
  username,
  account_type,
  moderator_scope,
  array_length(moderation_channels, 1) as channel_count,
  created_at
FROM users
WHERE account_type IN ('moderator', 'community_moderator', 'admin')
ORDER BY account_type, created_at DESC;
```

## Troubleshooting

### Common Issues

#### Issue 1: Community Moderator Cannot Access Channel

**Symptoms:**
- User gets 403 Forbidden when accessing channel moderation
- Error message: "Access denied: channel not in moderation scope"

**Diagnosis:**
```sql
-- Check user's moderation configuration
SELECT 
  username,
  account_type,
  moderator_scope,
  moderation_channels
FROM users
WHERE username = 'affected_user';
```

**Solutions:**

1. **Missing channel in moderation list:**
   ```sql
   -- Add the channel to the user's moderation list
   UPDATE users SET
     moderation_channels = array_append(
       moderation_channels,
       'channel-uuid-here'::uuid
     )
   WHERE username = 'affected_user';
   ```

2. **Wrong scope:**
   ```sql
   -- Verify scope is 'community'
   UPDATE users SET moderator_scope = 'community'
   WHERE username = 'affected_user';
   ```

3. **Empty moderation channels:**
   ```sql
   -- Ensure at least one channel is assigned
   UPDATE users SET
     moderation_channels = ARRAY['channel-uuid-here'::uuid]
   WHERE username = 'affected_user';
   ```

---

#### Issue 2: Permission Check Always Fails

**Symptoms:**
- User gets 403 Forbidden for actions they should be able to perform
- JWT token appears valid

**Diagnosis:**
```go
// Enable debug logging in middleware
logger.Debug("Permission check", map[string]interface{}{
    "user_id": user.ID,
    "username": user.Username,
    "account_type": user.AccountType,
    "role": user.Role,
    "required_permission": permission,
    "has_permission": user.Can(permission),
})
```

**Solutions:**

1. **Check account type:**
   ```sql
   -- Verify account type is correct
   SELECT username, account_type, role FROM users WHERE username = 'affected_user';
   
   -- Update if needed
   UPDATE users SET account_type = 'moderator' WHERE username = 'affected_user';
   ```

2. **Check permission definition:**
   ```go
   // Verify permission is in accountTypePermissions map
   permissions := models.GetAccountTypePermissions(user.AccountType)
   log.Println("User permissions:", permissions)
   ```

3. **Clear JWT token cache:**
   ```bash
   # User needs to log out and log back in
   # Or invalidate token server-side if implemented
   ```

---

#### Issue 3: Site Moderator Has Channel Scope Restrictions

**Symptoms:**
- Site moderator gets 403 when accessing certain channels
- Should have sitewide access but is being scope-checked

**Diagnosis:**
```sql
-- Check moderator configuration
SELECT 
  username,
  account_type,
  moderator_scope,
  moderation_channels
FROM users
WHERE username = 'site_mod_username';
```

**Solutions:**

1. **Wrong scope:**
   ```sql
   -- Set scope to 'site'
   UPDATE users SET 
     moderator_scope = 'site',
     moderation_channels = ARRAY[]::uuid[]
   WHERE username = 'site_mod_username';
   ```

2. **Channels array not empty:**
   ```sql
   -- Clear the channels array
   UPDATE users SET moderation_channels = ARRAY[]::uuid[]
   WHERE username = 'site_mod_username';
   ```

---

#### Issue 4: User Promoted But Permissions Don't Apply

**Symptoms:**
- User role/account type updated in database
- User still gets 403 Forbidden
- Permissions not reflected in API calls

**Root Cause:**
JWT token cached in client contains old role/account_type

**Solutions:**

1. **Force re-authentication:**
   - User must log out and log back in
   - This generates new JWT with updated claims

2. **Check JWT claims:**
   ```javascript
   // Decode JWT token (use jwt.io)
   // Verify 'role' and check if claims are stale
   ```

3. **Implement token refresh:**
   ```go
   // Server-side: Invalidate old tokens (if implemented)
   // Or: Implement token version checking
   ```

---

#### Issue 5: Admin Cannot Access Admin Panel

**Symptoms:**
- User has admin role but cannot access `/admin` routes
- Gets 401 or 403 errors

**Diagnosis:**
```sql
-- Verify admin status
SELECT username, role, account_type FROM users WHERE username = 'admin_user';
```

```bash
# Check JWT token claims
curl -H "Authorization: Bearer YOUR_TOKEN" http://api/v1/admin/health
# Should return user details with role=admin
```

**Solutions:**

1. **Both role and account_type should be admin:**
   ```sql
   UPDATE users SET 
     role = 'admin',
     account_type = 'admin'
   WHERE username = 'admin_user';
   ```

2. **User needs to re-authenticate:**
   - Log out and log back in to get new JWT token with admin claims

3. **Check middleware configuration:**
   ```go
   // Ensure admin routes use correct middleware
   admin.Use(middleware.RequirePermission(models.PermissionManageSystem))
   // OR
   admin.Use(middleware.RequireRole("admin"))
   ```

---

### Debugging Tools

#### Check User Permissions

```go
// In handler or test
func debugUserPermissions(user *models.User) {
    log.Printf("User: %s", user.Username)
    log.Printf("Account Type: %s", user.AccountType)
    log.Printf("Role: %s", user.Role)
    log.Printf("Moderator Scope: %s", user.ModeratorScope)
    log.Printf("Moderation Channels: %v", user.ModerationChannels)
    log.Printf("Permissions: %v", user.GetPermissions())
    log.Printf("Is Admin: %v", user.IsAdmin())
    log.Printf("Is Moderator: %v", user.IsModerator())
}
```

#### Validate Moderator Configuration

```go
// Use IsValidModerator helper
if !user.IsValidModerator() {
    log.Println("Invalid moderator configuration!")
    log.Printf("Scope: %s", user.ModeratorScope)
    log.Printf("Channels: %v", user.ModerationChannels)
}
```

#### Test Permission Checks

```bash
# Create test script: scripts/test-permissions.sh
#!/bin/bash
TOKEN="your-jwt-token"

# Test member access (should fail)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/admin/users

# Test moderator access (should succeed)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/moderation/queue
```

---

### Logging and Monitoring

**Permission Denial Events:**
```go
// Logged automatically by middleware
logger.Warn("Permission denied", map[string]interface{}{
    "user_id": user.ID.String(),
    "username": user.Username,
    "account_type": user.GetAccountType(),
    "permission": permission,
    "path": c.Request.URL.Path,
})
```

**Channel Scope Violations:**
```go
logger.Warn("Channel scope violation", map[string]interface{}{
    "user_id": user.ID.String(),
    "username": user.Username,
    "permission": permission,
    "channel_id": channelID.String(),
    "path": c.Request.URL.Path,
})
```

**Monitor these logs for:**
- Unusual permission denial patterns
- Potential security issues
- Configuration problems
- Unauthorized access attempts

---

### Prevention

**Before promoting users:**
1. ✅ Verify user identity and trustworthiness
2. ✅ Set correct account_type and role
3. ✅ Configure moderator_scope appropriately
4. ✅ Set moderation_channels for community mods
5. ✅ Test access before granting to production
6. ✅ Document reason for promotion
7. ✅ Add to audit log

**Best practices:**
- Use staging environment to test role changes
- Document all moderator/admin promotions
- Review permissions regularly
- Monitor audit logs for unusual activity
- Implement principle of least privilege

## Related Documentation

- RBAC - Role-based access control implementation
- [Authorization Framework](./authorization-framework.md) - IDOR prevention and authorization
- Authentication - Authentication flow and JWT
- [Moderation API](./moderation-api.md) - Moderation endpoints and features
- Security - General security practices
- [Audit Logging](./AUDIT_LOG_SERVICE.md) - Audit log implementation

## Support

For questions or issues:
- Create an issue with the `permissions` or `security` label
- Contact the security team
- Review the threat model documentation at `docs/product/threat-model.md`
