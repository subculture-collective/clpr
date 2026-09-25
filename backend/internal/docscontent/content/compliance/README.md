---
title: "README"
summary: "This directory contains comprehensive compliance documentation for Clipper's Twitch API integration, ensuring adherence to the [Twitch Developer Services Agreement](https://legal.twitch.com/legal/deve"
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Twitch Compliance Documentation

This directory contains comprehensive compliance documentation for Clipper's Twitch API integration, ensuring adherence to the [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/) and related policies.

## 📋 Quick Reference

| Document | Purpose | Audience |
|----------|---------|----------|
| **[twitch-api-usage.md](./twitch-api-usage.md)** | API endpoints, rate limits, caching | Backend Developers |
| **[twitch-embeds.md](./twitch-embeds.md)** | Embed compliance, video playback | Frontend Developers |
| **[data-retention.md](./data-retention.md)** | Data storage, retention policies | Backend Devs, Legal |
| **[oauth-scopes.md](./oauth-scopes.md)** | OAuth scopes, token security | Backend Devs, Security |
| **[guardrails.md](./guardrails.md)** | Developer guidelines, prohibited practices | All Developers |

## 🎯 Compliance Overview

### ✅ What We Do

- **Use official Twitch Helix API** exclusively for all data access
- **Embed clips and streams** using official Twitch embed URLs and SDK
- **Store only metadata** (titles, IDs, URLs) - never video files
- **Respect rate limits** via token bucket algorithm (800 req/min)
- **Encrypt OAuth tokens** at rest using AES-256
- **Request minimal OAuth scopes** (only `chat:read` and `chat:edit`)
- **Cache appropriately** to reduce API load (Redis with varying TTLs)
- **Handle errors gracefully** with circuit breaker and retry logic

### ❌ What We DO NOT Do

- ❌ **Scrape Twitch website** or use unofficial APIs
- ❌ **Re-host video files** or proxy Twitch content
- ❌ **Download clips** or offer download functionality
- ❌ **Bypass rate limits** or security measures
- ❌ **Request unnecessary OAuth scopes**
- ❌ **Sell or redistribute** Twitch data to third parties
- ❌ **Store raw video/audio** content
- ❌ **Use unauthorized embed methods**

## 📖 Document Summaries

### 1. Twitch API Usage (`twitch-api-usage.md`)

**What's Covered:**
- All 10 Twitch API endpoints we use (clips, users, games, streams, etc.)
- Authentication methods (app access token, user access token)
- Rate limiting implementation (token bucket, circuit breaker)
- Caching strategy (Redis, varying TTLs)
- Retry logic and error handling
- Prohibited API practices

**Key Takeaways:**
- Only official Helix API (`https://api.twitch.tv/helix/*`)
- 800 requests/minute limit strictly enforced
- Comprehensive caching reduces API load
- Circuit breaker prevents cascade failures

---

### 2. Twitch Embeds (`twitch-embeds.md`)

**What's Covered:**
- Clip embed implementation (`TwitchEmbed.tsx`)
- Live stream embed implementation (`TwitchPlayer.tsx`)
- Chat embed implementation (`TwitchChatEmbed.tsx`)
- Parent domain configuration
- Error handling for deleted clips
- Prohibited embed practices
- Sizing requirements (minimum 340x190px)

**Key Takeaways:**
- Official embed URLs only: `clips.twitch.tv/embed`, Twitch Embed SDK
- HTTPS required, proper `parent` parameter
- No video re-hosting, proxying, or downloading
- Graceful handling of deleted content

---

### 3. Data Retention (`data-retention.md`)

**What's Covered:**
- What Twitch data we store (clip metadata, OAuth tokens)
- What we DO NOT store (video files, audio files)
- Encryption practices (AES-256 for tokens)
- Retention policies (indefinite for clips, until revoked for tokens)
- User data rights (GDPR, CCPA)
- Third-party data sharing (minimal, no selling)

**Key Takeaways:**
- Only metadata stored, never raw media
- OAuth tokens encrypted at rest
- User can request deletion anytime
- No data reselling or sublicensing

---

### 4. OAuth Scopes (`oauth-scopes.md`)

**What's Covered:**
- Current scopes: `chat:read`, `chat:edit`
- Justification for each scope
- Scopes we DON'T request (and why)
- Token security (encryption, refresh, revocation)
- Scope change approval process
- User consent flow

**Key Takeaways:**
- Minimal scopes only (chat integration)
- User explicitly consents via OAuth
- Tokens encrypted, never logged
- Users can revoke anytime

---

### 5. Developer Guardrails (`guardrails.md`)

**What's Covered:**
- 8 absolute prohibitions (no scraping, no re-hosting, etc.)
- 6 required practices (official APIs, rate limits, etc.)
- Code review checklist
- Pre-commit checklist
- Incident response procedures
- Code examples (forbidden vs. correct)

**Key Takeaways:**
- Clear "do NOT do this" guidelines
- Concrete code examples
- Enforcement via code review
- Escalation path for questions

---

## 🚀 For Developers

### Before You Code

