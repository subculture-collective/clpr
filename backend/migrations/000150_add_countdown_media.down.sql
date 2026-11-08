DROP TABLE IF EXISTS countdown_edition_skips;
DROP TABLE IF EXISTS countdown_edition_media;
DROP INDEX IF EXISTS countdown_editions_status_idx;
UPDATE countdown_editions SET status = 'selected' WHERE status = 'fetched';
ALTER TABLE countdown_editions DROP CONSTRAINT countdown_editions_status_check;
ALTER TABLE countdown_editions ADD CONSTRAINT countdown_editions_status_check
    CHECK (status IN ('selected', 'fetching', 'rendering', 'rendered', 'published', 'failed'));
