-- Canonical tag taxonomy, safe glob blacklist matching, legacy aliases, and
-- bounded score refresh staging.

CREATE OR REPLACE FUNCTION tag_blacklist_glob_to_like(glob TEXT)
RETURNS TEXT
LANGUAGE plpgsql
IMMUTABLE
STRICT
AS $$
DECLARE
    result TEXT := '';
    ch TEXT;
    i INTEGER;
BEGIN
    FOR i IN 1..char_length(glob) LOOP
        ch := substr(glob, i, 1);
        result := result || CASE ch
            WHEN '*' THEN '%'
            WHEN '?' THEN '_'
            WHEN '%' THEN '\%'
            WHEN '_' THEN '\_'
            WHEN '\' THEN '\\'
            ELSE ch
        END;
    END LOOP;
    RETURN result;
END;
$$;

CREATE OR REPLACE FUNCTION is_tag_blacklisted(candidate_slug TEXT)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
STRICT
AS $$
    SELECT EXISTS (
        SELECT 1
        FROM blacklisted_tags
        WHERE candidate_slug ILIKE tag_blacklist_glob_to_like(pattern) ESCAPE '\'
    );
$$;

INSERT INTO tags (id, name, slug, parent_slug, usage_count, created_at) VALUES
    (gen_random_uuid(), 'Taxonomy: Game', 'game', NULL, 0, NOW()),
    (gen_random_uuid(), 'Taxonomy: Duration', 'duration', NULL, 0, NOW()),
    (gen_random_uuid(), 'Taxonomy: Language', 'lang', NULL, 0, NOW()),
    (gen_random_uuid(), 'Taxonomy: Content', 'content', NULL, 0, NOW()),
    (gen_random_uuid(), 'Taxonomy: Community', 'community', NULL, 0, NOW())
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tags (id, name, slug, parent_slug, usage_count, created_at)
SELECT gen_random_uuid(), seed.name, seed.slug, seed.parent_slug, 0, NOW()
FROM (VALUES
    ('Short (0-30s)', 'duration/short', 'duration'),
    ('Medium (31-90s)', 'duration/medium', 'duration'),
    ('Long (90s+)', 'duration/long', 'duration'),
    ('Language: English', 'lang/en', 'lang'),
    ('Language: Spanish', 'lang/es', 'lang'),
    ('Language: Portuguese', 'lang/pt', 'lang'),
    ('Language: French', 'lang/fr', 'lang'),
    ('Language: German', 'lang/de', 'lang'),
    ('Language: Russian', 'lang/ru', 'lang'),
    ('Language: Japanese', 'lang/ja', 'lang'),
    ('Language: Korean', 'lang/ko', 'lang'),
    ('Language: Chinese', 'lang/zh', 'lang'),
    ('Language: Italian', 'lang/it', 'lang'),
    ('Language: Turkish', 'lang/tr', 'lang'),
    ('Language: Arabic', 'lang/ar', 'lang'),
    ('Language: Other', 'lang/other', 'lang'),
    ('Content: Ace', 'content/ace', 'content'),
    ('Content: Clutch', 'content/clutch', 'content'),
    ('Content: Fail', 'content/fail', 'content'),
    ('Content: Rage', 'content/rage', 'content'),
    ('Content: Funny', 'content/funny', 'content'),
    ('Content: Insane', 'content/insane', 'content'),
    ('Content: Lucky', 'content/lucky', 'content'),
    ('Content: Bug', 'content/bug', 'content'),
    ('Content: Toxic', 'content/toxic', 'content'),
    ('Content: Epic', 'content/epic', 'content'),
    ('Content: Noob', 'content/noob', 'content'),
    ('Content: Pro', 'content/pro', 'content'),
    ('Content: Highlight', 'content/highlight', 'content'),
    ('Content: Highlights', 'content/highlights', 'content'),
    ('Content: Educational', 'content/educational', 'content'),
    ('Content: Reaction', 'content/reaction', 'content'),
    ('Content: Music', 'content/music', 'content'),
    ('Content: Creative', 'content/creative', 'content'),
    ('Content: IRL', 'content/irl', 'content'),
    ('Content: Speedrun', 'content/speedrun', 'content'),
    ('Content: Tutorial', 'content/tutorial', 'content')
) AS seed(name, slug, parent_slug)
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE tag_aliases (
    alias_slug VARCHAR(100) PRIMARY KEY,
    canonical_tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    retired_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX tag_aliases_canonical_tag_id_idx ON tag_aliases(canonical_tag_id);

CREATE UNLOGGED TABLE clip_score_refresh_stage (
    clip_id UUID PRIMARY KEY REFERENCES clips(id) ON DELETE CASCADE,
    trending_score DOUBLE PRECISION NOT NULL,
    hot_score DOUBLE PRECISION NOT NULL,
    popularity_index INTEGER NOT NULL,
    engagement_count INTEGER NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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
  AND NOT is_tag_blacklisted(t.slug)
  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id = t.id)
  AND NOT EXISTS (
      SELECT 1 FROM tag_promotion_queue q
      WHERE q.tag_slug = t.slug AND q.status IN ('approved', 'rejected')
  )
GROUP BY t.slug, t.name, t.parent_slug
HAVING COUNT(DISTINCT c.submitted_by_user_id) >= 3
   AND COUNT(DISTINCT ct.clip_id) >= 5;
