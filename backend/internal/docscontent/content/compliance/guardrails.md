---
title: "Guardrails"
summary: "**Last Updated:** 2025-12-29"
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Developer Guardrails for Twitch Compliance

**Last Updated:** 2025-12-29  
**Status:** Active  
**Owner:** Engineering Team, Legal

## Purpose

This document provides clear guidelines and constraints for developers working on Twitch integrations to prevent accidental violations of the [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/) and maintain long-term compliance.

## Core Principle

> **"When in doubt, don't do it. Ask first."**

If you're unsure whether a Twitch integration approach is compliant, STOP and consult:
1. This document
2. Engineering lead
3. Legal team (for significant changes)

---

## ⛔ Absolute Prohibitions

**These practices are STRICTLY FORBIDDEN and will result in immediate code rejection:**

### 1. ❌ NO WEB SCRAPING

**NEVER:**
- Scrape twitch.tv website HTML
- Parse Twitch web pages for data
- Use headless browsers (Puppeteer, Selenium) on Twitch
- Extract data from Twitch DOM
- Bypass robots.txt restrictions

**WHY:** Violates Twitch ToS, may cause IP ban, legal liability

**INSTEAD:**
- ✅ Use official Twitch Helix API
- ✅ Use documented API endpoints only
- ✅ Request API access if needed

**Code Review Red Flags:**
```javascript
// ❌ FORBIDDEN
const response = await fetch('https://twitch.tv/username');
const html = await response.text();
const clipUrl = html.match(/regex_pattern/);

// ❌ FORBIDDEN
const browser = await puppeteer.launch();
const page = await browser.newPage();
await page.goto('https://twitch.tv/...');
const data = await page.evaluate(() => { /* extract data */ });

// ✅ CORRECT
const clips = await twitchClient.GetClips(ctx, &twitch.ClipParams{
    BroadcasterID: broadcasterID,
});
```

---

### 2. ❌ NO UNOFFICIAL APIs

**NEVER:**
- Use undocumented Twitch API endpoints
- Access internal Twitch APIs (e.g., `gql.twitch.tv`)
- Reverse-engineer Twitch mobile app APIs
- Use third-party "Twitch API" wrappers that scrape

**WHY:** Violates ToS, may break without notice, security risk

**INSTEAD:**
- ✅ Use official Twitch Helix API only
- ✅ Use documented endpoints from dev.twitch.tv
- ✅ Submit feature requests to Twitch if endpoint missing

**Allowed Endpoints:**
- `https://api.twitch.tv/helix/*` - Official Helix API
- `https://id.twitch.tv/oauth2/*` - Official OAuth endpoints

**Code Review Red Flags:**
```javascript
// ❌ FORBIDDEN - Undocumented endpoint
fetch('https://gql.twitch.tv/gql', { 
    method: 'POST',
    body: JSON.stringify({ query: '...' })
});

// ❌ FORBIDDEN - Internal API
fetch('https://api.twitch.tv/v5/...'); // Deprecated v5 API

// ✅ CORRECT - Official Helix API
fetch('https://api.twitch.tv/helix/clips', {
    headers: {
        'Client-ID': clientID,
        'Authorization': `Bearer ${token}`
    }
});
```

---

### 3. ❌ NO VIDEO RE-HOSTING

**NEVER:**
- Download Twitch video files (.mp4, .webm, etc.)
- Store video files on our servers or CDN
- Parse HLS manifests (.m3u8) to extract video URLs
- Serve video from our infrastructure
- Create mirrors of Twitch videos
- Proxy video streams through our servers

**WHY:** Copyright infringement, DMCA violations, Twitch ToS breach

**INSTEAD:**
- ✅ Use official Twitch embed URLs only
- ✅ Let Twitch host and serve all video
- ✅ Link to Twitch for video playback

