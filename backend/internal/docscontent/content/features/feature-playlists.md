---
title: "Feature Playlists"
summary: "Summary"
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

## Feature: Playlist Creation & Management

Summary
Allow users to create, manage, and share playlists of clips. Playlists support reorder, privacy settings, and collaboration invites.

API
- `POST /api/v1/playlists` create
- `GET /api/v1/playlists/:id` get
- `PATCH /api/v1/playlists/:id` update
- `POST /api/v1/playlists/:id/clips` add clip
- `DELETE /api/v1/playlists/:id/clips/:clip_id` remove clip
- `POST /api/v1/playlists/:id/reorder` reorder
- Public/private flag and `share_token`

Frontend
- Playlist creation modal
- Drag-and-drop reorder
- Share link generator
- Collaboration invite (by username/email)

Acceptance Criteria
- Users can create and share playlists
- Reordering persists
- Privacy settings respected
- Playlist page load < 500ms

Testing
- Test playlist with 1000 clips
- Test share token security
