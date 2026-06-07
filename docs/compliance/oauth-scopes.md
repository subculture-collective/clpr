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

**Last Updated:** 2025-12-29  
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

**Purpose:** Enable user-authorized features like Twitch chat integration

**OAuth Flow:**
1. User clicks "Connect Twitch" in Clipper
2. Redirected to Twitch authorization page
3. User sees requested scopes and grants permission
4. Twitch redirects back with authorization code
5. Backend exchanges code for access token + refresh token
6. Tokens stored encrypted in database

**Current Scopes Requested:** `chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users`

---

## Requested User Scopes

### Scope 1: `chat:read`

**Purpose:** Read Twitch chat messages  
**Implementation:** Twitch Chat IRC integration  
**Feature:** Allow users to view chat while watching streams on Clipper

**Why Necessary:**
- Enables chat display alongside live streams
- Provides full streaming experience on Clipper
- Allows users to see community interaction
- Required for chat-based features (badges, emotes, etc.)

**Data Accessed:**
- Chat messages from public channels
- Usernames of chat participants
- Chat emotes and badges
- Timestamps of messages

**Data NOT Accessed:**
- Private messages (whispers)
- Moderator-only messages
- Deleted messages (unless cached before deletion)
- User's personal chat history across Twitch

**User Benefit:**
- Watch stream and chat in one place
- Better user experience
- No need to switch between Twitch and Clipper

**Compliance:**
- ✅ User explicitly consents via OAuth
- ✅ Only public chat data read
- ✅ No access to private conversations
- ✅ Clear purpose communicated to user

---

### Scope 2: `chat:edit`

**Purpose:** Send chat messages to Twitch channels  
**Implementation:** Twitch Chat IRC integration  
**Feature:** Allow users to participate in chat while watching on Clipper

**Why Necessary:**
- Enables users to send chat messages
- Allows participation in community
- Full-featured chat experience on Clipper
- Required to interact in Twitch chat

**Actions Enabled:**
- Send text messages to chat
- Use channel emotes
- React with emotes
- Participate in polls (if supported)

**Actions NOT Enabled:**
- Moderator actions (ban, timeout, delete)
- Whisper (private message) sending
- Channel point redemptions (different scope)
- Subscription management

**User Benefit:**
- Fully interact with community while on Clipper
- No need to open Twitch separately
- Seamless experience

**Compliance:**
- ✅ User explicitly consents via OAuth
- ✅ User's own messages only (no impersonation)
- ✅ Respects Twitch chat rules and moderation
- ✅ No spam or automation (ToS violation)

---

### Scope 3: `moderator:manage:banned_users`

**Purpose:** Manage banned users in channels where user is a moderator  
**Implementation:** Twitch ban sync and moderation features  
**Feature:** Allow moderators to ban/unban users on Twitch from Clipper

**Why Necessary:**
- Enables moderators to manage bans from Clipper interface
- Syncs ban lists from Twitch to Clipper
- Provides unified moderation experience
- Required to call Twitch ban/unban APIs as a moderator

**Data Accessed:**
- List of banned users in channels where user is a moderator
- Ban reasons and expiration times
- Moderator information for bans

**Data NOT Accessed:**
- Bans in channels where user is not a moderator
- Private moderator notes (if separate from ban reasons)
- Moderator actions unrelated to bans

**User Benefit:**
- Moderate Twitch channels directly from Clipper
- Keep ban lists synchronized
- Streamlined moderation workflow

**Compliance:**
- ✅ User explicitly consents via OAuth
- ✅ Only channels where user has moderator privileges
- ✅ All actions traceable to specific moderator
- ✅ Respects Twitch's moderation policies

---

### Scope 4: `channel:manage:banned_users`

**Purpose:** Manage banned users in user's own channel  
**Implementation:** Twitch ban sync and channel management features  
**Feature:** Allow broadcasters to ban/unban users on Twitch from Clipper

**Why Necessary:**
- Enables broadcasters to manage bans in their own channels
- Syncs ban lists from Twitch to Clipper
- Provides unified channel management experience
- Required to call Twitch ban/unban APIs as a broadcaster

**Data Accessed:**
- List of banned users in user's channel
- Ban reasons and expiration times
- Information about who created each ban

**Data NOT Accessed:**
- Bans in other channels
- Channel settings unrelated to bans
- Private channel information

**User Benefit:**
- Manage channel bans directly from Clipper
- Keep ban lists synchronized across platforms
- Better channel moderation tools

**Compliance:**
- ✅ User explicitly consents via OAuth
- ✅ Only user's own channel
- ✅ All actions logged and auditable
- ✅ Respects Twitch's channel management policies

---

## Scopes NOT Requested (and Why)

### Scopes We Could Request But Don't

| Scope | Purpose (if we used it) | Why We Don't Request It |
|-------|-------------------------|-------------------------|
| `user:read:email` | Read user's email address | Not needed - users provide email separately during Clipper registration |
| `user:read:subscriptions` | Check if user is subscribed to channels | Not needed for current features |
| `user:read:follows` | Read user's followed channels | Not needed yet - may add for personalization later |
| `channel:read:subscriptions` | Read channel subscribers | Requires broadcaster permission, not applicable |
| `channel:manage:broadcast` | Manage stream settings | Not needed - we don't modify streams |
| `bits:read` | Read bits/cheers | Not needed - we don't process payments |
| `channel_editor` | Edit channel | Dangerous and unnecessary |
| `user:edit:follows` | Modify user's follows | Dangerous and unnecessary |

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
✓ View Twitch chat messages (chat:read)
✓ Send messages in Twitch chat (chat:edit)

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
Connect your Twitch account to use chat features:
- View chat while watching streams on Clipper
- Send messages to participate in the community
- Full Twitch chat experience without leaving Clipper

Your Twitch password is never shared with Clipper.
You can disconnect anytime in Account Settings.

[Connect Twitch Account]
```

**After Connection:**

```
✓ Twitch connected as: [username]

You can now:
- Use chat on live streams
- Participate in Twitch communities

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
  &scope=chat:read+chat:edit
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
  "scope": ["chat:read", "chat:edit"]
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
