-- Remove multi-tag support from playlist_scripts
DROP INDEX IF EXISTS idx_playlist_scripts_tags;

ALTER TABLE playlist_scripts
    DROP COLUMN IF EXISTS tags_logic,
    DROP COLUMN IF EXISTS tags;