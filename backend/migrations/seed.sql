-- =============================================================================
-- CLIPPER - Ultimate Database Seed Script
-- =============================================================================
-- Populates ALL tables with realistic sample data so every page works.
-- Safe to run multiple times (uses ON CONFLICT DO NOTHING / upserts).
-- Run: make migrate-seed
-- =============================================================================

BEGIN;

-- ============================================================
-- 0. HELPER: grab existing user & clip IDs for FK references
-- ============================================================
-- We'll reference existing users/clips plus create new ones.
-- Store first 20 existing user IDs for FK references.
CREATE TEMP TABLE _seed_users AS
SELECT id, username, row_number() OVER (ORDER BY created_at) AS rn
FROM users ORDER BY created_at LIMIT 20;

CREATE TEMP TABLE _seed_clips AS
SELECT id, twitch_clip_id, broadcaster_name, title, row_number() OVER (ORDER BY created_at) AS rn
FROM clips ORDER BY created_at LIMIT 40;

-- ============================================================
-- 1. ENRICH EXISTING USERS - give them karma, roles, avatars
-- ============================================================

-- Make first user an admin
UPDATE users SET role = 'admin', karma_points = 15000,
    display_name = COALESCE(display_name, initcap(username)),
    avatar_url = COALESCE(avatar_url, 'https://api.dicebear.com/7.x/avataaars/svg?seed=' || username)
WHERE id = (SELECT id FROM _seed_users WHERE rn = 1);

-- Make users 2-3 moderators
UPDATE users SET role = 'moderator', karma_points = 7500 + (random() * 2500)::int,
    display_name = COALESCE(display_name, initcap(username)),
    avatar_url = COALESCE(avatar_url, 'https://api.dicebear.com/7.x/avataaars/svg?seed=' || username)
WHERE id IN (SELECT id FROM _seed_users WHERE rn IN (2, 3));

-- Give all users karma, display names, and avatars
UPDATE users SET
    karma_points = CASE
        WHEN karma_points > 0 THEN karma_points  -- preserve existing
        ELSE (random() * 12000)::int
    END,
    display_name = COALESCE(display_name, initcap(username)),
    avatar_url = COALESCE(avatar_url, 'https://api.dicebear.com/7.x/avataaars/svg?seed=' || username),
    bio = COALESCE(bio, CASE (random() * 5)::int
        WHEN 0 THEN 'Avid clip collector and Twitch enthusiast 🎮'
        WHEN 1 THEN 'Just here for the highlights!'
        WHEN 2 THEN 'Streaming is life. Clips are the memories.'
        WHEN 3 THEN 'Top-tier clip curator since day one.'
        WHEN 4 THEN 'I find the best moments so you don''t have to.'
        ELSE 'Gaming clips connoisseur 🎯'
    END),
    social_links = COALESCE(
        NULLIF(social_links::text, '{}')::jsonb,
        jsonb_build_object('twitch', 'https://twitch.tv/' || username)
    ),
    last_login_at = COALESCE(last_login_at, NOW() - (random() * 30 || ' days')::interval)
WHERE karma_points = 0 OR display_name IS NULL OR avatar_url IS NULL;

-- ============================================================
-- 2. ENRICH EXISTING CLIPS - add game data, vote scores, views
-- ============================================================

-- Add game data to clips that lack it
UPDATE clips SET
    game_name = CASE (random() * 9)::int
        WHEN 0 THEN 'Just Chatting'
        WHEN 1 THEN 'Fortnite'
        WHEN 2 THEN 'League of Legends'
        WHEN 3 THEN 'Valorant'
        WHEN 4 THEN 'Minecraft'
        WHEN 5 THEN 'Grand Theft Auto V'
        WHEN 6 THEN 'Counter-Strike 2'
        WHEN 7 THEN 'Apex Legends'
        WHEN 8 THEN 'Overwatch 2'
        ELSE 'Call of Duty: Warzone'
    END,
    game_id = CASE (random() * 9)::int
        WHEN 0 THEN '509658'
        WHEN 1 THEN '33214'
        WHEN 2 THEN '21779'
        WHEN 3 THEN '516575'
        WHEN 4 THEN '27471'
        WHEN 5 THEN '32982'
        WHEN 6 THEN '32399'
        WHEN 7 THEN '511224'
        WHEN 8 THEN '515025'
        ELSE '512710'
    END,
    language = COALESCE(language, 'en'),
    duration = COALESCE(duration, 15 + (random() * 45)::int)
WHERE game_name IS NULL OR game_name = '';

-- Randomize clip vote scores and engagement
UPDATE clips SET
    vote_score = CASE
        WHEN vote_score != 0 THEN vote_score
        ELSE (random() * 500)::int - 50  -- most positive, some negative
    END,
    view_count = CASE
        WHEN view_count > 1000 THEN view_count
        ELSE 100 + (random() * 10000)::int
    END,
    engagement_count = (random() * 200)::int,
    trending_score = random() * 100,
    hot_score = random() * 50,
    popularity_index = (random() * 100)::int
WHERE vote_score = 0 OR engagement_count = 0;

-- Feature a few clips
UPDATE clips SET is_featured = true
WHERE id IN (SELECT id FROM clips ORDER BY random() LIMIT 5);

-- ============================================================
-- 3. USER STATS (feeds engagement_leaderboard view)
-- ============================================================

INSERT INTO user_stats (user_id, trust_score, engagement_score, total_comments, total_votes_cast, total_clips_submitted, correct_reports, incorrect_reports, days_active, last_active_date)
SELECT
    u.id,
    30 + (random() * 70)::int,                              -- trust_score 30-100
    (random() * 5000)::int,                                   -- engagement_score
    (random() * 200)::int,                                    -- total_comments
    (random() * 500)::int,                                    -- total_votes_cast
    (random() * 50)::int,                                     -- total_clips_submitted
    (random() * 20)::int,                                     -- correct_reports
    (random() * 5)::int,                                      -- incorrect_reports
    1 + (random() * 365)::int,                                -- days_active
    (NOW() - (random() * 7 || ' days')::interval)::date      -- last_active_date
FROM users u
ON CONFLICT (user_id) DO UPDATE SET
    engagement_score = EXCLUDED.engagement_score,
    total_comments = EXCLUDED.total_comments,
    total_votes_cast = EXCLUDED.total_votes_cast,
    total_clips_submitted = EXCLUDED.total_clips_submitted,
    days_active = EXCLUDED.days_active,
    last_active_date = EXCLUDED.last_active_date;

-- ============================================================
-- 4. VOTES - create engagement on clips
-- ============================================================

INSERT INTO votes (user_id, clip_id, vote_type, created_at)
SELECT
    u.id,
    c.id,
    CASE WHEN random() > 0.2 THEN 1 ELSE -1 END,
    NOW() - (random() * 60 || ' days')::interval
FROM users u
CROSS JOIN clips c
WHERE random() < 0.15  -- ~15% of all user-clip combos get a vote
ON CONFLICT (user_id, clip_id) DO NOTHING;

-- ============================================================
-- 5. COMMENTS - realistic threaded comments
-- ============================================================

