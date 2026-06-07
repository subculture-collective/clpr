# Twitch Ban/Unban Service

This document describes the Twitch moderation service implementation that wraps Twitch Helix POST/DELETE `/moderation/bans` endpoints with comprehensive error handling, rate limiting, and retry logic.

## Features

### Per-Channel Rate Limiting
- **ChannelRateLimiter**: Implements per-channel token bucket rate limiting
- Default: 100 moderation actions per channel per minute
- Prevents abuse and ensures compliance with Twitch's rate limits
- Thread-safe with concurrent access support

### Structured Error Codes
The service maps Twitch API errors to structured error codes for consistent error handling:

- `insufficient_scope`: Token lacks required scopes
- `target_not_found`: Target user was not found
- `already_banned`: User is already banned in the channel
- `not_banned`: User is not currently banned (for unban operations)
- `rate_limited`: Rate limit exceeded
- `server_error`: Twitch server error (5xx)
- `invalid_request`: Malformed request
- `unknown`: Unknown error occurred

### Retry Logic with Jittered Backoff
- **Jittered exponential backoff**: Prevents thundering herd problem
- **Smart retry strategy**:
  - 4xx errors (client errors): Do NOT retry
  - 5xx errors (server errors): Retry up to 3 times
  - 429 (rate limit): Retry with backoff
  - Network errors: Retry with backoff
- **Configurable parameters**:
  - Max retries: 3
  - Base delay: 1 second
  - Max delay: 10 seconds

### Request ID Logging
- Extracts Twitch request IDs from response headers
- Logs request IDs for all operations
- Helps with debugging and support inquiries

## Usage

### Ban a User

```go
import (
    "context"
    "git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

// Initialize client (done during app startup)
client, err := twitch.NewClient(cfg, redis)
if err != nil {
    // handle error
}

// Ban a user
ctx := context.Background()
broadcasterID := "12345"
moderatorID := "12345"  // Must be the broadcaster or a mod
targetUserID := "67890"
userAccessToken := "user_token_with_scopes"

reason := "Spam"
duration := 600 // 10 minutes, nil for permanent

request := &twitch.BanUserRequest{
    UserID:   targetUserID,
    Duration: &duration,
    Reason:   &reason,
}

response, err := client.BanUser(ctx, broadcasterID, moderatorID, userAccessToken, request)
if err != nil {
    // Check if it's a moderation error
    var modErr *twitch.ModerationError
    if errors.As(err, &modErr) {
        switch modErr.Code {
        case twitch.ModerationErrorCodeAlreadyBanned:
            // User is already banned
        case twitch.ModerationErrorCodeInsufficientScope:
            // Token lacks required scopes
        case twitch.ModerationErrorCodeRateLimited:
            // Rate limit exceeded
        // ... handle other cases
        }
    }
    return err
}

// Success
log.Printf("User banned: %+v", response)
```

### Unban a User

```go
err := client.UnbanUser(ctx, broadcasterID, moderatorID, targetUserID, userAccessToken)
if err != nil {
    var modErr *twitch.ModerationError
    if errors.As(err, &modErr) {
        switch modErr.Code {
        case twitch.ModerationErrorCodeNotBanned:
            // User is not banned
        case twitch.ModerationErrorCodeTargetNotFound:
            // User not found
        // ... handle other cases
        }
    }
    return err
}

// Success
log.Printf("User unbanned successfully")
```

## Integration with TwitchModerationService

The `TwitchModerationService` in `internal/services` provides a higher-level interface with scope validation:

```go
service := services.NewTwitchModerationService(twitchClient, twitchAuthRepo, userRepo)

// The service handles:
// - OAuth token validation
// - Scope checking
// - Permission verification
// - Wrapping the lower-level client calls

err := service.BanUserOnTwitch(ctx, moderatorUserID, broadcasterID, targetUserID, &reason, &duration)
```

## Required Scopes

Users must have one of these scopes:
- `moderator:manage:banned_users` (for moderators)
- `channel:manage:banned_users` (for broadcasters)

## Rate Limits

- **Global**: 800 requests per minute (enforced by the base client)
- **Per-channel**: 100 moderation actions per channel per minute (enforced by ChannelRateLimiter)

## Error Handling Best Practices

1. **Always check for ModerationError**: Use `errors.As()` to extract structured error details
2. **Log request IDs**: Include request IDs in logs for debugging
3. **Handle idempotency**: Check for `already_banned` and `not_banned` errors gracefully
4. **Retry appropriately**: Don't retry 4xx errors; they won't succeed
5. **Rate limit awareness**: Handle rate limit errors by backing off

## Testing

Comprehensive tests are available in:
- `pkg/twitch/endpoints_test.go`: Ban/Unban endpoint tests
- `pkg/twitch/errors_test.go`: Error parsing tests
- `pkg/twitch/ratelimit_test.go`: Rate limiter tests
- `internal/services/twitch_moderation_service_test.go`: Integration tests

Run tests:
```bash
go test ./pkg/twitch/...
go test -run "Twitch" ./internal/services/...
```

## Compliance

This implementation follows Twitch's Developer Services Agreement:
- Uses ONLY official Twitch Helix API endpoints
- Respects rate limits
- Implements proper error handling
- Does not scrape or use unofficial endpoints

## Future Enhancements

- Support for moderator verification (checking if user is a mod for the channel)
- Ban reason templates
- Batch ban/unban operations
- Metrics and monitoring integration
