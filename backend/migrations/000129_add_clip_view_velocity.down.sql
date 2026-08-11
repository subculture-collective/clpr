ALTER TABLE clips
    DROP COLUMN IF EXISTS view_velocity,
    DROP COLUMN IF EXISTS view_count_observed_at,
    DROP COLUMN IF EXISTS previous_view_count;
