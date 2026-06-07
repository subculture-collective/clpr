-- Revert UNIQUE index back to regular index
DROP INDEX IF EXISTS idx_games_slug;
CREATE INDEX idx_games_slug ON games (slug);

-- Remove NOT NULL constraint from game slug
ALTER TABLE games ALTER COLUMN slug DROP NOT NULL;

-- Revert bot user role back to admin
UPDATE users SET role = 'admin' WHERE username = 'clpr-bot' AND role = 'service';
