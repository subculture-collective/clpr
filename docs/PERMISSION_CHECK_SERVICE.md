# PermissionCheckService

## Overview
The `PermissionCheckService` provides centralized permission checking and validation for community moderation with proper scope handling for both site and community moderators.

## Features
- Permission validation for ban/unban operations
- Channel-scoped moderation checks
- Multi-channel scope validation
- Redis caching for performance
- Clear, structured error messages

## Usage

### Initialization
```go
import (
    "git.subcult.tv/subculture-collective/clpr/internal/services"
    "git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// Initialize repositories and Redis client
communityRepo := repository.NewCommunityRepository(dbPool)
userRepo := repository.NewUserRepository(dbPool)

// Create the service
permService := services.NewPermissionCheckService(
    communityRepo,
    userRepo,
    redisClient,
)
```

### Check Ban Permission
```go
// Check if a user can ban another user in a channel
err := permService.CanBan(ctx, moderator, targetUserID, channelID)
if err != nil {
    if denialErr, ok := err.(*services.PermissionDenialReason); ok {
        // Handle permission denial with detailed error
        log.Printf("Ban denied: %s (code: %s)", denialErr.Message, denialErr.Code)
        log.Printf("Details: %v", denialErr.Details)
    }
    return err
}

// Permission granted, proceed with ban operation
```

### Check Unban Permission
```go
// Check if a user can unban by ban ID
err := permService.CanUnban(ctx, moderator, banID)
if err != nil {
    // Handle denial
    return err
}

// Permission granted, proceed with unban operation
```

### Check Moderation Permission
```go
// Check if a user can moderate a specific channel
err := permService.CanModerate(ctx, user, channelID)
if err != nil {
    // User cannot moderate this channel
    return err
}

// User has moderation access
```

### Validate Moderator Scope
```go
// Validate that a moderator has access to multiple channels
channelIDs := []uuid.UUID{channel1, channel2, channel3}
err := permService.ValidateModeratorScope(ctx, moderator, channelIDs)
if err != nil {
    // Moderator doesn't have access to all channels
    if denialErr, ok := err.(*services.PermissionDenialReason); ok {
        unauthorizedChannels := denialErr.Details["unauthorized_channels"]
        // Show which channels the moderator cannot access
    }
    return err
}

// Moderator has access to all requested channels
```

## Error Codes

The service returns structured `PermissionDenialReason` errors with the following codes:

- `CANNOT_BAN_OWNER` - Cannot ban the channel owner
- `ALREADY_BANNED` - User is already banned from this channel
- `NOT_BANNED` - User is not currently banned (for unban operations)
- `NOT_A_MEMBER` - Moderator is not a member of the channel
- `INSUFFICIENT_PERMISSIONS` - User lacks required permissions
- `NO_MODERATION_PRIVILEGES` - User has no moderation privileges
- `SCOPE_VIOLATION` - Moderator attempting to access unauthorized channels
- `INVALID_MODERATOR_CONFIG` - Invalid moderator configuration

## Permission Hierarchy

1. **Admin** (`AccountType=admin` or `Role=admin`)
   - Can moderate anywhere
   - Can ban/unban any user in any channel
   - No scope restrictions

2. **Site Moderator** (`AccountType=moderator`, `ModeratorScope=site`)
   - Can moderate across all channels
   - Can ban/unban users in any channel
   - No scope restrictions

3. **Community Moderator** (`AccountType=community_moderator`, `ModeratorScope=community`)
   - Can only moderate assigned channels (in `ModerationChannels`)
   - Must be a member with mod/admin role in the community
   - Limited to channels in their scope

4. **Regular User**
   - No moderation privileges
   - All moderation operations denied

## Caching

The service uses Redis caching to minimize database queries:

- **Permission checks**: 5 minutes TTL
- **User scope data**: 10 minutes TTL

### Cache Invalidation
```go
// Invalidate permission cache for a specific user/channel
err := permService.InvalidatePermissionCache(ctx, userID, channelID)

// Invalidate user's scope cache
err := permService.InvalidateUserScopeCache(ctx, userID)
```

Use cache invalidation when:
- User's role or account type changes
- User's moderation channels are updated
- User is added/removed from a community
- User's community role changes

## Performance

The service is designed for optimal performance:

- **No N+1 queries**: Uses efficient repository methods
- **Caching**: Reduces database load for repeated checks
- **Single query per check**: Most operations require 1-2 database queries
- **Batch validation**: `ValidateModeratorScope` checks multiple channels efficiently

## Testing

The service includes comprehensive tests:

- **Unit tests**: 14 tests covering all scenarios
- **Integration tests**: 7 scenarios with real database
- All tests verify permission logic, caching, and error handling

Run tests:
```bash
# Unit tests only
go test -v ./internal/services -run TestPermissionCheck

# Integration tests (requires test database)
INTEGRATION=1 go test -v ./tests/integration/rbac -run TestPermissionCheckService
```
