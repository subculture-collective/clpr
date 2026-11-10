UPDATE countdown_edition_media SET source = 'player' WHERE source = 'vod';
ALTER TABLE countdown_edition_media DROP CONSTRAINT IF EXISTS countdown_edition_media_source_check;
ALTER TABLE countdown_edition_media ADD CONSTRAINT countdown_edition_media_source_check
    CHECK (source IN ('helix', 'player'));
ALTER TABLE clips DROP COLUMN IF EXISTS twitch_vod_offset, DROP COLUMN IF EXISTS twitch_video_id;
