-- Automatically attach known Twitch games to the curated category taxonomy.
-- This is intentionally idempotent and also backfills games already present.
CREATE OR REPLACE FUNCTION sync_game_curated_categories()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO category_games (game_id, category_id)
    SELECT NEW.id, c.id
    FROM categories c
    WHERE
        (NEW.name = 'Just Chatting' AND c.slug IN ('just-chatting', 'irl', 'politics', 'news', 'variety', 'tech'))
        OR (NEW.name = 'Music' AND c.slug = 'music')
        OR (NEW.name IN ('Fortnite', 'Apex Legends', 'Call of Duty: Warzone') AND c.slug = 'battle-royale')
        OR (NEW.name IN ('Counter-Strike 2', 'Valorant', 'Overwatch 2') AND c.slug = 'fps')
        OR (NEW.name IN ('League of Legends', 'Dota 2', 'Teamfight Tactics') AND c.slug = 'moba')
        OR (NEW.name IN ('Baldur''s Gate 3', 'Elden Ring', 'Path of Exile') AND c.slug = 'rpg')
        OR (NEW.name IN ('Minecraft', 'Rust', 'DayZ') AND c.slug IN ('creative', 'art-design'))
        OR (NEW.name = 'Grand Theft Auto V' AND c.slug = 'other')
        OR (NEW.name = 'Sports' AND c.slug = 'sports')
        OR (NEW.name = 'Poker' AND c.slug = 'strategy')
        OR (NEW.name IN ('Counter-Strike 2', 'Valorant', 'Overwatch 2', 'League of Legends', 'Dota 2') AND c.slug = 'esports')
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS sync_game_curated_categories_trigger ON games;
CREATE TRIGGER sync_game_curated_categories_trigger
AFTER INSERT OR UPDATE OF name ON games
FOR EACH ROW EXECUTE FUNCTION sync_game_curated_categories();

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
