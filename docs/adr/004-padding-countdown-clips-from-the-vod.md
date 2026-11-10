---
title: "004 Padding Countdown Clips From The VOD"
summary: "Extends ADR-002: countdown media may include up to 30 seconds of the broadcast before and after each clip, cut from the VOD. Off by default."
tags: ["adr", "compliance", "media"]
area: "adr"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-10
---

# ADR-004: Padding countdown clips from the VOD

**Status**: Accepted when the pull request that adds it is merged
**Date**: 2026-11-10
**Decision maker**: the owner
**Extends**: [ADR-002](002-media-handling-for-countdowns.md)

## Context

A Twitch clip is what a viewer chose to cut, at most 60 seconds, and often starts after the setup or ends before the reaction. Helix reports, for each clip, the past broadcast (VOD) it came from and where in it the clip starts (`video_id`, `vod_offset`). With those, the render host can download a little more of the broadcast around the clip.

On 2026-11-10 the owner asked for an option to extend the length of what is downloaded, chose padding from the VOD, and asked that clpr keep as much of the clip data Helix gives as is useful.

## Decision

1. **Clip data.** The clips table stores the VOD ID and offset Helix reports, at import and on every refresh, since the offset is empty until Twitch processes the VOD and both go empty when the VOD is deleted. The only Helix clip field not stored is `is_featured`, which says the broadcaster pinned the clip on their channel and has no use in clpr.
2. **Padding.** The countdown media worker may download the clip plus up to 30 seconds of the broadcast before it and 30 after it, cut from the VOD with yt-dlp. The amounts are settings, `COUNTDOWN_VOD_PAD_BEFORE_SECONDS` and `COUNTDOWN_VOD_PAD_AFTER_SECONDS`, and both default to 0, so nothing changes until the operator sets them. Padding stops at the start of the broadcast.
3. **Fallback.** A clip whose VOD is unknown, deleted or not downloadable (subscriber-only, for example) is fetched alone, as before.
4. **Scope.** Padding applies to countdown videos and Shorts only. Search transcripts (ADR-003) cover the clip alone, so a snippet matches what the embedded clip plays. Weekly clip packs use the broadcaster's own Helix download and are never padded.
5. **Everything else in ADR-002 holds:** retention under 24 hours, credit and links to the original clip, exclusion on request, takedowns.

## Fair use

Padding moves away from one practice in ADR-002, "short excerpts of 5 to 60 seconds": a padded clip can last up to 120 seconds, and includes footage no viewer clipped. Each item in a countdown is still a short excerpt of a much longer broadcast, still ranked, framed and credited. Keep the padding small; a few seconds on each side gives the setup and the reaction.

## Consequences

- **More Twitch traffic through the unofficial player path.** VOD segments come through the same internal player API as clip files, with the same risk described in ADR-002.
- **Longer videos and Shorts.** A countdown of ten clips with 5 seconds of padding on each side runs about 100 seconds longer. Shorts stay under YouTube's three-minute limit.
- **Verification is looser for padded files.** A padded file is accepted when it lasts at least as long as the clip and at most the clip plus 60 seconds; an unpadded file must still match the clip within 2 seconds.
- **Disclosure.** The terms of service and `docs/compliance/twitch-api-usage.md` describe the VOD source.

## Alternatives considered

- **Look up the VOD at fetch time.** Avoids a schema change, but needs an app token on the render host and loses the data when the crawler already has it. Rejected in favour of storing what Helix returns.
- **Pad search transcripts too.** Rejected: the snippet would quote words the embedded clip does not play.
