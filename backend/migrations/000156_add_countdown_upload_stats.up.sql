-- YouTube view counts for countdown uploads (#84). The countdown worker
-- refreshes them with videos.list. youtube_privacy_status is the video's
-- privacy on YouTube now, which differs from privacy_status once a scheduled
-- upload goes public. youtube_missing marks a video YouTube no longer returns.
ALTER TABLE countdown_uploads
    ADD COLUMN view_count BIGINT CHECK (view_count IS NULL OR view_count >= 0),
    ADD COLUMN youtube_privacy_status TEXT
        CHECK (youtube_privacy_status IS NULL OR youtube_privacy_status IN ('private', 'unlisted', 'public')),
    ADD COLUMN youtube_missing BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN stats_refreshed_at TIMESTAMPTZ;

CREATE INDEX countdown_uploads_stats_idx ON countdown_uploads (stats_refreshed_at NULLS FIRST, uploaded_at)
    WHERE youtube_missing = FALSE;
