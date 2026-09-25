-- Migration: Remove advanced query performance indexes
-- Description: Drops indexes created for advanced query operations
-- Note: golang-migrate sends each file as one multi-statement query, which
-- PostgreSQL runs as a single transaction block, so DROP INDEX CONCURRENTLY is
-- rejected here. See 000020 for the same convention.

-- Drop pagination support indexes
DROP INDEX IF EXISTS idx_clips_pagination;

-- Drop user search indexes
DROP INDEX IF EXISTS idx_users_karma;
DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_users_search_vector;

-- Drop clip-tags indexes
DROP INDEX IF EXISTS idx_clip_tags_tag_id;
DROP INDEX IF EXISTS idx_clip_tags_clip_id;

-- Drop tags indexes
DROP INDEX IF EXISTS idx_tags_usage_count;
DROP INDEX IF EXISTS idx_tags_slug;

-- Drop equality indexes
DROP INDEX IF EXISTS idx_clips_is_nsfw;
DROP INDEX IF EXISTS idx_clips_is_featured;

-- Drop composite filtering indexes
DROP INDEX IF EXISTS idx_clips_active_by_language;
DROP INDEX IF EXISTS idx_clips_active_by_creator;
DROP INDEX IF EXISTS idx_clips_active_by_game;

-- Drop range query indexes
DROP INDEX IF EXISTS idx_clips_view_count;
DROP INDEX IF EXISTS idx_clips_created_at;
DROP INDEX IF EXISTS idx_clips_vote_score;

-- Drop full-text search indexes
DROP INDEX IF EXISTS idx_clips_search_vector;
