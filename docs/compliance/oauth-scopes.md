---
title: "OAuth Scopes"
summary: "**Last Updated:** 2025-12-29"
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Twitch OAuth Scopes Compliance Documentation

**Last Updated:** 2026-11-04  
**Status:** Active  
**Owner:** Backend Team, Security

## Purpose

This document defines all Twitch OAuth scopes requested by Clipper, justifies each scope's necessity, and ensures compliance with [Twitch's OAuth Documentation](https://dev.twitch.tv/docs/authentication/scopes/) and [Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/).

## Compliance Statement

Clipper's OAuth implementation complies with Twitch's guidelines:

✅ **Minimal scopes requested** - Only what's necessary for features  
✅ **User consent required** - Explicit authorization before accessing user data  
✅ **Scope purpose documented** - Clear explanation for each scope  
✅ **Secure token handling** - Encrypted storage, automatic refresh  
✅ **User revocation supported** - Users can disconnect anytime  
✅ **No scope creep** - Scopes not expanded without review

---

## Authentication Types

### 1. App Access Token (No User Scopes)

**Grant Type:** Client Credentials  
**Implementation:** `backend/pkg/twitch/auth.go`

**Used For:**
- Accessing public Twitch data
- Fetching clips via `/clips` endpoint
- Getting user/broadcaster public profiles
- Checking stream status
- Fetching game metadata

**Scopes:** NONE (app access token doesn't use scopes)

**Compliance:**
- ✅ No user data accessed without consent
- ✅ Only public API endpoints used
- ✅ No authentication required from users
- ✅ Appropriate for public data access

**Endpoints Accessed:**
- `GET /clips` - Public clip data
- `GET /users` - Public user profiles
- `GET /games` - Public game metadata
- `GET /streams` - Public live stream status
- `GET /channels` - Public channel information
- `GET /videos` - Public VODs
- `GET /games/top` - Public trending games

---

### 2. User Access Token (OAuth Authorization Code Flow)

**Grant Type:** Authorization Code  
**Implementation:** `backend/internal/handlers/twitch_oauth_handler.go`

**Purpose:** Let a broadcaster connect their channel so the Clpr bot can read its chat for the streamer clip room and so Clpr can fetch official clip download URLs

**OAuth Flow:**
1. User clicks "Connect Twitch" in Clipper
2. Redirected to Twitch authorization page
3. User sees requested scopes and grants permission
4. Twitch redirects back with authorization code
5. Backend exchanges code for access token + refresh token
6. Tokens stored encrypted in database

**Current Scopes Requested:** `channel:bot channel:manage:clips`

Sign-in is a separate flow (`backend/internal/services/auth_service.go`) that requests only `user:read:email` to identify the account; its access token is not stored.

---

## Requested User Scopes

### Scope: `channel:manage:clips`

**Purpose:** Allow Clpr to request Twitch's official temporary download URLs for clips captured from the broadcaster's channel.

**Feature:** Streamer Clip Rooms and authorized clip processing

**Behavior:**
- The streamer grants access explicitly through Twitch OAuth.
- Clpr limits this authorization to clips from the connected broadcaster's channel.
- The public status response exposes whether download authorization is present without exposing tokens or raw scopes.
- The streamer can revoke the authorization by disconnecting Twitch.

**Compliance:**
- ✅ Uses the scope Twitch documents for the Get Clips Download endpoint
- ✅ Streamer consent is required before the clip-room listener starts
- ✅ The dashboard explains why download access is requested
- ✅ Stored OAuth tokens are never returned to the frontend

### Scope: `channel:bot`

**Purpose:** Let the dedicated Clpr bot join the broadcaster's Twitch chat and detect supported clip links.

**Feature:** Streamer Clip Rooms

**Behavior:**
- The broadcaster grants access explicitly through Twitch OAuth.
- The bot only creates queue entries for supported Twitch or Clpr clip URLs found in public chat messages.
- Starting a listener is restricted to the Twitch channel connected to the authenticated Clpr account.
- Revoking the Twitch connection removes the stored broadcaster authorization.

**Bot identity:** The deployed listener uses a separate Twitch bot account configured with `TWITCH_BOT_USERNAME` and `TWITCH_BOT_OAUTH_TOKEN`; it does not authenticate to IRC as the broadcaster.

**Compliance:**
- ✅ Broadcaster consent is required before the listener starts
- ✅ Authorization is checked by exact OAuth scope name
- ✅ The bot cannot be started in an unrelated channel
- ✅ The broadcaster can stop the listener or disconnect Twitch

---

## Scopes NOT Requested (and Why)

### Scopes We Could Request But Don't

| Scope | Purpose (if we used it) | Why We Don't Request It |
|-------|-------------------------|-------------------------|
| `user:read:subscriptions` | Check if user is subscribed to channels | Not needed for current features |
| `user:read:follows` | Read user's followed channels | Not needed yet - may add for personalization later |
| `channel:read:subscriptions` | Read channel subscribers | Requires broadcaster permission, not applicable |
| `channel:manage:broadcast` | Manage stream settings | Not needed - we don't modify streams |
| `bits:read` | Read bits/cheers | Not needed - we don't process payments |
| `channel_editor` | Edit channel | Dangerous and unnecessary |
| `user:edit:follows` | Modify user's follows | Dangerous and unnecessary |
| `moderator:manage:banned_users`, `channel:manage:banned_users` | Ban and unban users on Twitch | Removed in November 2026 with ban sync and the Twitch ban actions; nothing uses them |
| `chat:read`, `chat:edit` | Read and send chat as the broadcaster | The Clpr bot reads chat with its own account under `channel:bot` |

**Principle:** Request ONLY what we actively use. Adding "nice to have" scopes violates user trust and Twitch guidelines.

---

## Scope Justification Process

**Before adding a new scope:**

1. **Identify Need**
   - What feature requires this scope?
   - Can we build the feature without it?
   - Is there an alternative approach?

2. **Document Purpose**
   - Write clear explanation of why scope is needed
   - Define exact data accessed
   - Explain user benefit

3. **Security Review**
   - Assess data sensitivity
   - Review storage requirements
   - Evaluate revocation process

4. **Legal Review**
   - Ensure compliance with Twitch ToS
   - Verify GDPR/CCPA compliance
   - Review privacy policy impact

5. **Approval Required**
   - Engineering lead approval
   - Legal approval
   - Update this document before implementation

---

## Token Security

### Token Storage

**Access Tokens:**
- Stored in `twitch_auth` table
- Encrypted with AES-256
- Encryption keys in a platform-managed secret store
- Never logged or sent to client

**Refresh Tokens:**
- Stored in `twitch_auth` table
- Encrypted with AES-256
- Used only for token refresh
- Never exposed via API

**In-Memory/Cache:**
- App access token cached in Redis
- Short-lived (expires_in - 5 min)
- Encrypted in transit (TLS)

**Code Implementation:**
```go
// backend/internal/models/models.go
type TwitchAuth struct {
    UserID         uuid.UUID `db:"user_id"`
    TwitchUserID   string    `db:"twitch_user_id"`
    TwitchUsername string    `db:"twitch_username"`
    AccessToken    string    `db:"access_token"`   // ENCRYPTED
    RefreshToken   string    `db:"refresh_token"`  // ENCRYPTED
    ExpiresAt      time.Time `db:"expires_at"`
    CreatedAt      time.Time `db:"created_at"`
    UpdatedAt      time.Time `db:"updated_at"`
}
```

---

### Token Refresh

**Automatic Refresh:**
```go
// backend/internal/handlers/twitch_oauth_handler.go
func (h *TwitchOAuthHandler) refreshTwitchToken(ctx context.Context, auth *models.TwitchAuth) error {
    // Exchange refresh token for new access token
    tokenResp, err := httpClient.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
        "client_id":     {clientID},
        "client_secret": {clientSecret},
        "refresh_token": {auth.RefreshToken},
        "grant_type":    {"refresh_token"},
    })
    // Update tokens in database
    h.twitchAuthRepo.RefreshToken(ctx, userID, newAccessToken, newRefreshToken, expiresAt)
}
```

**Triggers:**
- Token expires (checked before each use)
- 401 Unauthorized response from Twitch
- User initiates connection check
- Automatic before critical operations

**Compliance:**
- ✅ Smooth user experience (no re-auth needed)
- ✅ Security best practice
- ✅ Follows OAuth 2.0 spec
- ✅ No token expiry issues

---

### Token Revocation

**User-Initiated Revocation:**

**API Endpoint:** `DELETE /api/v1/twitch/auth`

```go
// backend/internal/handlers/twitch_oauth_handler.go
func (h *TwitchOAuthHandler) RevokeTwitchAuth(c *gin.Context) {
    userID := getUserIDFromContext(c)
    
    // Delete from database
    h.twitchAuthRepo.DeleteTwitchAuth(ctx, userID)
    
    // TODO: Call Twitch revocation endpoint (best practice)
    // POST https://id.twitch.tv/oauth2/revoke
}
```

**When User Revokes:**
1. Tokens deleted from database immediately
2. Chat connection closed (if active)
3. User can no longer send/read chat
4. User can re-authorize anytime

**Compliance:**
- ✅ User control over data access
- ✅ Immediate revocation
- ✅ GDPR right to withdraw consent
- ✅ CCPA right to opt-out

---

### Token Exposure Prevention

**Prohibited:**
- ❌ Never log access or refresh tokens
- ❌ Never send tokens to frontend/client
- ❌ Never include in error messages
- ❌ Never expose in API responses
- ❌ Never share with third parties
- ❌ Never store in plaintext
- ❌ Never include in URLs or query params

**Safeguards:**
```go
// Example: Redacting tokens in logs
logger.Info("OAuth token refreshed", map[string]interface{}{
    "user_id":    userID,
    "expires_at": expiresAt,
    // NO token field here!
})

// Example: API response
type TwitchAuthStatusResponse struct {
    Authenticated  bool    `json:"authenticated"`
    TwitchUsername *string `json:"twitch_username,omitempty"`
    // AccessToken NOT included
}
```

---

## User Communication

### OAuth Consent Screen

**What Users See:**

When authorizing Clipper to connect to Twitch:

```
Clipper wants to access your Twitch account

This application will be able to:
✓ Join your channel's chat as a bot (channel:bot)
✓ Manage clips on your channel (channel:manage:clips)

[Authorize] [Cancel]
```

**Twitch-Controlled:**
- Scope descriptions provided by Twitch
- User sees all requested scopes
- User can deny authorization
- User can revoke anytime on Twitch settings

---

### In-App Communication

**Before OAuth:**

```
Connect your Twitch channel to run a streamer clip room:
- The Clpr bot joins your chat and queues the clip links viewers post
- Clpr can fetch official download URLs for your channel's clips

Your Twitch password is never shared with Clipper.
You can disconnect anytime in Account Settings.

[Connect Twitch Account]
```

**After Connection:**

```
✓ Twitch connected as: [username]

You can now:
- Start a streamer clip room
- Let Clpr process your channel's clips

[Disconnect Twitch]
```

---

## OAuth Flow Security

### Authorization Code Flow

**Implementation:** `backend/internal/handlers/twitch_oauth_handler.go`

**Step 1: Initiate OAuth**
```
GET /api/v1/twitch/oauth/authorize
↓
Redirect to: https://id.twitch.tv/oauth2/authorize?
  client_id={CLIENT_ID}
  &redirect_uri={REDIRECT_URI}
  &response_type=code
  &scope=channel:bot+channel:manage:clips
```

**Step 2: User Authorizes**
- User logs into Twitch (if not already)
- User sees requested scopes
- User clicks "Authorize"

**Step 3: Callback**
```
GET /api/v1/twitch/oauth/callback?code={AUTH_CODE}
↓
Backend exchanges code for tokens:
POST https://id.twitch.tv/oauth2/token
  client_id={CLIENT_ID}
  &client_secret={CLIENT_SECRET}
  &code={AUTH_CODE}
  &grant_type=authorization_code
  &redirect_uri={REDIRECT_URI}
↓
Response:
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 3600,
  "scope": ["channel:bot", "channel:manage:clips"]
}
↓
Store encrypted tokens in database
↓
Redirect user to success page
```

**Security Features:**
- ✅ Authorization code used once, then discarded
- ✅ `client_secret` never exposed to client
- ✅ HTTPS enforced throughout
- ✅ Redirect URI validated
- ✅ State parameter (CSRF protection) - TODO: Add if not present
- ✅ PKCE (Proof Key for Code Exchange) - Consider adding for extra security

---

## Compliance Verification

### Pre-Launch Checklist

- [x] Only necessary scopes requested
- [x] Each scope justified and documented
- [x] User consent flow implemented
- [x] Tokens encrypted at rest
- [x] Tokens never logged or exposed
- [x] Revocation endpoint implemented
- [x] Automatic token refresh working
- [x] User communication clear
- [x] Privacy policy updated
- [x] GDPR/CCPA rights supported

### Quarterly Review Checklist

- [ ] No new scopes added without review
- [ ] Existing scopes still necessary
- [ ] No scope creep detected
- [ ] Token security intact
- [ ] Revocation process working
- [ ] User consent process clear
- [ ] Documentation up to date

---

## Scope Change Log

| Date | Change | Justification | Approved By |
|------|--------|---------------|-------------|
| 2025-12-29 | Initial scopes: `chat:read`, `chat:edit` | Enable Twitch chat integration | Engineering Lead, Legal |
| 2026-01-11 | Added scopes: `moderator:manage:banned_users`, `channel:manage:banned_users` | Enable ban management and sync from Clipper | Engineering Lead |
| 2026-08-11 | Replaced `chat:read`, `chat:edit` with `channel:bot`; added `channel:manage:clips` | The bot reads chat with its own account; official clip downloads for the clip room | Engineering Lead |
| 2026-11-04 | Removed `moderator:manage:banned_users`, `channel:manage:banned_users` | Ban sync and the Twitch ban actions were removed; no code path uses them | Engineering Lead |

**Future Scope Additions:**

If we need additional scopes in the future, they will be documented here BEFORE implementation, with full justification and approval.

---

## References

- [Twitch OAuth Scopes](https://dev.twitch.tv/docs/authentication/scopes/)
- [Twitch OAuth Documentation](https://dev.twitch.tv/docs/authentication/)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [Twitch Developer Agreement](https://legal.twitch.com/legal/developer-agreement/)

---

## Related Documents

- `twitch-api-usage.md` - API endpoints and authentication methods
- `data-retention.md` - How OAuth tokens are stored and retained
- `guardrails.md` - Developer guidelines for OAuth usage

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-29 | Initial OAuth scope documentation | Backend Team, Security |
| 2026-01-11 | Added ban management scopes documentation | Backend Team |

---

**Document Status:** ✅ COMPLETE  
**Next Review:** 2026-03-29 (Quarterly)
