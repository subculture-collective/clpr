DELETE FROM countdown_editions WHERE status = 'skipped';
ALTER TABLE countdown_editions DROP CONSTRAINT countdown_editions_status_check;
ALTER TABLE countdown_editions ADD CONSTRAINT countdown_editions_status_check
    CHECK (status IN ('selected', 'fetching', 'fetched', 'rendering', 'rendered', 'publishing', 'published', 'failed'));

ALTER TABLE countdown_series DROP COLUMN min_size;
