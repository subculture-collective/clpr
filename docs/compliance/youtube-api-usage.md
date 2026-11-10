---
title: "YouTube API Usage"
summary: "How the countdown worker uploads to clpr's own YouTube channel: OAuth, endpoints, quota, stored data and privacy until the API audit passes."
tags: ["compliance", "api", "countdowns"]
area: "compliance"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-10
---

# How clpr Uses the YouTube Data API

clpr uploads its daily countdown videos and Shorts to one YouTube channel that clpr owns (#79), and reads the view counts of those videos for the creator dashboard (#84). Nothing reads or writes any other channel, and no viewer signs in with Google.

## Authorization

- **Client:** a Google Cloud project with the YouTube Data API v3 enabled and an OAuth client of type "Desktop app".
- **Scope:** `https://www.googleapis.com/auth/youtube`. Uploads alone need `youtube.upload`, but playlists need the wider scope.
- **Consent:** the channel owner runs `countdown-media youtube-auth -out FILE` on a machine with a browser. It uses a loopback redirect with PKCE and a random state, and writes the refresh token to a file with mode 0600. The token is never printed. Put it in the secret store and give the worker `YOUTUBE_REFRESH_TOKEN_FILE`.
- **Worker settings:** `YOUTUBE_CLIENT_ID`, `YOUTUBE_CLIENT_SECRET` (or `YOUTUBE_CLIENT_SECRET_FILE`) and the refresh token. With none set, publishing is off; with only some set, the worker refuses to start.
- **Revoked or expired consent:** the worker pauses publishing for an hour at a time and logs an error. Editions stay `rendered`, so nothing is lost until the render retention (7 days) runs out.

## Calls

| Call | Use |
|------|-----|
| `POST oauth2.googleapis.com/token` | Refresh the access token, about once an hour |
| `videos.insert` (resumable upload) | Upload the countdown video and each Short |
| `thumbnails.set` | The countdown's custom thumbnail; needs a verified channel, and a failure leaves the video uploaded |
| `playlists.insert` | One playlist per series, created on its first video |
| `playlistItems.insert` | Add each countdown video to its series playlist |
| `videos.list` (`part=statistics,status`) | View count and privacy of clpr's own uploads, 50 videos per call, for the creator dashboard |

`backend/internal/services/youtube_client.go` makes all of them; `backend/internal/services/countdown_youtube.go` decides what to upload.

## Privacy and the audit

YouTube forces videos uploaded by an API project that has not passed its compliance audit to private. `YOUTUBE_PRIVACY` is therefore `private` by default. After the audit passes, set it to `scheduled` (private with a `publishAt` of `YOUTUBE_PUBLISH_HOUR_UTC`, default 15:00 UTC, on the edition date), `public` or `unlisted`. Every upload declares that it is not made for kids.

## Quota

- **Daily limits:** when YouTube answers `quotaExceeded`, `dailyLimitExceeded` or `uploadLimitExceeded`, the worker stops publishing until 08:00 UTC, the latest time YouTube resets daily quota (midnight Pacific).
- **Other temporary errors:** rate limits, 5xx responses and network errors pause publishing for 15 minutes.
- **No duplicates:** an uploaded render is recorded in `countdown_uploads` at once, so a retry does not upload it again.
- **View counts:** the worker reads them once an hour, only for videos uploaded in the last 90 days whose count is more than 6 hours old, 50 videos per `videos.list` call (1 unit). With four series of one countdown and five Shorts a day, that is under 200 units a day. Errors pause the reads the same way they pause publishing.

## What is stored

`countdown_uploads` keeps, per uploaded render, the YouTube video ID, the privacy status and the publish time, and from `videos.list` the public view count, the current privacy status, whether YouTube still returns the video, and when it was read. `countdown_youtube_playlists` keeps one playlist ID per series. No data about YouTube viewers and no YouTube Analytics data is read or stored.

The creator dashboard shows a broadcaster the view counts of the clpr videos their clips appear in, and links only videos that are public or unlisted. The counts are YouTube's public counts from the worker's last read.

## What each video says

The countdown's description links the clpr page of the edition, lists chapters (one per rank, then the credits), and credits every clip with its broadcaster's Twitch channel and the original clip. It also says how to have a clip removed. Each Short credits its broadcaster and links the original clip and the full countdown.

---

**See also:** [[twitch-api-usage|Twitch API Usage]] · [[../features/countdown-editions|Countdown editions]] · [[index|Compliance]]
