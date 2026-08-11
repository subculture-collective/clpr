ALTER TABLE clips
    ADD COLUMN previous_view_count INTEGER,
    ADD COLUMN view_count_observed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN view_velocity DOUBLE PRECISION NOT NULL DEFAULT 0;

UPDATE clips
SET view_count_observed_at = COALESCE(imported_at, created_at, NOW());

COMMENT ON COLUMN clips.previous_view_count IS
    'Twitch view count immediately before the latest increasing observation';
COMMENT ON COLUMN clips.view_count_observed_at IS
    'Time the current Twitch view count was first observed';
COMMENT ON COLUMN clips.view_velocity IS
    'Views gained per hour between the two latest increasing observations';
