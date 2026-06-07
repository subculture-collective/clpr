---
title: "Data Retention"
summary: "**Last Updated:** 2025-12-29"
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Twitch Data Storage & Retention Compliance

**Last Updated:** 2025-12-29  
**Status:** Active  
**Owner:** Backend Team, Legal

## Purpose

This document defines what Twitch-derived data Clipper stores, how long it's retained, and compliance with Twitch's [Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/) regarding data usage and storage.

## Compliance Statement

Clipper's data storage practices comply with Twitch's Developer Agreement:

✅ **Stores only metadata, not raw video or audio**  
✅ **No redistribution or sublicensing of Twitch data**  
✅ **No sale of Twitch-derived data to third parties**  
✅ **Data used solely for Service functionality**  
✅ **Proper data retention and deletion policies**  
✅ **User data rights respected (GDPR, CCPA)**

---

## What We Store

### 1. Clip Metadata (from Twitch API)

**Database Table:** `clips`  
**Source:** Twitch Helix API `/clips` endpoint  
**Schema:** `backend/migrations/000001_initial_schema.up.sql`

**Fields Stored:**

| Field | Type | Source | Purpose |
|-------|------|--------|---------|
| `id` | UUID | Generated | Internal primary key |
| `twitch_clip_id` | VARCHAR(100) | Twitch API | Twitch's unique clip ID |
| `twitch_clip_url` | TEXT | Twitch API | Original clip URL on Twitch |
| `embed_url` | TEXT | Twitch API | Official Twitch embed URL |
| `title` | VARCHAR(255) | Twitch API | Clip title set by creator |
| `creator_name` | VARCHAR(100) | Twitch API | Name of user who created clip |
| `creator_id` | VARCHAR(50) | Twitch API | Twitch user ID of creator |
| `broadcaster_name` | VARCHAR(100) | Twitch API | Broadcaster's username |
| `broadcaster_id` | VARCHAR(50) | Twitch API | Twitch user ID of broadcaster |
| `game_id` | VARCHAR(50) | Twitch API | Twitch game/category ID |
| `game_name` | VARCHAR(100) | Twitch API | Name of game/category |
| `language` | VARCHAR(10) | Twitch API | Clip language code |
| `thumbnail_url` | TEXT | Twitch API | Twitch CDN thumbnail URL |
| `duration` | FLOAT | Twitch API | Clip length in seconds |
| `view_count` | INT | Twitch API | View count (synced periodically) |
| `created_at` | TIMESTAMP | Twitch API | When clip was created on Twitch |
| `imported_at` | TIMESTAMP | Generated | When we fetched from Twitch |

**Compliance:**
- ✅ **ONLY METADATA** - No video files, no audio files
- ✅ **PUBLIC DATA ONLY** - All data is publicly accessible via Twitch
- ✅ **PROPER ATTRIBUTION** - Stores creator and broadcaster info
- ✅ **EXTERNAL REFERENCES** - URLs point to Twitch, we don't host media

**What We Do NOT Store:**
- ❌ Raw video files (`.mp4`, `.webm`, etc.)
- ❌ Video segments or HLS playlists
- ❌ Audio tracks
- ❌ Private or subscriber-only clips
- ❌ Deleted clips (auto-removed when embed fails)

---

### 2. User/Broadcaster Metadata (from Twitch API)

**Caching:** Redis  
**TTL:** 1 hour  
**Source:** Twitch Helix API `/users` endpoint

**Cached Fields:**

| Field | Purpose |
|-------|---------|
| `id` | Twitch user ID |
| `login` | Username (lowercase) |
| `display_name` | Display name (with caps) |
| `type` | Account type (empty, affiliate, partner) |
| `broadcaster_type` | Broadcaster type |
| `description` | User bio/description |
| `profile_image_url` | Avatar URL (Twitch CDN) |
| `offline_image_url` | Offline banner URL (Twitch CDN) |
| `created_at` | Account creation date |

**Compliance:**
- ✅ **CACHE ONLY** - Not persisted in database
- ✅ **PUBLIC DATA** - Available via public API
- ✅ **REASONABLE TTL** - 1 hour respects data freshness
- ✅ **EXTERNAL URLs** - Images hosted on Twitch CDN

