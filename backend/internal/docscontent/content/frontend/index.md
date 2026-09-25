---
title: "Frontend Documentation"
summary: "React frontend architecture, components, and development guides."
tags: ["frontend", "hub", "index"]
area: "frontend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-01
aliases: ["frontend hub", "react docs"]
---

# Frontend Documentation

This section covers the Clipper web frontend built with React 19, TypeScript, and Vite.

## Quick Links

- [[architecture|Architecture]] - Component structure and patterns
- [[dev-guide|Development Guide]] - Setup and workflow
- [[component-library|Component Library]] - Reusable UI components
- [[accessibility|Accessibility]] - WCAG compliance and best practices
- [[accessibility-test-results|Accessibility Test Results]] - Audit results
- [[performance-accessibility-audit|Performance & Accessibility Audit]] - Combined audit report
- [[css-injection-prevention|CSS Injection Prevention]] - Security best practices
- [[env-policy|Environment Policy]] - Environment variable handling
- [[docs-rendering|Docs Rendering]] - Documentation rendering system

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/frontend"
WHERE file.name != "index"
SORT title ASC
```

## Tech Stack

- **React 19** with TypeScript
- **Vite** for build tooling
- **TailwindCSS** for styling
- **React Router** for navigation
- **TanStack Query** for server state
- **Zustand** for client state
- **Axios** for HTTP client

---

**See also:**
[[../mobile/architecture|Mobile Architecture]] ·
[[../backend/api|API Reference]] ·
[[../index|Documentation Home]]
