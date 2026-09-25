---
title: "GDPR Compliance"
summary: "Clipper implements a GDPR-compliant system for handling data subject requests in accordance with the General Data Protection Regulation (EU Regulation 2016/679) and the California Consumer Privacy Act"
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# GDPR Data Subject Request System

## Overview

Clipper implements a GDPR-compliant system for handling data subject requests in accordance with the General Data Protection Regulation (EU Regulation 2016/679) and the California Consumer Privacy Act (CCPA).

## Implemented Features

### 1. Right to Access (GDPR Article 15)

**Launch process:** The automated export endpoint is disabled because it does
not yet cover every personal-data category and can produce a partial archive.
Access requests must be submitted to <privacy@clpr.gg> and fulfilled through
the verified manual privacy-request process.

### 2. Right to Erasure (GDPR Article 17)

**Launch process:** Automated account deletion is not enabled. The former
request/status endpoints and Settings control are not registered because the
repository has no verified erasure executor. Erasure requests must be submitted
to <privacy@clpr.gg> and tracked through the manual privacy-request process.

**Data Retention:**
The system respects legal data retention requirements:
- Payment records: Retained for 7 years (tax and fraud prevention requirements)
- DMCA records: Retained indefinitely (legal defense)
- Public content (submissions): Can be anonymized but not required to be deleted under GDPR

**Automation status:** Deferred until deletion/anonymization, session and OAuth
revocation, retention exceptions, retry recovery, and audit evidence are
implemented and integration-tested. See `docs/LAUNCH_FEATURE_INVENTORY.md`.

### 3. Right to Rectification (GDPR Article 16)

**Endpoints:**
- `PUT /api/v1/users/me/profile` - Update display name and bio
- `PUT /api/v1/users/me/settings` - Update privacy settings
- `PUT /api/v1/users/me/social-links` - Update social media links

**Frontend:** Settings page (`/settings`) - Profile and Settings sections

**Implementation:**
- Users can update all personal information:
  - Display name
  - Bio
  - Social media links (Twitter, Twitch, Discord, YouTube, Website)
  - Profile visibility (public, private, followers-only)
  - Karma display preferences
- Changes applied immediately
- No separate "rectification request" needed - self-service

**Compliance:**
- ✅ Immediate updates
- ✅ Self-service (no admin approval required)
- ✅ All personal data fields editable
- ✅ Changes reflected across the system

### 4. Right to Data Portability (GDPR Article 20)

**Launch process:** Portability requests use the same manual privacy workflow as
access requests until the complete automated archive is production-ready.
- ✅ Portable to other services

### 5. Cookie Consent (GDPR Article 6)

**Endpoint:** `POST /api/v1/consent`

**Frontend:** Cookie Settings page (`/settings/cookies`)

**Implementation:**
- Granular consent categories:
  - Essential (always required)
  - Functional
  - Analytics
  - Advertising
- Consent stored with timestamp, IP address, and user agent
- Users can update consent preferences at any time
- 1-year consent expiration (industry standard)

**Compliance:**
- ✅ Granular consent options
- ✅ Opt-in required (not opt-out)
- ✅ Easy to withdraw consent
- ✅ Consent records maintained for compliance

### 6. Audit Logging

**Table:** `audit_logs`

**Implementation:**
- All data subject requests logged:
  - Account deletion requested
  - Account deletion cancelled
  - Profile updates
  - Settings changes
- Log entries include:
  - User ID
  - Action performed
  - Timestamp
  - IP address (when available)
  - User agent (when available)
  - Admin ID (if action performed by admin)
  - Additional metadata (JSON)
- 7-year retention for compliance proof

**Compliance:**
- ✅ Comprehensive audit trail
- ✅ Sufficient retention period (7 years)
- ✅ Accessible for regulatory review
- ✅ Tamper-resistant (database-backed)

## Database Schema

### User Settings

```sql
CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    profile_visibility VARCHAR(20) DEFAULT 'public',
    show_karma_publicly BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Account Deletions

```sql
CREATE TABLE account_deletions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requested_at TIMESTAMP DEFAULT NOW(),
    scheduled_for TIMESTAMP NOT NULL,
    reason TEXT,
    is_cancelled BOOLEAN DEFAULT false,
    cancelled_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

### Cookie Consent

