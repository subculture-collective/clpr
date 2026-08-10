-- Add multi-tag support to playlist_scripts
ALTER TABLE playlist_scripts ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
ALTER TABLE playlist_scripts ADD COLUMN IF NOT EXISTS tags_logic VARCHAR(3) DEFAULT 'and';

-- Migrate existing single tag field into the new tags array
UPDATE playlist_scripts SET tags = ARRAY[tag] WHERE tag IS NOT NULL AND cardinality(tags) = 0;

-- Index for tag-based queries
CREATE INDEX IF NOT EXISTS idx_playlist_scripts_tags ON playlist_scripts USING GIN (tags);