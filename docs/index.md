---
title: "Clipper Documentation"
summary: "Complete documentation hub for the Clipper platform - a modern, community-driven Twitch clip curation platform."
tags: ["docs", "hub", "index"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["home", "docs home", "documentation"]
---

# Clipper Documentation

> A modern, community-driven Twitch clip curation platform

Welcome to the Clipper documentation! This is your comprehensive guide to using, developing, and deploying Clipper.

## 🚀 Quick Start

- **New Users?** Start with the [[users/user-guide|User Guide]]
- **Setting up dev environment?** See [[setup/development|Development Setup]]
- **Need API docs?** Check [[backend/api|API Reference]]
- **Deploying to production?** Read [[operations/deployment|Deployment Guide]]

## 📚 Documentation Sections

### For Users

Guides for using the Clipper platform:

- **[[users/user-guide|User Guide]]** - Complete guide to browsing, voting, commenting, and favoriting clips
- **[[users/faq|FAQ]]** - Frequently asked questions
- **[[users/community-guidelines|Community Guidelines]]** - Rules, content policy, and moderation

### Features

In-depth documentation for major platform features:

- **[[features/index|Features Hub]]** - Complete feature documentation index
- **[[features/comments|Comment System]]** - Reddit-style nested threading, E2E testing, and performance optimization
- **[[features/live-streams|Live Streams]]** - Watch Twitch streams with integrated chat, follow notifications, and clip creation
- **[[features/feature-feed-filtering|Feed Filtering]]** - Personalized content filtering
- **[[features/feature-playlists|Playlists]]** - User-created clip collections
- **[[features/feature-queue-history|Queue History]]** - Viewing history and playback queue
- **[[features/feature-theatre-mode|Theatre Mode]]** - Immersive fullscreen viewing

### For Developers

#### Getting Started

- **[[setup/development|Development Setup]]** - Environment setup, dependencies, and quick start
- **[[setup/environment|Environment Variables]]** - Configuration reference for all services
- **[[setup/troubleshooting|Troubleshooting]]** - Common issues and solutions
- **[[setup/development-workflow|Development Workflow]]** - Git workflow and best practices

#### Backend (Go + PostgreSQL)

- **[[backend/index|Backend Hub]]** - Complete backend documentation index
- **[[backend/architecture|Architecture]]** - System design, components, and data flow
- **[[backend/api|API Reference]]** - REST API endpoints, authentication, and examples
- **[[backend/api-moderation-index|Moderation API]]** - Complete moderation API with bans, moderators, and audit logs
- **[[backend/clip-submission-api-hub|Clip Submission API Hub]]** - Navigate all submission API resources
- **[[backend/clip-submission-api-guide|Clip Submission API Guide]]** - Complete guide with TypeScript and cURL examples
- **[[backend/clip-submission-api-quickref|Clip Submission Quick Reference]]** - Quick reference card for developers
- **[[backend/clip-api|Clip API]]** - Clip CRUD operations and management
- **[[backend/comment-api|Comment API]]** - Comment system with markdown support
- **[[backend/webhooks|Webhooks]]** - Outbound webhook integration
- **[[openapi/README|OpenAPI Specifications]]** - Machine-readable API specifications
- **[[backend/database|Database]]** - Schema, migrations, maintenance, and optimization
- **[[backend/search|Search Platform]]** - OpenSearch setup, indexing, and querying
- **[[backend/semantic-search|Semantic Search]]** - Vector search and hybrid BM25+embedding architecture
- **[[backend/caching-strategy|Caching Strategy]]** - Redis caching patterns
- **[[backend/rbac|RBAC]]** - Role-based access control and permissions
- **[[backend/security|Security]]** - Security best practices and threat mitigation
- **[[testing/index|Testing Hub]]** - Complete testing documentation index
- **[[testing/TESTING|Testing Strategy]]** - Comprehensive testing strategy for Roadmap 5.0 (unit, integration, E2E, load, scheduler, observability)
- **[[testing-gap-issues|Testing Gap Issues]]** - Known testing coverage gaps and tracking

#### Frontend (React + TypeScript)

- **[[frontend/index|Frontend Hub]]** - Complete frontend documentation index
- **[[frontend/architecture|Architecture]]** - Component structure, state management, and patterns
- **[[frontend/dev-guide|Development Guide]]** - Component creation, styling, and best practices
- **[[frontend/component-library|Component Library]]** - Reusable UI components
- **[[frontend/accessibility|Accessibility]]** - WCAG compliance and best practices

#### Mobile (React Native + Expo)

- **[[mobile/index|Mobile Hub]]** - Complete mobile app documentation index
- **[[mobile/architecture|Architecture]]** - App structure, navigation, and platform considerations
- **[[mobile/implementation|Implementation Guide]]** - Features, OAuth, search, comments, and submit flow
- **[[mobile/deep-linking|Deep Linking]]** - Universal links and app navigation
- **[[mobile/offline-caching|Offline Caching]]** - Offline-first architecture
- **[[mobile/i18n|Internationalization]]** - Multi-language support

#### Data Pipelines

- **[[pipelines/ingest|Data Ingestion]]** - Twitch API integration and clip importing
- **[[pipelines/clipping|Clip Processing]]** - Metadata extraction and enrichment
- **[[pipelines/analysis|Analytics Pipeline]]** - User behavior and engagement analytics

### Premium & Monetization

- **[[premium/index|Premium Hub]]** - Complete premium and monetization documentation index
- **[[premium/overview|Premium Overview]]** - Complete guide to subscription features
- **[[premium/tiers|Pricing Tiers]]** - Free vs Pro benefits, pricing strategy
- **[[premium/entitlements|Entitlements]]** - Feature gates and access control implementation
- **[[premium/stripe|Stripe Integration]]** - Billing, trials, discounts, and payment recovery
- **[[premium/dunning|Dunning]]** - Payment recovery process
- **[[premium/trials-and-discounts|Trials & Discounts]]** - Trial periods and discount codes
- **[[product/ad-slot-specification|Ad Slot Specification]]** - Ad taxonomy, placements, sizes, and fallback rules

### Deployment & Operations

- **[[operations/index|Operations Hub]]** - Complete operations documentation index
- **[[deployment/docker|Docker Deployment]]** - Container-based deployment and multi-stage builds
- **[[deployment/ci_cd|CI/CD Pipeline]]** - GitHub Actions workflows and automation
- **[[deployment/infra|Infrastructure]]** - Kubernetes, cloud providers, and scaling
- **[[deployment/runbook|Operations Runbook]]** - Day-to-day operational procedures
- **[[operations/preflight|Preflight Checklist]]** - Pre-deployment validation steps
- **[[operations/migration|Database Migrations]]** - Migration planning and execution
- **[[operations/monitoring|Monitoring]]** - Error tracking, logging, and observability
- **[[operations/feature-flags|Feature Flags]]** - Gradual rollout and feature toggles
- **[[operations/secrets-management|Secrets Management]]** - Secure credential handling
- **[[operations/security-scanning|Security Scanning]]** - Automated security checks
- **[[operations/observability|Observability]]** - Distributed tracing and metrics

### Architecture Decisions

- **[[decisions/adr-1-semantic-search-vector-db|ADR 001: Semantic Search & Vector DB]]** - Hybrid search architecture
- **[[decisions/adr-002-mobile-framework-selection|ADR 002: Mobile Framework]]** - React Native + Expo decision
- **[[decisions/adr-003-advanced-query-language|ADR 003: Query Language]]** - Advanced search syntax
- **[[decisions/index|All Architecture Decisions]]** - Complete ADR index
- **[[rfcs/index|Request for Comments (RFCs)]]** - Feature proposals and RFCs

### Product & Compliance

- **[[product/index|Product Hub]]** - Complete product documentation index
- **[[product/features|Features]]** - Complete feature list
- **[[product/roadmap|Roadmap]]** - High-level roadmap overview
- **[[product/roadmap-5.0|Roadmap 5.0]]** - Detailed current roadmap
- **[[product/reputation-system|Reputation System]]** - Karma and trust scores
- **[[product/trust-system|Trust System]]** - User trust and moderation
- **[[product/tagging-system|Tagging System]]** - Clip categorization
- **[[product/analytics|Analytics]]** - User engagement tracking
- **[[product/query-grammar|Query Grammar]]** - Advanced search syntax
- **[[compliance/index|Compliance]]** - Legal, regulatory, and compliance documentation
- **[[legal/index|Legal]]** - Terms of service and privacy policy

## 📖 Additional Resources

- **[[introduction|Introduction]]** - Project overview and key concepts
- **[[glossary|Glossary]]** - Terms and definitions
- **[[changelog|Changelog]]** - Version history and release notes
- **[[contributing|Contributing]]** - How to contribute to the project
- **[[adr/index|Legacy ADRs]]** - Older architecture decision records
- **[[examples/index|Code Examples]]** - Sample code and integration examples
- **[[openapi/index|OpenAPI Specifications]]** - Machine-readable API specs

## 🏗️ Project Status

**Current Version**: v0.x (Pre-release)  
**Status**: Active Development  
**Target**: MVP Release Q2 2025

### Implementation Status

- ✅ Core backend API (Go + Gin)
- ✅ PostgreSQL schema with migrations
- ✅ Twitch OAuth authentication
- ✅ OpenSearch integration
- ✅ Semantic vector search
- ✅ React frontend (web)
- ✅ React Native mobile apps
- ✅ Premium subscription system
- ✅ CI/CD pipeline
- 🚧 Production hardening
- 🚧 Mobile app release
- 📋 Advanced moderation tools
- 📋 Machine learning recommendations

## 🔗 External Links

- [GitHub Repository](https://git.subcult.tv/subculture-collective/clpr)
- [Issue Tracker](https://git.subcult.tv/subculture-collective/clpr/issues)
- [Discussions](https://git.subcult.tv/subculture-collective/clpr/discussions)
- [Twitch API Documentation](https://dev.twitch.tv/docs/api/)

## 💡 Using This Documentation

This documentation is structured for repository-based browsing, code review, and CI validation.

### Navigation Tips

- **Search**: Use your editor or repository search to find pages quickly
- **Indexes**: Start from this page or section hubs to navigate docs
- **Wikilinks**: Many docs still use `[[page-name]]` links for concise internal references

### Markdown Conventions

- **Wikilinks**: `[[page-name]]` or `[[page-name|Display Text]]`
- **Relative Links**: `[Link Text](./relative/path.md)`
- **Code Blocks**: Triple backticks with language identifier
- **Callouts**: Use `> [!note]`, `> [!warning]`, `> [!tip]` for emphasis
- **Tables**: GitHub-flavored markdown tables for structured data
- **Frontmatter**: All pages include concise YAML metadata for title, summary, area, status, owner, and review date

### Contributing to Docs

Found an error or want to improve the documentation?

1. Check [[contributing|Contributing Guide]] for guidelines
2. Match frontmatter and formatting conventions from nearby pages
3. Run quality checks locally: `npm run docs:check`
4. Submit a PR with your changes
5. Tag with `documentation` label
6. Documentation changes are validated via CI (see [[contributing/docs-quality-checks|Quality Checks]])

## 📝 Documentation Validation

All documentation is automatically validated on every commit:

- **Markdown Linting**: Ensures consistent formatting
- **Spell Checking**: Catches typos and errors
- **Link Validation**: Verifies all links work (excludes localhost)
- **Anchor Checking**: Confirms internal link targets exist
- **Orphan Detection**: Finds unreachable documentation (BFS from index.md)
- **Asset Hygiene**: Checks for unused or oversized images

📚 **See [[contributing/docs-quality-checks|Documentation Quality Checks]]** for:
- How to run checks locally
- Detailed explanation of each check
- Troubleshooting common issues
- Configuration file documentation
- Best practices

Related: [CI Workflow](.github/workflows/docs.yml) | Issues [#803](https://git.subcult.tv/subculture-collective/clpr/issues/803), [#845](https://git.subcult.tv/subculture-collective/clpr/issues/845), [#846](https://git.subcult.tv/subculture-collective/clpr/issues/846), [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805)

## 🆘 Getting Help

- **Users**: Check [[users/faq|FAQ]] or [[users/user-guide|User Guide]]
- **Developers**: See [[setup/troubleshooting|Troubleshooting]] or open an issue
- **Contributors**: Read [[contributing|Contributing Guide]]
- **Bugs**: [Report an issue](https://git.subcult.tv/subculture-collective/clpr/issues/new)

---

**Last Updated**: 2026-01-29  
**Maintained by**: [Subculture Collective](https://git.subcult.tv/subculture-collective)

**Related Issues**: [#803](https://git.subcult.tv/subculture-collective/clpr/issues/803), [#845](https://git.subcult.tv/subculture-collective/clpr/issues/845), [#846](https://git.subcult.tv/subculture-collective/clpr/issues/846)
