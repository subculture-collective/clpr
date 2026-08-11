UPDATE playlist_scripts
SET name = 'Across the Culture',
    description = 'A daily, lightly shuffled tour led by distinct creators and spread across the topics shaping live culture.',
    title_template = 'Across the Culture • {date}',
    updated_at = NOW()
WHERE name = 'Diversity Roulette';

UPDATE playlists
SET title = REPLACE(title, 'Diversity Roulette', 'Across the Culture'),
    description = 'A daily, lightly shuffled tour led by distinct creators and spread across the topics shaping live culture.',
    updated_at = NOW()
WHERE title LIKE 'Diversity Roulette%';

UPDATE playlist_scripts
SET description = 'A weekly quality-and-surprise mix capped at one clip per creator with a soft topic cap.',
    updated_at = NOW()
WHERE name = 'Weekend Mix';
