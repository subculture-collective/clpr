-- Transcript search (#87). The render host transcribes every countdown clip
-- while its media is on hand, not only the clips that get a Short, and
-- stores the English transcript for search. A transcript from the
-- broadcaster's own authorization is kept over one from the render host.
ALTER TABLE clip_transcripts DROP CONSTRAINT IF EXISTS clip_transcripts_source_check;
ALTER TABLE clip_transcripts ADD CONSTRAINT clip_transcripts_source_check
    CHECK (source IN ('twitch_authorized_whisper', 'countdown_render_whisper'));

-- Whisper runs on clpr's own hosts, so its cost is compute time. One row per
-- day and source adds up the clips, their audio length and the time spent.
CREATE TABLE transcription_usage_daily (
    day DATE NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('twitch_authorized_whisper', 'countdown_render_whisper')),
    clips INT NOT NULL DEFAULT 0 CHECK (clips >= 0),
    transcripts INT NOT NULL DEFAULT 0 CHECK (transcripts >= 0),
    audio_seconds DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (audio_seconds >= 0),
    compute_seconds DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (compute_seconds >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (day, source)
);

-- The API copies new and changed transcripts into the clips search index;
-- search_indexed_at is the transcript version it copied last.
ALTER TABLE clip_transcripts ADD COLUMN search_indexed_at TIMESTAMPTZ;
CREATE INDEX clip_transcripts_search_pending_idx ON clip_transcripts (updated_at)
    WHERE search_indexed_at IS NULL OR search_indexed_at < updated_at;
