---
title: "Roadmap"
summary: "Product roadmap and upcoming features for Clipper."
tags: ["product", "roadmap", "planning"]
area: "product"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["future", "planned features"]
---

# Roadmap

Product roadmap and upcoming features for Clipper.

> **Current Roadmap**: See [[roadmap-5.0|Roadmap 5.0]] for the detailed current roadmap with all phases and issues.

## Current Status

**Version**: v0.x (Pre-release)  
**Status**: Active Development  
**Target**: MVP Release Q2 2025

## Active Roadmap

- **[[roadmap-5.0|Roadmap 5.0]]** - Current detailed roadmap (Q4 2025 - Q2 2026)
  - Phase 0: Foundation
  - Phase 1: Testing Infrastructure
  - Phase 2: Mobile Feature Parity
  - Phase 3: Analytics & ML Enhancement
  - Phase 4: Documentation Excellence (In Progress)
  - Phase 5: Infrastructure Hardening

## Milestones

### v0.1.0 - MVP (Q1 2025)

Core functionality for public beta.

- [x] Core backend API (Go + Gin)
- [x] PostgreSQL schema with migrations
- [x] Twitch OAuth authentication
- [x] Clip browsing and filtering
- [x] Voting system (upvote/downvote)
- [x] Comment system with markdown
- [x] Basic search functionality
- [x] User profiles with karma
- [ ] Admin moderation tools

### v0.2.0 - Search & Discovery (Q1 2025)

Enhanced search and recommendation features.

- [x] OpenSearch integration
- [x] Semantic vector search (pgvector)
- [x] Advanced query language
- [x] Autocomplete suggestions
- [ ] Personalized recommendations
- [ ] Related clips feature

### v0.3.0 - Mobile Apps (Q2 2025)

Native mobile experience.

- [x] React Native + Expo setup
- [x] iOS app development
- [x] Android app development
- [x] Shared component library
- [ ] App store submission
- [ ] Push notifications

### v1.0.0 - General Availability (Q2 2025)

Production-ready release.

- [ ] Production hardening
- [ ] Full test coverage (>80%)
- [ ] Performance optimizations
- [ ] Security audit
- [ ] Documentation complete
- [ ] Marketing site

### v1.1.0 - Growth (Q3 2025)

Post-launch improvements.

- [ ] Machine learning recommendations
- [ ] Creator analytics dashboard
- [ ] API for third-party integrations
- [ ] Internationalization (i18n)
- [ ] Accessibility improvements

### v1.2.0 - Community (Q4 2025)

Enhanced community features.

- [ ] User-created collections (public)
- [ ] Follow users/streamers
- [ ] Clip compilation playlists
- [x] Community moderation tools
- [x] Twitch moderation actions (ban/unban)
- [ ] Achievements/badges

## Recent Additions (2026)

### Twitch Moderation Actions ✅

**Status**: Completed January 2026  
**Epic**: #1059

- [x] OAuth scope integration (`channel:manage:banned_users`, `moderator:manage:banned_users`)
- [x] Backend API endpoints (ban/unban)
- [x] Frontend UI components
- [x] Permission gating (broadcaster/Twitch mod only)
- [x] Site moderator read-only enforcement
- [x] Audit logging
- [x] E2E test coverage
- [x] Documentation and rollout plan

**Features:**
- Permanent bans and temporary timeouts (1s - 14 days)
- Ban reason tracking
- Rate limiting (10 actions/hour)
- Comprehensive error handling
- Audit trail for all actions

See: [Twitch Moderation Actions Docs](./twitch-moderation-actions.md)

## Future Considerations

These features are under evaluation:

- **Live Clips**: Real-time clip notifications
- **Clip Editor**: Basic clip trimming/cropping
- **Social Sharing**: Enhanced sharing to social platforms
- **Twitch Integration**: Deep linking to source streams
- **Creator Partnerships**: Featured creator program

## Feature Requests

Have a feature idea? [Open an issue](https://git.subcult.tv/subculture-collective/clpr/issues/new) with the `enhancement` label.

---

**See also:** [[features|Features]] · [[../changelog|Changelog]] · [[../index|Documentation Home]]
