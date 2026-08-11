DROP INDEX IF EXISTS idx_clips_auto_tag_queue;
ALTER TABLE clips DROP COLUMN IF EXISTS auto_tagged_at;
