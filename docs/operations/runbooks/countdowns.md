---
title: "Countdowns"
summary: "Running the daily countdown pipeline: series setup, the worker, its metrics and alerts, and what to do when an edition is missed or a stage fails."
tags: ["operations", "countdowns"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-08
---

# Countdowns Runbook

Each day clpr selects the fastest-rising Twitch clips for every enabled countdown series. It then fetches their media, renders a countdown video, a thumbnail and Shorts, and uploads them to YouTube. How each stage works is in [Countdown editions](../../features/countdown-editions.md); the YouTube side is in [YouTube API usage](../../compliance/youtube-api-usage.md). This page covers running it.

## The pipeline

| Stage | Runs in | Edition status |
|-------|---------|----------------|
| Selection, at each series' `selection_hour_utc` | API scheduler | `selected` (or `failed` when too few clips are eligible) |
| Media fetch | `countdown-media` worker on the render host | `fetching` → `fetched` |
| Render, with each clip's transcript stored for search | same worker | `rendering` → `rendered`; the page at `/countdowns/<date>/<series>` goes public |
| Publish to YouTube | same worker, once the channel is authorized | `publishing` → `published` |
| View counts for the creator dashboard, hourly | same worker, once the channel is authorized | unchanged |
| Weekly clip packs, from Monday 06:00 UTC | same worker, once `CLIP_PACK_LINK_SECRET` is set | separate `clip_packs` rows: `pending` → `building` → `ready` → `expired` after 7 days |

The worker runs a cycle every minute (`COUNTDOWN_MEDIA_INTERVAL_SECONDS`). A 10-clip edition takes about 3 minutes to render on 4 cores, plus about 30 seconds per Short and up to 10 seconds of transcription per Short for captions.

## Deploying the worker

The worker is image `backend/Dockerfile.render` with command `countdown-media`. Run it on the render host beside the database (Almaz), never on Dozor.

- **Volumes:** three named volumes: `/var/lib/clpr-countdown-media` (source clips, deleted within 24 hours), `/var/lib/clpr-countdown-renders` (renders, deleted after 7 days) and `/var/lib/clpr-clip-packs` (weekly clip packs, deleted after 7 days). The Whisper model downloads to `/var/lib/clpr-whisper-models` inside the container on first use.
- **Environment:**
  - `ENVIRONMENT=local` and `GIN_MODE=release`. The worker shares the API's config loader, and `ENVIRONMENT=production` would demand API-only secrets such as the JWT key. `GIN_MODE=release` gives JSON logs;
  - the API's database settings and `TWITCH_CLIENT_ID`;
  - `BASE_URL`, the same as the API's, for links in webhooks and video descriptions;
  - once the channel is authorized, `YOUTUBE_CLIENT_ID`, `YOUTUBE_CLIENT_SECRET_FILE` and `YOUTUBE_REFRESH_TOKEN_FILE` from the secret store.
  - for padding countdown clips from the VOD, `COUNTDOWN_VOD_PAD_BEFORE_SECONDS` and `COUNTDOWN_VOD_PAD_AFTER_SECONDS` (0 by default, at most 30; see [ADR-004](../../adr/004-padding-countdown-clips-from-the-vod.md)). The startup log shows `vod_pad_before` and `vod_pad_after`, and `clpr_countdown_media_fetches_total{source="vod"}` counts padded fetches;
  - for transcript search, nothing: popular clips are transcribed whenever Whisper is available ([ADR-003](../../adr/003-transcribing-clips-for-search.md)). `SEARCH_TRANSCRIPTS=off` stops it, and `SEARCH_TRANSCRIPT_MIN_VIEWS`, `SEARCH_TRANSCRIPT_MAX_AGE_DAYS` and `SEARCH_TRANSCRIPT_DAILY_LIMIT` change its bounds (see [Transcript search](../../features/transcript-search.md#settings));
  - for weekly clip packs, `CLIP_PACK_LINK_SECRET_FILE` (a random secret of at least 32 bytes from the secret store) and `CLIP_PACK_BASE_URL`, the public address of the pack server. Give the API the same two values as `CLIP_PACK_LINK_SECRET` and `CLIP_PACK_BASE_URL`. Packs also need `TWITCH_CLIENT_ID` and `TWITCH_CLIENT_SECRET`: they use only Helix downloads, and the worker renews broadcasters' Twitch tokens, which last about four hours, with the client secret.
- **Resources:** about 2 GB of memory and up to 4 CPUs while rendering. The image is about 1.6 GB.
- **Metrics:** port 9464 (`COUNTDOWN_MEDIA_METRICS_ADDR`). Publish it on the LAN address only, and add a Prometheus scrape job named `clpr-countdown-media` on Dozor.
- **Clip pack server:** port 9465 (`CLIP_PACK_ADDR`) serves only `GET /packs/<id>.zip` with a signed link. Route the `CLIP_PACK_BASE_URL` hostname to it through the edge (Caddy on Almaz), without Authelia, since broadcasters download with the link alone. It is off while `CLIP_PACK_LINK_SECRET` is unset.
- **Alert rules:** the `clpr-countdowns` group in `monitoring/prometheus/clpr-slo-rules.yml`. Copy it into Dozor's rules with the scrape job. Critical alerts page through the existing ntfy route; warnings go to the daily digest.

Check after deploying:

1. The log line "Countdown media worker started" shows `captions: true` and `search_transcripts: true`, and `youtube: true` once authorized.
2. `curl -s http://<render-host>:9464/metrics | grep clpr_countdown_edition_overdue_seconds` lists every enabled series.
3. In Prometheus, `up{job="clpr-countdown-media"}` is 1.

## Daily series

Plan of record: one overall series and two or three category series, with three to five Shorts a day in total. Series are rows in `countdown_series`, managed at `/api/v1/admin/countdowns/series` (admin with MFA). The same configuration in SQL:

```sql
INSERT INTO countdown_series (slug, name, game_id, size, window_hours, per_broadcaster_cap, selection_hour_utc, priority, enabled, shorts_count)
VALUES
  ('overall',       'Overall',       NULL,     10, 24, 2, 12, 0, true, 3),
  ('just-chatting', 'Just Chatting', '509658', 5,  24, 1, 12, 1, true, 1),
  ('<category>',    '<Category>',    '<id>',   5,  24, 1, 12, 2, true, 1)
ON CONFLICT (slug) DO NOTHING;
```

Size each series from its supply, not from a fixed number. In the week of 2026-11-01, Overall had about 800 usable clips a day and the largest category, Just Chatting, about 40. Top 10 with one clip per broadcaster would have failed most days in any category.

`min_size` makes a series follow the day's supply. With it set, an edition takes up to `size` clips, and a day with fewer than `min_size` is stored as `skipped`, which raises no alert. Without it, an edition needs exactly `size` clips or it fails and **Missed edition** fires. To set it: `UPDATE countdown_series SET size = 10, min_size = 5 WHERE slug = '<series>';`

A second list on the same day is another series row at a different `selection_hour_utc` and priority. A clip still goes to at most one edition a day.

The overall series picks first (priority 0); a clip goes to at most one edition a day. Choose the other categories from where the clips are, for example the categories with the most clips in the last week:

```sql
SELECT c.game_id, g.name, count(*) AS clips
FROM clips c LEFT JOIN games g ON g.twitch_game_id = c.game_id
WHERE c.source_type = 'twitch' AND c.created_at > NOW() - INTERVAL '7 days' AND NOT c.is_removed AND c.game_id IS NOT NULL
GROUP BY c.game_id, g.name ORDER BY clips DESC LIMIT 10;
```

A series needs `size` eligible clips, or `min_size` if it has one, in its window. A clip is eligible at 5 to 60 seconds, when it is not NSFW, removed or excluded, and once clpr has seen its view count rise. Below that number, an edition without `min_size` is stored as `failed` with the reason and **Missed edition** fires, and an edition with `min_size` is stored as `skipped`.

## Metrics

| Metric | Meaning |
|--------|---------|
| `clpr_countdown_edition_overdue_seconds{series,stage}` | How long the due edition has been missing at `rendered` (and at `published`, when publishing is on); 0 when it is there |
| `clpr_countdown_media_last_run_timestamp_seconds` | Last cycle that finished without errors |
| `clpr_countdown_media_editions_total{outcome}`, `clpr_countdown_media_fetches_total{source,outcome}`, `clpr_countdown_media_skips_total` | Fetch stage |
| `clpr_countdown_render_editions_total{outcome}`, `clpr_countdown_render_seconds{kind}`, `clpr_countdown_short_captions_total{outcome}` | Render stage |
| `clpr_countdown_publish_editions_total{outcome}`, `clpr_countdown_youtube_uploads_total{kind,outcome}` | Publish stage |
| `clpr_search_transcription_clips_total{outcome}` | Popular clips transcribed for search, by outcome (`transcribed`, `no_speech`, `failed`) |

## Alerts

### Worker down or stalled

`ClprCountdownWorkerDown` (critical): Prometheus cannot scrape the worker for 10 minutes. `ClprCountdownWorkerStalled` (warning): no error-free cycle for 30 minutes.

1. Check the container is running and read its last log lines.
2. A cycle that keeps failing names its stage in "Countdown media cycle failed". Database errors point at the database or its connection settings; "executable not found" at a broken image.
3. A worker killed mid-edition is safe: claims older than an hour (three for publishing) are taken again.

### Missed edition

`ClprCountdownEditionMissed` (critical): an enabled series has no rendered edition three hours after its selection hour.

1. Find the edition: `GET /api/v1/admin/countdowns/editions?series_id=<id>&limit=3`, or `SELECT edition_date, status, failure_reason FROM countdown_editions e JOIN countdown_series s ON s.id = e.series_id WHERE s.slug = '<series>' ORDER BY edition_date DESC LIMIT 3;`
2. **No row for the date:** selection did not run. The API's playlist script scheduler selects editions every 5 minutes. Check the API is up, that `job_last_success_timestamp_seconds{job_name="countdown_edition_select"}` is recent, and look for "Countdown edition selection failed" in the API log.
3. **`failed` with "only N of M clips were eligible":** the category had too few clips that day. Set `min_size`, lower `size`, widen `window_hours`, or accept the gap. A wider window can repeat a clip from the day before.
4. **`failed` in fetch or render:** see [Failed edition](#failed-edition).
5. **Stuck in `selected`, `fetched` or `rendering`:** the worker is not making progress; see [Worker down or stalled](#worker-down-or-stalled).
6. To try a date again after fixing the cause, select it again: `POST /api/v1/admin/countdowns/series/<id>/editions` with `{"date": "YYYY-MM-DD"}`. This works only while the edition is `selected`, `failed` or `skipped`.

### Failed edition

`ClprCountdownEditionFailed` (warning) names the stage. The edition's `failure_reason` says why.

- **fetch:** more than three clips could not be fetched or verified, or yt-dlp or ffprobe is missing. Twitch player changes break yt-dlp. Rebuild the image with a newer pinned `yt-dlp` in `backend/Dockerfile.render`.
- **render:** ffmpeg failed; the reason holds its last output. Source clips are already deleted, so select the date again to start over.
- **publish:** YouTube rejected the upload for a reason that will repeat, such as invalid metadata. Fix the cause, then select the date again or upload by hand.

### Not published

`ClprCountdownEditionNotPublished` (warning): an edition is rendered but not on YouTube a day after its selection hour. Look at the publish metrics and the worker log; quota, authorization and upload failures each have their own alert below. Renders stay on the host for 7 days.

### Upload failed

`ClprCountdownUploadFailed` and `ClprCountdownYouTubeQuotaReached` (warnings).

- **Quota or upload limit:** publishing pauses until 08:00 UTC and resumes on its own. If it happens daily, publish fewer Shorts (`shorts_count`) or ask Google for more quota.
- **Temporary errors** (YouTube 5xx, rate limits, network): publishing retries every 15 minutes. Nothing to do unless it persists.
- **Thumbnail failures:** these count as uploads of kind `thumbnail` and do not raise this alert. They usually mean the channel is not verified for custom thumbnails.

### YouTube authorization

`ClprCountdownYouTubeAuthFailed` (critical): Google rejected the refresh token (revoked, expired, or the OAuth client changed). Publishing pauses for an hour at a time; editions stay `rendered`.

1. On a machine with a browser, run `YOUTUBE_CLIENT_ID=… YOUTUBE_CLIENT_SECRET_FILE=… countdown-media youtube-auth -out yt-refresh-token` and sign in as the channel owner.
2. Replace the token in the secret store and restart the worker.
3. While the OAuth consent screen is in testing, Google expires refresh tokens after 7 days. Publish the consent screen to stop that.

## Takedowns

A broadcaster's request or a claim against a clip becomes a countdown exclusion (`/api/v1/admin/countdowns/exclusions`). It keeps the clip or broadcaster out of future editions and off the public countdown pages at once. On its next cycle the worker also empties the render host's transcripts of their clips, and the search index drops the text within five minutes. Videos already on YouTube must be edited or removed in YouTube Studio; the video IDs are in `countdown_uploads`.

## Rollback

Stop the worker: nothing else depends on it. Editions wait in their current status and resume when it runs again. To roll back the database, migrations 149 to 155 have down files; take a dump first, because down migrations drop the countdown tables.

---

**See also:** [Countdown editions](../../features/countdown-editions.md) · [Transcript search](../../features/transcript-search.md) · [YouTube API usage](../../compliance/youtube-api-usage.md) · [Runbooks](./README.md)
