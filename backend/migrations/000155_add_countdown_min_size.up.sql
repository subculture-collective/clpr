-- Lets a series publish a shorter list on a thin day. With min_size set, an
-- edition takes up to size clips, and a day with fewer than min_size is
-- stored as "skipped" instead of failing. NULL keeps the strict behavior:
-- exactly size clips or a failed edition.
ALTER TABLE countdown_series ADD COLUMN min_size INT
    CHECK (min_size IS NULL OR (min_size >= 1 AND min_size <= size));

ALTER TABLE countdown_editions DROP CONSTRAINT countdown_editions_status_check;
ALTER TABLE countdown_editions ADD CONSTRAINT countdown_editions_status_check
    CHECK (status IN ('selected', 'fetching', 'fetched', 'rendering', 'rendered', 'publishing', 'published', 'failed', 'skipped'));