**Code Review Red Flags:**
```javascript
// ❌ FORBIDDEN - Downloading video
const videoUrl = clip.thumbnail_url.replace('-preview', '.mp4');
const video = await fetch(videoUrl);
await fs.writeFile('clip.mp4', video.body);

// ❌ FORBIDDEN - HLS parsing
const m3u8 = await fetch(clip.manifest_url);
const segments = parseM3U8(m3u8);
segments.forEach(seg => downloadSegment(seg));

// ❌ FORBIDDEN - Proxying video
app.get('/video/:clipId', async (req, res) => {
    const video = await fetchTwitchVideo(req.params.clipId);
    res.send(video); // Serving Twitch content from our server
});

// ✅ CORRECT - Official embed
<iframe src={`https://clips.twitch.tv/embed?clip=${clipId}&parent=${domain}`} />
```

**File System Red Flags:**
- Video files in `/uploads/`, `/media/`, `/videos/`
- CDN buckets containing `.mp4`, `.webm`, `.flv`
- Video processing pipelines (FFmpeg, handbrake, etc.)

---

### 4. ❌ NO CLIP DOWNLOAD FEATURES

**NEVER:**
- Provide "Download Clip" buttons
- Generate download links for videos
- Link to third-party clip download services
- Offer browser extensions to download clips
- Archive clips for offline viewing
- Enable "Save As" functionality

**WHY:** Violates creator rights, Twitch ToS, copyright law

**INSTEAD:**
- ✅ Direct users to Twitch for viewing
- ✅ Let creators download their own clips via Twitch

**Code Review Red Flags:**
```javascript
// ❌ FORBIDDEN
<button onClick={() => downloadClip(clipId)}>
    Download Clip
</button>

// ❌ FORBIDDEN
<a href={`https://third-party-downloader.com/clip/${clipId}`}>
    Download
</a>

// ✅ CORRECT
<a href={`https://clips.twitch.tv/${clipId}`} target="_blank">
    View on Twitch
</a>
```

---

### 5. ❌ NO RATE LIMIT BYPASS

**NEVER:**
- Rotate API keys to exceed rate limits
- Use multiple IP addresses to bypass limits
- Disable or circumvent rate limiter
- Make API calls faster than 800/minute
- Retry immediately after 429 responses

**WHY:** Violates Twitch ToS, may result in API ban

**INSTEAD:**
- ✅ Respect 800 requests/minute limit
- ✅ Use token bucket rate limiter
- ✅ Implement exponential backoff on 429
- ✅ Cache API responses appropriately

**Code Review Red Flags:**
```go
// ❌ FORBIDDEN - Disabling rate limiter
// if err := c.rateLimiter.Wait(ctx); err != nil {
//     return nil, err
// }

// ❌ FORBIDDEN - Immediate retry on 429
if resp.StatusCode == 429 {
    return c.doRequest(ctx, method, endpoint, params) // Immediate retry
}

// ❌ FORBIDDEN - Multiple client IDs
clientIDs := []string{clientID1, clientID2, clientID3}
currentClient := clientIDs[requestCount % len(clientIDs)]

// ✅ CORRECT - Respect rate limit
if err := c.rateLimiter.Wait(ctx); err != nil {
    return nil, fmt.Errorf("rate limit wait cancelled: %w", err)
}

// ✅ CORRECT - Exponential backoff
if resp.StatusCode == 429 {
    delay := baseDelay * time.Duration(1<<uint(attempt))
    time.Sleep(delay)
    continue
}
```

---

### 6. ❌ NO UNAUTHORIZED OAUTH SCOPES

**NEVER:**
- Request OAuth scopes not needed for current features
- Request "nice to have" scopes speculatively
- Add scopes without documentation and approval
- Use scopes for purposes other than documented

**WHY:** Violates user trust, Twitch ToS, privacy regulations

**INSTEAD:**
- ✅ Request minimal necessary scopes only
- ✅ Document each scope's purpose
- ✅ Get approval before adding new scopes
- ✅ Review scopes quarterly

**Code Review Red Flags:**
```javascript
// ❌ FORBIDDEN - Unnecessary scopes
const scopes = 'chat:read chat:edit user:read:email user:read:subscriptions channel:manage:broadcast';

// ❌ FORBIDDEN - Scope without justification
// TODO: Might need this later
const scopes = 'chat:read chat:edit user:read:follows';

