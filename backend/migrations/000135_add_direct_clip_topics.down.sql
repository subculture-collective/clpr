DROP INDEX IF EXISTS clips_topics_unclassified_idx;
ALTER TABLE clips DROP COLUMN IF EXISTS topics_classified_at;
DROP TABLE IF EXISTS clip_topics;
ALTER TABLE categories DROP COLUMN IF EXISTS merged_into_id, DROP COLUMN IF EXISTS is_active;