-- Top-level comments
INSERT INTO comments (id, clip_id, user_id, content, vote_score, created_at)
SELECT
    gen_random_uuid(),
    c.id,
    u.id,
    CASE (random() * 14)::int
        WHEN 0 THEN 'This clip is absolutely insane! 🔥'
        WHEN 1 THEN 'How does this not have more upvotes?'
        WHEN 2 THEN 'Peak gaming moment right here'
        WHEN 3 THEN 'I was watching this live and my jaw dropped'
        WHEN 4 THEN 'This is why I love Twitch clips'
        WHEN 5 THEN 'The timing on this is unreal'
        WHEN 6 THEN 'Legendary play, nothing else to say'
        WHEN 7 THEN 'Can we talk about how clean this was?'
        WHEN 8 THEN 'The reaction makes this 10x better 😂'
        WHEN 9 THEN 'Saved for later, this is gold'
        WHEN 10 THEN 'This entire stream was content but THIS moment...'
        WHEN 11 THEN 'My ADHD brain replayed this 5 times'
        WHEN 12 THEN 'Actual content right here, not like the other clips'
        WHEN 13 THEN 'I need more clips from this streamer'
        ELSE 'GG well played 👏'
    END,
    (random() * 50)::int - 5,
    NOW() - (random() * 30 || ' days')::interval
FROM clips c
CROSS JOIN LATERAL (
    SELECT id FROM users ORDER BY random() LIMIT (2 + (random() * 4)::int)
) u;

-- Reply comments (grab some parent comments)
INSERT INTO comments (clip_id, user_id, parent_comment_id, content, vote_score, created_at)
SELECT
    pc.clip_id,
    u.id,
    pc.id,
    CASE (random() * 9)::int
        WHEN 0 THEN 'Totally agree with this ^'
        WHEN 1 THEN 'Facts!'
        WHEN 2 THEN 'Underrated comment right here'
        WHEN 3 THEN 'This is exactly what I was thinking'
        WHEN 4 THEN 'Nah you''re tripping, this was mid'
        WHEN 5 THEN 'W take honestly'
        WHEN 6 THEN 'I can''t stop laughing at this thread 😭'
        WHEN 7 THEN 'Someone finally said it'
        WHEN 8 THEN 'The real clip is always in the comments'
        ELSE 'Based'
    END,
    (random() * 20)::int - 3,
    pc.created_at + (random() * 48 || ' hours')::interval
FROM comments pc
CROSS JOIN LATERAL (
    SELECT id FROM users WHERE id != pc.user_id ORDER BY random() LIMIT (1 + (random() * 2)::int)
) u
WHERE pc.parent_comment_id IS NULL  -- only reply to top-level comments
AND random() < 0.6;

-- Update clip comment counts
UPDATE clips SET comment_count = sub.cnt
FROM (SELECT clip_id, count(*) as cnt FROM comments GROUP BY clip_id) sub
WHERE clips.id = sub.clip_id;

-- ============================================================
-- 6. COMMENT VOTES
-- ============================================================

INSERT INTO comment_votes (user_id, comment_id, vote_type, created_at)
SELECT
    u.id,
    c.id,
    CASE WHEN random() > 0.25 THEN 1 ELSE -1 END,
    c.created_at + (random() * 24 || ' hours')::interval
FROM comments c
CROSS JOIN LATERAL (
    SELECT id FROM users WHERE id != c.user_id ORDER BY random() LIMIT (1 + (random() * 3)::int)
) u
WHERE random() < 0.4
ON CONFLICT (user_id, comment_id) DO NOTHING;

-- ============================================================
-- 7. FAVORITES
-- ============================================================

INSERT INTO favorites (user_id, clip_id, created_at)
SELECT
    u.id,
    c.id,
    NOW() - (random() * 60 || ' days')::interval
FROM users u
CROSS JOIN clips c
WHERE random() < 0.08  -- ~8% of user-clip combos
ON CONFLICT (user_id, clip_id) DO NOTHING;

-- Update clip favorite counts
UPDATE clips SET favorite_count = sub.cnt
FROM (SELECT clip_id, count(*) as cnt FROM favorites GROUP BY clip_id) sub
WHERE clips.id = sub.clip_id;

-- ============================================================
-- 8. KARMA HISTORY
-- ============================================================

INSERT INTO karma_history (user_id, amount, source, source_id, created_at)
SELECT
    u.id,
    CASE s.source
        WHEN 'clip_vote' THEN (random() * 10)::int + 1
        WHEN 'comment_vote' THEN (random() * 5)::int + 1
        WHEN 'awarded_comment' THEN 25
        WHEN 'clip_submission' THEN 10
        WHEN 'daily_login' THEN 5
        ELSE (random() * 15)::int
    END,
    s.source,
    gen_random_uuid(),
    NOW() - (random() * 90 || ' days')::interval
FROM users u
CROSS JOIN (VALUES ('clip_vote'), ('comment_vote'), ('clip_submission'), ('daily_login'), ('awarded_comment')) AS s(source)
WHERE random() < 0.6;

-- ============================================================
-- 9. USER BADGES
-- ============================================================

INSERT INTO user_badges (user_id, badge_id, awarded_at)
SELECT u.id, b.badge_id, NOW() - (random() * 180 || ' days')::interval
FROM users u
CROSS JOIN (VALUES
    ('early_adopter'),
    ('first_clip'),
    ('first_comment'),
    ('helpful'),
    ('top_clpr'),
    ('streak_7'),
    ('streak_30'),
    ('community_builder'),
    ('trusted_reporter'),
    ('verified')
) AS b(badge_id)
WHERE random() < 0.15
ON CONFLICT (user_id, badge_id) DO NOTHING;

-- ============================================================
-- 9.5 CATEGORIES (featured + extended)
-- ============================================================

INSERT INTO categories (name, slug, description, icon, position, category_type, is_featured, is_custom)
VALUES
    ('Just Chatting', 'just-chatting', 'Real-life streams and conversations', '💬', 1, 'game', true, false),
    ('IRL', 'irl', 'In real life streams, travel, and day-to-day moments', '🌍', 2, 'topic', true, false),
    ('Politics', 'politics', 'Politics, commentary, and debates', '🗳️', 3, 'topic', true, false),
    ('News', 'news', 'Breaking news, analysis, and current events', '📰', 4, 'topic', true, false),
    ('Music', 'music', 'Music performances and DJ sets', '🎵', 5, 'game', true, false),
    ('Creative', 'creative', 'Art, design, and creative content', '🎨', 6, 'game', true, false),
    ('Sports', 'sports', 'Sports games and athletic competitions', '⚽', 7, 'game', true, false),
    ('FPS', 'fps', 'First-person shooter games', '🎯', 8, 'game', true, false),
    ('MOBA', 'moba', 'Multiplayer online battle arenas', '🏆', 9, 'game', true, false),
    ('RPG', 'rpg', 'Role-playing games', '⚔️', 10, 'game', true, false),
    ('Battle Royale', 'battle-royale', 'Battle royale games', '🎮', 11, 'game', true, false),
    ('Esports', 'esports', 'Competitive events and tournament highlights', '🏅', 12, 'topic', true, false),
    ('Variety', 'variety', 'Mixed content and variety streaming', '🎲', 13, 'topic', true, false),
    ('Strategy', 'strategy', 'Strategy and tactics games', '🧠', 14, 'game', false, false),
    ('Tech', 'tech', 'Programming, gadgets, and tech talk', '💻', 15, 'topic', false, false),
    ('Art & Design', 'art-design', 'Digital art, design, and illustration', '🖌️', 16, 'topic', false, false),
    ('Simulation', 'simulation', 'Simulation and management games', '🚗', 17, 'game', false, false),
    ('Horror', 'horror', 'Horror games and spooky streams', '👻', 18, 'game', false, false),
    ('Indie', 'indie', 'Indie games and creative projects', '🌟', 19, 'game', false, false),
    ('Other', 'other', 'Other games and content', '🎲', 20, 'game', false, false)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    position = EXCLUDED.position,
    category_type = EXCLUDED.category_type,
    is_featured = EXCLUDED.is_featured,
    is_custom = EXCLUDED.is_custom,
    updated_at = NOW();

