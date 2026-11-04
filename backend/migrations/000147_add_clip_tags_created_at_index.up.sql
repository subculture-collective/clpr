-- Migration: Index clip_tags by assignment time
-- Description: The trending tag list aggregates the last seven days of tag
-- assignments. Without this index that is a sequential scan of every
-- clip_tags row (574k on production, about 2 s). The index covers the
-- aggregation columns so the window can be read without touching the heap.

CREATE INDEX IF NOT EXISTS idx_clip_tags_created_at
    ON clip_tags (created_at, tag_id, clip_id);
