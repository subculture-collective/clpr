-- Helix returns, for each clip, the past broadcast (VOD) it came from and
-- where in it the clip starts. Both change: the offset is empty until Twitch
-- processes the VOD, and both go empty when the VOD is deleted. Every import
-- and refresh stores the current values.
ALTER TABLE clips
    ADD COLUMN twitch_video_id TEXT CHECK (twitch_video_id IS NULL OR twitch_video_id ~ '^[0-9]{1,20}$'),
    ADD COLUMN twitch_vod_offset INT CHECK (twitch_vod_offset IS NULL OR twitch_vod_offset >= 0);

-- Countdown media may come from the VOD, with a few seconds of the
-- broadcast before and after the clip (ADR-004).
ALTER TABLE countdown_edition_media DROP CONSTRAINT IF EXISTS countdown_edition_media_source_check;
ALTER TABLE countdown_edition_media ADD CONSTRAINT countdown_edition_media_source_check
    CHECK (source IN ('helix', 'player', 'vod'));
