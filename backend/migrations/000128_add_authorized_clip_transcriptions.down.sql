DROP INDEX IF EXISTS idx_clips_transcription_queue;

ALTER TABLE clip_enrichments
    DROP CONSTRAINT IF EXISTS clip_enrichments_basis_check;
ALTER TABLE clip_enrichments
    ADD CONSTRAINT clip_enrichments_basis_check
    CHECK (basis IN ('source_title', 'visible', 'metadata', 'insufficient'));

DROP TABLE IF EXISTS clip_transcripts;

ALTER TABLE clips
    DROP COLUMN IF EXISTS transcription_error,
    DROP COLUMN IF EXISTS transcription_attempt_count,
    DROP COLUMN IF EXISTS transcription_attempted_at,
    DROP COLUMN IF EXISTS transcription_processed_at;
