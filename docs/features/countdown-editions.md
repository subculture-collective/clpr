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

A countdown edition is one day's ranked list of clips for a countdown video or Short. This page covers selection (#76), fetching the media (#77), rendering (#78), the public pages (#82), publishing to YouTube (#79) and the creator dashboard (#84). The media policy is [[../adr/002-media-handling-for-countdowns|ADR-002]].

## Series

A series is one daily list. Admins configure series under `/api/v1/admin/countdowns/series`:

| Field | Meaning | Default |
| --- | --- | --- |
| `slug`, `name` | Identifier and display name | required |
| `game_id` | Limit to one Twitch category | every category |
| `language` | Limit to one clip language | every language |
| `size` | Most clips in an edition | 10 |
| `min_size` | Fewest clips for an edition. Below it the day is `skipped`. Unset means exactly `size` clips or `failed` | unset |
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

A series with `min_size` takes up to `size` clips and stores a `skipped` edition, which raises no alert, when fewer than `min_size` are eligible. Without `min_size`, when fewer clips are eligible than the series size, the edition is stored as `failed` with a reason such as "only 6 of 10 clips were eligible" and no clips. It is not retried on its own. An admin can select a past date again with `POST /api/v1/admin/countdowns/series/{id}/editions`, which replaces an edition that is still `selected` or `failed`. Editions that have moved on to `fetching` or later are left alone.

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

## Public pages

An edition becomes public once it is `rendered` or `published`. Selected, in-progress and failed editions stay admin-only.

- `/countdowns` lists the public editions by day, newest first, 14 days at a time.
- `/countdowns/<date>/<series>` (for example `/countdowns/2026-11-07/overall`) plays one edition as a playlist, from the lowest rank to #1: one Twitch player, previous and next buttons, and the ranked list beside it. Opening it from `/countdowns` starts playing on screens at least 768 pixels wide; a shared link waits for a press. A Twitch clip embed cannot report its end, so the page moves on three seconds after the clip's duration, and only while the player is on screen.
- Clip pages and broadcaster pages show "Featured in countdown" with links to the editions the clip, or the broadcaster's clips, placed in.

The pages read `GET /api/v1/countdowns`, `GET /api/v1/countdowns/{date}/{series}` and `GET /api/v1/countdowns/featured`. They need no sign-in, are rate-limited to 120 requests a minute per client and may be cached for five minutes. A clip removed, hidden, taken down or excluded after its edition was made drops out of these pages, so the ranks shown can skip a number. The sitemap lists `/countdowns` and the newest 1,000 edition pages. The YouTube link joins each page once #79 stores the video ID.

