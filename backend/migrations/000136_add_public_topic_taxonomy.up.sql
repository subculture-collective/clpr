ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE categories
SET is_public = slug IN (
        'news-politics', 'irl-travel', 'reactions-commentary', 'gaming',
        'music-performance', 'sports', 'creative-making', 'tech', 'culture-drama'
    ),
    icon = CASE WHEN slug = 'sports' THEN '⚽' ELSE icon END,
    description = CASE WHEN slug = 'sports'
        THEN 'Sports, matches, competition, and analysis.' ELSE description END
WHERE category_type = 'topic';

WITH redirects(source_slug, target_slug) AS (
    VALUES
        ('news', 'news-politics'), ('politics', 'news-politics'),
        ('irl', 'irl-travel'), ('drama', 'culture-drama'),
        ('music', 'music-performance'), ('creative', 'creative-making'),
        ('esports', 'gaming'), ('highlights', 'gaming'),
        ('fails', 'reactions-commentary'), ('variety', 'reactions-commentary'),
        ('just-chatting', 'reactions-commentary')
)
UPDATE categories source
SET merged_into_id = target.id, is_public = FALSE, is_featured = FALSE
FROM redirects r
JOIN categories target ON target.slug = r.target_slug
WHERE source.slug = r.source_slug AND source.id <> target.id;
