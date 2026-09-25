---
title: "Product Documentation"
summary: "Product features, roadmap, and glossary."
tags: ["product", "hub", "index"]
area: "product"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["product hub"]
---

# Product Documentation

This section covers product-level documentation including features, roadmap, and terminology.

## Quick Links

- [[features|Features]] - Complete feature list
- [[feature-inventory|Feature Inventory]] - Detailed feature tracking
- [[feature-test-coverage|Feature Test Coverage]] - Test coverage by feature
- [[feature-audit-issues-tracking|Feature Audit Issues]] - Known issues tracking
- [[FEATURE_INVENTORY_SWEEP_II_SUMMARY|Feature Inventory Sweep II]] - Comprehensive review
- [[FEATURE_INVENTORY_SWEEP_II_VERIFICATION|Feature Inventory Verification]] - Verification results
- [[roadmap|Roadmap]] - Upcoming features and milestones
- [[roadmap-5.0|Roadmap 5.0]] - Current roadmap details
- [[rfc-003-implementation-summary|RFC-003 Implementation]] - Infrastructure modernization
- [[../glossary|Glossary]] - Terms and definitions
- [[quick-reference|Quick Reference]] - Command and keyboard shortcuts
- [[keyboard-shortcuts|Keyboard Shortcuts]] - Interface shortcuts

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/product"
WHERE file.name != "index"
SORT title ASC
```

## Product Areas

### Features & Discovery
- [[features|Features]] - Complete feature list
- [[feature-inventory|Feature Inventory]] - Detailed feature tracking
- [[discovery-lists|Discovery Lists]] - Content discovery features
- [[clip-serving-strategy|Clip Serving Strategy]] - Content delivery optimization
- [[scraped-clips|Scraped Clips]] - Automated clip ingestion
- [[ad-slot-specification|Ad Slot Specification]] - Advertising placement

### Search & Query
- [[query-grammar|Query Grammar]] - Advanced search syntax
- [[query-language-examples|Query Language Examples]] - Search examples

### User Systems
- [[reputation-system|Reputation System]] - Karma and trust scores
- [[trust-system|Trust System]] - User trust and moderation
- [[tagging-system|Tagging System]] - Clip categorization
- [[twitch-moderation-actions|Twitch Moderation Actions]] - Moderation tools

### Analytics
- [[analytics|Analytics]] - User engagement tracking
- [[analytics-setup|Analytics Setup]] - Analytics configuration
- [[engagement-metrics|Engagement Metrics]] - Metric definitions

### Security & Threat Model
- [[threat-model|Threat Model]] - Security threat analysis
- [[threat-model-quickref|Threat Model Quick Reference]] - Quick security reference

## Project Status

**Current Version**: v0.x (Pre-release)  
**Status**: Active Development  
**Target**: MVP Release Q2 2025

---

**See also:** [[../users/index|User Documentation]] · [[../premium/overview|Premium]] · [[../index|Documentation Home]]