// ✅ CORRECT - Minimal scopes with documentation
// COMPLIANCE: These scopes are required for chat integration
// chat:read - View chat messages (user consent required)
// chat:edit - Send chat messages (user consent required)
const scopes = 'chat:read chat:edit';
```

---

### 7. ❌ NO TOKEN SHARING OR EXPOSURE

**NEVER:**
- Send OAuth tokens to frontend/client
- Log access or refresh tokens
- Include tokens in error messages
- Store tokens in plaintext
- Share tokens between users
- Expose tokens via API responses

**WHY:** Security violation, account compromise risk

**INSTEAD:**
- ✅ Encrypt tokens at rest (AES-256)
- ✅ Keep tokens backend-only
- ✅ Redact tokens in logs
- ✅ Use secure token storage (vault, encrypted DB)

**Code Review Red Flags:**
```go
// ❌ FORBIDDEN - Logging token
logger.Info("User authenticated", map[string]interface{}{
    "access_token": token,
})

// ❌ FORBIDDEN - Sending token to client
c.JSON(200, gin.H{
    "access_token": auth.AccessToken,
})

// ❌ FORBIDDEN - Plaintext storage
INSERT INTO twitch_auth (user_id, access_token) VALUES (?, ?)

// ✅ CORRECT - Encrypted storage
encryptedToken := encrypt(auth.AccessToken, encryptionKey)
INSERT INTO twitch_auth (user_id, access_token) VALUES (?, ?)

// ✅ CORRECT - Token redaction
logger.Info("Token refreshed", map[string]interface{}{
    "user_id": userID,
    "expires_at": expiresAt,
    // No token here
})
```

---

### 8. ❌ NO DATA RESELLING

**NEVER:**
- Sell Twitch data to third parties
- License Twitch data to other companies
- Create derivative databases for commercial sale
- Provide Twitch data APIs to external clients
- Monetize Twitch data outside Service scope

**WHY:** Violates Twitch ToS, legal liability

**INSTEAD:**
- ✅ Use Twitch data only for Clipper features
- ✅ Display data in context of Service
- ✅ Do not commoditize Twitch data

---

## ✅ Required Practices

**These practices are MANDATORY for all Twitch integrations:**

### 1. ✅ USE OFFICIAL APIs ONLY

**ALWAYS:**
- Use Twitch Helix API (`https://api.twitch.tv/helix/*`)
- Use documented endpoints from dev.twitch.tv
- Include proper `Client-ID` and `Authorization` headers
- Follow API pagination correctly

**Example:**
```go
// ✅ CORRECT
resp, err := c.doRequest(ctx, "GET", "/clips", urlParams)

func (c *Client) doRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Client-Id", c.clientID)
    return c.httpClient.Do(req)
}
```

---

### 2. ✅ RESPECT RATE LIMITS

**ALWAYS:**
- Enforce 800 requests/minute limit
- Use token bucket algorithm
- Implement exponential backoff on 429
- Cache API responses appropriately

**Example:**
```go
// ✅ CORRECT - Rate limiting
if err := c.rateLimiter.Wait(ctx); err != nil {
    return nil, fmt.Errorf("rate limit wait cancelled: %w", err)
}

// ✅ CORRECT - Caching
if cachedUser, err := c.cache.CachedUser(ctx, userID); err == nil {
    return cachedUser, nil // Use cache
}
// Cache miss, fetch from API
user, err := c.GetUser(ctx, userID)
c.cache.CacheUser(ctx, user, time.Hour) // Cache for 1 hour
```

---

### 3. ✅ USE OFFICIAL EMBEDS ONLY

**ALWAYS:**
- Use `clips.twitch.tv/embed` for clips
- Use Twitch Embed SDK (`embed.twitch.tv/embed/v1.js`) for streams
- Include `parent` parameter with actual domain
- Use HTTPS only

**Example:**
```tsx
// ✅ CORRECT - Clip embed
<iframe 
    src={`https://clips.twitch.tv/embed?clip=${clipId}&parent=${window.location.hostname}`}
    allowFullScreen
/>

