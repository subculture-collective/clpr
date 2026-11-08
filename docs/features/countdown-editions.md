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

A countdown edition is one day's ranked list of clips for a countdown video or Short. This page covers selection (#76), fetching the media (#77) and rendering (#78). Publishing (#79) builds on the renders stored here. The media policy is [[../adr/002-media-handling-for-countdowns|ADR-002]].

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
| `shorts_count` | How many top clips of each edition also get a 9:16 Short, at most `size` | 0 |

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
| `COUNTDOWN_RENDER_DIR` | `/var/lib/clpr-countdown-renders` |
| `COUNTDOWN_RENDER_RETENTION_DAYS` | 7 |
| `COUNTDOWN_RENDER_TEMPLATE` | the built-in template |
| `FFMPEG_PATH` | `ffmpeg` |

The worker also reads the usual database settings and `TWITCH_CLIENT_ID`; without a client ID it uses only the player path. Metrics: `clpr_countdown_media_fetches_total{source,outcome}`, `clpr_countdown_media_skips_total`, `clpr_countdown_media_editions_total{outcome}`, `clpr_countdown_media_files_deleted_total`, `clpr_countdown_render_seconds{kind}`, `clpr_countdown_render_editions_total{outcome}`, `clpr_countdown_render_files_deleted_total` and `clpr_countdown_media_last_run_timestamp_seconds`, which covers both stages.

## Rendering

The same worker then renders each `fetched` edition (`backend/internal/countdownrender`, `CountdownRenderWorker`). It moves the edition to `rendering`, and then to `rendered`, or to `failed` with the reason. A claim left in `rendering` for an hour is taken again.

- **Countdown video**, 1920×1080, 30 fps H.264 and 48 kHz stereo AAC. For each clip from the lowest rank to #1, it shows a rank card (rank, title, broadcaster), then the clip, scaled and padded to 16:9, with a lower third (rank, broadcaster, title) for its first seconds and the site label in the corner. An outro lists every clip and says that each clip belongs to its broadcaster. Audio is normalized to the template loudness.
- **Thumbnail**, 1280×720 JPEG: the middle frame of the #1 clip with "TOP N", the series name and the date.
- **Shorts**, 1080×1920: one for each of the top `shorts_count` clips. The clip sits over a blurred, cropped copy of itself, with its title at the top and the broadcaster and site label below. English clips get captions (#81), below a landscape clip or over a portrait one.

Layout, colours, fonts, timings and encoder settings are data: `backend/internal/countdownrender/template.json`, overridden field by field by the file in `COUNTDOWN_RENDER_TEMPLATE`. Titles reach ffmpeg through text files rather than filter-graph escaping. Emoji and other characters the bundled Noto Sans CJK font cannot draw are dropped, and long titles are wrapped or cut with an ellipsis.

The source clips are deleted as soon as the edition finishes rendering, whether it succeeded or not. Intermediate segments hold clip footage, so they are written under the media directory, where the 24-hour sweep covers anything a crash leaves behind. Rendered files go to `COUNTDOWN_RENDER_DIR` and `countdown_renders` (kind, SHA-256, size, duration). They are deleted after `COUNTDOWN_RENDER_RETENTION_DAYS`, or at once when the edition fails.

### Captions

Before rendering a Short, the worker transcribes its clip with the faster-whisper runner the API uses for transcription (`backend/whisper/whisper_runner.py --words`), with word timings. Lines of at most `caption_max_chars` columns (template, default 22) show from their first word until the next line starts. Segments Whisper marks as probably not speech (no-speech probability above 0.6, or average log-probability below -1.0) are left out, and clips Whisper does not detect as English get no captions. A failed transcription renders the Short without captions and counts `clpr_countdown_short_captions_total{outcome="failed"}`; the other outcomes are `captioned` and `none`. The temporary WAV is written under the media directory and deleted at once.

The render image sets `WHISPER_PYTHON_PATH` and `WHISPER_RUNNER_DIR`, so captions are on by default; `COUNTDOWN_CAPTIONS=off` turns them off. `WHISPER_MODEL` picks the model (default `base`), which is downloaded on first use to `/var/lib/clpr-whisper-models`. A 12-second clip took 7 seconds to transcribe on a 4-core host, model download included.

Two renders of the same inputs are byte-identical. On a 4-core host, a 10-clip edition of 30-second 1080p60 clips renders in about 2.5 minutes, and a 30-second Short in under 30 seconds.

## Exclusions

`/api/v1/admin/countdowns/exclusions` holds the broadcasters (by Twitch ID) and single clips left out of every future edition. Each has a reason (`request`, `dmca` or `takedown`) and a source (`admin` now, `self_service` once #75 adds the broadcaster's own toggle). Clips removed after a DMCA notice are already ineligible without a row. A YouTube claim against a clip is recorded as a `takedown` exclusion.

Excluding a broadcaster affects editions selected afterwards. Removing them from videos already published is #75.

## Tests

- `backend/internal/repository/countdown_repository_integration_test.go`: every eligibility rule, ranking, the per-broadcaster cap, cross-series dedupe for the day, short editions stored as failed, reselection and exclusions, against a migrated database.
- `backend/internal/services/countdown_service_test.go`: due dates and windows, skipped series, input validation.
- `backend/internal/repository/countdown_media_repository_integration_test.go`: claiming, reclaiming a stale claim, backfill with the cap and the position shift, skips across series, recording media, and deletion.
- `backend/internal/services/countdown_media_test.go`: the worker's fetch, backfill, give-up and failure paths, the 24-hour sweep, ffprobe parsing, the Twitch URL check, and source choice between Helix and yt-dlp.
- `backend/internal/countdownrender/render_test.go`: golden frames of the countdown (cards, clips with and without the lower third, a portrait clip, the outro), both Shorts, a captioned Short (shown while its words are spoken, gone after) and the thumbnail; stream formats and durations; byte-identical output across runs; and, with `COUNTDOWN_RENDER_TIMING=1`, a full-size timing run. These tests need ffmpeg and the fonts, so `backend/Dockerfile.render` runs them while building the image. To regenerate the goldens, run the test binary in the image's `runtime` stage with `UPDATE_GOLDEN=1` and the source `testdata` directory mounted.
- `backend/internal/countdownrender/text_test.go` and `backend/internal/services/countdown_captions_test.go`: grouping words into caption lines; keeping only confident English speech; Shorts rendered without captions when transcription fails.
- `backend/internal/repository/countdown_render_repository_integration_test.go` and `backend/internal/services/countdown_render_worker_test.go`: claiming, recording and finishing renders, source deletion on success and failure, missing media, shutdown, and the render sweep.
- `backend/internal/handlers/countdown_handler_test.go` and `backend/internal/scheduler/playlist_script_scheduler_test.go`: status codes and the scheduler hook.

---

**See also:** [[archive-handoff-import|Archive handoff import]] · [[index|Features]]