-- ============================================================
-- 10. GAMES
-- ============================================================

INSERT INTO games (id, twitch_game_id, name, box_art_url)
VALUES
    (gen_random_uuid(), '509658', 'Just Chatting', 'https://static-cdn.jtvnw.net/ttv-boxart/509658-144x192.jpg'),
    (gen_random_uuid(), '33214', 'Fortnite', 'https://static-cdn.jtvnw.net/ttv-boxart/33214-144x192.jpg'),
    (gen_random_uuid(), '21779', 'League of Legends', 'https://static-cdn.jtvnw.net/ttv-boxart/21779-144x192.jpg'),
    (gen_random_uuid(), '516575', 'Valorant', 'https://static-cdn.jtvnw.net/ttv-boxart/516575-144x192.jpg'),
    (gen_random_uuid(), '27471', 'Minecraft', 'https://static-cdn.jtvnw.net/ttv-boxart/27471-144x192.jpg'),
    (gen_random_uuid(), '32982', 'Grand Theft Auto V', 'https://static-cdn.jtvnw.net/ttv-boxart/32982-144x192.jpg'),
    (gen_random_uuid(), '32399', 'Counter-Strike 2', 'https://static-cdn.jtvnw.net/ttv-boxart/32399-144x192.jpg'),
    (gen_random_uuid(), '511224', 'Apex Legends', 'https://static-cdn.jtvnw.net/ttv-boxart/511224-144x192.jpg'),
    (gen_random_uuid(), '515025', 'Overwatch 2', 'https://static-cdn.jtvnw.net/ttv-boxart/515025-144x192.jpg'),
    (gen_random_uuid(), '512710', 'Call of Duty: Warzone', 'https://static-cdn.jtvnw.net/ttv-boxart/512710-144x192.jpg'),
    (gen_random_uuid(), '29595', 'Dota 2', 'https://static-cdn.jtvnw.net/ttv-boxart/29595-144x192.jpg'),
    (gen_random_uuid(), '26936', 'Music', 'https://static-cdn.jtvnw.net/ttv-boxart/26936-144x192.jpg'),
    (gen_random_uuid(), '518203', 'Sports', 'https://static-cdn.jtvnw.net/ttv-boxart/518203-144x192.jpg'),
    (gen_random_uuid(), '488190', 'Poker', 'https://static-cdn.jtvnw.net/ttv-boxart/488190-144x192.jpg'),
    (gen_random_uuid(), '460630', 'Baldur''s Gate 3', 'https://static-cdn.jtvnw.net/ttv-boxart/460630-144x192.jpg'),
    (gen_random_uuid(), '512953', 'Elden Ring', 'https://static-cdn.jtvnw.net/ttv-boxart/512953-144x192.jpg'),
    (gen_random_uuid(), '29307', 'Path of Exile', 'https://static-cdn.jtvnw.net/ttv-boxart/29307-144x192.jpg'),
    (gen_random_uuid(), '513143', 'Teamfight Tactics', 'https://static-cdn.jtvnw.net/ttv-boxart/513143-144x192.jpg'),
    (gen_random_uuid(), '65632', 'DayZ', 'https://static-cdn.jtvnw.net/ttv-boxart/65632-144x192.jpg'),
    (gen_random_uuid(), '263490', 'Rust', 'https://static-cdn.jtvnw.net/ttv-boxart/263490-144x192.jpg')
ON CONFLICT (twitch_game_id) DO NOTHING;

-- Map games to categories
INSERT INTO category_games (game_id, category_id)
SELECT g.id, c.id
FROM games g
CROSS JOIN categories c
WHERE
    (g.name = 'Just Chatting' AND c.slug IN ('just-chatting', 'irl', 'politics', 'news', 'variety', 'tech'))
    OR (g.name = 'Music' AND c.slug = 'music')
    OR (g.name IN ('Fortnite', 'Apex Legends', 'Call of Duty: Warzone') AND c.slug = 'battle-royale')
    OR (g.name IN ('Counter-Strike 2', 'Valorant', 'Overwatch 2') AND c.slug = 'fps')
    OR (g.name IN ('League of Legends', 'Dota 2', 'Teamfight Tactics') AND c.slug = 'moba')
    OR (g.name IN ('Baldur''s Gate 3', 'Elden Ring', 'Path of Exile') AND c.slug = 'rpg')
    OR (g.name IN ('Minecraft', 'Rust', 'DayZ') AND c.slug IN ('creative', 'art-design'))
    OR (g.name = 'Grand Theft Auto V' AND c.slug = 'other')
    OR (g.name = 'Sports' AND c.slug = 'sports')
    OR (g.name = 'Poker' AND c.slug = 'strategy')
    OR (g.name IN ('Counter-Strike 2', 'Valorant', 'Overwatch 2', 'League of Legends', 'Dota 2') AND c.slug = 'esports')
ON CONFLICT DO NOTHING;

-- ============================================================
-- 11. CLIP TAGS (link existing clips to existing tags)
-- ============================================================

INSERT INTO clip_tags (clip_id, tag_id, created_at)
SELECT c.id, t.id, NOW() - (random() * 30 || ' days')::interval
FROM clips c
CROSS JOIN LATERAL (
    SELECT id FROM tags ORDER BY random() LIMIT (1 + (random() * 3)::int)
) t
ON CONFLICT DO NOTHING;

-- ============================================================
-- 12. STREAMS (live streamers)
-- ============================================================

INSERT INTO streams (id, streamer_username, display_name, is_live, last_went_live, game_name, title, viewer_count, created_at)
VALUES
    (gen_random_uuid(), 'xqc', 'xQc', true, NOW() - interval '3 hours', 'Just Chatting', '🔴 LIVE - REACTING TO CLIPS', 45000, NOW() - interval '6 months'),
    (gen_random_uuid(), 'pokimane', 'Pokimane', true, NOW() - interval '1 hour', 'Valorant', 'Ranked grind w/ friends', 22000, NOW() - interval '6 months'),
    (gen_random_uuid(), 'shroud', 'shroud', false, NOW() - interval '1 day', 'Counter-Strike 2', 'FPL matches', 0, NOW() - interval '6 months'),
    (gen_random_uuid(), 'summit1g', 'summit1g', true, NOW() - interval '4 hours', 'DayZ', 'THE JUICER IS LOOSE', 18000, NOW() - interval '6 months'),
    (gen_random_uuid(), 'hasanabi', 'HasanAbi', true, NOW() - interval '2 hours', 'Just Chatting', 'News & Politics + Viewer clips', 32000, NOW() - interval '6 months'),
    (gen_random_uuid(), 'caedrel', 'Caedrel', false, NOW() - interval '12 hours', 'League of Legends', 'WORLDS WATCH PARTY 🌍', 0, NOW() - interval '3 months'),
    (gen_random_uuid(), 'timthetatman', 'TimTheTatman', false, NOW() - interval '2 days', 'Fortnite', 'OLD MAN GAMING', 0, NOW() - interval '6 months'),
    (gen_random_uuid(), 'lirik', 'LIRIK', true, NOW() - interval '5 hours', 'Baldur''s Gate 3', 'BG3 Day 400', 15000, NOW() - interval '6 months'),
    (gen_random_uuid(), 'ludwig', 'ludwig', false, NOW() - interval '3 days', 'Just Chatting', 'MOGUL MONEY', 0, NOW() - interval '6 months'),
    (gen_random_uuid(), 'lacy', 'Lacy', true, NOW() - interval '30 minutes', 'Just Chatting', 'late night vibes 🌙', 8500, NOW() - interval '3 months'),
    (gen_random_uuid(), 'cloakzy', 'cloakzy', false, NOW() - interval '6 hours', 'Fortnite', 'SOLO DROPS', 0, NOW() - interval '3 months'),
    (gen_random_uuid(), 'hachubby', 'HAchubby', true, NOW() - interval '2 hours', 'Just Chatting', '한국에서 안녕! Hello from Korea! 🇰🇷', 12000, NOW() - interval '3 months')
