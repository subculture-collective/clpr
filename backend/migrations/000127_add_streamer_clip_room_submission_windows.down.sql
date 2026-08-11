DROP INDEX IF EXISTS streamer_clip_rooms_submission_window_idx;

ALTER TABLE streamer_clip_rooms
    DROP COLUMN IF EXISTS submissions_close_at,
    DROP COLUMN IF EXISTS submissions_open;
