---
title: "Twitch Embeds"
summary: "**Last Updated:** 2025-12-29"
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Twitch Embed Compliance Documentation

**Last Updated:** 2025-12-29  
**Status:** Active  
**Owner:** Frontend Team

## Purpose

This document verifies that all Twitch clip and stream playback on Clipper complies with [Twitch's Embedding Guidelines](https://dev.twitch.tv/docs/embed/video-and-clips/) and the [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/).

## Compliance Statement

Clipper's embed implementation is fully compliant with Twitch's embedding policies:

✅ **Uses official Twitch embed URLs and players only**  
✅ **No re-hosting, proxying, or serving of video files**  
✅ **HTTPS-only embed URLs**  
✅ **Proper parent domain configuration**  
✅ **Standard iframe embedding**  
✅ **Respects minimum sizing requirements**  
✅ **No unauthorized download functionality**  
✅ **Proper attribution to Twitch and creators**

---

## Embed Implementations

### 1. Clip Embeds (`TwitchEmbed.tsx`)

**File:** `frontend/src/components/clip/TwitchEmbed.tsx`  
**Purpose:** Embed and play Twitch clips

#### Official Embed URL

```typescript
const embedUrl = `https://clips.twitch.tv/embed?clip=${clipId}&parent=${parentDomain}&autoplay=${autoplay}&muted=${muted}`;
```

**Compliance Verification:**
- ✅ Uses official `clips.twitch.tv/embed` URL
- ✅ HTTPS protocol enforced
- ✅ Includes `parent` parameter with actual domain
- ✅ Respects autoplay and mute parameters
- ✅ No custom video player or re-hosting

#### Embed Characteristics

**Implementation:**
```tsx
<iframe
  src={embedUrl}
  className="absolute inset-0 w-full h-full"
  allowFullScreen
  title={title}
  onError={handleError}
  allow="autoplay; fullscreen"
/>
```

**Compliance:**
- ✅ Standard HTML5 iframe element
- ✅ Allows fullscreen as per Twitch guidelines
- ✅ Allows autoplay with user interaction
- ✅ Proper title attribute for accessibility
- ✅ Error handling for unavailable clips

#### Feed Loading Strategy

**Implementation:** Controlled thumbnail click-to-play with optional remembered muted autoplay

```tsx
// Initial state: every inactive clip is a lightweight thumbnail
<button onClick={onActivate}>
  <img src={thumbnailUrl} alt={title} />
  <div>Click to play</div>
</button>

// The feed owns activeClipId and mounts at most one official embed
{active && <iframe src={embedUrl} ... />}
```

**Benefits:**
1. **Performance** - Reduces initial page load
2. **User Control** - Explicit activation by default; muted autoplay is opt-in
3. **Bandwidth** - Saves data for users
4. **Resource Control** - Unmounting the prior iframe prevents off-screen playback

**Compliance:**
- ✅ Uses official Twitch thumbnail URLs
- ✅ No custom video preview generation
- ✅ Embed loaded only on user interaction
- ✅ Respects user's choice to view content
- ✅ Only one feed iframe is mounted at a time

The mobile feed uses the full available viewport width. On narrow portrait
screens it recommends landscape orientation while retaining the official
Twitch player, branding, controls, attribution, and direct Twitch fallback.

#### Parent Domain Configuration

```typescript
const parentDomain = typeof window !== 'undefined' 
  ? window.location.hostname 
  : 'localhost';
```

**Compliance:**
- ✅ Dynamically determines actual parent domain
- ✅ No domain spoofing or manipulation
- ✅ Works correctly in all deployment environments
- ✅ Required by Twitch for embed security

#### Error Handling

```tsx
if (hasError) {
  return (
    <div>
      <p>This clip is no longer available</p>
      <a href={`https://clips.twitch.tv/${clipId}`}>
        Try viewing on Twitch
      </a>
    </div>
  );
}
```

**Compliance:**
- ✅ Gracefully handles deleted clips
- ✅ Links back to Twitch for verification
- ✅ Does not cache or serve unavailable content
- ✅ Respects creator's right to delete clips

---

### 2. Live Stream Embeds (`TwitchPlayer.tsx`)

**File:** `frontend/src/components/stream/TwitchPlayer.tsx`  
**Purpose:** Embed live Twitch streams

#### Official Twitch Embed SDK

```typescript
// Load official Twitch Embed SDK
<script src="https://embed.twitch.tv/embed/v1.js" async />

// Initialize embed
const embed = new window.Twitch.Embed(elementId, {
  width: '100%',
  height: '100%',
  channel: channel,
  layout: showChat ? 'video-with-chat' : 'video',
  autoplay: true,
  muted: false,
  parent: [parentDomain],
});
```

**Compliance Verification:**
- ✅ Uses official Twitch Embed SDK (v1.js)
- ✅ HTTPS-only script loading
- ✅ Proper parent domain whitelisting
- ✅ Respects Twitch's layout options
- ✅ No custom streaming implementation

#### SDK Loading Best Practices

```typescript
// Script loading with reference counting
const existingScript = document.querySelector('script[src="https://embed.twitch.tv/embed/v1.js"]');
if (existingScript) {
  // Reuse existing script, don't create duplicate
  existingScript.addEventListener('load', handleLoad);
} else {
  // Create new script if needed
  const script = document.createElement('script');
  script.src = 'https://embed.twitch.tv/embed/v1.js';
  script.async = true;
  document.body.appendChild(script);
}
// Script persists across component unmounts (shared resource)
```

**Benefits:**
1. **Performance** - Single SDK load for entire app
2. **Reliability** - Avoids duplicate script conflicts
3. **Best Practices** - Follows Twitch's recommendations
4. **Resource Efficiency** - No redundant downloads

**Compliance:**
- ✅ Loads SDK from official Twitch CDN only
- ✅ No modification of Twitch's JavaScript
- ✅ No bundling or re-hosting of SDK
- ✅ Respects Twitch's CDN and versioning

#### Stream Status Integration

```typescript
// Fetch stream status before showing embed
const streamInfo = await fetchStreamStatus(channel);

if (!streamInfo?.is_live) {
  return <StreamOfflineScreen channel={channel} />;
}

// Only load embed if stream is actually live
if (isLive) {
  new window.Twitch.Embed(...);
}
```

**Compliance:**
- ✅ Verifies stream is live via official API
- ✅ Doesn't show embed for offline streams
- ✅ Graceful offline state with stream schedule
- ✅ No fake "live" indicators

#### Chat Integration

```typescript
layout: showChat ? 'video-with-chat' : 'video'
```

**Options:**
- `video` - Stream only (no chat)
- `video-with-chat` - Stream + Twitch chat

**Compliance:**
- ✅ Uses Twitch's built-in chat embed
- ✅ No custom chat implementation
- ✅ Respects chat moderation and rules
- ✅ Proper Twitch branding maintained

---

### 3. Twitch Chat Embed (`TwitchChatEmbed.tsx`)

**File:** `frontend/src/components/stream/TwitchChatEmbed.tsx`  
**Purpose:** Standalone Twitch chat embed

#### Official Chat Embed URL

```typescript
const embedUrl = `https://www.twitch.tv/embed/${channel}/chat?parent=${parentDomain}${darkMode ? '&darkpopout' : ''}`;
```

**Compliance:**
- ✅ Uses official `twitch.tv/embed/{channel}/chat` URL
- ✅ HTTPS protocol
- ✅ Proper parent domain parameter
- ✅ Respects theme preferences
- ✅ No custom chat client

---

## Prohibited Practices

**The following practices are STRICTLY FORBIDDEN:**

### ❌ Re-hosting Video Files

**NEVER:**
- Download Twitch clip `.mp4` files
- Store video files on our servers
- Serve video files from our CDN
- Cache video segments locally
- Create mirrors of Twitch video content

**WHY:** Violates Twitch ToS, copyright law, and DMCA

### ❌ Custom Video Players

**NEVER:**
- Build custom video player for Twitch content
- Use third-party players (Video.js, Plyr, etc.) for Twitch clips
- Parse HLS/m3u8 streams directly
- Decode or transcode Twitch video
- Extract audio from clips

**WHY:** Bypasses Twitch's embed requirements and tracking

### ❌ Clip Download Functionality

**NEVER:**
- Provide "Download Clip" buttons
- Link to third-party download services
- Offer browser extensions to download clips
- Archive clips for offline viewing
- Enable "Save Video As" functionality

**WHY:** Violates creator rights and Twitch ToS

**EXCEPTION:** Clips may be downloaded by the original creator through official Twitch tools, not through our platform.

### ❌ Embed Manipulation

**NEVER:**
- Strip Twitch branding from embeds
- Hide attribution or creator names
- Modify embed appearance beyond allowed parameters
- Inject custom controls into Twitch player
- Intercept or modify player events

**WHY:** Misrepresentation and ToS violation

### ❌ Domain Spoofing

**NEVER:**
- Fake the `parent` parameter
- Use proxy domains to bypass restrictions
- Rotate domains to evade detection
- Embed from unauthorized domains

**WHY:** Security violation and ToS breach

---

## Sizing Requirements

### Minimum Embed Dimensions

**Twitch Requirements:**
- **Minimum Width:** 340px
- **Minimum Height:** 190px
- **Recommended Aspect Ratio:** 16:9

**Our Implementation:**
```css
/* Responsive 16:9 aspect ratio */
.embed-container {
  position: relative;
  width: 100%;
  padding-top: 56.25%; /* 16:9 aspect ratio */
}

.embed-iframe {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}
```

**Compliance:**
- ✅ Exceeds minimum width on all viewports
- ✅ Maintains 16:9 aspect ratio
- ✅ Responsive design works on mobile and desktop
- ✅ No squished or stretched embeds

---

## Attribution and Branding

### Creator Attribution

**All clip displays include:**
- Broadcaster name (from Twitch API metadata)
- Creator name (if different from broadcaster)
- Link to broadcaster's Twitch channel
- Link to original clip on Twitch
- Game/category information

**Implementation:**
```tsx
<div className="clip-metadata">
  <a href={`https://twitch.tv/${broadcasterName}`}>
    {broadcasterDisplayName}
  </a>
  <span>playing {gameName}</span>
  <a href={twitchClipURL}>View on Twitch</a>
</div>
```

**Compliance:**
- ✅ Proper attribution to creators
- ✅ Links to Twitch for full context
- ✅ No claim of ownership over content
- ✅ Respects creator branding

### Twitch Branding

**Maintained:**
- Twitch logo visible in embeds (native to embed)
- "Watch on Twitch" links where appropriate
- Twitch color scheme in chat embed
- No modification of Twitch player UI

**Compliance:**
- ✅ Twitch branding not obscured
- ✅ No white-labeling of Twitch content
- ✅ Clear association with Twitch platform

---

## Content Availability and Removal

### Handling Deleted Clips

**When a clip is deleted on Twitch:**

1. **Embed Fails** - Twitch returns error
2. **Error Handler Triggered** - Our `onError` handler catches it
3. **Fallback UI Shown** - "Clip no longer available" message
4. **Link to Twitch** - User can verify on Twitch directly

**Implementation:**
```tsx
if (hasError) {
  return (
    <div>
      <p>This clip is no longer available</p>
      <a href={`https://clips.twitch.tv/${clipId}`}>
        Try viewing on Twitch
      </a>
    </div>
  );
}
```

**Compliance:**
- ✅ Respects creator's right to delete content
- ✅ No caching of deleted clips
- ✅ No attempts to preserve removed content
- ✅ Graceful degradation

### DMCA Compliance

**If content is DMCA'd:**
- Twitch removes the clip
- Embed fails automatically
- We do NOT maintain cached copies
- We do NOT have mechanisms to "restore" clips
- User is directed to Twitch for resolution

**Compliance:**
- ✅ No copyright circumvention
- ✅ Respects DMCA takedowns
- ✅ No archival of taken-down content

---

## Mobile and Responsive Considerations

### Mobile App Embed Strategy

**iOS/Android Apps:**
- Use Twitch's official mobile SDKs where available
- Fall back to web embeds in WebView if needed
- Respect platform-specific guidelines

**Compliance:**
- ✅ Official SDK usage prevents violations
- ✅ Platform policies respected
- ✅ No native video playback of Twitch content

### Responsive Web Design

**Implementation:**
```css
/* Desktop */
@media (min-width: 768px) {
  .clip-embed { width: 100%; max-width: 1280px; }
}

/* Tablet */
@media (min-width: 640px) and (max-width: 767px) {
  .clip-embed { width: 100%; }
}

/* Mobile */
@media (max-width: 639px) {
  .clip-embed { width: 100%; min-height: 190px; }
}
```

**Compliance:**
- ✅ Meets minimum size on all devices
- ✅ Maintains aspect ratio
- ✅ No layout breaking on small screens

---

## Code Comments for Compliance

### Added Compliance Comments

**In `TwitchEmbed.tsx`:**
```typescript
// COMPLIANCE: Uses official Twitch embed URL only.
// See: https://dev.twitch.tv/docs/embed/video-and-clips/
// Per Twitch Developer Agreement, we MUST NOT:
// - Re-host or proxy video files
// - Use unofficial embed methods
// - Strip Twitch branding or attribution
const embedUrl = `https://clips.twitch.tv/embed?clip=${clipId}&parent=${parentDomain}`;
```

**In `TwitchPlayer.tsx`:**
```typescript
// COMPLIANCE: Uses official Twitch Embed SDK (v1.js)
// See: https://dev.twitch.tv/docs/embed/video-and-clips/
// Must load from official Twitch CDN, never bundle or re-host
script.src = 'https://embed.twitch.tv/embed/v1.js';
```

**In `TwitchChatEmbed.tsx`:**
```typescript
// COMPLIANCE: Official Twitch chat embed
// See: https://dev.twitch.tv/docs/embed/chat/
const embedUrl = `https://www.twitch.tv/embed/${channel}/chat?parent=${parentDomain}`;
```

---

## Security Considerations

### Content Security Policy (CSP)

**Required CSP Directives:**
```http
frame-src: https://clips.twitch.tv https://player.twitch.tv https://www.twitch.tv
script-src: https://embed.twitch.tv
```

**Implementation:** Configured in production web server (nginx/Caddy)

**Compliance:**
- ✅ Allows only official Twitch embed domains
- ✅ Prevents unauthorized iframe sources
- ✅ Blocks malicious embed attempts

### HTTPS Enforcement

**All embeds:**
- Use HTTPS protocol exclusively
- No mixed content warnings
- Secure parent parameter transmission

**Compliance:**
- ✅ Meets Twitch's HTTPS requirement
- ✅ Protects user privacy
- ✅ Prevents man-in-the-middle attacks

---

## Testing and Verification

### Manual Testing Checklist

**Before each release:**
- [ ] Clip embeds load correctly
- [ ] Live stream embeds work
- [ ] Chat embeds display properly
- [ ] Error states show correctly (deleted clips)
- [ ] Parent domain parameter is correct
- [ ] No console errors related to embeds
- [ ] Fullscreen works on all embeds
- [ ] Mobile responsive behavior correct
- [ ] No video re-hosting detected
- [ ] Twitch branding visible and intact

### Automated Testing

**E2E Tests:**
```typescript
// frontend/e2e/tests/clips.spec.ts
test('should load Twitch clip embed', async ({ page }) => {
  await page.goto('/clips/test-clip-id');
  const iframe = page.locator('iframe[src*="clips.twitch.tv/embed"]');
  await expect(iframe).toBeVisible();
});

test('should use correct parent parameter', async ({ page }) => {
  await page.goto('/clips/test-clip-id');
  const iframe = page.locator('iframe');
  const src = await iframe.getAttribute('src');
  expect(src).toContain(`parent=${process.env.DOMAIN}`);
});
```

---

## References

- [Twitch Embed Documentation](https://dev.twitch.tv/docs/embed/video-and-clips/)
- [Twitch Chat Embed](https://dev.twitch.tv/docs/embed/chat/)
- [Twitch Developer Agreement](https://legal.twitch.com/legal/developer-agreement/)
- [Twitch Embedding Requirements](https://dev.twitch.tv/docs/embed/#embedding-twitch-on-your-site)

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-29 | Initial embed compliance audit | Frontend Team |
| 2025-12-29 | Added compliance code comments | Frontend Team |

---

**Document Status:** ✅ COMPLETE  
**Next Review:** 2026-03-29 (Quarterly)
