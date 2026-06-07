# Twitch API Client

Production-ready Twitch API client library with rate limiting, error handling, and caching.

## Features

- **Authentication Management**: Automatic OAuth token refresh with Redis caching
- **Rate Limiting**: Token bucket algorithm (800 requests/minute globally, 100/minute per-channel for moderation)
- **Caching**: Redis-backed caching for users, games, channels, and streams
- **Error Handling**: Structured error types for better debugging
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Circuit Breaker**: Automatic API health monitoring
- **Comprehensive Logging**: Detailed request and error logging with request IDs
- **Moderation Support**: Ban/unban users with advanced error handling

## Documentation

- [Main API Client](./README.md) - This file
- [**Moderation Service**](./README_MODERATION.md) - Ban/unban functionality with rate limits

## Architecture

```
twitch/
├── client.go       # Main client and circuit breaker
├── auth.go         # Authentication management
├── ratelimit.go    # Rate limiting logic
├── cache.go        # Caching layer
├── endpoints.go    # API endpoint methods
├── models.go       # Data structures
└── errors.go       # Error types
```

## Usage

### Initialize Client

```go
import (
    "git.subcult.tv/subculture-collective/clpr/config"
    "git.subcult.tv/subculture-collective/clpr/pkg/twitch"
    "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// Create Redis client
redisClient, err := redis.NewClient(&cfg.Redis)
if err != nil {
    log.Fatal(err)
}

// Create Twitch client
twitchClient, err := twitch.NewClient(&cfg.Twitch, redisClient)
if err != nil {
    log.Fatal(err)
}
```

### Get User Information

```go
// Get single user (with caching)
user, err := twitchClient.GetUser(ctx, userID)

// Get multiple users
users, err := twitchClient.GetUsers(ctx, []string{userID1, userID2}, nil)

// Get users by login name
users, err := twitchClient.GetUsers(ctx, nil, []string{"username1", "username2"})

// Get from cache only
cachedUser, err := twitchClient.GetCachedUser(ctx, userID)
```

### Get Live Stream Status

```go
// Get live status for users (cached for 30 seconds)
streams, err := twitchClient.GetStreams(ctx, []string{userID1, userID2})

// Check if specific user is live
for _, stream := range streams.Data {
    if stream.UserID == targetUserID {
        log.Printf("%s is live with %d viewers", stream.UserName, stream.ViewerCount)
    }
}
```

### Get Clips

```go
params := &twitch.ClipParams{
    BroadcasterID: broadcasterID,
    First:         20,
    StartedAt:     time.Now().Add(-24 * time.Hour),
    EndedAt:       time.Now(),
}

clips, err := twitchClient.GetClips(ctx, params)
if err != nil {
    log.Printf("Error fetching clips: %v", err)
}

for _, clip := range clips.Data {
    log.Printf("Clip: %s - %d views", clip.Title, clip.ViewCount)
}
```

### Get Game Information

```go
// Get games by ID
games, err := twitchClient.GetGames(ctx, []string{gameID1, gameID2}, nil)

// Get games by name
games, err := twitchClient.GetGames(ctx, nil, []string{"Fortnite", "League of Legends"})

// Get from cache only (cached for 4 hours)
cachedGame, err := twitchClient.GetCachedGame(ctx, gameID)
```

### Get Channel Information

```go
// Get channel info (cached for 1 hour)
channels, err := twitchClient.GetChannels(ctx, []string{broadcasterID1, broadcasterID2})

for _, channel := range channels.Data {
    log.Printf("Channel: %s - %s", channel.BroadcasterName, channel.Title)
}
```

### Get Videos

```go
// Get videos for a user
videos, err := twitchClient.GetVideos(ctx, userID, "", nil, 20, "", "")

// Get specific videos by ID
videos, err := twitchClient.GetVideos(ctx, "", "", []string{videoID1, videoID2}, 0, "", "")
```

### Get Channel Followers

```go
// Get followers (requires user:read:follows scope)
followers, err := twitchClient.GetChannelFollowers(ctx, broadcasterID, 100, "")

log.Printf("Total followers: %d", followers.Total)
for _, follower := range followers.Data {
    log.Printf("Follower: %s followed at %s", follower.UserName, follower.FollowedAt)
}
```

## Error Handling

The client provides structured error types:

```go
clips, err := twitchClient.GetClips(ctx, params)
if err != nil {
    switch e := err.(type) {
    case *twitch.AuthError:
        log.Printf("Authentication failed: %v", e)
        // Handle auth error (e.g., invalid credentials)
    case *twitch.RateLimitError:
        log.Printf("Rate limited, retry after %d seconds", e.RetryAfter)
        // Wait and retry
    case *twitch.APIError:
        log.Printf("API error (status %d): %s", e.StatusCode, e.Message)
        // Handle specific status codes
    case *twitch.CircuitBreakerError:
        log.Printf("Service unavailable: %v", e)
        // Twitch API is down, use fallback or wait
    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

## Rate Limiting

The client automatically handles rate limiting:

- Local rate limiter: 800 requests/minute (token bucket)
- Automatic backoff on 429 responses from Twitch
- Request queuing when limit approaching
- Exponential backoff on repeated failures

## Caching Strategy

Different TTLs for different data types:

| Data Type | TTL | Reason |
|-----------|-----|--------|
| User data | 1 hour | User profiles change infrequently |
| Stream status | 30 seconds | Live indicator needs to be near real-time |
| Game data | 4 hours | Game catalog is relatively stable |
| Channel info | 1 hour | Channel info changes occasionally |

## Circuit Breaker

Protects against cascading failures when Twitch API is unavailable:

- Opens after 5 consecutive failures
- Remains open for 30 seconds
- Automatically tries to recover (half-open state)
- Closes on successful request

## Performance

- Target: < 100ms per request (excluding Twitch API latency)
- Caching reduces redundant API calls
- Rate limiting prevents hitting Twitch limits
- Concurrent safe for high-throughput applications

## Testing

Run tests with:

```bash
cd backend
go test ./pkg/twitch/... -v
go test ./pkg/twitch/... -cover
```

Current test coverage: 15.9% (33 tests)

## Configuration

Required environment variables:

```
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret
REDIS_HOST=localhost
REDIS_PORT=6379
```

## Dependencies

- `github.com/redis/go-redis/v9` - Redis client
- `git.subcult.tv/subculture-collective/clpr/config` - Configuration management
- `git.subcult.tv/subculture-collective/clpr/pkg/redis` - Redis wrapper

## Best Practices

1. **Reuse the client**: Create one client instance and reuse it
2. **Use context timeouts**: Always pass context with timeout
3. **Handle errors gracefully**: Check error types and retry appropriately
4. **Monitor circuit breaker**: Log circuit breaker state changes
5. **Cache when possible**: Use `GetCachedUser` and `GetCachedGame` for frequently accessed data

## Future Improvements

- [ ] Increase test coverage to >80%
- [ ] Add batch operations for bulk requests
- [ ] Add webhook support
- [ ] Add EventSub integration
- [ ] Performance benchmarking
- [ ] Distributed rate limiting for multi-instance deployments
