-- Migration: 000123_add_tag_promotion_queue
-- Add user tag promotion queue with threshold gate for moderator review

CREATE TABLE IF NOT EXISTS tag_promotion_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_slug VARCHAR(100) NOT NULL,
    usage_count INTEGER NOT NULL DEFAULT 0,
    unique_users INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    promoted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Prevent duplicate pending entries for the same tag
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_promotion_slug_status
    ON tag_promotion_queue(tag_slug) WHERE status = 'pending';

-- View: tags that have crossed the promotion threshold
-- Community tags (parent_slug IS NULL or 'community') are candidates when
-- used by ≥3 unique submitters on ≥5 distinct clips.
CREATE OR REPLACE VIEW tag_promotion_candidates AS
SELECT
    t.slug,
    t.name,
    COUNT(DISTINCT ct.clip_id) AS clip_count,
    COUNT(DISTINCT c.submitted_by_user_id) AS unique_users,
    t.parent_slug
FROM clip_tags ct
JOIN tags t ON t.id = ct.tag_id
JOIN clips c ON c.id = ct.clip_id
WHERE (t.parent_slug IS NULL OR t.parent_slug = 'community')
  AND c.submitted_by_user_id IS NOT NULL
GROUP BY t.slug, t.name, t.parent_slug
HAVING COUNT(DISTINCT c.submitted_by_user_id) >= 3
   AND COUNT(DISTINCT ct.clip_id) >= 5;