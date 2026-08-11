ALTER TABLE streamer_clip_rooms
    ADD COLUMN submissions_open BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN submissions_close_at TIMESTAMPTZ;

CREATE INDEX streamer_clip_rooms_submission_window_idx
    ON streamer_clip_rooms (submissions_close_at)
    WHERE submissions_open = true AND submissions_close_at IS NOT NULL;
