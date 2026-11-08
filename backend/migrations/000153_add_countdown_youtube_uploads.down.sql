DROP TABLE IF EXISTS countdown_youtube_playlists;
DROP TABLE IF EXISTS countdown_uploads;
UPDATE countdown_editions SET status = 'rendered' WHERE status = 'publishing';
ALTER TABLE countdown_editions DROP CONSTRAINT countdown_editions_status_check;
ALTER TABLE countdown_editions ADD CONSTRAINT countdown_editions_status_check
    CHECK (status IN ('selected', 'fetching', 'fetched', 'rendering', 'rendered', 'published', 'failed'));