1. **Read the relevant docs** for your area:
   - Backend API work → `twitch-api-usage.md`, `oauth-scopes.md`
   - Frontend embeds → `twitch-embeds.md`
   - Data storage → `data-retention.md`
   - General rules → `guardrails.md` (READ THIS FIRST)

2. **Review code examples** in the docs showing correct vs. incorrect patterns

3. **Check the prohibited practices** - if unsure, ask first!

### During Development

1. **Add compliance comments** to code referencing these docs
2. **Use official APIs only** - no shortcuts!
3. **Test rate limiting** doesn't break
4. **Verify embedstill use official URLs**

### Before Committing

Run through the **Pre-Commit Checklist** in `guardrails.md`:
- [ ] No absolute prohibitions violated
- [ ] Required practices followed
- [ ] Compliance comments added
- [ ] No tokens in logs
- [ ] Official APIs/embeds only

### Code Review

Reviewers: Use the **Code Review Checklist** in `guardrails.md` to verify compliance.

---

## 🔍 Common Questions

### "Can I download clips for users?"
**NO.** This violates creator rights and Twitch ToS. Users must view clips via official Twitch embeds only.

### "Can I use Puppeteer to scrape Twitch?"
**NO.** Web scraping is strictly prohibited. Use official Twitch API only.

### "Can I add a new OAuth scope for [feature]?"
**MAYBE.** Follow the scope approval process in `oauth-scopes.md`:
1. Document why it's needed
2. Get engineering lead approval
3. Get legal approval
4. Update `oauth-scopes.md` before implementing

### "Can I cache video files to improve performance?"
**NO.** You can cache metadata, but NEVER video/audio files. That's re-hosting and violates ToS.

### "The API rate limit is slowing us down. Can I rotate API keys?"
**NO.** Rate limit bypass is strictly forbidden. Instead:
- Implement better caching
- Reduce unnecessary API calls
- Batch requests where possible
- Accept the 800/min limit as designed

---

## 📞 Who to Ask

| Question Type | Contact |
|---------------|---------|
| "Is this approach compliant?" | Engineering Lead (before coding) |
| "Can I add this OAuth scope?" | Engineering Lead + Legal |
| "Discovered violation in code" | Engineering Lead (immediately) |
| "Twitch sent us a warning" | Engineering Lead + Legal + Management (urgent) |

**Golden Rule:** When in doubt, DON'T do it. Ask first.

---

## 🔄 Maintenance

### Quarterly Reviews

Every 3 months (March, June, September, December):
- [ ] Review all 5 compliance docs
- [ ] Verify no new Twitch API usage
- [ ] Check for scope creep
- [ ] Audit codebase for violations
- [ ] Update docs if Twitch policies changed
- [ ] Update next review date

**Next Review:** 2026-03-29

### When to Update

Update these docs when:
- New Twitch API endpoint added
- New OAuth scope requested
- Twitch updates their ToS or API docs
- Violation discovered and remediated
- New compliance requirement identified

---

## 📚 External References

### Twitch Official Docs
- [Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
- [Helix API Documentation](https://dev.twitch.tv/docs/api/)
- [Embed Guidelines](https://dev.twitch.tv/docs/embed/video-and-clips/)
- [OAuth Scopes Reference](https://dev.twitch.tv/docs/authentication/scopes/)
- [Rate Limits](https://dev.twitch.tv/docs/api/guide/#rate-limits)

### Internal Docs
- [Privacy Policy](../legal/privacy-policy.md)
- [Terms of Service](../legal/terms-of-service.md)
- [DMCA Policy](../legal/dmca-policy.md)

---

## ✅ Compliance Verification

**Last Full Audit:** 2025-12-29  
**Status:** ✅ COMPLIANT  
**Next Audit:** 2026-03-29 (Quarterly)

### Audit Results

- ✅ All API calls use official Helix endpoints
- ✅ No scraping or unofficial API usage
- ✅ All embeds use official Twitch URLs/SDK
- ✅ No video re-hosting infrastructure
- ✅ Rate limiting properly implemented (800/min)
- ✅ OAuth tokens encrypted at rest
- ✅ Minimal scopes requested (chat only)
- ✅ User data rights supported
- ✅ Legal disclosures present
- ✅ Developer guardrails documented

**Violations Found:** None  
**Action Items:** None

---

## 📝 Document Status

| Document | Status | Last Updated | Next Review |
|----------|--------|--------------|-------------|
| twitch-api-usage.md | ✅ Complete | 2025-12-29 | 2026-03-29 |
| twitch-embeds.md | ✅ Complete | 2025-12-29 | 2026-03-29 |
| data-retention.md | ✅ Complete | 2025-12-29 | 2026-03-29 |
| oauth-scopes.md | ✅ Complete | 2025-12-29 | 2026-03-29 |
| guardrails.md | ✅ Complete | 2025-12-29 | 2026-03-29 |

---

**Questions?** Read the docs. Still have questions? Ask before coding.

**Remember:** Twitch compliance is non-negotiable. Violations can result in API key revocation, legal action, or platform shutdown. When in doubt, err on the side of caution.
