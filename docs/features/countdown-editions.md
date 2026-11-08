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

A countdown edition is one day's ranked list of clips for a countdown video or Short. This page covers selection (#76). Fetching the media (#77), rendering (#78) and publishing (#79) build on the editions stored here. The media policy is [[../adr/002-media-handling-for-countdowns|ADR-002]].

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

## Exclusions

`/api/v1/admin/countdowns/exclusions` holds the broadcasters (by Twitch ID) and single clips left out of every future edition. Each has a reason (`request`, `dmca` or `takedown`) and a source (`admin` now, `self_service` once #75 adds the broadcaster's own toggle). Clips removed after a DMCA notice are already ineligible without a row. A YouTube claim against a clip is recorded as a `takedown` exclusion.

Excluding a broadcaster affects editions selected afterwards. Removing them from videos already published is #75.

## Tests

- `backend/internal/repository/countdown_repository_integration_test.go`: every eligibility rule, ranking, the per-broadcaster cap, cross-series dedupe for the day, short editions stored as failed, reselection and exclusions, against a migrated database.
- `backend/internal/services/countdown_service_test.go`: due dates and windows, skipped series, input validation.
- `backend/internal/handlers/countdown_handler_test.go` and `backend/internal/scheduler/playlist_script_scheduler_test.go`: status codes and the scheduler hook.

---

**See also:** [[archive-handoff-import|Archive handoff import]] · [[index|Features]]
