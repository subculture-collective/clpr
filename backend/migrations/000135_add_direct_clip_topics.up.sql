ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS merged_into_id UUID REFERENCES categories(id) ON DELETE SET NULL;

INSERT INTO categories (name, slug, description, position, category_type, is_featured, is_custom)
VALUES
    ('News & Politics', 'news-politics', 'News, public affairs, elections, and political commentary.', 10, 'topic', TRUE, FALSE),
    ('IRL & Travel', 'irl-travel', 'Real-world streams, travel, food, and life away from the desk.', 20, 'topic', TRUE, FALSE),
    ('Reactions & Commentary', 'reactions-commentary', 'Creator reactions, analysis, and commentary.', 30, 'topic', TRUE, FALSE),
    ('Gaming', 'gaming', 'Games, speedruns, esports, and gaming culture.', 40, 'topic', TRUE, FALSE),
    ('Music & Performance', 'music-performance', 'Music, concerts, singing, and live performance.', 50, 'topic', TRUE, FALSE),
    ('Sports', 'sports', 'Sports, matches, competition, and analysis.', 60, 'topic', TRUE, FALSE),
    ('Creative & Making', 'creative-making', 'Art, cooking, crafting, building, and creative process.', 70, 'topic', TRUE, FALSE),
    ('Tech', 'tech', 'Technology, coding, hardware, and science.', 80, 'topic', TRUE, FALSE),
    ('Culture & Drama', 'culture-drama', 'Internet culture, creator stories, and unfolding drama.', 90, 'topic', TRUE, FALSE)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    position = EXCLUDED.position,
    category_type = 'topic',
    is_featured = TRUE,
    is_active = TRUE,
    merged_into_id = NULL,
    updated_at = NOW();

CREATE TABLE clip_topics (
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    topic_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    source VARCHAR(40) NOT NULL CHECK (source IN ('twitch_category', 'metadata', 'tag', 'transcript', 'manual', 'backfill')),
    confidence NUMERIC(4,3) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    assigned_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (clip_id, topic_id)
);

CREATE INDEX clip_topics_topic_confidence_idx ON clip_topics(topic_id, confidence DESC, created_at DESC);
CREATE INDEX clip_topics_clip_idx ON clip_topics(clip_id, confidence DESC);
ALTER TABLE clips ADD COLUMN IF NOT EXISTS topics_classified_at TIMESTAMPTZ;
CREATE INDEX clips_topics_unclassified_idx ON clips(imported_at)
    WHERE topics_classified_at IS NULL AND is_removed = FALSE AND is_hidden = FALSE;

INSERT INTO clip_topics (clip_id, topic_id, source, confidence, evidence)
SELECT c.id, cg.category_id, 'backfill', 0.550, jsonb_build_object('twitch_category_id', c.game_id)
FROM clips c
JOIN games g ON g.twitch_game_id = c.game_id
JOIN category_games cg ON cg.game_id = g.id
WHERE c.is_removed = FALSE AND c.is_hidden = FALSE
ON CONFLICT (clip_id, topic_id) DO NOTHING;
