---
title: "Twitch Api Usage"
summary: "**Last Updated:** 2025-12-29"
tags: ["compliance","api"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Twitch API Usage Compliance Documentation

**Last Updated:** 2025-12-29  
**Status:** Active  
**Owner:** Engineering Team

## Purpose

This document provides a comprehensive audit of all Twitch API usage within the Clipper platform to ensure full compliance with the [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/) and [Twitch API Documentation](https://dev.twitch.tv/docs/api/).

## Compliance Statement

Clipper is fully compliant with Twitch's Developer Services Agreement:

✅ **Uses only official Twitch Helix API endpoints**  
✅ **No scraping or unofficial APIs**  
✅ **All API calls use valid Client ID and OAuth tokens**  
✅ **Rate limits are respected via token bucket algorithm**  
✅ **Caching is implemented to reduce API load**  
✅ **No attempts to bypass rate limits or security measures**

## Twitch API Endpoints Used

### 1. Helix API Base URL

**Endpoint:** `https://api.twitch.tv/helix`  
**Purpose:** Base URL for all Twitch API requests  
**Authentication:** Client ID + OAuth Bearer Token  
**Implementation:** `backend/pkg/twitch/client.go`

---

### 2. GET /clips

**Endpoint:** `GET https://api.twitch.tv/helix/clips`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetClips()`  
**Purpose:** Fetch Twitch clips based on various filters

**Parameters Supported:**
- `broadcaster_id` - Filter clips by broadcaster
- `game_id` - Filter clips by game/category
- `id` - Fetch specific clips by ID
- `started_at` - Filter clips created after this timestamp
- `ended_at` - Filter clips created before this timestamp
- `first` - Number of results per page (max 100)
- `after` - Pagination cursor
- `before` - Pagination cursor

**OAuth Scopes Required:** None (uses App Access Token)

**Rate Limit Considerations:**
- Respects 800 requests/minute limit via token bucket
- Implements exponential backoff on 429 responses
- Caches results where appropriate

**Use Cases:**
1. **Trending Clips Discovery** - Fetch popular clips from top games
2. **Broadcaster Sync** - Fetch clips for specific streamers
3. **User Submissions** - Validate and fetch user-submitted clip URLs
4. **Game-based Discovery** - Fetch clips from specific categories

**Compliance Notes:**
- Only fetches publicly available clip metadata
- Does not access private or restricted clips
- Respects Twitch's pagination and filtering rules

---

### 3. GET /users

**Endpoint:** `GET https://api.twitch.tv/helix/users`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetUsers()`, `GetUser()`  
**Purpose:** Fetch user/broadcaster information

**Parameters Supported:**
- `id` - User IDs (up to 100)
- `login` - Usernames (up to 100)

**OAuth Scopes Required:** None (uses App Access Token for public data)

**Caching:**
- Results cached in Redis for 1 hour
- Reduces API calls for frequently accessed users
- Cache key: `twitch:user:{user_id}`

**Use Cases:**
1. **Broadcaster Information** - Display broadcaster names, avatars
2. **User Profile Enrichment** - Fetch public Twitch profile data
3. **Stream Status Validation** - Verify user exists before operations

**Compliance Notes:**
- Only accesses public user information
- Implements caching to minimize API load
- Does not access private user data

---

### 4. GET /games

**Endpoint:** `GET https://api.twitch.tv/helix/games`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetGames()`  
**Purpose:** Fetch game/category metadata

**Parameters Supported:**
- `id` - Game IDs (up to 100)
- `name` - Game names (up to 100)

**OAuth Scopes Required:** None (uses App Access Token)

**Caching:**
- Results cached in Redis for 4 hours
- Longer cache due to infrequent game metadata changes
- Cache key: `twitch:game:{game_id}`

**Use Cases:**
1. **Category Display** - Show game names and artwork
2. **Trending Games** - Display popular game categories
3. **Clip Categorization** - Associate clips with games

**Compliance Notes:**
- Only accesses public game metadata
- Extended caching reduces API load
- Game data rarely changes, caching is appropriate

---

### 5. GET /games/top

**Endpoint:** `GET https://api.twitch.tv/helix/games/top`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetTopGames()`  
**Purpose:** Fetch currently trending games on Twitch

**Parameters Supported:**
- `first` - Number of results (max 100)
- `after` - Pagination cursor

**OAuth Scopes Required:** None (uses App Access Token)

**Use Cases:**
1. **Trending Clip Discovery** - Identify popular games for clip curation
2. **Category Browsing** - Display popular categories to users

**Compliance Notes:**
- Public data only
- Used for content discovery, not surveillance

---

### 6. GET /streams

**Endpoint:** `GET https://api.twitch.tv/helix/streams`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetStreams()`, `GetStreamStatusByUsername()`  
**Purpose:** Check if broadcasters are currently live

**Parameters Supported:**
- `user_id` - User IDs (up to 100)

**OAuth Scopes Required:** None (uses App Access Token)

**Caching:**
- Results cached for 30 seconds
- Short TTL due to rapidly changing live status
- Cache key: `twitch:streams:{sorted_user_ids}`

**Use Cases:**
1. **Live Status Display** - Show if broadcaster is currently streaming
2. **Stream Page** - Display live stream information
3. **Broadcaster Following** - Track live status of followed creators

**Compliance Notes:**
- Only accesses public live stream status
- Short cache duration respects real-time nature
- No access to private or subscriber-only streams

---

### 7. GET /channels

**Endpoint:** `GET https://api.twitch.tv/helix/channels`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetChannels()`  
**Purpose:** Fetch channel information (title, game, tags)

**Parameters Supported:**
- `broadcaster_id` - Broadcaster IDs (up to 100)

**OAuth Scopes Required:** None (uses App Access Token)

**Caching:**
- Results cached for 1 hour
- Cache key: `twitch:channels:{sorted_broadcaster_ids}`

**Use Cases:**
1. **Channel Information** - Display current stream title and category
2. **Broadcaster Profiles** - Show channel metadata

**Compliance Notes:**
- Public channel data only
- Appropriate caching reduces API load

---

### 8. GET /videos

**Endpoint:** `GET https://api.twitch.tv/helix/videos`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetVideos()`  
**Purpose:** Fetch VOD information

**Parameters Supported:**
- `user_id` - Filter by broadcaster
- `game_id` - Filter by game
- `id` - Fetch specific videos
- `first` - Number of results
- `after` / `before` - Pagination

**OAuth Scopes Required:** None (uses App Access Token for public videos)

**Use Cases:**
1. **VOD Discovery** - Display broadcaster VODs
2. **Content Archival** - Reference VOD sources for clips

**Compliance Notes:**
- Only accesses public VODs
- No access to subscriber-only or private videos

---

### 9. GET /channels/followers

**Endpoint:** `GET https://api.twitch.tv/helix/channels/followers`  
**Implementation:** `backend/pkg/twitch/endpoints.go` - `GetChannelFollowers()`  
**Purpose:** Fetch follower information for a broadcaster

**Parameters Supported:**
- `broadcaster_id` - Required broadcaster ID
- `first` - Number of results (max 100)
- `after` - Pagination cursor

**OAuth Scopes Required:** `moderator:read:followers` (requires User Access Token)

**Use Cases:**
1. **Follower Count** - Display follower statistics
2. **Community Metrics** - Analyze broadcaster popularity

**Compliance Notes:**
- Requires appropriate OAuth scope
- Used only for public follower counts
- Respects Twitch's follower privacy settings

---

### 10. OAuth Token Endpoint

**Endpoint:** `POST https://id.twitch.tv/oauth2/token`  
**Implementation:** `backend/pkg/twitch/auth.go` - `RefreshToken()`  
**Purpose:** Obtain and refresh OAuth access tokens

**Grant Types:**
1. **Client Credentials** - App Access Token (for public API access)
2. **Authorization Code** - User Access Token (for user-specific features)
3. **Refresh Token** - Token refresh

**OAuth Scopes Requested:** See `oauth-scopes.md` for full details

**Token Management:**
- Access tokens cached until 5 minutes before expiry
- Automatic token refresh on 401 Unauthorized responses
- Tokens encrypted at rest in database
- Tokens stored in Redis cache for performance

**Compliance Notes:**
- Follows OAuth 2.0 best practices
- Tokens are never logged or exposed
- Automatic rotation prevents token expiry issues
- Secure storage via encryption

---

## Authentication Methods

### 1. App Access Token (Client Credentials)

**Purpose:** Access public Twitch data without user context  
**Implementation:** `backend/pkg/twitch/auth.go`  
**Grant Type:** `client_credentials`

**Used For:**
- Fetching public clips
- Getting user/broadcaster information
- Accessing game metadata
- Checking stream status

**Token Lifecycle:**
1. Token requested on client initialization
2. Token cached in Redis until near expiry (expires_in - 300 seconds)
3. Token automatically refreshed on expiry or 401 error
4. Circuit breaker prevents cascading failures

**Compliance Notes:**
- Uses official OAuth 2.0 client credentials flow
- Token secured in memory and cache
- No user data accessed without user consent

---

### 2. User Access Token (Authorization Code)

**Purpose:** Access user-specific features (chat integration)  
**Implementation:** `backend/internal/handlers/twitch_oauth_handler.go`  
**Grant Type:** `authorization_code`

**OAuth Flow:**
1. User initiates OAuth via `/api/v1/twitch/oauth/authorize`
2. Redirects to Twitch authorization page
3. User approves requested scopes
4. Twitch redirects to callback with code
5. Backend exchanges code for access token + refresh token
6. Tokens stored encrypted in database

**Used For:**
- Twitch chat integration (`chat:read`, `chat:edit`)
- User-authorized features only

**Compliance Notes:**
- Uses standard OAuth 2.0 authorization code flow
- User explicitly consents to scopes
- Tokens stored encrypted
- Users can revoke access anytime
- Refresh tokens used to maintain access

---

## Rate Limiting Implementation

### Token Bucket Algorithm

**Implementation:** `backend/pkg/twitch/ratelimit.go`  
**Capacity:** 800 tokens (requests)  
**Refill Rate:** 800 tokens per minute (13.33 per second)

**How It Works:**
1. Each API request consumes 1 token
2. Tokens refill at constant rate
3. If no tokens available, request waits
4. Context cancellation respects timeouts

**Compliance:**
✅ Respects Twitch's 800 requests/minute limit  
✅ No burst abuse - smooth rate distribution  
✅ Implements exponential backoff on 429 responses  
✅ Circuit breaker prevents cascade failures

**Code Reference:**
```go
// backend/pkg/twitch/ratelimit.go
type RateLimiter struct {
    rate       float64  // Tokens per second
    capacity   int      // Maximum tokens
    tokens     float64  // Current tokens
    lastUpdate time.Time
    mu         sync.Mutex
}
```

---

## Retry and Error Handling

**Implementation:** `backend/pkg/twitch/client.go` - `doRequest()`

### Retryable Errors

**401 Unauthorized:**
- **Action:** Refresh access token automatically
- **Max Retries:** 3
- **Use Case:** Token expired or invalidated

**429 Too Many Requests:**
- **Action:** Exponential backoff (1s, 2s, 4s)
- **Max Retries:** 3
- **Use Case:** Rate limit exceeded (fallback if token bucket fails)

**503 Service Unavailable / 502 Bad Gateway / 504 Gateway Timeout:**
- **Action:** Exponential backoff
- **Max Retries:** 3
- **Use Case:** Twitch service temporarily down

### Non-Retryable Errors

**404 Not Found:**
- **Action:** Return error immediately
- **Use Case:** Clip/user/game doesn't exist

**400 Bad Request:**
- **Action:** Return error immediately
- **Use Case:** Invalid parameters

**Other Errors:**
- **Action:** Return error immediately
- **Use Case:** Unexpected API responses

### Circuit Breaker

**Implementation:** `backend/pkg/twitch/client.go` - `CircuitBreaker`

**States:**
- **Closed** - Normal operation
- **Open** - Too many failures, block requests
- **Half-Open** - Test if service recovered

**Thresholds:**
- **Failure Limit:** 5 consecutive failures
- **Timeout:** 30 seconds (time before retry)

**Purpose:**
- Prevent cascading failures
- Protect Twitch API from excessive retries
- Graceful degradation when Twitch is down

---

## Caching Strategy

### Cache Implementation

**Backend:** Redis  
**Client:** `backend/pkg/twitch/cache.go`

### Cache Policies

| Data Type | TTL | Reason |
|-----------|-----|--------|
| Access Token | Until expiry - 5min | Security + automatic refresh |
| User Data | 1 hour | Infrequent changes |
| Game Data | 4 hours | Very infrequent changes |
| Stream Status | 30 seconds | Rapidly changing |
| Channel Info | 1 hour | Moderate change frequency |

### Cache Keys

**Format:** `twitch:{type}:{identifier}`

**Examples:**
- `twitch:access_token` - App access token
- `twitch:user:12345` - User data for ID 12345
- `twitch:game:67890` - Game data for ID 67890
- `twitch:streams:123,456,789` - Stream status (sorted IDs)
- `twitch:stream_status:username` - Stream status by username

### Cache Benefits

1. **Reduced API Load** - Fewer requests to Twitch
2. **Faster Responses** - Serve from cache instead of API
3. **Rate Limit Headroom** - More capacity for real-time requests
4. **Cost Efficiency** - Lower infrastructure costs
5. **Twitch Compliance** - Good API citizenship

---

## Prohibited Practices

**The following practices are STRICTLY FORBIDDEN to maintain Twitch API compliance:**

❌ **No Web Scraping** - Never scrape twitch.tv website  
❌ **No Unofficial APIs** - Only use documented Helix API  
❌ **No Rate Limit Bypass** - Never attempt to circumvent rate limits  
❌ **No Token Sharing** - Never share OAuth tokens between users  
❌ **No Unauthorized Scopes** - Only request necessary OAuth scopes  
❌ **No Data Reselling** - Never sell or redistribute Twitch data  
❌ **No Media Re-hosting** - Never download/re-host clip video files  
❌ **No Impersonation** - Never impersonate Twitch or users  
❌ **No Automation Abuse** - No bots for artificial engagement  
❌ **No Security Bypass** - Never attempt to bypass Twitch security

**Violation of these prohibitions may result in:**
- Twitch API key revocation
- Legal action from Twitch
- Platform shutdown
- User trust damage

---

## Verification Checklist

**Regular compliance checks (quarterly):**

- [ ] All API calls use official Helix endpoints
- [ ] No scraping or unofficial API usage detected
- [ ] Rate limiter functioning correctly (800/min)
- [ ] Caching working as designed
- [ ] No rate limit bypass attempts
- [ ] OAuth tokens managed securely
- [ ] All scopes justified and documented
- [ ] Error handling following best practices
- [ ] Circuit breaker functioning
- [ ] No prohibited practices detected

---

## References

- [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
- [Twitch Helix API Documentation](https://dev.twitch.tv/docs/api/)
- [Twitch API Reference](https://dev.twitch.tv/docs/api/reference)
- [Twitch OAuth Documentation](https://dev.twitch.tv/docs/authentication/)
- [Twitch Rate Limits](https://dev.twitch.tv/docs/api/guide/#rate-limits)

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-29 | Initial compliance audit | Engineering Team |

---

**Document Status:** ✅ COMPLETE  
**Next Review:** 2026-03-29 (Quarterly)
