-- Add generalized clip source and upload support

ALTER TABLE clips
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_platform VARCHAR(30) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_url TEXT,
    ADD COLUMN IF NOT EXISTS source_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS source_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS duration_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS duration_verified BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS storage_provider VARCHAR(30),
    ADD COLUMN IF NOT EXISTS storage_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_key TEXT,
    ADD COLUMN IF NOT EXISTS original_filename TEXT,
    ADD COLUMN IF NOT EXISTS mime_type VARCHAR(255),
    ADD COLUMN IF NOT EXISTS file_size_bytes BIGINT;

ALTER TABLE clip_submissions
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_platform VARCHAR(30) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_url TEXT,
    ADD COLUMN IF NOT EXISTS source_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS source_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS duration_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS duration_verified BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS storage_provider VARCHAR(30),
    ADD COLUMN IF NOT EXISTS storage_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_key TEXT,
    ADD COLUMN IF NOT EXISTS original_filename TEXT,
    ADD COLUMN IF NOT EXISTS mime_type VARCHAR(255),
    ADD COLUMN IF NOT EXISTS file_size_bytes BIGINT,
    ADD COLUMN IF NOT EXISTS upload_status VARCHAR(30) NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS duration_validation_error TEXT,
    ADD COLUMN IF NOT EXISTS storage_visibility VARCHAR(20) NOT NULL DEFAULT 'private';

ALTER TABLE clips DROP CONSTRAINT IF EXISTS clips_stream_source_check;
ALTER TABLE clips ADD CONSTRAINT clips_stream_source_check
    CHECK (stream_source IN ('twitch', 'stream', 'external', 'upload'));

ALTER TABLE clips ADD CONSTRAINT clips_source_type_check
    CHECK (source_type IN ('twitch', 'external', 'upload'));
ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_source_type_check
    CHECK (source_type IN ('twitch', 'external', 'upload'));

ALTER TABLE clips ADD CONSTRAINT clips_source_platform_check
    CHECK (source_platform IN ('twitch', 'kick', 'youtube', 'youtube_shorts', 'tiktok', 'upload'));
ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_source_platform_check
    CHECK (source_platform IN ('twitch', 'kick', 'youtube', 'youtube_shorts', 'tiktok', 'upload'));

ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_upload_status_check
    CHECK (upload_status IN ('none', 'pending', 'uploaded', 'validated', 'rejected'));
ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_storage_visibility_check
    CHECK (storage_visibility IN ('private', 'public'));

CREATE INDEX IF NOT EXISTS idx_clips_source_platform ON clips(source_platform);
CREATE INDEX IF NOT EXISTS idx_clips_source_id ON clips(source_platform, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clip_submissions_source_id ON clip_submissions(source_platform, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clip_submissions_upload_status ON clip_submissions(upload_status);
