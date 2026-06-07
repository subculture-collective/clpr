-- Fix bot user role from admin to service
UPDATE users SET role = 'service' WHERE username = 'clpr-bot' AND role = 'admin';

-- Backfill any NULL game slugs from game name
UPDATE games SET slug = LOWER(REGEXP_REPLACE(name, '[^a-zA-Z0-9]+', '-', 'g')) WHERE slug IS NULL;

-- Add NOT NULL constraint to game slug
ALTER TABLE games ALTER COLUMN slug SET NOT NULL;

-- Replace existing index with a UNIQUE index
DROP INDEX IF EXISTS idx_games_slug;
CREATE UNIQUE INDEX idx_games_slug ON games (slug);
