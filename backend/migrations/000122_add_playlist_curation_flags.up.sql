-- Add curation flags to playlists table (idempotent; columns may already exist from prior migrations)
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS is_curated BOOLEAN DEFAULT false;
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS is_featured BOOLEAN DEFAULT false;
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS display_order INTEGER DEFAULT 0;
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS script_id UUID REFERENCES playlist_scripts(id);