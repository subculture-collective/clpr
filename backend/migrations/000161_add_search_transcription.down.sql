DROP TABLE IF EXISTS search_transcription_attempts;
DELETE FROM transcription_usage_daily WHERE source = 'search_render_whisper';
DELETE FROM clip_transcripts WHERE source = 'search_render_whisper';
ALTER TABLE transcription_usage_daily DROP CONSTRAINT IF EXISTS transcription_usage_daily_source_check;
ALTER TABLE transcription_usage_daily ADD CONSTRAINT transcription_usage_daily_source_check
    CHECK (source IN ('twitch_authorized_whisper', 'countdown_render_whisper'));
ALTER TABLE clip_transcripts DROP CONSTRAINT IF EXISTS clip_transcripts_source_check;
ALTER TABLE clip_transcripts ADD CONSTRAINT clip_transcripts_source_check
    CHECK (source IN ('twitch_authorized_whisper', 'countdown_render_whisper'));
