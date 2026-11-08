---
title: "Countdown editions"
summary: "Daily ranked selection of Twitch clips for countdown videos, with per-broadcaster caps, one edition per clip per day, and exclusions."
tags: ["features", "countdowns", "compliance"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-07
---

# Countdown editions

A countdown edition is one day's ranked list of clips for a countdown video or Short. This page covers selection (#76) and fetching the media (#77). Rendering (#78) and publishing (#79) build on the editions stored here. The media policy is [[../adr/002-media-handling-for-countdowns|ADR-002]].

## Series

A series is one daily list. Admins configure series under `/api/v1/admin/countdowns/series`:

| Field | Meaning | Default |
| --- | --- | --- |
| `slug`, `name` | Identifier and display name | required |
| `game_id` | Limit to one Twitch category | every category |
| `language` | Limit to one clip language | every language |
| `size` | Clips per edition | 10 |
| `window_hours` | Clips created this many hours before the selection hour are eligible | 24 |
| `per_broadcaster_cap` | Most clips from one broadcaster in an edition | 1 |
| `selection_hour_utc` | Hour (UTC) when the day's edition is selected; the window ends here | 12 |
| `priority` | Lower selects first each day | 100 |
| `enabled` | Disabled series are skipped | true |

## Selection

The playlist script scheduler checks every five minutes, under its existing database lock. For each enabled series in priority order, it selects the most recent edition whose selection hour has passed, unless that edition already exists. A missed day is not backfilled beyond that one date.

A clip is eligible when it:

- is a Twitch clip (not an upload or external link);
- is not removed, hidden, NSFW or removed after a DMCA notice;
- lasts 5 to 60 seconds;
- was created inside the series window and matches its category and language;
- has a view velocity above zero;
- is not excluded, and its broadcaster is not excluded;
- is not already in another edition on the same date.

Eligible clips are ranked by `view_velocity`, then view count. Position 1 is the top clip. The database refuses a second row for the same clip and date (`UNIQUE (edition_date, clip_id)`), so a clip is in at most one edition per day. Series with a lower priority get first pick.

When fewer clips are eligible than the series size, the edition is stored as `failed` with a reason such as "only 6 of 10 clips were eligible" and no clips. It is not retried on its own. An admin can select a past date again with `POST /api/v1/admin/countdowns/series/{id}/editions`, which replaces an edition that is still `selected` or `failed`. Editions that have moved on to `fetching` or later are left alone.

## Media

The `countdown-media` worker (`backend/cmd/countdown-media`, built from `backend/Dockerfile.render`) fetches the clips of each edition onto the render host. It claims the oldest `selected` edition and moves it to `fetching`, fetches and verifies every clip, and then moves it to `fetched`, or to `failed` with the reason. A claim left in `fetching` for an hour, for example by a worker that died, is taken again. [[../compliance/twitch-api-usage#countdown-media|Twitch API usage]] describes the two sources and the checks.

When a clip cannot be fetched or fails verification, it is recorded in `countdown_edition_skips`, the clips below it move up one place, and the next eligible clip is appended, still under the per-broadcaster cap. A skipped clip is left out of every edition that day. After three skips in one edition the worker gives up and fails it, because that many usually means the source changed.

Files live in the worker's volume (`COUNTDOWN_MEDIA_DIR`, default `/var/lib/clpr-countdown-media`). Each cycle deletes files fetched more than 24 hours ago (`COUNTDOWN_MEDIA_RETENTION_HOURS`, at most 24), the files of failed editions, and any leftover file older than that. `countdown_edition_media` keeps the source, SHA-256, size, duration and frame size of each file, and when it was deleted.

| Setting | Default |
| --- | --- |
| `COUNTDOWN_MEDIA_DIR` | `/var/lib/clpr-countdown-media` |
| `COUNTDOWN_MEDIA_INTERVAL_SECONDS` | 60 |
| `COUNTDOWN_MEDIA_RETENTION_HOURS` | 24 (the maximum) |
| `YTDLP_PATH`, `FFPROBE_PATH` | `yt-dlp`, `ffprobe` |
| `COUNTDOWN_MEDIA_METRICS_ADDR` | `:9464`; `off` disables `/metrics` and `/healthz` |

The worker also reads the usual database settings and `TWITCH_CLIENT_ID`; without a client ID it uses only the player path. Metrics: `clpr_countdown_media_fetches_total{source,outcome}`, `clpr_countdown_media_skips_total`, `clpr_countdown_media_editions_total{outcome}`, `clpr_countdown_media_files_deleted_total` and `clpr_countdown_media_last_run_timestamp_seconds`.

## Exclusions

`/api/v1/admin/countdowns/exclusions` holds the broadcasters (by Twitch ID) and single clips left out of every future edition. Each has a reason (`request`, `dmca` or `takedown`) and a source (`admin` now, `self_service` once #75 adds the broadcaster's own toggle). Clips removed after a DMCA notice are already ineligible without a row. A YouTube claim against a clip is recorded as a `takedown` exclusion.

Excluding a broadcaster affects editions selected afterwards. Removing them from videos already published is #75.

## Tests

- `backend/internal/repository/countdown_repository_integration_test.go`: every eligibility rule, ranking, the per-broadcaster cap, cross-series dedupe for the day, short editions stored as failed, reselection and exclusions, against a migrated database.
- `backend/internal/services/countdown_service_test.go`: due dates and windows, skipped series, input validation.
- `backend/internal/repository/countdown_media_repository_integration_test.go`: claiming, reclaiming a stale claim, backfill with the cap and the position shift, skips across series, recording media, and deletion.
- `backend/internal/services/countdown_media_test.go`: the worker's fetch, backfill, give-up and failure paths, the 24-hour sweep, ffprobe parsing, the Twitch URL check, and source choice between Helix and yt-dlp.
- `backend/internal/handlers/countdown_handler_test.go` and `backend/internal/scheduler/playlist_script_scheduler_test.go`: status codes and the scheduler hook.

---

**See also:** [[archive-handoff-import|Archive handoff import]] · [[index|Features]]
