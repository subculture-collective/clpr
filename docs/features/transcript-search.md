---
title: "Transcript search"
summary: "Find clips by what was said in them: which clips are transcribed, the search index, the said filter and snippets, and the daily transcription cost."
tags: ["features", "search", "countdowns"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-10
---

# Transcript search

Search finds clips by what was said in them (#87). A query matches clip transcripts as well as titles and names, the `said` filter keeps only clips whose transcript contains a phrase, and each clip that matched in its transcript shows the matching words.

## Which clips have transcripts

Transcripts come from three places, all running Whisper locally:

- **Popular clips** ([[../adr/003-transcribing-clips-for-search|ADR-003]]). The `countdown-media` worker on the render host transcribes public Twitch clips with at least 1,000 views from the last 30 days, most viewed first, up to 500 a day and 5 per one-minute cycle. It skips clips that are removed, hidden, NSFW or taken down, clips whose Twitch language is set to something other than English, and clips and broadcasters on the countdown exclusion list. Each file is downloaded the same way as countdown media, transcribed and deleted before the next one. `search_transcription_attempts` keeps the last attempt per clip: a clip without English speech is not tried again, and a failed one is retried after 6 hours, three times at most.
- **Countdown clips.** The worker transcribes every clip of each countdown edition while its media is on the render host, not only the clips that get a Short (see [[countdown-editions#captions|Captions]]). A clip that already has a transcript is not transcribed again unless it gets a Short, which needs the word timings for captions. `COUNTDOWN_TRANSCRIPTS=off` limits the worker to the Shorts' captions again.
- **Authorized clips.** The API's existing transcription of clips whose broadcaster granted `channel:manage:clips`.

Only confident English speech is kept, the same rule as for captions. A transcript from the broadcaster's own authorization replaces one from the render host, never the other way round, and a search transcript never replaces any other. Transcripts are stored in `clip_transcripts` (`source` is `twitch_authorized_whisper`, `countdown_render_whisper` or `search_render_whisper`) and are never sent to clients in full.

When a clip or broadcaster is added to the exclusion list, the worker empties the render host's transcripts of those clips on its next cycle. The emptied transcript counts as changed, so the index scheduler removes the text from the search index within five minutes.

### Settings

Popular-clip transcription runs whenever the worker has Whisper (see [[countdown-editions#captions|Captions]]).

| Variable | Default | Meaning |
| --- | --- | --- |
| `SEARCH_TRANSCRIPTS` | on | `off` stops popular-clip transcription; countdown and authorized clips are still transcribed |
| `SEARCH_TRANSCRIPT_MIN_VIEWS` | `1000` | Fewest views a clip needs |
| `SEARCH_TRANSCRIPT_MAX_AGE_DAYS` | `30` | Oldest clip, by creation date |
| `SEARCH_TRANSCRIPT_DAILY_LIMIT` | `500` | Most clips attempted per UTC day, failures included |

`clpr_search_transcription_clips_total{outcome}` counts clips by outcome (`transcribed`, `no_speech`, `failed`) on the worker's metrics port.

## Indexing

The clips search index has a `transcript` field. Full index runs (`backfill-search`, `search-index-manager rebuild`) read it from `clip_transcripts`. Between them, an API scheduler copies new and changed transcripts into the index every five minutes, 200 at a time, with partial updates; `clip_transcripts.search_indexed_at` records the version it copied. A clip the index does not hold yet is skipped and picked up by the next full run. When a full run finishes, the transcripts written while it ran are queued again, since the run read them before they changed.

An index created before this change gets the field from OpenSearch's dynamic mapping (text with a keyword subfield), which serves phrase queries and highlights. Rebuild the index with `search-index-manager rebuild -index clips` to get the declared mapping, which adds an English-analyzed `transcript.en` subfield.

## Searching

- **Text query (`q`):** the clip query's `multi_match` includes `transcript` with the same weight as the Twitch category name.
- **`said`:** `GET /api/v1/search?said=<phrase>` adds a `match_phrase` on `transcript`, so the words must appear next to each other in that order, as in the database fallback. It works with or without `q`; without `q`, only clips are searched. At most 200 characters.
- **Snippets:** for clips that matched in their transcript, the response carries `transcript_snippet`: about 140 characters around the match, with each matched word between U+E000 and U+E001. The search page shows it under the clip with the words marked. Hybrid reranking keeps the snippets.
- **Database fallback:** when OpenSearch is unavailable and search falls back to PostgreSQL, `said` uses `phraseto_tsquery` on the transcript and `ts_headline` for the snippet, with the same markers.

The search page has a "Said in the clip" box above the filters, which sets the `said` URL parameter.

## Cost

Whisper runs on clpr's own hosts, so its cost is compute time. Each transcription adds to `transcription_usage_daily`, one row per UTC day and source, with the clips attempted, transcripts stored, seconds of audio and seconds of compute. Clips the API skipped because the broadcaster has not authorized transcription, or because the provider is down, are not counted. A search with only `said` is not recorded in search history or analytics.

```sql
SELECT day, source, clips, transcripts, round(audio_seconds) AS audio_s, round(compute_seconds) AS compute_s
FROM transcription_usage_daily ORDER BY day DESC, source;
```

## Tests

- `backend/internal/repository/transcript_repository_integration_test.go`: storing render-host transcripts without replacing authorized ones, daily usage totals across a UTC midnight, the index queue, and the database `said` search with its snippet.
- `backend/internal/services/transcript_search_test.go` and `backend/internal/scheduler/transcript_index_scheduler_test.go`: the OpenSearch request body for `said` and highlights, snippet parsing, snippets kept through reranking, transcripts never serialized, and the index scheduler's batches and failure handling.
- `backend/internal/services/countdown_captions_test.go`: every edition clip transcribed once, Shorts keeping their captions, and usage recorded.
- `backend/internal/repository/search_transcription_repository_integration_test.go` and `backend/internal/services/search_transcription_worker_test.go`: which popular clips are picked, retries, the daily limit, files deleted after each clip, failures that stop the pass, and emptying excluded broadcasters' transcripts.
- `frontend/src/components/search/TranscriptSearch.test.tsx` and `frontend/src/pages/SearchPage.test.tsx`: the marked snippet, the said box, and a said-only search.

---

**See also:** [[countdown-editions|Countdown editions]] · [[../compliance/twitch-api-usage|Twitch API usage]] · [[index|Features]]
