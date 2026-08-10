-- Migration: 000120_add_tag_parent_slug
-- Add hierarchical parent_slug to tags table

ALTER TABLE tags ADD COLUMN IF NOT EXISTS parent_slug VARCHAR(100);
CREATE INDEX IF NOT EXISTS idx_tags_parent_slug ON tags(parent_slug);
CREATE INDEX IF NOT EXISTS idx_tags_slug_parent ON tags(slug, parent_slug);

-- Note: No FK constraint on parent_slug -> tags(slug) is added here
-- to avoid circular dependency during insert. Parent validation is
-- performed at the application layer.