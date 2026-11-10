---
title: "Twitch API Usage"
summary: "Which Twitch APIs clpr calls, with which tokens and scopes, and how requests are limited, retried, and cached."
tags: ["compliance","api"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "2.1"
last_reviewed: 2026-11-07
---

# How clpr Uses the Twitch API

This page lists the Twitch APIs the clpr backend calls and how it
authenticates, limits, retries, and caches those calls. It was checked against
the code on 2026-09-25; file paths are relative to the repository root.
Playback is covered separately in [How clpr embeds Twitch content](twitch-embeds.md).

## Summary

- All Twitch data comes from the Helix API (`https://api.twitch.tv/helix`),
  Twitch OAuth (`https://id.twitch.tv/oauth2`), and, for one opt-in feature,
  Twitch chat over IRC.
- Public data is read with an app access token. User tokens are used only for
  actions a broadcaster has authorized on their own channel.
- Sign-in requests one scope, `user:read:email`.
- App-token requests share a limit of 800 per minute per API process.
- The countdown media worker on the render host downloads the clips of each
  countdown edition, and popular public clips to transcribe for search, through
  Helix where the broadcaster authorized it and otherwise from Twitch's player
  with yt-dlp. See [Countdown media](#countdown-media).

## Helix endpoints

Most calls go through the client in `backend/pkg/twitch/`. The table lists
every Helix endpoint that production code calls, the client function, and
where it is called from.

| Endpoint | Function | Token | Called from | Purpose |
| --- | --- | --- | --- | --- |
| `GET /clips` | `GetClips` (`backend/pkg/twitch/endpoints.go`) | App | `internal/services/clip_sync_service.go`, `internal/services/submission_service.go`, `internal/scheduler/engagement_scheduler.go` | Import clips by category, broadcaster, or ID; validate submitted clip URLs; refresh view counts |
| `GET /games` | `GetGames` | App | `internal/services/submission_service.go` | Resolve the category name for a submitted clip |
| `GET /games/top` | `GetTopGames` | App | `internal/services/clip_sync_service.go`, `internal/services/playlist_strategies.go` | Choose categories for trending clip imports and generated playlists |
| `GET /streams` | `GetStreams`, `GetStreamStatusByUsername` | App | `internal/services/live_status_service.go`, `internal/handlers/stream_handler.go` | Live status for broadcasters and the stream page |
| `GET /users` | `GetUsers`, `GetStreamStatusByUsername` | App | `internal/handlers/broadcaster_handler.go`, `internal/handlers/stream_handler.go` | Broadcaster display name, avatar, and description; resolve a login to a user ID |
| `GET /channels` | `GetChannels` | App | `internal/services/clip_sync_service.go` | Channel tags applied to imported clips |
| `GET /clips/downloads` | `GetDownloadURL` (`backend/internal/services/twitch_clip_download_service.go`) | Broadcaster | `internal/services/clip_transcription_service.go`, `internal/services/countdown_media.go` | Optional clip transcription and countdown media, described below |
| `GET /users` | `fetchTwitchUser` (`backend/internal/services/auth_service.go`) and `TwitchOAuthCallback` (`backend/internal/handlers/twitch_oauth_handler.go`) | User | Sign-in and channel connection | Identify the Twitch account that just authorized clpr |

Backend paths in the "Called from" column are relative to `backend/`.

`GetClips` is called with an app token, so it only returns clips Twitch makes
public. `GetVideos`, `GetChannelFollowers`, and `GetUser` exist in
`endpoints.go` but no production code calls them.

Twitch ban actions are limited to the broadcaster's own channel. Twitch
moderators of that channel cannot use them yet
(`backend/internal/services/twitch_moderation_service.go`, `ValidateTwitchBanScope`).

## Sign-in

"Continue with Twitch" uses the authorization code flow in
`backend/internal/services/auth_service.go`. It requests only
`user:read:email`. The OAuth `state` is stored in Redis for 5 minutes. The
backend exchanges the code, reads the user from `GET /users`, and then issues
clpr's own session tokens. The Twitch access token from sign-in is not stored.

## Connecting a Twitch channel

A signed-in user can separately connect their Twitch channel through
`GET /api/v1/twitch/oauth/authorize`
(`backend/internal/handlers/twitch_oauth_handler.go`, `InitiateTwitchOAuth`).
The `state` value is signed and bound to the clpr user. This flow requests:

| Scope | Purpose stated in the code | Used by |
| --- | --- | --- |
| `channel:bot` | Let the clpr bot account join and read this broadcaster's chat | Streamer clip room, below |
| `channel:manage:clips` | Official temporary download URLs for this broadcaster's clips | Optional transcription, below; also required to start the clip room, and the broadcaster's "feature my clips" opt-in and weekly clip packs |

The callback exchanges the code, reads the account from `GET /users`, and
stores the access token, refresh token, granted scopes, and expiry in the
`twitch_auth` table. `GET /api/v1/twitch/auth/status` refreshes an expired
token with the refresh token grant.

"Feature my clips" in settings starts the same flow with `purpose=feature`. It asks for
`channel:manage:clips` only, plus any scope the broadcaster granted before, because a new
token replaces the old one. When the grant includes the scope, the callback records an
opt-in in `broadcaster_feature_optins` (broadcaster, user, scope, consent text version,
time). Countdown media for that broadcaster then comes through `GET /clips/downloads`.

`DELETE /api/v1/twitch/auth` deletes the stored tokens and the opt-in. clpr does not call
Twitch's token revocation endpoint. Users can also disconnect clpr in their
Twitch connection settings, after which the stored tokens stop working.

See Twitch's [scope reference](https://dev.twitch.tv/docs/authentication/scopes/).

## Streamer clip room

A broadcaster who has connected their channel with `channel:bot` and
`channel:manage:clips` can start a clip room for that channel only
(`backend/internal/handlers/streamer_clip_room_handler.go`, `StartRoom`).
clpr then connects its bot account, configured by `TWITCH_BOT_USERNAME` and
`TWITCH_BOT_OAUTH_TOKEN`, to `irc.chat.twitch.tv:6697` over TLS and joins
that channel (`backend/internal/services/twitch_chat_listener.go`). The
listener reads chat messages and does not post to chat.

## Optional clip transcription

Transcription is off unless the deployment sets `WHISPER_ENABLED=true` and
configures the runner (`backend/cmd/api/services.go`). When it is on, it
applies only to clips whose broadcaster has connected clpr with
`channel:manage:clips` and has an unexpired token
(`backend/internal/services/clip_transcription_service.go`, `TranscribeClip`).
clpr asks `GET /clips/downloads` for the clip's official download URL,
accepts only HTTPS URLs on Twitch media hosts, streams the audio into a
temporary WAV file, transcribes it, and deletes the file. The transcript is
stored with the clip and used by the auto-tagging job
(`backend/internal/scheduler/auto_tag_scheduler.go`). If AI tagging is also
enabled (`VISION_ENABLED=true`), the transcript is sent with the clip's
metadata and public thumbnail to the configured AI provider.

## Countdown media

Countdown videos and Shorts are rendered from clip files, under the media
policy in [ADR-002](../adr/002-media-handling-for-countdowns.md). The
`countdown-media` worker (`backend/cmd/countdown-media`, image
`backend/Dockerfile.render`) runs on the render host beside the database, not
in the API process. For each selected edition it fetches every clip
(`backend/internal/services/countdown_media.go`, `TwitchClipMediaFetcher`):

- **Helix, when authorized.** If the broadcaster connected clpr with
  `channel:manage:clips` and the token has not expired, it asks
  `GET /clips/downloads` for the official download URL with the broadcaster's
  token, as transcription does, and downloads the file over HTTPS.
- **Twitch's player otherwise.** It runs `yt-dlp --no-config` on the clip's
  `clips.twitch.tv` or `twitch.tv/<channel>/clip/` URL. yt-dlp reads the
  playback files Twitch serves to its own player through Twitch's internal
  player API (`gql.twitch.tv`), using yt-dlp's own client settings. clpr's
  client ID and tokens are never sent to that API. This is the only use of
  the internal API.

Each file is checked with ffprobe (a video stream, and a duration within 2 s
of the clip's), hashed with SHA-256 and recorded in
`countdown_edition_media`. A clip that fails is skipped and the next eligible
clip takes its place; after three skips in one edition, or if yt-dlp or
ffprobe is missing, the edition is marked failed instead of going out short.
Files stay in the worker's own volume, are never served by clpr.tv or put in
object storage, and are deleted within 24 hours of fetching, or as soon as
the edition fails.

The worker also transcribes each fetched clip on the render host
(`backend/internal/services/countdown_captions.go`): to burn captions into the
Shorts, and to store the English transcript for transcript search (#87). It
uses the same local faster-whisper runner as the optional transcription above,
writes a temporary WAV under the media volume and deletes it at once. No audio
or video leaves the render host. The transcript text is stored in
`clip_transcripts` and indexed for search, and only English speech is kept.

Under [ADR-003](../adr/003-transcribing-clips-for-search.md) the worker also
downloads public clips for transcript search alone
(`backend/internal/services/search_transcription_worker.go`): clips with at
least 1,000 views from the last 30 days, most viewed first, up to 500 a day,
leaving out clips and broadcasters on the countdown exclusion list. It fetches
each clip the same way as countdown media, transcribes it and deletes the file
before fetching the next. Only the transcript is kept, and full transcripts
are never sent to clients. When a clip or broadcaster is excluded, the worker
empties the render host's transcripts of their clips.

The same worker builds weekly clip packs for broadcasters who feature their
clips (`backend/internal/services/clip_pack_worker.go`). A pack is made from
the broadcaster's own clips and fetched only through Helix with their own
`channel:manage:clips` token; a clip without a working Helix download is left
out, and the player fallback is never used. The downloaded clips are deleted
as soon as the pack is built. The zip is served only to the broadcaster, by a
signed link that expires within a day, and is deleted 7 days after it is made.

## App access token

`backend/pkg/twitch/auth.go` obtains the app token with the
`client_credentials` grant (`RefreshToken`). The token is kept in memory and
in Redis under `twitch:access_token`, and it is replaced 5 minutes before the
expiry Twitch reports. If a Helix request returns 401, `doRequest` in
`backend/pkg/twitch/client.go` fetches a new app token and retries.

## Rate limiting

`backend/pkg/twitch/ratelimit.go` (`RateLimiter`) allows 800 app-token
requests per minute in each API process. The allowance resets to 800 one
minute after the previous reset; when it runs out, requests wait for the next
reset. Waiting respects request cancellation.

Ban and unban calls are also limited to 100 per channel per minute
(`ChannelRateLimiter`, `backend/pkg/twitch/client.go` `NewClient`).

See Twitch's [rate limit documentation](https://dev.twitch.tv/docs/api/guide/).

## Retries and circuit breaker

`doRequest` in `backend/pkg/twitch/client.go` makes up to three attempts:

| Response | Behavior |
| --- | --- |
| Network error, 502, 503, 504 | Retry after 1 s, then 2 s |
| 429 | Retry after 1 s, then 2 s; then return a rate-limit error |
| 401 | Get a new app token and retry |
| 404 and other statuses | Return the response to the caller without retrying |

The circuit breaker in the same file opens after 5 consecutive failures and
rejects requests for 30 seconds. After that, requests are allowed again; a
success closes the breaker and another failure reopens it.

`BanUser` and `UnbanUser` use their own retry loop with jittered exponential
backoff capped at 10 seconds.

## Caching

The client caches some responses in Redis (`backend/pkg/twitch/endpoints.go`):

| Data | Key | TTL | Read back before calling Twitch |
| --- | --- | --- | --- |
| App access token | `twitch:access_token` | Until 5 minutes before expiry | Yes |
| Streams by user IDs | `twitch:streams:{sorted ids}` | 30 seconds | Yes |
| Stream status by login | `twitch:stream_status:{login}` | 60 seconds | Yes |
| Channels by broadcaster IDs | `twitch:channels:{sorted ids}` | 1 hour | Yes |
| Users | `twitch:user:{id}` | 1 hour | No; `GetUsers` always calls Twitch |
| Games | `twitch:game:{id}` | 4 hours | No; `GetGames` always calls Twitch |

## Clip video

clpr.tv does not serve Twitch clip video. Clips submitted from Twitch are
stored with a `clips.twitch.tv/embed` URL and no video URL
(`backend/internal/services/submission_service.go`, the `SourceTypeTwitch`
case), and playback uses Twitch's players. Server-side media access is
limited to the optional transcription and the countdown media worker, both
described above.

## References

- [Twitch API reference](https://dev.twitch.tv/docs/api/reference/)
- [Twitch API guide](https://dev.twitch.tv/docs/api/guide/)
- [Authentication](https://dev.twitch.tv/docs/authentication/)
- [Scopes](https://dev.twitch.tv/docs/authentication/scopes/)
- [Chat over IRC](https://dev.twitch.tv/docs/chat/irc/)
- [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
