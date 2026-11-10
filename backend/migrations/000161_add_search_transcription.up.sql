-- Transcribing public clips for search (ADR-003, extending ADR-002). The
-- render host downloads public clips above a view threshold, transcribes
-- them and deletes the file at once; only the transcript is kept. Its
-- transcripts and their cost have their own source.
ALTER TABLE clip_transcripts DROP CONSTRAINT IF EXISTS clip_transcripts_source_check;
ALTER TABLE clip_transcripts ADD CONSTRAINT clip_transcripts_source_check
    CHECK (source IN ('twitch_authorized_whisper', 'countdown_render_whisper', 'search_render_whisper'));

ALTER TABLE transcription_usage_daily DROP CONSTRAINT IF EXISTS transcription_usage_daily_source_check;
ALTER TABLE transcription_usage_daily ADD CONSTRAINT transcription_usage_daily_source_check
    CHECK (source IN ('twitch_authorized_whisper', 'countdown_render_whisper', 'search_render_whisper'));

-- The last attempt per clip, so a clip without English speech is not
-- downloaded again and a failed one is retried a few times only.
CREATE TABLE search_transcription_attempts (
    clip_id UUID PRIMARY KEY REFERENCES clips(id) ON DELETE CASCADE,
    outcome TEXT NOT NULL CHECK (outcome IN ('transcribed', 'no_speech', 'failed')),
    attempts INT NOT NULL DEFAULT 1 CHECK (attempts > 0),
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
