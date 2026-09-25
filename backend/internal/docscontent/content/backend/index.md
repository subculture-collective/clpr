---
title: "Backend Documentation"
summary: "Go backend architecture, API, database, and services documentation."
tags: ["backend", "hub", "index"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["backend hub", "api docs"]
---

# Backend Documentation

This section covers the Clipper backend implemented in Go with PostgreSQL, Redis, and OpenSearch.

## Quick Links

- [[architecture|Architecture]] - System design and components
- [[api|API Reference]] - REST API endpoints and usage
- [[clip-submission-api-hub|Clip Submission API Hub]] - Complete submission API documentation
- [[database|Database]] - Schema, migrations, and queries
- [[search|Search Platform]] - OpenSearch setup and querying
- [[semantic-search|Semantic Search]] - Vector search architecture
- [[testing|Testing Guide]] - Testing strategy and tools

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/backend"
WHERE file.name != "index"
SORT title ASC
```

## Core Components

### Architecture & Design

- [[architecture|Backend Architecture]] - High-level system design
- [[api|API Reference]] - Complete REST API documentation

### APIs

- [[clip-submission-api-hub|Clip Submission API Hub]] - Submission API navigation
- [[clip-submission-api-guide|Clip Submission API Guide]] - Complete guide with examples
- [[clip-submission-api-quickref|Clip Submission Quick Reference]] - Quick reference card
- [[clip-api|Clip API]] - Clip CRUD operations
- [[comment-api|Comment API]] - Comment system with markdown
- [[api-playlist-sharing|Playlist Sharing API]] - Playlist sharing endpoints
- [[watch-parties-api|Watch Parties API]] - Watch party features
- [[watch-parties-api-features|Watch Parties API Features]] - Feature documentation

### Data & Storage

- [[database|Database]] - PostgreSQL schema and operations
- [[search|Search Platform]] - OpenSearch integration
- [[search-evaluation|Search Evaluation]] - Search quality evaluation
- [[search-feature-completion|Search Feature Completion]] - Search feature status
- [[semantic-search|Semantic Search]] - Vector similarity search
- [[semantic-search-arch|Semantic Search Architecture]] - Detailed architecture
- [[caching-strategy|Caching Strategy]] - Redis caching patterns
- [[redis-operations|Redis Operations]] - Redis management
- [[PGBOUNCER|PgBouncer]] - Connection pooling
- [[PGBOUNCER_QUICKSTART|PgBouncer Quick Start]] - Quick setup guide

### Security

- [[authentication|Authentication]] - OAuth and JWT flow
- [[rbac|RBAC]] - Role-based access control
- [[authorization-framework|Authorization Framework]] - Authorization system design
- [[security|Security]] - Security best practices
- [[trust-score-implementation|Trust Score]] - Trust scoring system
- [[trust-score-security|Trust Score Security]] - Security considerations
- [[submission-rate-limiting|Rate Limiting]] - API rate limiting
- [[ABUSE_DETECTION|Abuse Detection]] - Automated abuse detection
- [[AUTHORIZATION_AUDIT_LOGGING|Authorization Audit Logging]] - Audit logging for authorization
- [[SENDGRID_WEBHOOK_SECURITY|SendGrid Webhook Security]] - Webhook signature verification

### Integrations

- [[twitch-integration|Twitch Integration]] - Twitch API integration
- [[twitch-moderation-api|Twitch Moderation API]] - Twitch moderation integration
- [[email-service|Email Service]] - Email notifications
- [[email-templates|Email Templates]] - Email template documentation
- [[broadcaster-live-sync-implementation|Broadcaster Live Sync]] - Live status tracking
- [[broadcaster-live-sync-testing|Broadcaster Live Sync Testing]] - Testing guide
- [[webhooks|Webhooks]] - Outbound webhook system
- [[webhook-retry|Webhook Retry]] - Retry logic and DLQ
- [[webhook-signature-verification|Webhook Signature Verification]] - Webhook security
- [[webhook-subscription-management|Webhook Subscription Management]] - Subscription management
- [[websocket-configuration|WebSocket Configuration]] - WebSocket setup

### Quality & Performance

- [[testing|Testing Guide]] - Unit, integration, and load testing
- [[test-roadmap|Test Roadmap]] - Testing coverage roadmap
- [[testing-performance|Performance Testing]] - Load testing strategies
- [[profiling|Profiling]] - Performance profiling guide
- [[optimization-analysis|Optimization Analysis]] - Performance optimization
- [[CF-OPTIMIZATION-RESULTS|Cloudflare Optimization Results]] - CDN optimization

### Content Moderation

- [[api-moderation-index|Moderation API Quick Start]] - Quick start guide for moderation API
- [[moderation-api|Moderation API Reference]] - Complete moderation API documentation
- [[CHAT_MODERATION|Chat Moderation]] - Chat moderation features
- [[NSFW_DETECTION|NSFW Detection]] - Content safety detection
- [[NSFW_ENV_VARS|NSFW Environment Variables]] - Configuration
- [[TOXICITY_CLASSIFICATION|Toxicity Classification]] - Toxicity detection
- [[TOXICITY_RULES|Toxicity Rules]] - Moderation rules
- [[AUDIT_LOG_SERVICE|Audit Log Service]] - Audit logging service

### Infrastructure

- [[APPLICATION_LOGS|Application Logs]] - Logging infrastructure
- [[FFMPEG_JOB_QUEUE|FFmpeg Job Queue]] - Video processing queue
- [[cdn-integration|CDN Integration]] - CDN setup and configuration
- [[mirror-hosting|Mirror Hosting]] - Video mirror hosting

---

**See also:**
[[../setup/development|Development Setup]] ·
[[../operations/deployment|Deployment]] ·
[[../index|Documentation Home]]