---

### 3. Game/Category Metadata (from Twitch API)

**Caching:** Redis  
**TTL:** 4 hours  
**Source:** Twitch Helix API `/games` endpoint

**Cached Fields:**

| Field | Purpose |
|-------|---------|
| `id` | Twitch game ID |
| `name` | Game/category name |
| `box_art_url` | Box art image URL (Twitch CDN) |

**Compliance:**
- ✅ **CACHE ONLY** - Not persisted in database
- ✅ **PUBLIC DATA** - Game metadata is public
- ✅ **LONG TTL** - 4 hours appropriate (games rarely change)
- ✅ **EXTERNAL URLs** - Images hosted on Twitch CDN

---

### 4. OAuth Access Tokens (User Authorization)

**Database Table:** `twitch_auth`  
**Source:** User OAuth consent flow  
**Migration:** `backend/migrations/000068_add_twitch_auth.up.sql`

**Fields Stored:**

| Field | Type | Encrypted | Purpose |
|-------|------|-----------|---------|
| `user_id` | UUID | No | Internal user ID (FK) |
| `twitch_user_id` | VARCHAR(50) | No | User's Twitch ID |
| `twitch_username` | VARCHAR(100) | No | User's Twitch username |
| `access_token` | TEXT | **YES** | OAuth access token |
| `refresh_token` | TEXT | **YES** | OAuth refresh token |
| `expires_at` | TIMESTAMP | No | Token expiration time |
| `created_at` | TIMESTAMP | No | When connection established |
| `updated_at` | TIMESTAMP | No | Last token refresh |

**Scopes Granted:**
- `chat:read` - Read Twitch chat messages
- `chat:edit` - Send Twitch chat messages

**Compliance:**
- ✅ **USER CONSENT** - User explicitly authorizes via OAuth
- ✅ **ENCRYPTED** - Tokens encrypted at rest (AES-256)
- ✅ **MINIMAL SCOPES** - Only necessary scopes requested
- ✅ **USER REVOCABLE** - User can disconnect Twitch anytime
- ✅ **SECURE HANDLING** - Never logged or exposed
- ✅ **AUTO REFRESH** - Tokens refreshed before expiry

**User Rights:**
- Users can revoke access via `/api/v1/twitch/auth` DELETE
- Revocation deletes tokens immediately
- Users can disconnect in account settings
- Deletion removes all OAuth data

---

### 5. Stream Status (Live/Offline)

**Caching:** Redis  
**TTL:** 30 seconds  
**Source:** Twitch Helix API `/streams` endpoint

**Cached Fields:**

| Field | Purpose |
|-------|---------|
| `id` | Stream ID |
| `user_id` | Broadcaster's user ID |
| `user_name` | Broadcaster's username |
| `game_id` | Current game being streamed |
| `game_name` | Current game name |
| `type` | Stream type (always "live") |
| `title` | Current stream title |
| `viewer_count` | Current viewer count |
| `started_at` | When stream went live |
| `language` | Stream language |
| `thumbnail_url` | Stream thumbnail URL |

**Compliance:**
- ✅ **CACHE ONLY** - Very short TTL (30 sec)
- ✅ **PUBLIC DATA** - Stream status is public
- ✅ **REAL-TIME NATURE** - Short cache respects live status
- ✅ **NO ARCHIVAL** - Not stored long-term

---

## What We Do NOT Store

### ❌ Raw Video/Audio Files

**NEVER STORED:**
- Video files (`.mp4`, `.webm`, `.flv`, etc.)
- Audio files (`.mp3`, `.aac`, `.opus`, etc.)
- Video segments or chunks
- HLS manifests (`.m3u8`) or playlists
- DASH manifests
- Subtitles or captions (unless via Twitch API metadata)

**WHY:**
- Violates Twitch ToS
- Copyright infringement
- DMCA violations
- Massive legal liability

**VERIFICATION:**
- No CDN or object storage for Twitch media
- No video processing pipelines
- No transcoding or encoding infrastructure
- Database schema contains NO BLOB fields for media

