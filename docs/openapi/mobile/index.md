---
title: "Native viewer API contract fixtures"
summary: "Shared wire examples and offline checks for the separate iOS and Android viewer apps."
tags: ["openapi", "mobile", "testing"]
area: "openapi"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-17
---

# Native viewer API contract fixtures

Both native clients consume the JSON response files in this directory. The
`manifest.json` file identifies each response's method, registered path, status,
source boundary and provenance. These are deterministic examples using owned
fixture values. They contain no production records, credentials or real account
identifiers.

This pack implements the contract prerequisite in the
[native app plan](../../product/native-apps-plan.md) and
[issue #1444](https://github.com/subculture-collective/clpr/issues/1444).
The authoritative schema is [openapi.yaml](../openapi.yaml).

## Response families

| Journey | Route | Decoder and paging contract |
| --- | --- | --- |
| Ranked clips | `GET /api/v1/clips` | `success`, array `data`, and `meta` with `page`, `limit`, `total`, `total_pages`, `has_next`, `has_prev`; defaults page 1 and limit 25 |
| Clip detail | `GET /api/v1/clips/{id}` | `success` and object `data`; the identifier accepts an internal UUID or Twitch clip slug |
| Text and spoken phrases | `GET /api/v1/search?q=...&said=...` | `query`, `results`, `counts`, `meta`; all five result arrays exist, including empty arrays; paging uses `total_items` rather than `total` |
| Tags and taxonomy lanes | `GET /api/v1/tags` | `tags`, `total`, `page`, `limit`, `has_more`; default limit 50; retain the server's lane/display labels |
| Twitch categories | `GET /api/v1/games/trending` | Legacy route and `games` key; `page`, `limit`, `has_more`; default limit 20 |
| Clip topics | `GET /api/v1/clips/{id}/topics` | `topics` array with topic identity, confidence, evidence and dates |
| Creator rails | `GET /api/v1/broadcasters/discover` | `success`, object `data` containing `trending`, `rising`, `new`; bounded rail limit, without a page cursor |
| Countdown editions | `GET /api/v1/countdowns` | `success`, array `data`, `meta.next_before`; pass that date as `before` for the next page; null ends paging |
| Countdown detail | `GET /api/v1/countdowns/{date}/{series}` | `success`, edition `data`; missing or unpublished editions return 404 |
| Public playlist detail | `GET /api/v1/playlists/{id}` | `success`, playlist `data` and page metadata; default limit 20; an empty playlist can omit `clips` |

Ordinary timestamps are RFC 3339 strings and may contain fractional seconds.
Countdown dates and cursors use `YYYY-MM-DD`. UUIDs and upstream Twitch category
identifiers are distinct types: category IDs are strings, not internal UUIDs.
The fixtures include retained `game_*` names and the `twitch_category_*` aliases.

Treat omitted optional metadata as absent. Do not require optional thumbnails,
duration, descriptions, permissions or user-specific fields just because one
populated example contains them. `next_before` is a required key whose value can
be null. Do not turn an empty search result array into null. Ignore additional
response fields so an additive server change does not break decoding.

Search snippets are optional excerpts supplied by the API. Matched words are
delimited by U+E000 and U+E001. Render those markers as native text emphasis;
never interpret the snippet as HTML or request the full stored transcript.
Countdown ranks can have gaps after exclusions or removals. A rendered edition
does not establish that its YouTube video has been published.

## Errors and account boundaries

Clients must inspect the HTTP status and decode both existing error families:
`{"error":"message"}` and
`{"success":false,"error":{"code":"CODE","message":"message"}}`.
Rate-limit bodies can include a numeric `retry_after` in seconds. The constructed
429 example exercises decoding; it does not claim that a live rate limit fired.
Retain the server's retry guidance and avoid automatic write retries that could
duplicate an action.

A missing, private or removed resource must not reveal its hidden contents.
The playlist 404 fixture represents both missing and inaccessible playlists.
Personal libraries, playlist writes and submissions still require authenticated
authorization. These fixtures grant no access and contain no session response.
The native authentication/MFA exchange remains a separate prerequisite.

## Provenance and limits

- `handler-with-owned-stub`: an actual Gin handler produced the body using a
  deterministic provider or validation path, without a database or live provider.
- `constructed-from-wire-types`: the current Go response envelope and model
  serializers produced a synthetic body. This checks serialization and decoder
  shape; it is not a capture of the complete database-backed request.
- `constructed-negative-case`: a deliberate removed/private/rate-limit example
  based on its named handler or middleware branch.

Real staging requests, provider playback, OAuth/MFA and physical-phone tests
remain separate evidence. The fixture generator checks that every route is
registered and every named source file exists; the existing OpenAPI route test
checks that the route manifest remains current.

## Verify and update

From the repository root, with its documented Go and Node toolchains and root
dependencies installed:

```bash
task mobile:contracts
npm run openapi:lint
npm run openapi:embed:check
```

The task compares the files with current Go serializers/handler captures, then
validates each body against its current OpenAPI response schema. Negative checks
reject missing envelopes, incorrect counters, invalid UUIDs/dates, malformed
errors and missing/null search arrays. The source gate runs the Node checks in
its API documentation phase; backend tests also check fixture drift.

After an intentional contract change, regenerate and review the fixtures:

```bash
cd backend
UPDATE_NATIVE_FIXTURES=1 go test ./cmd/api -run TestNativeContractFixtures -count=1
cd ..
npm run openapi:embed
task mobile:contracts
```

Update the schema, fixtures, consumer decoders and tests together. Do not add a
proposed native-session endpoint to this pack before that contract exists.
