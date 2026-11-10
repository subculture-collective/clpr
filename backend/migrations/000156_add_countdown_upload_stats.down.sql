DROP INDEX IF EXISTS countdown_uploads_stats_idx;
ALTER TABLE countdown_uploads
    DROP COLUMN IF EXISTS stats_refreshed_at,
    DROP COLUMN IF EXISTS youtube_missing,
    DROP COLUMN IF EXISTS youtube_privacy_status,
    DROP COLUMN IF EXISTS view_count;