// ✅ CORRECT - Stream embed
new window.Twitch.Embed(elementId, {
    channel: channel,
    parent: [window.location.hostname],
});
```

---

### 4. ✅ ENCRYPT OAUTH TOKENS

**ALWAYS:**
- Encrypt tokens at rest (AES-256)
- Store encryption keys in secure vault
- Never log or expose tokens
- Implement secure token refresh

**Example:**
```go
// ✅ CORRECT - Encrypted storage
type TwitchAuth struct {
    UserID       uuid.UUID
    AccessToken  string // ENCRYPTED before storage
    RefreshToken string // ENCRYPTED before storage
}

// Before insert/update
encryptedAccess := encryptAES256(auth.AccessToken, key)
encryptedRefresh := encryptAES256(auth.RefreshToken, key)
```

---

### 5. ✅ HANDLE ERRORS GRACEFULLY

**ALWAYS:**
- Implement circuit breaker pattern
- Retry with exponential backoff
- Handle token expiry automatically
- Display user-friendly error messages

**Example:**
```go
// ✅ CORRECT - Retry logic
for attempt := 0; attempt < maxRetries; attempt++ {
    resp, err := c.httpClient.Do(req)
    if err != nil {
        if attempt < maxRetries-1 {
            delay := baseDelay * time.Duration(1<<uint(attempt))
            time.Sleep(delay)
            continue
        }
        return nil, err
    }
    
    switch resp.StatusCode {
    case 401:
        // Refresh token and retry
        c.authManager.RefreshToken(ctx)
        continue
    case 429:
        // Rate limited, backoff and retry
        time.Sleep(delay)
        continue
    case 200:
        return resp, nil
    }
}
```

---

### 6. ✅ DOCUMENT COMPLIANCE

**ALWAYS:**
- Add compliance comments to code
- Reference Twitch docs in comments
- Document OAuth scope purposes
- Maintain compliance documentation

**Example:**
```typescript
// COMPLIANCE: Uses official Twitch embed URL only.
// See: https://dev.twitch.tv/docs/embed/video-and-clips/
// Per Twitch Developer Agreement, we MUST NOT:
// - Re-host or proxy video files
// - Use unofficial embed methods
// - Strip Twitch branding or attribution
const embedUrl = `https://clips.twitch.tv/embed?clip=${clipId}&parent=${parentDomain}`;
```

```go
// COMPLIANCE: OAuth scopes required for chat integration
// chat:read - View Twitch chat messages (user consent required)
// chat:edit - Send messages in Twitch chat (user consent required)
// See: docs/compliance/oauth-scopes.md
const scopes = "chat:read chat:edit"
```

---

## 🚨 Code Review Checklist

**Reviewers must verify:**

### Twitch API Changes
- [ ] Uses official Helix API endpoints only
- [ ] No web scraping or HTML parsing
- [ ] Proper Client-ID and Authorization headers
- [ ] Rate limiting enforced
- [ ] Caching implemented where appropriate
- [ ] Error handling with retry logic
- [ ] No rate limit bypass attempts

### Embed Changes
- [ ] Uses official embed URLs (clips.twitch.tv, Twitch Embed SDK)
- [ ] Includes `parent` parameter correctly
- [ ] HTTPS only
- [ ] No video re-hosting or proxying
- [ ] No download functionality
- [ ] Twitch branding intact

### OAuth Changes
- [ ] Only minimal necessary scopes requested
- [ ] Each scope documented and justified
- [ ] Tokens encrypted at rest
- [ ] Tokens never logged or exposed
- [ ] Token refresh implemented
- [ ] User revocation supported

### Data Storage Changes
- [ ] No video or audio files stored
- [ ] Only metadata stored in database
- [ ] Public data only (or with user consent)
- [ ] Appropriate data retention policies
- [ ] GDPR/CCPA compliance maintained

---

## 📝 Pre-Commit Checklist

**Before committing Twitch-related code:**

- [ ] Reviewed absolute prohibitions (no violations)
- [ ] Followed required practices
- [ ] Added compliance comments to code
- [ ] Tested rate limiting works
- [ ] Verified no tokens in logs
- [ ] Checked embed URLs are official
- [ ] Confirmed no video re-hosting
- [ ] Updated documentation if needed

---

## 🔍 Automated Checks (Future)

**Lint rules to add (optional but recommended):**

```yaml
# .eslintrc.yml or similar
rules:
  no-restricted-imports:
    - error
    - paths:
      - name: 'puppeteer'
        message: 'Web scraping is prohibited. Use official Twitch API.'
      - name: 'cheerio'
        message: 'HTML parsing is prohibited. Use official Twitch API.'
  
  no-restricted-syntax:
    - error
    - selector: 'CallExpression[callee.name="fetch"][arguments.0.value=/twitch\\.tv(?!\\/embed)/]'
      message: 'Direct fetching from twitch.tv is prohibited. Use official API.'
