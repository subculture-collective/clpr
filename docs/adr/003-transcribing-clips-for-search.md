---
title: "003 Transcribing Clips For Search"
summary: "Extends ADR-002: the render host may download public clips above a view threshold to transcribe them for search, keeping only the transcript."
tags: ["adr", "compliance", "media", "search"]
area: "adr"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-10
---

# ADR-003: Transcribing public clips for search

**Status**: Accepted when the pull request that adds it is merged
**Date**: 2026-11-10
**Decision maker**: the owner
**Extends**: [ADR-002](002-media-handling-for-countdowns.md)
**Related issues**: #87

## Context

Transcript search (#87) asked for transcription of "clips above a view threshold". ADR-002 allows the render host to download clips only for countdown videos and Shorts, so the first version of #87 transcribed only countdown clips and the clips of broadcasters who granted `channel:manage:clips`. That covers a few dozen clips a day at most, too few for search to find much.

On 2026-11-10 the owner told the agent to extend ADR-002 to cover search.

## Decision

1. **Which clips.** Public Twitch clips with at least 1,000 views, created in the last 30 days, in English or with no language set. Clips that are removed, hidden, NSFW or taken down are left out, as are clips and broadcasters excluded on request (the countdown exclusion list from ADR-002 and #75). The most-viewed clips go first. Broadcaster consent is not a condition, as in ADR-002.
2. **How many.** At most 500 clips per UTC day, and 5 per one-minute worker cycle, so countdown work in the same worker is not held up. A clip without English speech is not tried again. A clip that fails is tried again after 6 hours, three times at most.
3. **Fetching.** The same path as ADR-002: Helix `GET /clips/downloads` where the broadcaster authorized it, and otherwise yt-dlp on Twitch's player files, with no clpr credentials sent to the player API.
4. **Retention.** The worker deletes each file as soon as it has transcribed it, before it fetches the next. The media directory's 24-hour sweep covers a file left behind by a crash. Only the transcript is kept.
5. **Use of the transcript.** Transcripts feed search: the search index, the `said` filter and short snippets around a match. Full transcripts are never sent to clients, and nothing is published from them.
6. **Exclusion on request.** When a broadcaster or clip is excluded, the worker empties the render host's transcripts of their clips (search and countdown) on its next cycle, and the search index drops the text within minutes. Transcripts from the broadcaster's own authorization are theirs to manage and stay.

The view threshold, age and daily limit are settings (`SEARCH_TRANSCRIPT_MIN_VIEWS`, `SEARCH_TRANSCRIPT_MAX_AGE_DAYS`, `SEARCH_TRANSCRIPT_DAILY_LIMIT`), and `SEARCH_TRANSCRIPTS=off` turns the feature off. The numbers above are the defaults chosen when the decision was made, not limits the owner set. Raising them needs no new decision; downloading for another purpose does.

## Why the fair use position still holds

This use is further from republishing than the countdowns: the clip itself is never shown, edited or published, only words from it appear, a few at a time, next to an embed of the original clip that Twitch serves. Searching by speech sends viewers to the original clip.

## Consequences

- **More downloads from Twitch.** Up to 500 more clip fetches a day through the same unofficial player path, so the Twitch API risk in ADR-002 grows with volume. If Twitch objects, setting `SEARCH_TRANSCRIPTS=off` stops it without touching countdowns.
- **Compute.** Whisper runs on the render host. Each transcription is recorded in `transcription_usage_daily` under the source `search_render_whisper`, so its cost is visible separately from countdowns.
- **Disclosure.** The terms of service and `docs/compliance/twitch-api-usage.md` describe this use.

## Alternatives considered

- **Countdown and authorized clips only** (the first version of #87). Allowed under ADR-002, but too few clips for search to find much. Kept as the fallback with `SEARCH_TRANSCRIPTS=off`.
- **Twitch's own captions.** Twitch publishes no clip captions or transcripts through Helix.
