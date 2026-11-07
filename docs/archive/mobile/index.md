---
title: "Mobile Documentation"
summary: "React Native mobile app architecture and implementation guides."
tags: ["mobile", "hub", "index"]
area: "mobile"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["mobile hub", "app docs"]
---

# Mobile Documentation

This section covers the Clipper mobile apps built with React Native 0.76 and Expo 52.

## Quick Links

- [[architecture|Architecture]] - App structure and patterns
- [[implementation|Implementation Guide]] - Setup and development
- [[deep-linking|Deep Linking]] - Universal links and app navigation
- [[deep-linking-examples|Deep Linking Examples]] - Code examples
- [[deep-linking-test-cases|Deep Linking Test Cases]] - Testing guide
- [[oauth-pkce|OAuth PKCE]] - Secure authentication flow
- [[offline-caching|Offline Caching]] - Offline-first architecture
- [[i18n|Internationalization]] - Multi-language support
- [[submit-flow-documentation|Submit Flow Documentation]] - Clip submission flow
- [[submit-flow-diagram|Submit Flow Diagram]] - Visual flow diagram
- [[responsive-qa|Responsive QA]] - Responsive design testing
- [[first-visual-examples|First Visual Examples]] - First contentful paint optimization
- [[analytics-readme|Analytics README]] - Analytics overview
- [[analytics-dashboards|Analytics Dashboards]] - PostHog dashboard setup and configuration
- [[analytics-dashboard-setup|Quick Setup Guide]] - Step-by-step dashboard implementation
- [[posthog-dashboard-queries|Dashboard Queries]] - Quick reference for PostHog queries
- [[dashboard-implementation-summary|Dashboard Summary]] - Implementation status and overview

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/mobile"
WHERE file.name != "index"
SORT title ASC
```

## Tech Stack

- **React Native 0.76** with Expo 52
- **Expo Router** for navigation
- **NativeWind** (TailwindCSS) for styling
- **TanStack Query** for server state
- **Zustand** for client state

## Platform Support

- **iOS**: 13+
- **Android**: 8.0+ (API 26+)

---

**See also:** [[../frontend/architecture|Web Frontend]] · [[../backend/api|API Reference]] · [[../index|Documentation Home]]
