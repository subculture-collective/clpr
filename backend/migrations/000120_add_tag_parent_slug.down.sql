DROP INDEX IF EXISTS idx_tags_slug_parent;
DROP INDEX IF EXISTS idx_tags_parent_slug;
ALTER TABLE tags DROP COLUMN IF EXISTS parent_slug;