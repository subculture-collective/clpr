---
title: "Transcript search"
summary: "Find clips by what was said in them: stored transcripts, the search index, the said filter and snippets, and the daily transcription cost."
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

Transcripts come from two places, both already running Whisper locally:

- **Countdown clips.** The `countdown-media` worker transcribes every clip of each countdown edition while its media is on the render host, not only the clips that get a Short (see [[countdown-editions#captions|Captions]]). Countdown clips are the day's fastest-rising clips, so they serve as the view threshold. Only confident English speech is kept, the same rule as for captions. `COUNTDOWN_TRANSCRIPTS=off` limits the worker to the Shorts' captions again.
- **Authorized clips.** The API's existing transcription of clips whose broadcaster granted `channel:manage:clips`.

No clip is downloaded for search alone: [[../adr/002-media-handling-for-countdowns|ADR-002]] allows downloading clips only for countdowns. A transcript from the broadcaster's own authorization replaces one from the render host, never the other way round. Transcripts are stored in `clip_transcripts` (`source` is `twitch_authorized_whisper` or `countdown_render_whisper`) and are never sent to clients in full.

## Indexing

The clips search index has a `transcript` field. Full index runs (`backfill-search`, `search-index-manager rebuild`) read it from `clip_transcripts`. Between them, an API scheduler copies new and changed transcripts into the index every five minutes, 200 at a time, with partial updates; `clip_transcripts.search_indexed_at` records the version it copied. A clip the index does not hold yet is skipped and picked up by the next full run.

An index created before this change gets the field from OpenSearch's dynamic mapping (text with a keyword subfield), which serves phrase queries and highlights. Rebuild the index with `search-index-manager rebuild -index clips` to get the declared mapping, which adds an English-analyzed `transcript.en` subfield.

## Searching

- **Text query (`q`):** the clip query's `multi_match` includes `transcript` with the same weight as the Twitch category name.
- **`said`:** `GET /api/v1/search?said=<phrase>` adds a `match_phrase` on `transcript` (slop 1), so the words must appear in that order. It works with or without `q`; without `q`, only clips are searched. At most 200 characters.
- **Snippets:** for clips that matched in their transcript, the response carries `transcript_snippet`: about 140 characters around the match, with each matched word between U+E000 and U+E001. The search page shows it under the clip with the words marked. Hybrid reranking keeps the snippets.
- **Database fallback:** when OpenSearch is unavailable and search falls back to PostgreSQL, `said` uses `phraseto_tsquery` on the transcript and `ts_headline` for the snippet, with the same markers.

The search page has a "Said in the clip" box above the filters, which sets the `said` URL parameter.

## Cost

Whisper runs on clpr's own hosts, so its cost is compute time. Each transcription adds to `transcription_usage_daily`, one row per UTC day and source, with the clips attempted, transcripts stored, seconds of audio and seconds of compute. Clips the API skipped because the broadcaster has not authorized transcription, or because the provider is down, are not counted.

```sql
SELECT day, source, clips, transcripts, round(audio_seconds) AS audio_s, round(compute_seconds) AS compute_s
FROM transcription_usage_daily ORDER BY day DESC, source;
```

## Tests

- `backend/internal/repository/transcript_repository_integration_test.go`: storing render-host transcripts without replacing authorized ones, daily usage totals across a UTC midnight, the index queue, and the database `said` search with its snippet.
- `backend/internal/services/transcript_search_test.go` and `backend/internal/scheduler/transcript_index_scheduler_test.go`: the OpenSearch request body for `said` and highlights, snippet parsing, snippets kept through reranking, transcripts never serialized, and the index scheduler's batches and failure handling.
- `backend/internal/services/countdown_captions_test.go`: every edition clip transcribed once, Shorts keeping their captions, and usage recorded.
- `frontend/src/components/search/TranscriptSearch.test.tsx` and `frontend/src/pages/SearchPage.test.tsx`: the marked snippet, the said box, and a said-only search.

---

**See also:** [[countdown-editions|Countdown editions]] · [[../compliance/twitch-api-usage|Twitch API usage]] · [[index|Features]]
