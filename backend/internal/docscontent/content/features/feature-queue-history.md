---
title: "Feature Queue History"
summary: "Summary"
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

## Feature: Queue & Watch History

Summary
Allow users to add clips to a personal queue, reorder, persist across sessions, and view their watch history for resuming playback.

API
- `POST /api/v1/users/:id/queue` add
- `GET /api/v1/users/:id/queue` list
- `PATCH /api/v1/users/:id/queue/reorder` reorder
- `DELETE /api/v1/users/:id/queue/:clip_id` remove
- Watch history API: `GET /api/v1/users/:id/history?limit=50`

Frontend
- Queue drawer accessible from player
- Reorder with drag-and-drop
- Resume button on history items to start from last position
- Clear history button with confirmation

Acceptance Criteria
- Queue persists across devices when logged in
- Reorder latency < 300ms
- History shows timestamp of last watched position
- Privacy: users can opt-out of history tracking

Testing
- Sync queue across multiple devices (web + mobile)
- Test history resume at correct timestamp
