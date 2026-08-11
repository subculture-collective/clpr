-- The original discovery collections were display-only shells. Once automated
-- curated playlists are enabled, empty shells should not occupy featured slots.
UPDATE playlists p
SET is_featured = false,
    is_curated = false,
    updated_at = NOW()
WHERE p.script_id IS NULL
  AND p.title IN ('Epic Gaming Moments', 'Funny Fails', 'Speedrun Records', 'Community Favorites')
  AND NOT EXISTS (
      SELECT 1 FROM playlist_items pi WHERE pi.playlist_id = p.id
  );
