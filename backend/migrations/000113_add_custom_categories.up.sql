-- Add custom topic categories alongside existing game categories
-- Use ON CONFLICT on both name and slug to handle duplicates gracefully
INSERT INTO categories (name, slug, description, icon, position, category_type, is_custom, is_featured)
VALUES
    ('News', 'news', 'Breaking news and current events from streams', 'newspaper', 100, 'topic', true, true),
    ('Politics', 'politics', 'Political commentary and discussion', 'landmark', 101, 'topic', true, true),
    ('Drama', 'drama', 'Streamer drama and community moments', 'flame', 102, 'topic', true, true),
    ('IRL', 'irl', 'Real life streaming moments', 'camera', 103, 'topic', true, true),
    ('Esports', 'esports', 'Competitive gaming and tournaments', 'trophy', 105, 'topic', true, true),
    ('Highlights', 'highlights', 'Best plays and standout moments', 'star', 106, 'topic', true, true),
    ('Fails', 'fails', 'Epic fails and funny mistakes', 'skull', 107, 'topic', true, false)
ON CONFLICT (name) DO UPDATE SET
    category_type = EXCLUDED.category_type,
    is_custom = EXCLUDED.is_custom,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    position = EXCLUDED.position,
    slug = EXCLUDED.slug;

-- Music and Creative already exist as game categories — update them to also be topic categories
UPDATE categories SET
    category_type = 'topic',
    is_custom = true,
    is_featured = false,
    description = 'Music performances and reactions',
    position = 104
WHERE name = 'Music';

UPDATE categories SET
    category_type = 'topic',
    is_custom = true,
    is_featured = false,
    description = 'Art, music production, and creative streams',
    position = 108
WHERE name = 'Creative';
