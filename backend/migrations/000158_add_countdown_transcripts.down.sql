DROP INDEX IF EXISTS clip_transcripts_search_pending_idx;
ALTER TABLE clip_transcripts DROP COLUMN IF EXISTS search_indexed_at;
DROP TABLE IF EXISTS transcription_usage_daily;
DELETE FROM clip_transcripts WHERE source = 'countdown_render_whisper';
ALTER TABLE clip_transcripts DROP CONSTRAINT IF EXISTS clip_transcripts_source_check;
ALTER TABLE clip_transcripts ADD CONSTRAINT clip_transcripts_source_check
    CHECK (source = 'twitch_authorized_whisper');