ON CONFLICT (streamer_username) DO NOTHING;

-- ============================================================
-- 13. COMMUNITIES
-- ============================================================

INSERT INTO communities (id, name, slug, description, icon, owner_id, is_public, rules, created_at)
SELECT
    gen_random_uuid(),
    v.name,
    v.slug,
    v.description,
    v.icon,
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    true,
    v.rules,
    NOW() - (random() * 90 || ' days')::interval
FROM (VALUES
    ('Clip Legends', 'clip-legends', 'The best clips from across Twitch. Only top-tier content.', '🏆', 'Be respectful. No spam. Only high-quality clips.'),
    ('FPS Highlights', 'fps-highlights', 'Insane headshots, clutch plays, and ace moments.', '🎯', 'FPS clips only. No self-promo.'),
    ('Just Chatting Moments', 'just-chatting-moments', 'The funniest, most awkward, and most wholesome IRL moments.', '💬', 'Keep it wholesome. No drama-baiting.'),
    ('Speedrun Clips', 'speedrun-clips', 'World records, PBs, and incredible speedrun moments.', '⚡', 'Speedrun content only. Include game name in title.'),
    ('Fails & Funny', 'fails-and-funny', 'When it goes wrong, it goes VERY wrong. Gaming fails compilation.', '😂', 'Funny clips only. No bullying. Tag NSFW if needed.'),
    ('Esports Central', 'esports-central', 'Pro plays, tournament highlights, and competitive gaming clips.', '🏟️', 'Competitive clips only. Credit the tournament/event.'),
    ('Music & Creative', 'music-and-creative', 'Live performance highlights and creative stream moments.', '🎵', 'Original content encouraged. Credit artists.'),
    ('Horror Gaming', 'horror-gaming', 'Jump scares, creepy moments, and horror game highlights.', '👻', 'Horror content only. Spoiler warnings required.')
) AS v(name, slug, description, icon, rules)
ON CONFLICT (name) DO NOTHING;

-- Add members to communities
INSERT INTO community_members (community_id, user_id, role, joined_at)
SELECT c.id, u.id,
    CASE
        WHEN u.id = c.owner_id THEN 'admin'
        WHEN random() < 0.1 THEN 'mod'
        ELSE 'member'
    END,
    NOW() - (random() * 60 || ' days')::interval
FROM communities c
CROSS JOIN LATERAL (
    SELECT id FROM users ORDER BY random() LIMIT (5 + (random() * 25)::int)
) u
ON CONFLICT DO NOTHING;

-- Update community member counts
UPDATE communities SET member_count = sub.cnt
FROM (SELECT community_id, count(*) as cnt FROM community_members GROUP BY community_id) sub
WHERE communities.id = sub.community_id;

-- Add clips to communities
INSERT INTO community_clips (community_id, clip_id, added_by_user_id, added_at)
SELECT c.id, cl.id,
    (SELECT user_id FROM community_members WHERE community_id = c.id ORDER BY random() LIMIT 1),
    NOW() - (random() * 30 || ' days')::interval
FROM communities c
CROSS JOIN LATERAL (
    SELECT id FROM clips ORDER BY random() LIMIT (3 + (random() * 10)::int)
) cl
ON CONFLICT DO NOTHING;

-- Community discussions
INSERT INTO community_discussions (id, community_id, user_id, title, content, vote_score, comment_count, created_at)
SELECT
    gen_random_uuid(),
    c.id,
    (SELECT user_id FROM community_members WHERE community_id = c.id ORDER BY random() LIMIT 1),
    v.title,
    v.content,
    (random() * 30)::int,
    0,
    NOW() - (random() * 30 || ' days')::interval
FROM communities c
CROSS JOIN (VALUES
    ('What''s the best clip you''ve ever seen?', 'I''ll start - that xQc fail clip from last week had me rolling. What are your all-time favorites?'),
    ('Community clip of the week thread', 'Drop your nominations below! Most upvoted clip gets pinned.'),
    ('New member introductions', 'Welcome to the community! Tell us about yourself and your favorite streamers.'),
    ('Should we allow NSFW clips?', 'Been seeing some borderline clips lately. Let''s discuss the rules around this.'),
    ('Clip submission guidelines update', 'We''re updating the guidelines for clip submissions. Please read and give feedback.')
) AS v(title, content);

-- ============================================================
-- 14. FEEDS (custom user feeds)
-- ============================================================

