---
title: "Twitch Embeds"
summary: "How clpr embeds Twitch clips, live streams, and chat, and the rules the frontend enforces."
tags: ["compliance"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "2.0"
last_reviewed: 2026-09-25
---

# How clpr Embeds Twitch Content

This page describes how clpr plays Twitch clips, live streams, and chat. It
was checked against the code on 2026-09-25; file paths are relative to the
repository root. The rules come from Twitch's
[video and clip embedding guidelines](https://dev.twitch.tv/docs/embed/video-and-clips/),
the [chat embed documentation](https://dev.twitch.tv/docs/embed/chat/), and
the [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/).

## Summary

- All Twitch playback uses Twitch's own players: the `clips.twitch.tv/embed`
  iframe or the Twitch Embed SDK loaded from `https://embed.twitch.tv/embed/v1.js`.
- Every embed passes the page's hostname as the `parent` parameter.
- A player is created only when its box is at least 400×300 pixels. Smaller
  boxes show a "Watch on Twitch" link instead.
- Floating page elements are hidden, and the consent banner moves into the
  page, while they would cover a player.
- Live streams autoplay muted.
- Feed clips stay thumbnails until the viewer plays them, unless the viewer
  has turned on feed autoplay.

## Players

| Content | Component | Player |
| --- | --- | --- |
| Clips in feeds and cards | `frontend/src/components/clip/TwitchEmbed.tsx` | `https://clips.twitch.tv/embed` iframe |
| Clips on the clip page, in playlists, and in the queue | `frontend/src/components/video/VideoPlayer.tsx` | Twitch Embed SDK with the `clip` option; falls back to the `clips.twitch.tv/embed` iframe if the SDK fails to load or to create the player |
| Live streams | `frontend/src/components/stream/TwitchPlayer.tsx` | Twitch Embed SDK with the `channel` option |
| Live chat | `frontend/src/components/stream/TwitchChatEmbed.tsx` | `https://www.twitch.tv/embed/{channel}/chat` iframe |

### Clip iframe

`TwitchEmbed.tsx` builds the URL from the clip ID and the current hostname:

```typescript
const embedUrl = `https://clips.twitch.tv/embed?clip=${encodeURIComponent(clipId)}&parent=${encodeURIComponent(parentDomain)}&autoplay=${shouldAutoplay ? 'true' : 'false'}&muted=${embedMuted ? 'true' : 'false'}`;
```

The iframe has `allowFullScreen` and `allow="autoplay; fullscreen"`, and its
`title` is the clip title. Nothing is drawn over it. The "Sound off · Turn on
for future clips" control sits below the player box, not on it.

### Live stream player

`TwitchPlayer.tsx` loads the SDK once from Twitch's CDN and reuses an
existing script tag if another component already added it. It creates the
player only after `fetchStreamStatus` reports the channel as live; offline
channels get `StreamOfflineScreen` with a link to the channel on Twitch. The
status is refetched every 60 seconds.

```typescript
new window.Twitch.Embed(container.id, {
  width: '100%',
  height: '100%',
  channel: channel,
  layout: showChat ? 'video-with-chat' : 'video',
  autoplay: true,
  // Autoplay is always muted; viewers unmute with Twitch's own control.
  muted: true,
  parent: [parentDomain],
});
```

The stream page (`frontend/src/pages/StreamPage.tsx`) uses the `video`
layout and shows chat in a separate `TwitchChatEmbed` beside or below the
player. The live badge and viewer count (`LiveIndicator`) are rendered in the
page heading, outside the player.

### Chat

`TwitchChatEmbed.tsx` frames
`https://www.twitch.tv/embed/{channel}/chat?parent={hostname}&darkpopout`.
The viewer can hide chat or move it below the player. Signed-in clpr users
who have not connected Twitch see a "Login to Chat" button, which starts the
channel connection flow described in
[How clpr uses the Twitch API](twitch-api-usage.md#connecting-a-twitch-channel).

## Size

Twitch requires clip, video, and live embeds to be at least 400×300 pixels.
`frontend/src/hooks/useTwitchEmbedFits.ts` measures each player's box with a
`ResizeObserver` and exports the limits as `TWITCH_EMBED_MIN_WIDTH = 400` and
`TWITCH_EMBED_MIN_HEIGHT = 300`.

`TwitchEmbed`, `VideoPlayer`, and `TwitchPlayer` create a player only while
the box meets both limits. Below that, they show "Not enough room for the
Twitch player." with a "Watch on Twitch" link to the clip or channel. If the
box is resized below the limit, the player is removed.

Feed clip boxes use a 4:3 aspect ratio on small screens and 16:9 from the
`md` breakpoint. On small screens the clip thumbnail shows a "Landscape
recommended" hint.

## Nothing covers a player

Twitch does not allow page elements over an embed.
`frontend/src/hooks/useTwitchPlayerLayer.ts` keeps a registry of mounted
players. Each player component registers its box while the embed is shown;
`TwitchChatEmbed` registers the chat frame too.

Floating elements call `useOverlapsTwitchPlayer`, which reports when the
element comes within 8 pixels of a registered player. The check runs on
scroll, resize, element resize, and every 500 ms. While it reports an overlap:

- `ScrollToTop`, `MiniFooter`, `OfflineIndicator`, and the collapsed
  `QueueWidget` set `visibility: hidden`.
- `ConsentBanner` stops floating over the page and renders in its slot in the
  page flow.

A player inside the floating element itself, such as the queue miniplayer,
does not count as covered.

## Feed playback

`frontend/src/components/clip/ClipFeed.tsx` tracks one `activeClipId`. Every
other clip is a thumbnail button, so at most one feed iframe is mounted.
Activating another clip, or scrolling the active clip out of view, unmounts
the previous iframe.

By default a clip plays only when the viewer clicks it. The feed header offers
an autoplay setting, stored in `localStorage` by
`frontend/src/hooks/useFeedAutoplayPreference.ts`. With autoplay on, the clip
that scrolls into view becomes active and autoplays. Clip embeds start muted
until the viewer chooses "Turn on for future clips", which is stored by
`frontend/src/hooks/useVolumePreference.ts`.

## Unavailable clips

If a clip iframe fires an error event, `TwitchEmbed` shows "This clip is
unavailable here." with a "Watch on Twitch" link to
`https://clips.twitch.tv/{clipId}`. clpr has no copy of the video to show in
place of a clip removed from Twitch: clips submitted from Twitch are stored
with an embed URL and no video URL
(`backend/internal/services/submission_service.go`, the `SourceTypeTwitch`
case).

## Attribution

Clip cards and the clip page show the broadcaster, the Twitch category, and
the person who clipped or submitted the clip. These names link to clpr's
broadcaster, category, and profile pages. The Twitch player keeps its own
branding and controls, and every size or error fallback links to the clip or
channel on Twitch.

## Content Security Policy

`Caddyfile` and `frontend/nginx.conf` send the same `frame-src` and
`script-src` sources for Twitch:

```text
frame-src 'self' https://clips.twitch.tv https://player.twitch.tv https://embed.twitch.tv https://www.twitch.tv
script-src 'self' https://embed.twitch.tv (plus analytics sources)
```

`scripts/test-edge-contracts.sh` fails if any of the four Twitch frame
sources is missing from either file.

## Tests

- `frontend/src/components/clip/TwitchEmbed.test.tsx`: thumbnail until
  activated, iframe unmounted when deactivated, sound control outside the
  player, link to Twitch below 400×300.
- `frontend/src/components/stream/TwitchPlayer.test.tsx`: no embed for an
  offline channel, muted autoplay with nothing over the player, link to Twitch
  below 400×300.
- `frontend/src/components/stream/TwitchChatEmbed.test.tsx`: chat framed from
  `www.twitch.tv` with this site as `parent`.
- `frontend/src/hooks/useTwitchPlayerLayer.test.tsx`: overlap detection,
  clearance, and the miniplayer exception.
- `frontend/src/components/video/VideoPlayer.test.tsx`: iframe keeps `parent`
  and the mute preference, sound control below the player, link to Twitch
  below 400×300, fallback to the iframe when the SDK fails.

## References

- [Embedding video and clips](https://dev.twitch.tv/docs/embed/video-and-clips/)
- [Embedding chat](https://dev.twitch.tv/docs/embed/chat/)
- [Twitch embed documentation](https://dev.twitch.tv/docs/embed/)
- [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
- [How clpr uses the Twitch API](twitch-api-usage.md)
