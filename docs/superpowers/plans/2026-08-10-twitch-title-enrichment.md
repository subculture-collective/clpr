# Twitch-Constrained Title Enrichment

**Status:** Approved for implementation

**Created:** 2026-08-10

## Seed

Create accurate, useful titles for automatically ingested Twitch clips using
only information Twitch officially exposes: the source title, clip metadata,
and thumbnail URL. The system must not imply that it heard dialogue or watched
the full clip.

This is an error-rich seed. Audio/video extraction failed because arbitrary
clips do not expose an authorized raw media URL, while Twitch's current
thumbnail URLs are valid and publicly retrievable.

## Defaults considered

1. Rewrite every title with a language model.
2. Scrape Twitch's private playback endpoints and transcribe the video.
3. Keep Twitch titles unchanged.
4. Generate titles from the thumbnail alone.
5. Ask users to title every automated clip.

The first, second, and fourth options can confidently invent context. The
third wastes useful evidence; the fifth defeats automated ingestion.

## Functions and constraints

- Preserve Twitch's source title as evidence and provenance.
- Improve only automated clips; never overwrite a human title.
- Use the source title as the primary signal and the thumbnail as supporting
  visual evidence.
- Include broadcaster, creator, game, language, and duration as context.
- Do not invent dialogue, identities, outcomes, or events outside those inputs.
- Fall back to the Twitch title whenever confidence or evidence is weak.
- Attach conservative content tags independently of title acceptance.
- Record structural and vision processing separately; failures remain retryable.

## Adjacent possibilities

1. **Conservative enrichment (selected):** one structured vision request returns
   a candidate title, confidence, evidence basis, and controlled tags. A local
   policy gate decides whether the title is safe to publish.
2. Title-quality scoring without rewriting, for later moderator review.
3. Multiple title candidates with community selection.
4. Creator-authorized caption enrichment if Twitch exposes captions through an
   official API in the future.

## Axis map

| Axis | Default | Selected rotation |
|---|---|---|
| Who decides | Model | Model proposes; deterministic policy accepts |
| When | Every sync | Once per thumbnail revision, with explicit retry state |
| Scope | Replace every title | Automated clips only; meaningful source titles normally remain |
| Evidence | Assumed full video | Twitch title + metadata + one official thumbnail |

## Stress test and safeguards

- A thumbnail cannot prove spoken dialogue, a clutch outcome, or why someone
  reacted. The prompt forbids these claims unless already present in the source
  title.
- Model confidence is not sufficient. Candidates must pass length, content,
  provenance, and source-quality gates.
- Provider refreshes must not overwrite accepted AI or human titles.
- A failed API call must not mark vision processing complete.
- The existing 719 false-complete rows are reset only after the corrected worker
  is deployed.

## Decision

Implement a deep `AnalyzeClipThumbnail` interface that accepts clip metadata
and returns a structured enrichment result. Persist title provenance separately
from Twitch's source title. Publish a generated title only when the source title
is demonstrably low-quality and the candidate passes the conservative gate.
Use the same successful response for controlled content tags.

This approach is genuinely different from the failed media pipeline: it treats
the model as an evidence-bounded editor, not a substitute viewer or transcriber.

## Behavioral tests

1. A vision request uses the official remote thumbnail plus Twitch metadata and
   returns a validated structured result.
2. A meaningful Twitch title is preserved even when the model proposes another.
3. A low-quality Twitch title can be replaced by a high-confidence,
   evidence-bounded candidate.
4. Human titles are never overwritten by ingestion or enrichment.
5. Failed vision attempts remain pending; successful attempts are completed.
6. The scheduler no longer downloads Twitch video or invokes Whisper/FFmpeg.
