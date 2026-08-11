UPDATE playlist_scripts
SET description = CASE name
    WHEN 'Diversity Roulette' THEN 'A daily, lightly shuffled tour across the catalog with one strong clip from each Twitch category.'
    WHEN 'Weekend Mix' THEN 'A weekly quality-and-surprise mix capped at one clip per creator and two clips per Twitch category.'
    WHEN 'Trending Now' THEN 'Automatically imports and publishes a fresh mix of top clips from Twitch''s current hottest creators and categories.'
    ELSE description
END,
updated_at = NOW()
WHERE name IN ('Diversity Roulette', 'Weekend Mix', 'Trending Now');

UPDATE playlists p
SET description = s.description,
    updated_at = NOW()
FROM playlist_scripts s
WHERE p.script_id = s.id
  AND s.name IN ('Diversity Roulette', 'Weekend Mix', 'Trending Now');