```

**Git hooks:**

```bash
# pre-commit hook
#!/bin/bash
# Check for forbidden patterns
if git diff --cached | grep -E "(puppeteer|cheerio|twitch\.tv[^/]|\.mp4|\.webm)"; then
    echo "ERROR: Potential Twitch ToS violation detected!"
    echo "Review code against docs/compliance/guardrails.md"
    exit 1
fi
```

---

## 📚 Training and Onboarding

**For new developers:**

1. **Read Compliance Docs** (mandatory)
   - `twitch-api-usage.md`
   - `twitch-embeds.md`
   - `data-retention.md`
   - `oauth-scopes.md`
   - `guardrails.md` (this document)

2. **Review Example Code**
   - `backend/pkg/twitch/` - Correct API usage
   - `frontend/src/components/clip/TwitchEmbed.tsx` - Correct embed usage

3. **Understand Prohibitions**
   - No scraping, no re-hosting, no unofficial APIs
   - If unsure, ask first

4. **Attend Compliance Training** (quarterly)
   - Review recent violations (if any)
   - Discuss new Twitch policies
   - Update practices as needed

---

## 🚑 Incident Response

**If Twitch ToS violation is discovered:**

### Immediate Actions (within 1 hour)
1. **Stop the violating code** - Disable feature, rollback deploy
2. **Notify stakeholders** - Engineering lead, legal, management
3. **Document the violation** - What, when, how, impact
4. **Assess risk** - API ban risk, legal liability, user impact

### Short-Term Actions (within 24 hours)
5. **Fix the violation** - Remove code, update implementation
6. **Test the fix** - Ensure compliance
7. **Deploy the fix** - Emergency deployment if needed
8. **Notify Twitch** - If appropriate, proactively disclose and remediate

### Long-Term Actions (within 1 week)
9. **Root cause analysis** - Why did this happen?
10. **Update guardrails** - Add checks to prevent recurrence
11. **Team training** - Educate team on the violation
12. **Process improvement** - Update code review, testing

---

## 📞 Escalation Path

**When to escalate:**

| Situation | Contact | Urgency |
|-----------|---------|---------|
| Unsure if approach is compliant | Engineering Lead | Before coding |
| Potential ToS violation in PR | Engineering Lead | Before merging |
| Discovered violation in production | Engineering Lead + Legal | Immediately |
| Twitch API ban or warning | Engineering Lead + Legal + Management | Immediately |
| New feature needing scopes | Engineering Lead + Legal | Before implementation |

---

## 🔗 References

- [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
- [Twitch API Documentation](https://dev.twitch.tv/docs/api/)
- [Twitch Embed Guidelines](https://dev.twitch.tv/docs/embed/)
- [Twitch OAuth Scopes](https://dev.twitch.tv/docs/authentication/scopes/)
- Internal: `docs/compliance/twitch-api-usage.md`
- Internal: `docs/compliance/twitch-embeds.md`
- Internal: `docs/compliance/data-retention.md`
- Internal: `docs/compliance/oauth-scopes.md`

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-29 | Initial developer guardrails | Engineering Team, Legal |

---

**Document Status:** ✅ COMPLETE  
**Next Review:** 2026-03-29 (Quarterly)

---

**REMEMBER: When in doubt, don't do it. Ask first.**
