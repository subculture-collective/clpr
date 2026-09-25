---
title: "Compliance Documentation"
summary: "Legal, regulatory, and compliance documentation for Clipper including GDPR, data retention, and third-party API usage."
tags: ["compliance", "hub", "index", "legal", "gdpr"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
aliases: ["compliance hub", "legal compliance"]
---

# Compliance Documentation

Legal, regulatory, and compliance documentation for Clipper including GDPR, data retention, and third-party API usage.

## Quick Links

- [[README|Compliance Overview]] - Introduction to compliance requirements
- [[data-retention|Data Retention Policy]] - User data storage and deletion policies
- [[gdpr-compliance|GDPR Compliance]] - European data protection compliance
- [[guardrails|Content Guardrails]] - Content moderation and safety policies
- [[oauth-scopes|OAuth Scopes]] - Third-party authentication permissions
- [[twitch-api-usage|Twitch API Usage]] - Twitch API compliance requirements
- [[twitch-embeds|Twitch Embeds]] - Twitch embed policy and implementation

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/compliance"
WHERE file.name != "index" AND file.name != "README"
SORT title ASC
```

## Compliance Areas

### Data Protection

- [[gdpr-compliance|GDPR Compliance]] - European data protection regulations
- [[data-retention|Data Retention]] - Data storage and deletion policies

### Third-Party Integration

- [[twitch-api-usage|Twitch API Usage]] - Compliance with Twitch Developer Agreement
- [[twitch-embeds|Twitch Embeds]] - Proper usage of Twitch embed players
- [[oauth-scopes|OAuth Scopes]] - Authentication scope requirements

### Content & Safety

- [[guardrails|Content Guardrails]] - Content moderation policies and implementation

## Related Documentation

- [[../legal/index|Legal Documentation]] - Terms of Service, Privacy Policy
- [[../backend/security|Security Best Practices]]
- [[../operations/secrets-management|Secrets Management]]

---

**See also:**
[[../legal/index|Legal]] ·
[[../backend/security|Security]] ·
[[../index|Documentation Home]]
