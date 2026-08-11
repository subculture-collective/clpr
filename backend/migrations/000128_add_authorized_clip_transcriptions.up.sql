ALTER TABLE clips
    ADD COLUMN transcription_processed_at TIMESTAMPTZ,
    ADD COLUMN transcription_attempted_at TIMESTAMPTZ,
    ADD COLUMN transcription_attempt_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN transcription_error TEXT;

CREATE TABLE clip_transcripts (
    clip_id UUID PRIMARY KEY REFERENCES clips(id) ON DELETE CASCADE,
    language TEXT,
    full_text TEXT NOT NULL,
    segments JSONB NOT NULL DEFAULT '[]'::jsonb,
    source TEXT NOT NULL DEFAULT 'twitch_authorized_whisper',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (source = 'twitch_authorized_whisper')
);

ALTER TABLE clip_enrichments
    DROP CONSTRAINT IF EXISTS clip_enrichments_basis_check;
ALTER TABLE clip_enrichments
    ADD CONSTRAINT clip_enrichments_basis_check
    CHECK (basis IN ('source_title', 'transcript', 'visible', 'metadata', 'insufficient'));

CREATE INDEX idx_clips_transcription_queue
    ON clips (created_at DESC)
    WHERE transcription_processed_at IS NULL
      AND broadcaster_id IS NOT NULL
      AND is_removed = FALSE;