---

### ❌ Private or Restricted Content

**NEVER ACCESSED OR STORED:**
- Subscriber-only VODs
- Private clips
- Deleted clips (removed when detected)
- Banned channel content
- Suspended account data
- DMCA'd content

**ENFORCEMENT:**
- Only use App Access Token (public data only)
- User Access Token limited to authorized scopes
- No attempts to access restricted endpoints
- Deleted clips purged when embed fails

---

### ❌ User Private Information

**NOT STORED FROM TWITCH:**
- Email addresses (unless user provides separately)
- Phone numbers
- Payment information
- Private messages
- Follower lists (except public counts)
- Subscription status (private)
- Channel analytics (private)

**COMPLIANCE:**
- Only public Twitch data via API
- User consent required for any OAuth scopes
- No scraping of private profile pages

---

## Data Retention Policies

### 1. Clip Metadata

**Retention:** Indefinite (while clip exists on Twitch)

**Deletion Triggers:**
1. **Clip Deleted on Twitch**
   - Embed fails to load
   - User reports clip unavailable
   - Manual verification confirms deletion
   - **Action:** Mark as removed, purge within 30 days

2. **DMCA Takedown**
   - Twitch removes clip
   - Embed fails
   - **Action:** Immediate removal from database

3. **User Request** (GDPR/CCPA)
   - User requests their submission data deleted
   - **Action:** Delete within 30 days

4. **Broadcaster Request**
   - Broadcaster requests their clips removed
   - Verified via Twitch connection or email
   - **Action:** Remove within 7 days

**Verification:**
```sql
-- Clips marked as removed
UPDATE clips SET is_removed = true, removed_reason = 'deleted_on_twitch'
WHERE twitch_clip_id = ?;

-- Purge removed clips older than 30 days
DELETE FROM clips WHERE is_removed = true AND updated_at < NOW() - INTERVAL '30 days';
```

---

### 2. Cached User/Game Data

**Retention:** Redis TTL (auto-expires)

**TTL Policies:**
- User metadata: 1 hour
- Game metadata: 4 hours  
- Stream status: 30 seconds
- Channel info: 1 hour

**No Manual Deletion Needed:**
- Redis automatically evicts on TTL expiry
- No persistent storage of this data
- Fresh data fetched when cache misses

---

### 3. OAuth Tokens

**Retention:** Until user revokes or account deleted

**Deletion Triggers:**
1. **User Revokes**
   - DELETE `/api/v1/twitch/auth`
   - **Action:** Immediate deletion

2. **Account Deletion**
   - User deletes Clipper account
   - **Action:** All OAuth tokens deleted within 30 days

3. **Invalid Token**
   - Token refresh fails (user revoked on Twitch)
   - **Action:** Delete invalid tokens

4. **Inactivity**
   - User hasn't logged in for 2 years
   - **Action:** Delete account and all tokens

**Encryption:**
- Tokens encrypted with AES-256
- Encryption keys rotated annually
- Keys stored in a secure platform-managed secret store

---

### 4. Access Tokens (App Credentials)

**Retention:** Until expiry (then refreshed)

**TTL:** Expires_in - 5 minutes (safe margin)

**Storage:**
- Redis cache only
- Not persisted in database
- Auto-refreshed on expiry
- New token replaces old token

---

## Data Usage Restrictions

### Permitted Uses

✅ **Display clips on Clipper platform**  
✅ **Enable search and discovery features**  
✅ **Show broadcaster/game information**  
✅ **Track clip engagement (votes, comments)**  
✅ **Provide clip recommendations**  
✅ **Display live stream status**  
✅ **Facilitate Twitch chat integration (with user consent)**

### Prohibited Uses

❌ **Sell or license Twitch data to third parties**  
❌ **Create derivative databases for resale**  
❌ **Provide clip download services**  
❌ **Re-host or mirror Twitch videos**  
❌ **Scrape additional data not provided by API**  
❌ **Use for training AI/ML models without consent**  
❌ **Share user OAuth tokens with third parties**  
❌ **Use data for purposes outside Service scope**

