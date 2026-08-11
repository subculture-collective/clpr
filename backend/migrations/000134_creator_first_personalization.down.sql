UPDATE playlist_scripts
SET name = 'Diversity Roulette',
    description = 'A daily, lightly shuffled tour across the catalog with one strong clip from each Twitch category.',
    title_template = 'Diversity Roulette • {date}',
    updated_at = NOW()
WHERE name = 'Across the Culture';

UPDATE playlists
SET title = REPLACE(title, 'Across the Culture', 'Diversity Roulette'),
    description = 'A daily, lightly shuffled tour across the catalog with one strong clip from each Twitch category.',
    updated_at = NOW()
WHERE title LIKE 'Across the Culture%';

UPDATE playlist_scripts
SET description = 'A weekly quality-and-surprise mix capped at one clip per creator and two clips per Twitch category.',
    updated_at = NOW()
WHERE name = 'Weekend Mix';
