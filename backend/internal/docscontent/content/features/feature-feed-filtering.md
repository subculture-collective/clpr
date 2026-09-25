---
title: "Feature Feed Filtering"
summary: "Summary"
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

## Feature: Feed Filtering UI & API

Summary
Provide a robust feed filtering UI and corresponding API to let users filter content by game, streamer, tags, date range, and sort order. Include saved presets.

API
- `GET /api/v1/feed?game=...&streamer=...&tags=...&from=...&to=...&sort=trending|new|comments&page=1&limit=20`
- Server-side filtering and pagination
- Persistent saved filters: `POST /api/v1/users/:id/filters` and `GET /api/v1/users/:id/filters`

Frontend
- Filter bar on homepage with multi-select tags, streamer input, date picker
- Presets dropdown to save/load filters
- Toggle for compact/list/grid view
- Client-side debounce on inputs

Acceptance Criteria
- Filters apply within 200ms for cached queries
- Server returns correct counts and pages
- Saved filters persist per user
- No N+1 queries

Testing
- Test with large dataset (100k clips)
- Validate edge cases for empty filters
