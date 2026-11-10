DROP INDEX IF EXISTS streamer_clip_room_items_room_approved_idx;
ALTER TABLE countdown_series DROP COLUMN IF EXISTS clip_room_id, DROP COLUMN IF EXISTS source;
