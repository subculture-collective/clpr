-- Seed: structural and content tag hierarchy
-- Run: psql $DATABASE_URL -f backend/scripts/seed_tag_hierarchy.sql
-- Idempotent: safe to run multiple times (ON CONFLICT DO NOTHING)

-- Root tags (structural + content + community)
INSERT INTO tags (id, name, slug, parent_slug) VALUES
  (gen_random_uuid(), 'Game', 'game', NULL),
  (gen_random_uuid(), 'Duration', 'duration', NULL),
  (gen_random_uuid(), 'Language', 'lang', NULL),
  (gen_random_uuid(), 'Broadcaster Tier', 'tier', NULL)
ON CONFLICT (slug) DO NOTHING;

-- Duration children
INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'duration'
  FROM (VALUES
    ('Short (0-30s)', 'duration/short'),
    ('Medium (30-90s)', 'duration/medium'),
    ('Long (90s+)', 'duration/long')
  ) AS v(name, slug)
ON CONFLICT (slug) DO NOTHING;

-- Language children (common Twitch languages)
INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'lang'
  FROM (VALUES
    ('English', 'lang/en'),
    ('Spanish', 'lang/es'),
    ('Portuguese', 'lang/pt'),
    ('French', 'lang/fr'),
    ('German', 'lang/de'),
    ('Russian', 'lang/ru'),
    ('Japanese', 'lang/ja'),
    ('Korean', 'lang/ko')
  ) AS v(name, slug)
ON CONFLICT (slug) DO NOTHING;

-- Broadcaster tier children
INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'tier'
  FROM (VALUES
    ('Partner', 'tier/partner'),
    ('Affiliate', 'tier/affiliate'),
    ('Non-affiliate', 'tier/non-affiliate')
  ) AS v(name, slug)
ON CONFLICT (slug) DO NOTHING;

-- Content tags: assigned by AI (roots for future children)
INSERT INTO tags (id, name, slug, parent_slug) VALUES
  (gen_random_uuid(), 'Content', 'content', NULL)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'content'
  FROM (VALUES
    ('Clutch Play', 'content/clutch'),
    ('Funny Moment', 'content/funny'),
    ('Fail', 'content/fail'),
    ('Educational', 'content/educational'),
    ('Highlights', 'content/highlights'),
    ('Reaction', 'content/reaction'),
    ('Speedrun', 'content/speedrun'),
    ('Music', 'content/music'),
    ('Art/Creative', 'content/creative'),
    ('IRL/Just Chatting', 'content/irl')
  ) AS v(name, slug)
ON CONFLICT (slug) DO NOTHING;

-- Community tags root (user-promoted tags go here)
INSERT INTO tags (id, name, slug, parent_slug) VALUES
  (gen_random_uuid(), 'Community', 'community', NULL)
ON CONFLICT (slug) DO NOTHING;