When an edition is rendered, the worker also queues the `countdown.published` webhook event, which Discord subscriptions post to their channel (see [[../backend/webhook-subscription-management#discord-delivery|Discord delivery]]). The links in it use `BASE_URL`, so the worker needs the same value as the API.

## Publishing

Once the channel is authorized (see [[../compliance/youtube-api-usage|YouTube API usage]]), the same worker uploads each `rendered` edition. It moves the edition to `publishing`, and then:

1. uploads the countdown video, sets its thumbnail and adds it to the series playlist, creating the playlist on first use;
2. uploads each Short;
3. moves the edition to `published`.

Every upload is recorded in `countdown_uploads` as soon as it finishes, so a retry skips it.

- **Errors:** a quota error returns the edition to `rendered` and pauses publishing until 08:00 UTC; temporary errors pause it for 15 minutes, and an authorization error for an hour. An error YouTube will repeat, such as rejected metadata, fails the edition with the reason. A failed thumbnail or playlist step leaves the video published and counts a failure.
- **Privacy:** uploads are private until the API audit passes; `YOUTUBE_PRIVACY` then switches to `scheduled`, `public` or `unlisted`.
- **Metrics:** `clpr_countdown_youtube_uploads_total{kind,outcome}` and `clpr_countdown_publish_editions_total{outcome}`.

**Titles and descriptions:**

- The countdown is titled "Top N Twitch Clips of <date>: <series>".
- Its description links the edition page and lists chapters: one per rank, timed from the rank card length and each clip's fetched duration, then the credits. It credits every clip's broadcaster and original clip, and says how to have a clip removed.
- Shorts are titled with the clip title and `#Shorts`.
- Angle brackets, which YouTube rejects, become ‹ and ›.
- Titles stop at 100 characters, descriptions at 5,000 bytes and tags at 500 characters.

## Creator dashboard

`/countdowns` carries an offer to streamers: connect the channel, get featured in the daily countdowns, and see where each clip placed. It links a signed-in Twitch account to its creator dashboard (`/creator/<twitch id>/dashboard`, also in the user menu) and asks anyone else to sign in.

On their own dashboard, a broadcaster sees "Your clips in countdowns": each placement with its rank, series and date, linked to the edition page, and the countdown video and Short it appears in with their YouTube view counts. Until they have featured their clips, the card also offers "Connect my channel", the same `channel:manage:clips` grant as in Settings, returning to the dashboard. The card reads `GET /api/v1/users/me/countdowns/placements` (the newest 50, under the public-page rules) and hides itself for accounts without Twitch.

A video is linked only once viewers can open it: YouTube reports it public or unlisted, or, before the worker first reads it, its scheduled publish time has passed. The worker reads view counts with `videos.list` once an hour, for videos uploaded in the last 90 days whose count is more than 6 hours old, 50 per call. A video YouTube no longer returns is marked missing and no longer linked or read. Quota and authorization errors pause the reads as they pause publishing; other errors pause them until the next period.

## Weekly clip packs

Broadcasters who feature their clips get a weekly clip pack (#86): their most-viewed clips of the past week, ready to post.

- **When:** from Monday 06:00 UTC, the `countdown-media` worker queues a pack for the week that ended (Monday to Monday, UTC) for each broadcaster with a "feature my clips" opt-in and at least one clip that week. It checks again every hour, so a broadcaster who opts in later in the week still gets last week's pack. Broadcasters left out of countdowns, and clips removed, hidden, taken down or excluded, are skipped.
- **What:** up to 10 clips, most views first. Each becomes a captioned 9:16 video (the Short layout), and together they make a 16:9 highlight reel (the countdown layout, with a "This week's top clips" outro). The zip holds the reel, the 9:16 files and `credits.txt` with each clip's title, views and Twitch link.
- **Source:** clips are fetched only through Helix with the broadcaster's own `channel:manage:clips` token. A clip that cannot be downloaded is left out; a pack with none fails. Downloads and intermediate files are deleted once the pack is built.
- **Delivery:** the zip stays on the render host in `COUNTDOWN_PACK_DIR` (default `/var/lib/clpr-clip-packs`) for 7 days, then the worker deletes it and marks the pack `expired`. `GET /api/v1/users/me/clip-packs` lists the broadcaster's packs, and the creator dashboard shows them with a Download button. The link points at the worker's pack server (`CLIP_PACK_BASE_URL/packs/<id>.zip`), is signed with `CLIP_PACK_LINK_SECRET` (HMAC-SHA256), lasts 24 hours and never outlasts the pack. The server answers 403 for a bad or expired link and 410 for a pack that is gone.
- **Metrics:** `clpr_clip_packs_total{outcome}` and `clpr_clip_pack_files_deleted_total`.

## Weekly streamer digest

Broadcasters who feature their clips get a digest email every Monday (#85): their five most-viewed clips of the past week, where their clips placed in the daily countdowns, and whether their clip pack is ready, with a button to the creator dashboard.

- **Who:** a "feature my clips" opt-in, an email address, email notifications turned on (`email_enabled`, off by default), the "Weekly Streamer Digest" preference on (on by default), an email frequency other than "never", and not left out of countdowns. A broadcaster with no clips, placements or pack that week gets no email.
- **When:** the API checks hourly and sends last week's digests from Monday 09:00 UTC, after the clip packs. It runs only with `EMAIL_ENABLED=true`. Each digest is recorded in `streamer_digest_sends` before it is sent, so it goes out at most once even with several API replicas; the row ends as `sent`, `skipped` or `failed` with the reason.
- **Unsubscribe:** the email's "Stop the weekly digest" link turns off only the digest; Settings → Notifications has the same switch.
- **Tracking:** the email goes through SendGrid with the category `streamer_digest`, so the SendGrid event webhook records its deliveries, opens and clicks under that template in the email metrics. It is also logged in `email_notification_logs` with the type `streamer_digest`.
- **Turning it on:** set `EMAIL_SANDBOX_MODE=true` with `EMAIL_ENABLED=true` first, so digests are logged but not sent, and check a log line "SANDBOX MODE: Email would be sent". Then turn sandbox mode off.

## Operations

Series setup, the worker deployment, metrics, alerts and what to do when they fire are in the [[../operations/runbooks/countdowns|countdowns runbook]]. The alert rules are the `clpr-countdowns` group in `monitoring/prometheus/clpr-slo-rules.yml`, with controlled-failure tests in `monitoring/prometheus/clpr-countdown-rules.test.yml` (`promtool test rules`).

## Exclusions

`/api/v1/admin/countdowns/exclusions` holds the broadcasters (by Twitch ID) and single clips left out of every future edition and off the public countdown pages. Each has a reason (`request`, `dmca` or `takedown`) and a source:

- **`admin`:** recorded on `/admin/countdowns`, for example for a broadcaster who asked by email and has no clpr account.
- **`self_service`:** the broadcaster's own "Leave my clips out" toggle in Settings (`PUT /api/v1/users/me/countdowns/exclusion`), which also records the consent text version. Turning the toggle off removes only their own exclusion; an admin's stays and locks the toggle.

Clips removed after a DMCA notice are already ineligible without a row. A YouTube claim against a clip is recorded as a `takedown` exclusion.

An exclusion affects editions selected afterwards. Clips it covers in videos already on YouTube appear on `/admin/countdowns` as removal tasks, due 7 days after the exclusion. An admin cuts the clip in YouTube Studio or removes the video, then marks the task done (`removed`, `edited`, or `kept` with a note). The video IDs are in `countdown_uploads`.

**Feature my clips.** A broadcaster can opt in from Settings by granting `channel:manage:clips` through Twitch (see [[../compliance/twitch-api-usage|Twitch API usage]]). Their clips are then fetched through Helix rather than Twitch's player. "Stop featuring" or disconnecting Twitch removes the opt-in. The broadcaster API object (`GET /api/v1/broadcasters/{id}`) shows both controls as `countdown_excluded` and `countdown_featured`.

## Tests

- `backend/internal/repository/countdown_repository_integration_test.go`: every eligibility rule, ranking, the per-broadcaster cap, cross-series dedupe for the day, short editions stored as failed, reselection and exclusions, against a migrated database.
- `backend/internal/services/countdown_service_test.go`: due dates and windows, skipped series, input validation.
- `backend/internal/repository/countdown_media_repository_integration_test.go`: claiming, reclaiming a stale claim, backfill with the cap and the position shift, skips across series, recording media, and deletion.
- `backend/internal/services/countdown_media_test.go`: the worker's fetch, backfill, give-up and failure paths, the 24-hour sweep, ffprobe parsing, the Twitch URL check, and source choice between Helix and yt-dlp.
- `backend/internal/countdownrender/render_test.go`: golden frames of the countdown (cards, clips with and without the lower third, a portrait clip, the outro), both Shorts, a captioned Short (shown while its words are spoken, gone after) and the thumbnail; stream formats and durations; byte-identical output across runs; and, with `COUNTDOWN_RENDER_TIMING=1`, a full-size timing run. These tests need ffmpeg and the fonts, so `backend/Dockerfile.render` runs them while building the image. To regenerate the goldens, run the test binary in the image's `runtime` stage with `UPDATE_GOLDEN=1` and the source `testdata` directory mounted.
- `backend/internal/countdownrender/text_test.go` and `backend/internal/services/countdown_captions_test.go`: grouping words into caption lines; keeping only confident English speech; Shorts rendered without captions when transcription fails.
- `backend/internal/repository/countdown_render_repository_integration_test.go` and `backend/internal/services/countdown_render_worker_test.go`: claiming, recording and finishing renders, source deletion on success and failure, missing media, shutdown, and the render sweep.
- `backend/internal/repository/countdown_public_repository_integration_test.go`, `backend/internal/handlers/countdown_public_handler_test.go`, `frontend/src/pages/CountdownPages.test.tsx` and `frontend/e2e/mocked/countdowns.spec.ts`: which editions and clips are public, paging by day, placements, request validation, the sitemap entries, and both pages in a browser, including a phone-width check.
- `backend/internal/services/youtube_client_test.go`, `backend/internal/services/countdown_youtube_test.go` and `backend/internal/repository/countdown_publish_repository_integration_test.go`: the YouTube calls against a fake API (token refresh and caching, resumable upload, thumbnail, playlist, error reasons); titles, chapters, credits and limits; publishing with quota, authorization, temporary and permanent errors without uploading anything twice; scheduled publish times; claiming, recording and finishing in the database.
- `backend/internal/repository/countdown_controls_repository_integration_test.go`, `backend/internal/handlers/countdown_controls_handler_test.go`, `frontend/src/components/countdown/CountdownSettingsCard.test.tsx`, `frontend/src/pages/admin/AdminCountdownsPage.test.tsx` and `frontend/e2e/mocked/broadcaster-controls.spec.ts`: the self-service toggle (an admin's exclusion survives it), removal tasks for countdowns and Shorts and their resolution, the admin path for a broadcaster without an account, and the "feature my clips" grant and its removal.
- `backend/internal/repository/countdown_stats_repository_integration_test.go`, `backend/internal/services/countdown_youtube_stats_test.go`, `frontend/src/components/countdown/CountdownPlacementsCard.test.tsx`, `frontend/src/components/countdown/CreatorConnectOffer.test.tsx` and the dashboard test in `frontend/e2e/mocked/broadcaster-controls.spec.ts`: which placements and videos the dashboard shows before and after a view-count read, missing videos, the refresher's batching, interval and pauses, the connect offer, and the card at phone width.
- `backend/internal/repository/clip_pack_repository_integration_test.go`, `backend/internal/services/clip_pack_worker_test.go`, `backend/internal/handlers/clip_pack_handler_test.go` and `frontend/src/components/countdown/ClipPacksCard.test.tsx`: which broadcasters and clips get a pack, claiming and finishing, the zip's contents, a clip that cannot be downloaded, the week boundaries, deleting expired zips, signed links (tampered, expired, past the pack's expiry) and the pack server's answers, and the dashboard list.
- `backend/internal/services/streamer_digest_test.go`, `backend/internal/services/streamer_digest_integration_test.go` and `frontend/src/pages/NotificationPreferencesPage.test.tsx`: the email's content and escaping, who gets a digest, sending once per week, skipped and failed digests, the digest-only unsubscribe, preference updates that leave the field out, and the settings switch.
- `backend/internal/handlers/countdown_handler_test.go` and `backend/internal/scheduler/playlist_script_scheduler_test.go`: status codes and the scheduler hook.

---

**See also:** [[archive-handoff-import|Archive handoff import]] · [[index|Features]]
