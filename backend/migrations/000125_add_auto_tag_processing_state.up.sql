ALTER TABLE clips
    ADD COLUMN IF NOT EXISTS auto_tagged_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_clips_auto_tag_queue
    ON clips (imported_at)
    WHERE auto_tagged_at IS NULL AND is_removed = FALSE;
