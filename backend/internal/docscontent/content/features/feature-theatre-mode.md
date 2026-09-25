---
title: "Feature Theatre Mode"
summary: "Summary"
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

## Feature: Theatre Mode & Player

Summary
Implement immersive theatre mode player with full-screen playback, keyboard controls, and playlist auto-play.

API/Backend
- Ensure player endpoints support range requests and adaptive streaming (HLS)
- Player analytics events: `player.play`, `player.pause`, `player.seek`, `player.fullscreen`, `player.quality_change`

Frontend
- Full-screen layout with dark backdrop
- Playlist sidebar toggle
- Controls: play/pause, next, previous, speed, quality, subtitles toggle
- Keyboard shortcuts: space (play/pause), arrow keys (seek), F (fullscreen), N (next)
- Theatre mode persists user preference

Acceptance Criteria
- Full-screen toggles without layout shift
- Auto-play next clip in playlist with 3s gap
- Keyboard shortcuts responsive < 50ms
- HLS playback works across supported browsers

Testing
- Test on Chrome, Firefox, Safari
- Test HLS fallback to progressive when needed
