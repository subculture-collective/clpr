-- Restore the original featured flags only for legacy collections that remain empty.
UPDATE playlists p
SET is_featured = p.title IN ('Epic Gaming Moments', 'Funny Fails', 'Speedrun Records'),
    is_curated = true,
    updated_at = NOW()
WHERE p.script_id IS NULL
  AND p.title IN ('Epic Gaming Moments', 'Funny Fails', 'Speedrun Records', 'Community Favorites')
  AND NOT EXISTS (
      SELECT 1 FROM playlist_items pi WHERE pi.playlist_id = p.id
  );
