DROP TABLE IF EXISTS clip_enrichments;
DROP INDEX IF EXISTS idx_clips_vision_queue;
DROP INDEX IF EXISTS idx_clips_structural_tag_queue;

ALTER TABLE clips
    DROP CONSTRAINT IF EXISTS clips_title_source_check,
    DROP COLUMN IF EXISTS vision_error,
    DROP COLUMN IF EXISTS vision_attempt_count,
    DROP COLUMN IF EXISTS vision_attempted_at,
    DROP COLUMN IF EXISTS vision_processed_at,
    DROP COLUMN IF EXISTS structural_tagged_at,
    DROP COLUMN IF EXISTS title_source;

CREATE INDEX IF NOT EXISTS idx_clips_auto_tag_queue
    ON clips (imported_at)
    WHERE auto_tagged_at IS NULL AND is_removed = FALSE;