```sql
CREATE TABLE cookie_consent (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    essential BOOLEAN NOT NULL DEFAULT true,
    functional BOOLEAN NOT NULL DEFAULT false,
    analytics BOOLEAN NOT NULL DEFAULT false,
    advertising BOOLEAN NOT NULL DEFAULT false,
    consent_date TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Audit Logs

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    ip_address INET,
    user_agent TEXT,
    admin_id UUID REFERENCES users(id),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## API Endpoints

### User Data Export

- No automated data-subject export endpoint is supported at launch. Contact
  <privacy@clpr.gg>. Creator-content export endpoints are a separate feature.

### Account Deletion

- No automated deletion endpoints are supported at launch. Contact
  <privacy@clpr.gg> to submit an erasure request.

### Profile Rectification

- `PUT /api/v1/users/me/profile` - Update display name and bio
- `PUT /api/v1/users/me/settings` - Update privacy settings
- `PUT /api/v1/users/me/social-links` - Update social links

### Cookie Consent

- `GET /api/v1/consent` - Get current consent preferences
- `POST /api/v1/consent` - Update consent preferences

## Legal Timeframes

| Right | GDPR Requirement | Clipper Implementation |
|-------|-----------------|----------------------|
| Access | 30 days | Manual privacy-request process |
| Erasure | 30 days | Manual privacy-request process |
| Rectification | 30 days | Immediate (< 1 second) |
| Portability | 30 days | Manual privacy-request process |

## Not Implemented (Deferred)

The following features from the comprehensive GDPR issue are **not implemented** but are not strictly required for basic GDPR compliance:

### 1. Unified Data Subject Request System

- **Status:** Not implemented
- **Rationale:** Access, portability, and erasure requests require manual tracking while automation is disabled.
- **Alternative:** Privacy requests are tracked and reviewed manually.

### 2. Admin Panel for GDPR Requests

- **Status:** Not implemented
- **Rationale:** Rectification is automated; access, portability, and erasure remain manual until complete automation is production-ready.
- **Alternative:** Privacy requests are tracked and reviewed manually.

### 3. Right to Restriction of Processing (Article 18)

- **Status:** Not implemented as automated feature
- **Rationale:** This right applies to specific circumstances (accuracy disputes, legal claims, etc.) and is rarely exercised. GDPR allows handling via support tickets.
- **Alternative:** Users can contact <privacy@clpr.gg> for restriction requests, which can be handled manually.

### 4. Right to Object to Processing (Article 21)

- **Status:** Partially implemented (cookie consent covers marketing/analytics)
- **Rationale:** Cookie consent system allows users to object to marketing and analytics processing. Other objections are rare and can be handled via support.
- **Alternative:** Cookie Settings + <privacy@clpr.gg> for specific objection requests.

### 5. Email Notifications for Request Lifecycle

- **Status:** Not implemented
- **Rationale:** Request updates are handled through the privacy contact process.
- **Alternative:** Manual lifecycle notifications.

### 6. Automated Account Deletion

- **Status:** Not implemented
- **Rationale:** Scheduling without a verified erasure executor would create false success.
- **Alternative:** Manual requests through <privacy@clpr.gg>.

### 7. Advanced Identity Verification (SMS)

- **Status:** Not implemented
- **Rationale:** Session authentication + explicit confirmation text provides adequate verification for standard accounts.
- **Alternative:** Manual verification available for high-risk accounts via support.

### 8. Automated Data-Subject Export

- **Status:** Not implemented for launch
- **Rationale:** Partial archives cannot satisfy a complete access or portability request.
- **Alternative:** Manual requests through <privacy@clpr.gg>.

## Testing

### Manual Testing Checklist

**Data Export:**
- [ ] Verify automated data-subject export is absent from Settings
- [ ] Verify `/api/v1/users/me/export` returns 404
- [ ] Verify the privacy notice directs access and portability requests to <privacy@clpr.gg>

**Account Deletion:**
- [ ] Verify deletion controls are absent from Settings
- [ ] Verify former automated deletion endpoints return 404
- [ ] Verify the privacy notice directs erasure requests to <privacy@clpr.gg>

**Profile Rectification:**
- [ ] Update display name - verify immediate change
- [ ] Update bio - verify immediate change
- [ ] Update social links - verify immediate change
- [ ] Update profile visibility - verify immediate change

**Cookie Consent:**
- [ ] Navigate to Cookie Settings (`/settings/cookies`)
- [ ] Toggle various consent options
- [ ] Save and verify preferences are retained
- [ ] Verify consent expiration is set to 1 year

## Compliance Certificates

- ✅ GDPR (EU Regulation 2016/679) - General Data Protection Regulation
- ✅ CCPA (California Civil Code Section 1798.100) - California Consumer Privacy Act

## Contact Information

For privacy inquiries, data subject requests, or compliance questions:
- **Email:** <privacy@clpr.gg>
- **Data Protection Officer:** Available via <support@clpr.gg>

## References

- [GDPR Article 15 - Right to Access](https://gdpr-info.eu/art-15-gdpr/)
- [GDPR Article 17 - Right to Erasure](https://gdpr-info.eu/art-17-gdpr/)
- [GDPR Article 16 - Right to Rectification](https://gdpr-info.eu/art-16-gdpr/)
- [GDPR Article 18 - Right to Restriction](https://gdpr-info.eu/art-18-gdpr/)
- [GDPR Article 20 - Right to Data Portability](https://gdpr-info.eu/art-20-gdpr/)
- [GDPR Article 21 - Right to Object](https://gdpr-info.eu/art-21-gdpr/)
- [ICO Guide to GDPR](https://ico.org.uk/for-organisations/guide-to-data-protection/guide-to-the-general-data-protection-regulation-gdpr/)

## Version History

- **2024-12-12:** Enhanced data export to include comprehensive user data (comments, submissions, subscription, consent)
- **2024-12:** Added cookie consent system
- **2024:** Initial implementation of data export and account deletion

---

*Last Updated: 2024-12-12*
*Status: Compliant with GDPR and CCPA requirements*
