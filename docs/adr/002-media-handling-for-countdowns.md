---
title: "002 Media Handling For Countdowns"
summary: "Clips for countdown videos and Shorts are downloaded and edited under a fair use position, without asking each broadcaster first."
tags: ["adr", "compliance", "media"]
area: "adr"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-07
---

# ADR-002: Media handling for countdown videos and Shorts

**Status**: Accepted when the pull request that adds it is merged
**Date**: 2026-11-07
**Decision maker**: the owner (issue #74)
**Related issues**: #74; changes the scope of #75, #76 and #77
**Extended by**: [ADR-003](003-transcribing-clips-for-search.md), which also allows downloading public clips to transcribe them for search, and [ADR-004](004-padding-countdown-clips-from-the-vod.md), which allows padding countdown clips from the VOD

## Context

The second milestone publishes daily countdown videos and Shorts on YouTube, built from Twitch clips (#76 to #82). Rendering needs the clip files. Until now the guardrails forbade downloading Twitch video at all, and #74 proposed one narrow exception: Helix `GET /clips/downloads`, only for broadcasters who granted `channel:manage:clips` and agreed to republishing.

That route needs every featured broadcaster to sign in and opt in first. No broadcaster has granted the scope so far, so a countdown limited to consenting broadcasters would have nothing to show.

## Decision

The owner decided on 2026-11-07:

> If it's more efficient, we should just use ffmpeg to download what we want. I am going to operate under the idea that it is fair use, so we don't have to ask for permission.

So:

1. **Source.** Countdowns and Shorts may use any public Twitch clip that passes the edition rules in #76. Broadcaster consent is not a condition.
2. **Fetching.** The render host downloads the selected clip files with yt-dlp and ffmpeg, from the same playback files Twitch serves to its own player. Helix `GET /clips/downloads` is still used where a broadcaster has granted `channel:manage:clips`. Getting a playback URL without that grant goes through Twitch's internal player API (`gql.twitch.tv`). This is the only use of that API the guardrails allow.
3. **Retention.** Files are kept on the render host only until the render finishes, and deleted within 24 hours whether or not it succeeds. They are not served from clpr.tv, and users get no download button.
4. **clpr.tv itself** keeps playing Twitch clips through Twitch's embed player. Countdown pages (#82) embed the original clips, not the rendered video.

## Fair use position and the practices that support it

Fair use is a defense that is weighed case by case, not a permission. These practices make the countdowns more transformative and lower the chance and cost of a dispute. They are part of the decision, not optional:

- **Transform the material.** Each edition adds ranking, title cards, lower-thirds and an outro. Clips are used as an item in a list, not reposted on their own. Shorts carry captions and clpr framing.
- **Credit and link.** Every video names each broadcaster on screen and in the description, with a link to their channel and to the original clip.
- **Short excerpts.** Editions use clips of 5 to 60 seconds (#76 eligibility).
- **Exclusion on request.** A broadcaster who asks to be left out is excluded from future editions, and their clip is removed or the video is taken down within 7 days. #75 becomes the self-service way to ask, and to opt in to the creator features in #84 to #86.
- **Takedowns.** A DMCA notice or a YouTube copyright claim removes the clip from future editions, and the affected video is edited or taken down. The existing DMCA flow applies.
- **No bare re-uploads.** A Short shows one clip with clpr captions and framing, never the clip file on its own.

## Consequences

The decision accepts these risks:

- **Twitch API access.** The Twitch Developer Services Agreement restricts downloading content and using undocumented endpoints. If Twitch objects, it can revoke clpr's application credentials. Those credentials also run the catalog crawler, so the whole site would stop getting new clips, not only the countdowns. Running the downloader on a separate host and without clpr's client ID limits how directly it is tied to the application, but does not remove the risk.
- **YouTube strikes.** Broadcasters or rights holders, for example of music playing on stream, can file claims. Three copyright strikes terminate a YouTube channel. Content ID claims can demonetize or block a video without a strike.
- **Breakage.** The internal player API is undocumented and can change without notice. The downloader must fail visibly (#80 alerts) instead of quietly producing empty editions.
- **Disclosure.** The terms of service and the public Twitch API usage page must describe this honestly. This change updates the terms. `docs/compliance/twitch-api-usage.md` is updated when the downloader lands (#77).

## Alternatives considered

- **Helix download with broadcaster consent only** (#74 as filed). It needs no unofficial API and is clearly allowed for authorized clips, but no broadcaster has opted in, so editions would be empty for the foreseeable future. Rejected for now. The Helix path stays for broadcasters who connect.
- **Embed-only countdown pages, no video.** No media handling at all, but no YouTube distribution, which is the point of the milestone. Rejected.

## Follow-up scope changes

- **#75:** consent is no longer required for inclusion. It becomes the opt-out and opt-in record for creator features.
- **#76:** eligibility drops "consenting broadcasters only" and adds "not excluded by request".
- **#77:** fetch through yt-dlp and ffmpeg, with Helix where authorized; checksum, ffprobe verification and deletion as already planned.
