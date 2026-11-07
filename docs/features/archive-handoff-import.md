---
title: "Archive handoff import"
summary: "Import a Rekolekt clip handoff into an upload draft, with its source passage, transcript revision and editable captions."
tags: ["features", "submissions", "uploads"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-07
---

# Archive handoff import

A Rekolekt archive can export a passage as an MP4 together with a JSON handoff (`rekolekt.clip-handoff`, version 1). The handoff records the recording, the selected range, the transcript revision, clip-relative captions and the MP4's SHA-256. On the submit page, in **Upload video** mode, a submitter can import that JSON into the normal submission draft. There is no separate submission flow.

The producer documents the contract in Rekolekt's `docs/product/clip-handoff.md`. clpr keeps a byte-identical copy of its example as the test fixture in `backend/internal/services/testdata/rekolekt-clip-handoff.v1.json`, with a second copy in `frontend/src/test/fixtures/`. A backend test fails if the two copies differ.

## What the submitter does

1. Choose **Upload video**, then choose the handoff JSON under **Archive handoff**.
2. Review the summary: archive, passage range, transcript source and revision, cue count and expected file. **Use in draft** applies it. If the draft already has a title, reason, tags or an earlier import, the button reads **Replace draft** and a warning says what is cleared. **Cancel** leaves the draft unchanged.
3. Edit the caption text or remove cues. **Restore the archive's captions** brings back the originals.
4. Choose the MP4 downloaded with the handoff. The page compares its size and SHA-256 with the handoff and shows whether they match.
5. Press **Upload Clip**. Importing never submits anything on its own.

The draft, including the import and caption edits, is saved in the browser as soon as the handoff is applied, and again by the normal 30-second autosave. A resumed draft restores the import, but the MP4 has to be chosen again because browsers do not keep file selections.

## What the server checks

`POST /api/v1/submissions/upload` accepts two optional form fields alongside the file:

- `archive_handoff`: the handoff document as imported, at most 512 KiB.
- `caption_cues`: a JSON array of the submitter's cues, each with `start_ms`, `end_ms` (or `null`) and `text`. It is accepted only together with `archive_handoff`.

The handler rejects these with a 400 and a message naming the problem, before the upload is validated or stored:

- an unknown schema, or any version other than 1;
- an invalid export ID, a passage whose end is not after its start or whose duration does not match;
- passage or player URLs that are not `https`, contain credentials, or are not on the archive's declared origin;
- more than 200 cues, a cue over 2,000 characters, cues out of order or outside the clip, or over 64 KiB of caption text;
- captions that do not match the handoff's `text_sha256`, which the server recomputes with the producer's algorithm;
- a media type other than `video/mp4`, or a provenance basis other than `declared`;
- an uploaded file whose size or SHA-256 differs from `media.bytes` and `media.sha256` (`field: "file"`).

After that, the upload goes through the usual pipeline unchanged: MIME, size and duration validation, storage under the pending prefix, ban and karma checks, abuse and rate limits, and moderation. The server never fetches a URL from the handoff and never receives archive cookies.

## What is stored

The submission's `source_metadata.archive_provenance` holds a typed copy of the handoff fields clpr uses, plus:

- `captions.source_cues` and `captions.source_text_sha256`: the archive's cues, unchanged;
- `captions.cues` and `captions.edited`: the cues shown to viewers. These are the submitter's edits when they differ from the archive's, and the archive's cues otherwise;
- `import.media_digest_matched` and `import.imported_at`: what clpr checked.

Approval copies `source_metadata` to the clip. Playback stays the uploaded file. The passage URL is shown as a link and is never used as `video_url` or `source_url`.

## What viewers see

- **Source link:** the clip page shows a **Source** row linking to the passage, with a note that the archive declared the provenance and clpr did not verify it.
- **Captions:** when the clip has cues, the player gets a caption track from `GET /api/v1/clips/{id}/captions`. That route returns WebVTT, escapes markup in the cue text, and shows a cue without an end time until the next cue starts or the clip ends.

## Provenance is not rights

The handoff says which recording, range and transcript the archive owner selected. It is not proof of rights to the media. An imported upload gets no exemption from moderation, duration limits or other submission rules. Uploads have no content-hash duplicate check today, so a second submission of the same export is not detected automatically.

## Tests

- `backend/internal/services/archive_handoff_test.go`: the fixture contract, the producer digest, each rejection, and edited cues kept apart from the archive's.
- `backend/internal/services/archive_provenance_upload_test.go`: provenance carries from the upload to the approved clip, and playback stays the uploaded file.
- `backend/internal/handlers/submission_archive_upload_test.go`: a matching upload, plus mismatched bytes, unsupported versions and unsafe URLs rejected before storage; WebVTT rendering.
- `frontend/src/lib/archive-handoff.test.ts` and `frontend/src/pages/SubmitClipPage.archive.test.tsx`: import, review, cancel, replace, draft resume, the submitted fields and a mismatched MP4.
- `frontend/src/pages/ClipDetailPage.test.tsx`: the source link and caption track.

---

**See also:** [[clip-submission-rate-limiting|Clip submission rate limiting]] · [[index|Features]]
