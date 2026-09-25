---
title: "Twitch API Usage"
summary: "Which Twitch APIs clpr calls, with which tokens and scopes, and how requests are limited, retried, and cached."
tags: ["compliance","api"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "2.0"
last_reviewed: 2026-09-25
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
| `GET /moderation/banned` | `GetBannedUsers` | Broadcaster | `internal/services/twitch_ban_sync_service.go` | Sync a broadcaster's Twitch bans into clpr (`POST /api/v1/moderation/sync-bans`) |
| `POST /moderation/bans` | `BanUser` | Broadcaster | `internal/services/twitch_moderation_service.go` | Ban a user on the broadcaster's channel (`POST /api/v1/moderation/twitch/ban`) |
| `DELETE /moderation/bans` | `UnbanUser` | Broadcaster | `internal/services/twitch_moderation_service.go` | Unban a user (`DELETE /api/v1/moderation/twitch/ban`) |
| `GET /clips/downloads` | `GetDownloadURL` (`backend/internal/services/twitch_clip_download_service.go`) | Broadcaster | `internal/services/clip_transcription_service.go` | Optional clip transcription, described below |
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
| `chat:read`, `chat:edit` | Chat functionality | No current server code sends or reads chat with the user's token |
| `channel:bot` | Let the clpr bot account join and read this broadcaster's chat | Streamer clip room, below |
| `channel:manage:clips` | Official temporary download URLs for this broadcaster's clips | Optional transcription, below; also required to start the clip room |
| `moderator:manage:banned_users` | Moderators banning and unbanning users | Requested; ban actions are currently limited to broadcasters |
| `channel:manage:banned_users` | Broadcasters banning and unbanning users | Ban sync, ban, and unban |

The callback exchanges the code, reads the account from `GET /users`, and
stores the access token, refresh token, granted scopes, and expiry in the
`twitch_auth` table. `GET /api/v1/twitch/auth/status` refreshes an expired
token with the refresh token grant.

`DELETE /api/v1/twitch/auth` deletes the stored tokens. clpr does not call
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

clpr does not copy Twitch clip video. Clips submitted from Twitch are stored
with a `clips.twitch.tv/embed` URL and no video URL
(`backend/internal/services/submission_service.go`, the `SourceTypeTwitch`
case), and playback uses Twitch's players. The only server-side media access
is the optional transcription described above.

## References

- [Twitch API reference](https://dev.twitch.tv/docs/api/reference/)
- [Twitch API guide](https://dev.twitch.tv/docs/api/guide/)
- [Authentication](https://dev.twitch.tv/docs/authentication/)
- [Scopes](https://dev.twitch.tv/docs/authentication/scopes/)
- [Chat over IRC](https://dev.twitch.tv/docs/chat/irc/)
- [Twitch Developer Services Agreement](https://legal.twitch.com/legal/developer-agreement/)