---

## Third-Party Data Sharing

### Who We Share Twitch Data With

**NO SHARING** except:

1. **Stripe (Payment Processing)**
   - **Data:** User's Twitch username (optional, for verification)
   - **Purpose:** Link subscription to Twitch creator
   - **Legal Basis:** User consent, contract performance

2. **Twitch (API Calls)**
   - **Data:** OAuth tokens for authorized operations
   - **Purpose:** Chat integration, stream embeds
   - **Legal Basis:** User authorization, Twitch ToS

3. **Law Enforcement (Legal Requirement)**
   - **Data:** Minimal necessary data per court order
   - **Purpose:** Comply with legal obligations
   - **Legal Basis:** Legal obligation

**NOT SHARED WITH:**
- ❌ Advertising networks
- ❌ Analytics platforms (no Twitch data in analytics)
- ❌ Marketing partners
- ❌ Data brokers
- ❌ AI training companies
- ❌ Any other third parties

---

## User Data Rights

### GDPR Rights (EU/UK/Swiss)

**Right to Access:**
- Users can request all Twitch-derived data we hold
- Provided in JSON format within 30 days
- Includes: Clip submissions, OAuth connection status

**Right to Deletion ("Right to be Forgotten"):**
- Users can request deletion of their Twitch connection
- OAuth tokens deleted immediately
- Clip submissions can be disassociated or deleted
- Action within 30 days

**Right to Portability:**
- Export all clip submission metadata
- Export Twitch connection info (username, connection date)
- JSON format provided

**Right to Object:**
- Users can object to data processing
- Twitch connection can be revoked anytime
- Service functionality may be limited

---

### CCPA Rights (California)

**Right to Know:**
- Categories of Twitch data collected
- Sources (Twitch API, user OAuth)
- Purposes (Service functionality)
- Third parties shared with (none, except service providers)

**Right to Delete:**
- Request deletion of Twitch-derived data
- 45-day response time
- Some data retained for legal obligations

**Right to Opt-Out:**
- We DO NOT SELL Twitch data
- No opt-out needed

---

## Data Security

### Encryption

**At Rest:**
- OAuth tokens: AES-256 encryption
- Database: Encrypted volumes (AWS RDS)
- Backups: Encrypted

**In Transit:**
- TLS 1.3 for all API calls
- HTTPS only for embeds and frontend
- Encrypted Redis connections

### Access Control

**Who Can Access Twitch Data:**
- Backend services (authenticated via API keys)
- Database administrators (MFA required, audit logged)
- Support team (view-only, no tokens)

**Who CANNOT Access:**
- Frontend JavaScript (tokens never sent to client)
- Third-party services (except Twitch API)
- Unauthenticated requests

---

## Compliance Verification

### Quarterly Audit Checklist

- [ ] No video/audio files stored
- [ ] Only metadata in database
- [ ] OAuth tokens encrypted
- [ ] Deleted clips purged
- [ ] Cache TTLs appropriate
- [ ] No unauthorized data sharing
- [ ] User data rights process functional
- [ ] GDPR/CCPA export working
- [ ] Data retention policies followed
- [ ] No scraping or unofficial APIs

---

## Incident Response

### Data Breach Procedure

**If Twitch data compromised:**
1. **Immediate:** Revoke all OAuth tokens
2. **Within 24 hours:** Notify affected users
3. **Within 72 hours:** Notify Twitch and regulators (if GDPR applies)
4. **Within 7 days:** Provide breach details and remediation plan

**If unauthorized access detected:**
1. Lock down access immediately
2. Audit access logs
3. Notify Twitch if API credentials compromised
4. Rotate all tokens and keys

---

## References

- [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
- [Twitch Privacy Notice](https://www.twitch.tv/p/legal/privacy-notice/)
- [GDPR Compliance](https://gdpr.eu/)
- [CCPA Compliance](https://oag.ca.gov/privacy/ccpa)

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-29 | Initial data retention policy | Backend Team, Legal |

---

**Document Status:** ✅ COMPLETE  
**Next Review:** 2026-03-29 (Quarterly)