INSERT INTO feeds (id, user_id, name, description, icon, is_public, follower_count, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    v.name,
    v.description,
    v.icon,
    true,
    (random() * 100)::int,
    NOW() - (random() * 60 || ' days')::interval
FROM (SELECT id FROM users ORDER BY random() LIMIT 6) u
CROSS JOIN LATERAL (
    SELECT name, description, icon FROM (VALUES
        ('Top Plays Today', 'Curated selection of the best plays from today', '🔥'),
        ('Chill Vibes', 'Relaxing and wholesome stream moments', '☕'),
        ('Esports Highlights', 'Best moments from competitive gaming', '🏆')
    ) AS f(name, description, icon)
    ORDER BY random() LIMIT 1
) v;

-- Add clips to feeds
INSERT INTO feed_items (id, feed_id, clip_id, position, added_at)
SELECT
    gen_random_uuid(),
    f.id,
    c.id,
    row_number() OVER (PARTITION BY f.id ORDER BY random()),
    NOW() - (random() * 14 || ' days')::interval
FROM feeds f
CROSS JOIN LATERAL (
    SELECT id FROM clips ORDER BY random() LIMIT (3 + (random() * 8)::int)
) c
ON CONFLICT (feed_id, clip_id) DO NOTHING;

-- Feed follows
INSERT INTO feed_follows (id, user_id, feed_id, followed_at)
SELECT gen_random_uuid(), u.id, f.id, NOW() - (random() * 30 || ' days')::interval
FROM feeds f
CROSS JOIN LATERAL (
    SELECT id FROM users WHERE id != f.user_id ORDER BY random() LIMIT (2 + (random() * 10)::int)
) u
ON CONFLICT (user_id, feed_id) DO NOTHING;

-- ============================================================
-- 15. DISCOVERY LISTS (enrich existing + add clips)
-- ============================================================

-- Ensure discovery lists have clips
INSERT INTO discovery_list_clips (list_id, clip_id, display_order, added_at)
SELECT
    dl.id,
    c.id,
    row_number() OVER (PARTITION BY dl.id ORDER BY random()),
    NOW()
FROM discovery_lists dl
CROSS JOIN LATERAL (
    SELECT id FROM clips ORDER BY random() LIMIT (5 + (random() * 10)::int)
) c
ON CONFLICT DO NOTHING;

-- ============================================================
-- 16. PLAYLISTS
-- ============================================================

INSERT INTO playlists (id, user_id, title, description, visibility, like_count, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    v.title,
    v.description,
    CASE WHEN random() > 0.3 THEN 'public' ELSE 'private' END,
    (random() * 50)::int,
    NOW() - (random() * 60 || ' days')::interval
FROM (SELECT id FROM users ORDER BY random() LIMIT 12) u
CROSS JOIN LATERAL (
    SELECT title, description FROM (VALUES
        ('My All-Time Favorites', 'The clips that never get old'),
        ('Daily Highlights', 'Best clips from my daily browsing'),
        ('Stream Fails Compilation', 'The best fails I could find'),
        ('Clutch Plays', 'When gamers go GOD MODE'),
        ('Wholesome Moments', 'Twitch clips that restore your faith in humanity'),
        ('Top 10 This Week', 'My weekly curation of the best content')
    ) AS p(title, description)
    ORDER BY random() LIMIT 1
) v;

-- Add clips to playlists
INSERT INTO playlist_items (playlist_id, clip_id, order_index, added_at)
SELECT
    p.id,
    c.id,
    row_number() OVER (PARTITION BY p.id ORDER BY random()),
    NOW()
FROM playlists p
CROSS JOIN LATERAL (
    SELECT id FROM clips ORDER BY random() LIMIT (3 + (random() * 12)::int)
) c
ON CONFLICT (playlist_id, clip_id) DO NOTHING;

-- Playlist likes
INSERT INTO playlist_likes (id, user_id, playlist_id, created_at)
SELECT gen_random_uuid(), u.id, p.id, NOW() - (random() * 30 || ' days')::interval
FROM playlists p
CROSS JOIN LATERAL (
    SELECT id FROM users WHERE id != p.user_id ORDER BY random() LIMIT (1 + (random() * 5)::int)
) u
WHERE p.visibility = 'public'
ON CONFLICT (user_id, playlist_id) DO NOTHING;

-- ============================================================
-- 17. BROADCASTER FOLLOWS
-- ============================================================

INSERT INTO broadcaster_follows (id, user_id, broadcaster_id, broadcaster_name, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    b.broadcaster_id,
    b.broadcaster_name,
    NOW() - (random() * 90 || ' days')::interval
FROM users u
CROSS JOIN (
    SELECT DISTINCT broadcaster_id, broadcaster_name
    FROM clips
    WHERE broadcaster_id IS NOT NULL AND broadcaster_id != ''
) b
WHERE random() < 0.12
ON CONFLICT (user_id, broadcaster_id) DO NOTHING;

-- ============================================================
-- 18. STREAM FOLLOWS
-- ============================================================

INSERT INTO stream_follows (id, user_id, streamer_username, notifications_enabled, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    s.streamer_username,
    random() > 0.3,
    NOW() - (random() * 60 || ' days')::interval
FROM users u
CROSS JOIN streams s
WHERE random() < 0.15
ON CONFLICT (user_id, streamer_username) DO NOTHING;

-- ============================================================
-- 19. USER FOLLOWS (social graph)
-- ============================================================

INSERT INTO user_follows (id, follower_id, following_id, created_at)
SELECT
    gen_random_uuid(),
    u1.id,
    u2.id,
    NOW() - (random() * 90 || ' days')::interval
FROM users u1
CROSS JOIN LATERAL (
    SELECT id FROM users WHERE id != u1.id ORDER BY random() LIMIT (1 + (random() * 5)::int)
) u2
WHERE random() < 0.3
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- Update follower/following counts
UPDATE users SET follower_count = sub.cnt
FROM (SELECT following_id, count(*) as cnt FROM user_follows GROUP BY following_id) sub
WHERE users.id = sub.following_id;

UPDATE users SET following_count = sub.cnt
FROM (SELECT follower_id, count(*) as cnt FROM user_follows GROUP BY follower_id) sub
WHERE users.id = sub.follower_id;

-- ============================================================
-- 20. WATCH HISTORY
-- ============================================================

INSERT INTO watch_history (id, user_id, clip_id, progress_seconds, duration_seconds, completed, watched_at)
SELECT
    gen_random_uuid(),
    u.id,
    c.id,
    CASE WHEN random() > 0.3 THEN c_dur ELSE (random() * c_dur)::int END,
    c_dur,
    random() > 0.3,
    NOW() - (random() * 30 || ' days')::interval
FROM users u
CROSS JOIN LATERAL (
    SELECT id, COALESCE(duration, 30)::int AS c_dur FROM clips ORDER BY random() LIMIT (2 + (random() * 8)::int)
) c
WHERE random() < 0.5
ON CONFLICT (user_id, clip_id) DO NOTHING;

-- ============================================================
-- 21. ANALYTICS EVENTS
-- ============================================================

INSERT INTO analytics_events (id, event_type, user_id, clip_id, metadata, created_at)
SELECT
    gen_random_uuid(),
    v.event_type,
    u.id,
    c.id,
    jsonb_build_object('source', 'seed', 'platform',
        CASE (random() * 2)::int WHEN 0 THEN 'web' WHEN 1 THEN 'mobile' ELSE 'api' END),
    NOW() - (random() * 30 || ' days')::interval
FROM (VALUES ('clip_view'), ('vote'), ('comment'), ('favorite'), ('share'), ('search')) AS v(event_type)
CROSS JOIN LATERAL (SELECT id FROM users ORDER BY random() LIMIT 10) u
CROSS JOIN LATERAL (SELECT id FROM clips ORDER BY random() LIMIT 5) c
WHERE random() < 0.4;

-- ============================================================
-- 22. DAILY ANALYTICS (last 30 days)
-- ============================================================

INSERT INTO daily_analytics (id, date, metric_type, entity_type, entity_id, value, metadata)
SELECT
    gen_random_uuid(),
    d.day,
    m.metric_type,
    m.entity_type,
    m.entity_id,
    (random() * m.max_val)::bigint + 1,
    '{}'::jsonb
FROM generate_series(NOW() - interval '30 days', NOW(), interval '1 day') AS d(day)
CROSS JOIN (VALUES
    ('clip_views', 'platform', 'global', 5000),
    ('votes', 'platform', 'global', 1000),
    ('comments', 'platform', 'global', 500),
    ('users_active', 'platform', 'global', 200),
    ('new_users', 'platform', 'global', 50),
    ('favorites', 'platform', 'global', 300),
    ('shares', 'platform', 'global', 150),
    ('searches', 'platform', 'global', 800)
) AS m(metric_type, entity_type, entity_id, max_val)
ON CONFLICT (date, metric_type, entity_type, entity_id) DO NOTHING;

-- Per-game daily analytics
INSERT INTO daily_analytics (id, date, metric_type, entity_type, entity_id, value, metadata)
SELECT
    gen_random_uuid(),
    d.day,
    'clip_views',
    'game',
    g.twitch_game_id,
    (random() * 500)::bigint + 10,
    '{}'::jsonb
FROM generate_series(NOW() - interval '14 days', NOW(), interval '1 day') AS d(day)
CROSS JOIN (SELECT twitch_game_id FROM games ORDER BY random() LIMIT 5) g
ON CONFLICT (date, metric_type, entity_type, entity_id) DO NOTHING;

-- ============================================================
-- 23. PLATFORM ANALYTICS (last 30 days)
-- ============================================================

INSERT INTO platform_analytics (id, date, total_users, active_users_daily, active_users_weekly, active_users_monthly,
    new_users_today, total_clips, new_clips_today, total_votes, votes_today, total_comments, comments_today,
    total_views, views_today, avg_session_duration, metadata)
SELECT
    gen_random_uuid(),
    d.day::date,
    118 + (row_number() OVER (ORDER BY d.day))::int,       -- growing user base
    (50 + random() * 80)::int,                               -- daily active
    (100 + random() * 80)::int,                              -- weekly active
    (150 + random() * 100)::int,                             -- monthly active
    (random() * 15)::int,                                     -- new users today
    83 + (row_number() OVER (ORDER BY d.day))::int,          -- growing clip base
    (random() * 10)::int,                                     -- new clips today
    5 + (row_number() OVER (ORDER BY d.day) * 20)::bigint,  -- total votes
    (random() * 100)::bigint,                                 -- votes today
    (row_number() OVER (ORDER BY d.day) * 5)::bigint,       -- total comments
    (random() * 40)::bigint,                                  -- comments today
    (row_number() OVER (ORDER BY d.day) * 500)::bigint,     -- total views
    (random() * 3000)::bigint,                                -- views today
    120 + random() * 300,                                     -- avg session seconds
    '{}'::jsonb
FROM generate_series(NOW() - interval '30 days', NOW(), interval '1 day') AS d(day)
ON CONFLICT (date) DO NOTHING;

-- ============================================================
-- 24. CLIP ANALYTICS
-- ============================================================

INSERT INTO clip_analytics (clip_id, total_views, unique_viewers, avg_view_duration, total_shares,
    peak_concurrent_viewers, retention_rate)
SELECT
    c.id,
    c.view_count + (random() * 1000)::bigint,
    (c.view_count * 0.7 + random() * 200)::bigint,
    COALESCE(c.duration, 30) * (0.5 + random() * 0.5),
    (random() * 50)::bigint,
    (random() * 30)::int + 1,
    0.3 + random() * 0.7
FROM clips c
ON CONFLICT (clip_id) DO UPDATE SET
    total_views = EXCLUDED.total_views,
    unique_viewers = EXCLUDED.unique_viewers;

-- ============================================================
-- 25. CREATOR ANALYTICS
-- ============================================================

INSERT INTO creator_analytics (creator_name, creator_id, total_clips, total_views, total_upvotes,
    total_downvotes, total_comments, total_favorites, avg_engagement_rate, follower_count)
SELECT
    broadcaster_name,
    MIN(broadcaster_id),
    count(*),
    sum(view_count)::bigint,
    sum(GREATEST(vote_score, 0))::bigint,
    abs(sum(LEAST(vote_score, 0)))::bigint,
    sum(comment_count)::bigint,
    sum(favorite_count)::bigint,
    (random() * 10 + 2)::float,
    (random() * 50000)::int
FROM clips
WHERE broadcaster_name IS NOT NULL
GROUP BY broadcaster_name
ON CONFLICT (creator_name) DO UPDATE SET
    total_clips = EXCLUDED.total_clips,
    total_views = EXCLUDED.total_views;

-- ============================================================
-- 26. USER ANALYTICS
-- ============================================================

INSERT INTO user_analytics (user_id, clips_upvoted, clips_downvoted, comments_posted, clips_favorited,
    searches_performed, days_active, total_karma_earned)
SELECT
    u.id,
    COALESCE(uv.up, 0),
    COALESCE(uv.down, 0),
    COALESCE(uc.cnt, 0),
    COALESCE(uf.cnt, 0),
    (random() * 100)::int,
    COALESCE(us.days_active, 1 + (random() * 90)::int),
    u.karma_points
FROM users u
LEFT JOIN (SELECT user_id, count(*) FILTER (WHERE vote_type = 1) as up, count(*) FILTER (WHERE vote_type = -1) as down FROM votes GROUP BY user_id) uv ON u.id = uv.user_id
LEFT JOIN (SELECT user_id, count(*) as cnt FROM comments GROUP BY user_id) uc ON u.id = uc.user_id
LEFT JOIN (SELECT user_id, count(*) as cnt FROM favorites GROUP BY user_id) uf ON u.id = uf.user_id
LEFT JOIN user_stats us ON u.id = us.user_id
ON CONFLICT (user_id) DO UPDATE SET
    clips_upvoted = EXCLUDED.clips_upvoted,
    clips_downvoted = EXCLUDED.clips_downvoted,
    comments_posted = EXCLUDED.comments_posted,
    clips_favorited = EXCLUDED.clips_favorited,
    total_karma_earned = EXCLUDED.total_karma_earned;

-- ============================================================
-- 27. NOTIFICATIONS (sample for each user)
-- ============================================================

INSERT INTO notifications (id, user_id, type, title, message, link, is_read, source_user_id, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    v.type,
    v.title,
    v.message,
    v.link,
    random() > 0.4,
    (SELECT id FROM users WHERE id != u.id ORDER BY random() LIMIT 1),
    NOW() - (random() * 14 || ' days')::interval
FROM (SELECT id FROM users ORDER BY random() LIMIT 30) u
CROSS JOIN (VALUES
    ('upvote', 'Your clip was upvoted!', 'Someone liked your clip submission.', '/clips'),
    ('comment', 'New comment on your clip', 'Someone commented on a clip you shared.', '/clips'),
    ('badge', 'Badge earned! 🏅', 'You earned the "First Clip" badge!', '/profile'),
    ('follow', 'New follower', 'Someone started following you.', '/profile'),
    ('rank_up', 'Rank up! ⬆️', 'You reached a new reputation rank!', '/leaderboards'),
    ('mention', 'You were mentioned', 'Someone mentioned you in a comment.', '/notifications'),
    ('system', 'Welcome to Clipper!', 'Thanks for joining! Start exploring clips now.', '/')
) AS v(type, title, message, link)
WHERE random() < 0.4;

-- ============================================================
-- 28. USER SETTINGS
-- ============================================================

INSERT INTO user_settings (user_id, profile_visibility, show_karma_publicly)
SELECT id, 'public', true FROM users
ON CONFLICT (user_id) DO NOTHING;

-- ============================================================
-- 29. NOTIFICATION PREFERENCES
-- ============================================================

INSERT INTO notification_preferences (user_id, in_app_enabled, email_enabled, email_digest,
    notify_replies, notify_mentions, notify_votes, notify_badges, notify_moderation, notify_rank_up)
SELECT id, true, random() > 0.4, 'daily', true, true, random() > 0.3, true, true, true
FROM users
ON CONFLICT (user_id) DO NOTHING;

-- ============================================================
-- 30. WATCH PARTIES
-- ============================================================

INSERT INTO watch_parties (id, host_user_id, title, playlist_id, current_clip_id, is_playing,
    visibility, invite_code, max_participants, created_at, started_at)
SELECT
    gen_random_uuid(),
    u.id,
    v.title,
    (SELECT id FROM playlists WHERE user_id = u.id LIMIT 1),
    (SELECT id FROM clips ORDER BY random() LIMIT 1),
    v.is_playing,
    v.visibility,
    substr(md5(random()::text), 1, 8),
    v.max_p,
    NOW() - (random() * 7 || ' days')::interval,
    CASE WHEN v.is_playing THEN NOW() - (random() * 3 || ' hours')::interval ELSE NULL END
FROM (SELECT id FROM users ORDER BY random() LIMIT 5) u
CROSS JOIN (VALUES
    ('Friday Night Clips 🎬', true, 'public', 50),
    ('Chill viewing sesh', true, 'public', 20),
    ('Best of the Week Watch Party', false, 'public', 100),
    ('Private squad viewing', true, 'private', 10),
    ('Community clip night', true, 'public', 75)
) AS v(title, is_playing, visibility, max_p);

-- Watch party participants
INSERT INTO watch_party_participants (party_id, user_id, role, joined_at)
SELECT
    wp.id,
    u.id,
    CASE WHEN u.id = wp.host_user_id THEN 'host' ELSE 'viewer' END,
    wp.started_at + (random() * 30 || ' minutes')::interval
FROM watch_parties wp
CROSS JOIN LATERAL (
    SELECT id FROM users ORDER BY random() LIMIT (3 + (random() * 10)::int)
) u
ON CONFLICT (party_id, user_id) DO NOTHING;

-- Watch party events
INSERT INTO watch_party_events (id, time, party_id, user_id, event_type, metadata)
SELECT
    gen_random_uuid(),
    wp.created_at + (random() * 120 || ' minutes')::interval,
    wp.id,
    (SELECT user_id FROM watch_party_participants WHERE party_id = wp.id ORDER BY random() LIMIT 1),
    v.event_type,
    CASE v.event_type
        WHEN 'chat' THEN jsonb_build_object('message', 'This clip is fire! 🔥')
        WHEN 'reaction' THEN jsonb_build_object('emoji', '😂')
        ELSE '{}'::jsonb
    END
FROM watch_parties wp
CROSS JOIN (VALUES ('join'), ('chat'), ('reaction'), ('sync'), ('chat'), ('reaction')) AS v(event_type);

-- ============================================================
-- 31. CHAT CHANNELS
-- ============================================================

INSERT INTO chat_channels (id, name, description, creator_id, channel_type, is_active, max_participants, created_at)
SELECT
    gen_random_uuid(),
    v.name,
    v.description,
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    v.channel_type,
    true,
    v.max_p,
    NOW() - (random() * 60 || ' days')::interval
FROM (VALUES
    ('General', 'General discussion about clips and streams', 'public', 500),
    ('Clip Submissions', 'Discuss and review clip submissions', 'public', 200),
    ('Off-Topic', 'Anything goes (within reason)', 'public', 500),
    ('Moderators', 'Moderator coordination channel', 'private', 50),
    ('FPS Discussion', 'Talk about FPS clips and games', 'public', 300),
    ('Just Chatting Fans', 'For fans of IRL and Just Chatting content', 'public', 300)
) AS v(name, description, channel_type, max_p);

-- Chat messages
INSERT INTO chat_messages (id, channel_id, user_id, content, created_at)
SELECT
    gen_random_uuid(),
    ch.id,
    u.id,
    v.content,
    NOW() - (random() * 7 || ' days')::interval
FROM chat_channels ch
CROSS JOIN LATERAL (SELECT id FROM users ORDER BY random() LIMIT 8) u
CROSS JOIN LATERAL (
    SELECT content FROM (VALUES
        ('Has anyone seen that new xQc clip? Absolutely insane'),
        ('Just submitted a clip from last night''s stream'),
        ('The leaderboard competition is getting intense!'),
        ('Who else is watching the tournament right now?'),
        ('GG everyone, great clips today'),
        ('Can someone explain how the karma system works?'),
        ('New community just dropped, check it out!'),
        ('This platform is so much better than Reddit for clips')
    ) AS m(content) ORDER BY random() LIMIT 1
) v
WHERE random() < 0.3;

-- ============================================================
-- 32. FORUM THREADS
-- ============================================================

INSERT INTO forum_threads (id, user_id, title, content, reply_count, view_count, created_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    v.title,
    v.content,
    0,
    (random() * 500)::int + 10,
    NOW() - (random() * 30 || ' days')::interval
FROM (VALUES
    ('Best clip submission practices', 'What makes a great clip? Let''s discuss the qualities that get the most engagement on Clipper.'),
    ('Introduce yourself!', 'New to Clipper? Drop by and say hello! Tell us your favorite streamers and games.'),
    ('Weekly clip roundup thread', 'Share your favorite clips from this week. Let''s curate the best of the best together.'),
    ('Feature request: Dark mode improvements', 'The dark mode is great but could use some tweaks. Here are my suggestions...'),
    ('Guide: How to earn karma fast', 'After months on the platform, here are my tips for building karma and reaching Legend rank.'),
    ('Tournament clip highlights discussion', 'Let''s discuss the best clips from this weekend''s tournaments across all games.'),
    ('Moderation feedback thread', 'Got feedback on moderation? Share it here constructively.'),
    ('Bug report: Video player issues on mobile', 'Anyone else experiencing buffering on mobile? Here''s what I''ve tried so far...'),
    ('Community spotlight: FPS Highlights', 'Shoutout to the FPS Highlights community for consistently great content!'),
    ('What features would you like to see next?', 'The dev team wants YOUR input. What should we build next?')
) AS v(title, content);

-- Forum replies
INSERT INTO forum_replies (id, thread_id, user_id, content, created_at)
SELECT
    gen_random_uuid(),
    ft.id,
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    v.content,
    ft.created_at + (random() * 72 || ' hours')::interval
FROM forum_threads ft
CROSS JOIN (VALUES
    ('Great post, totally agree with this!'),
    ('I have a different perspective on this...'),
    ('Thanks for sharing, very helpful!'),
    ('+1 on this, been wanting this forever'),
    ('Here''s my experience with this topic...')
) AS v(content)
WHERE random() < 0.5;

-- Update forum thread reply counts
UPDATE forum_threads SET reply_count = sub.cnt
FROM (SELECT thread_id, count(*) as cnt FROM forum_replies GROUP BY thread_id) sub
WHERE forum_threads.id = sub.thread_id;

-- ============================================================
-- 33. USER ACTIVITY (recent activity feed)
-- ============================================================

INSERT INTO user_activity (id, user_id, activity_type, target_id, target_type, metadata, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    v.activity_type,
    c.id,
    'clip',
    jsonb_build_object('clip_title', c.title),
    NOW() - (random() * 14 || ' days')::interval
FROM (SELECT id FROM users ORDER BY random() LIMIT 20) u
CROSS JOIN LATERAL (SELECT id, title FROM clips ORDER BY random() LIMIT 3) c
CROSS JOIN (VALUES ('upvote'), ('comment'), ('favorite'), ('share'), ('clip_submitted')) AS v(activity_type)
WHERE random() < 0.3;

-- ============================================================
-- 34. GAME FOLLOWS
-- ============================================================

INSERT INTO game_follows (id, user_id, game_id, followed_at)
SELECT
    gen_random_uuid(),
    u.id,
    g.id,
    NOW() - (random() * 60 || ' days')::interval
FROM users u
CROSS JOIN LATERAL (
    SELECT id FROM games ORDER BY random() LIMIT (1 + (random() * 4)::int)
) g
WHERE random() < 0.2
ON CONFLICT (user_id, game_id) DO NOTHING;

-- ============================================================
-- 35. CLIP SUBMISSIONS
-- ============================================================

INSERT INTO clip_submissions (id, user_id, twitch_clip_id, twitch_clip_url, title, status, created_at)
SELECT
    gen_random_uuid(),
    u.id,
    'seed-submission-' || row_number() OVER (),
    'https://clips.twitch.tv/seed-submission-' || row_number() OVER (),
    v.title,
    v.status,
    NOW() - (random() * 30 || ' days')::interval
FROM (SELECT id FROM users ORDER BY random() LIMIT 15) u
CROSS JOIN (VALUES
    ('Amazing clutch play in ranked', 'approved'),
    ('Funniest fail of the stream', 'approved'),
    ('Wholesome streamer moment', 'pending'),
    ('Insane 1v5 ace', 'approved'),
    ('Streamer reacts to donation', 'rejected')
) AS v(title, status)
WHERE random() < 0.4;

-- ============================================================
-- 36. TOP STREAMERS
-- ============================================================

INSERT INTO top_streamers (id, broadcaster_id, broadcaster_name, rank, follower_count, view_count, created_at)
VALUES
    (gen_random_uuid(), '71092938', 'xQc', 1, 12500000, 850000000, NOW()),
    (gen_random_uuid(), '44445592', 'Pokimane', 2, 9800000, 420000000, NOW()),
    (gen_random_uuid(), '37402112', 'shroud', 3, 10200000, 530000000, NOW()),
    (gen_random_uuid(), '26490481', 'summit1g', 4, 6200000, 380000000, NOW()),
    (gen_random_uuid(), '207813352', 'HasanAbi', 5, 2800000, 310000000, NOW()),
    (gen_random_uuid(), '31239503', 'Caedrel', 6, 1900000, 180000000, NOW()),
    (gen_random_uuid(), '36769016', 'TimTheTatman', 7, 7100000, 290000000, NOW()),
    (gen_random_uuid(), '23161357', 'LIRIK', 8, 2900000, 400000000, NOW()),
    (gen_random_uuid(), '67955580', 'ludwig', 9, 4200000, 330000000, NOW()),
    (gen_random_uuid(), '118070440', 'Lacy', 10, 950000, 45000000, NOW())
ON CONFLICT (broadcaster_id) DO NOTHING;

-- ============================================================
-- 37. USER PREFERENCES (recommendations)
-- ============================================================

INSERT INTO user_preferences (user_id, favorite_games, followed_streamers, preferred_categories, onboarding_completed, cold_start_source, created_at)
SELECT
    u.id,
    ARRAY(SELECT name FROM games ORDER BY random() LIMIT (1 + (random() * 3)::int)),
    ARRAY(SELECT streamer_username FROM streams ORDER BY random() LIMIT (1 + (random() * 3)::int)),
    ARRAY(SELECT slug FROM categories ORDER BY random() LIMIT (1 + (random() * 2)::int)),
    random() > 0.3,
    CASE WHEN random() > 0.5 THEN 'onboarding' ELSE 'inferred' END,
    NOW() - (random() * 30 || ' days')::interval
FROM users u
WHERE random() < 0.5
ON CONFLICT (user_id) DO NOTHING;

-- ============================================================
-- 38. USER CLIP INTERACTIONS (for recommendations)
-- ============================================================

INSERT INTO user_clip_interactions (id, user_id, clip_id, interaction_type, dwell_time, timestamp)
SELECT
    gen_random_uuid(),
    u.id,
    c.id,
    v.itype,
    CASE WHEN v.itype = 'dwell' THEN (random() * 60)::int + 5 ELSE NULL END,
    NOW() - (random() * 30 || ' days')::interval
FROM (SELECT id FROM users ORDER BY random() LIMIT 30) u
CROSS JOIN LATERAL (SELECT id FROM clips ORDER BY random() LIMIT 5) c
CROSS JOIN (VALUES ('view'), ('like'), ('dwell')) AS v(itype)
WHERE random() < 0.4
ON CONFLICT (user_id, clip_id, interaction_type) DO NOTHING;

-- ============================================================
-- 39. STREAM SESSIONS
-- ============================================================

INSERT INTO stream_sessions (id, user_id, stream_id, started_at, ended_at, watch_duration_seconds)
SELECT
    gen_random_uuid(),
    u.id,
    s.id,
    start_time,
    start_time + (duration || ' seconds')::interval,
    duration
FROM (SELECT id FROM users ORDER BY random() LIMIT 20) u
CROSS JOIN LATERAL (SELECT id FROM streams ORDER BY random() LIMIT 2) s
CROSS JOIN LATERAL (
    SELECT
        NOW() - (random() * 14 || ' days')::interval AS start_time,
        (300 + random() * 7200)::int AS duration
) timing
WHERE random() < 0.5;

-- ============================================================
-- 40. REPORTS (sample moderation data)
-- ============================================================

INSERT INTO reports (id, reporter_id, reportable_type, reportable_id, reason, description, status, created_at)
SELECT
    gen_random_uuid(),
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    v.rtype,
    CASE v.rtype
        WHEN 'clip' THEN (SELECT id FROM clips ORDER BY random() LIMIT 1)
        WHEN 'comment' THEN COALESCE((SELECT id FROM comments ORDER BY random() LIMIT 1), gen_random_uuid())
        ELSE (SELECT id FROM users ORDER BY random() LIMIT 1)
    END,
    v.reason,
    v.description,
    v.status,
    NOW() - (random() * 30 || ' days')::interval
FROM (VALUES
    ('clip', 'spam', 'This clip appears to be spam/self-promotion', 'pending'),
    ('clip', 'nsfw', 'Contains inappropriate content', 'reviewed'),
    ('comment', 'harassment', 'Rude and offensive comment', 'actioned'),
    ('comment', 'spam', 'Bot spam in comments', 'dismissed'),
    ('clip', 'copyright', 'Contains copyrighted music', 'pending'),
    ('user', 'ban_evasion', 'This user appears to be evading a ban', 'pending')
) AS v(rtype, reason, description, status);

-- ============================================================
-- 41. REFRESH MATERIALIZED VIEWS
-- ============================================================

REFRESH MATERIALIZED VIEW CONCURRENTLY hot_clips_materialized;

-- ============================================================
-- CLEANUP TEMP TABLES
-- ============================================================

DROP TABLE IF EXISTS _seed_users;
DROP TABLE IF EXISTS _seed_clips;

COMMIT;

-- Final report
DO $$
DECLARE
    u_count int;
    c_count int;
    v_count int;
    cm_count int;
    f_count int;
    ks_count int;
    comm_count int;
    pl_count int;
    st_count int;
    wp_count int;
    ft_count int;
    ch_count int;
BEGIN
    SELECT count(*) INTO u_count FROM users;
    SELECT count(*) INTO c_count FROM clips;
    SELECT count(*) INTO v_count FROM votes;
    SELECT count(*) INTO cm_count FROM comments;
    SELECT count(*) INTO f_count FROM favorites;
    SELECT count(*) INTO ks_count FROM user_stats;
    SELECT count(*) INTO comm_count FROM communities;
    SELECT count(*) INTO pl_count FROM playlists;
    SELECT count(*) INTO st_count FROM streams;
    SELECT count(*) INTO wp_count FROM watch_parties;
    SELECT count(*) INTO ft_count FROM forum_threads;
    SELECT count(*) INTO ch_count FROM chat_channels;

    RAISE NOTICE '';
    RAISE NOTICE '=== SEED COMPLETE ===';
    RAISE NOTICE 'Users:        %', u_count;
    RAISE NOTICE 'Clips:        %', c_count;
    RAISE NOTICE 'Votes:        %', v_count;
    RAISE NOTICE 'Comments:     %', cm_count;
    RAISE NOTICE 'Favorites:    %', f_count;
    RAISE NOTICE 'User Stats:   %', ks_count;
    RAISE NOTICE 'Communities:  %', comm_count;
    RAISE NOTICE 'Playlists:    %', pl_count;
    RAISE NOTICE 'Streams:      %', st_count;
    RAISE NOTICE 'Watch Parties:%', wp_count;
    RAISE NOTICE 'Forum Threads:%', ft_count;
    RAISE NOTICE 'Chat Channels:%', ch_count;
    RAISE NOTICE '=====================';
END $$;
