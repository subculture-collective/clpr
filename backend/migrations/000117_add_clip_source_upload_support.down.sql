-- Roll back generalized clip source and upload support

DROP INDEX IF EXISTS idx_clip_submissions_upload_status;
DROP INDEX IF EXISTS idx_clip_submissions_source_id;
DROP INDEX IF EXISTS idx_clips_source_id;
DROP INDEX IF EXISTS idx_clips_source_platform;

ALTER TABLE clip_submissions DROP CONSTRAINT IF EXISTS clip_submissions_storage_visibility_check;
ALTER TABLE clip_submissions DROP CONSTRAINT IF EXISTS clip_submissions_upload_status_check;
ALTER TABLE clip_submissions DROP CONSTRAINT IF EXISTS clip_submissions_source_platform_check;
ALTER TABLE clip_submissions DROP CONSTRAINT IF EXISTS clip_submissions_source_type_check;

ALTER TABLE clips DROP CONSTRAINT IF EXISTS clips_source_platform_check;
ALTER TABLE clips DROP CONSTRAINT IF EXISTS clips_source_type_check;
ALTER TABLE clips DROP CONSTRAINT IF EXISTS clips_stream_source_check;

ALTER TABLE clip_submissions
    DROP COLUMN IF EXISTS storage_visibility,
    DROP COLUMN IF EXISTS duration_validation_error,
    DROP COLUMN IF EXISTS upload_status,
    DROP COLUMN IF EXISTS file_size_bytes,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS original_filename,
    DROP COLUMN IF EXISTS storage_key,
    DROP COLUMN IF EXISTS storage_bucket,
    DROP COLUMN IF EXISTS storage_provider,
    DROP COLUMN IF EXISTS duration_verified,
    DROP COLUMN IF EXISTS duration_seconds,
    DROP COLUMN IF EXISTS source_metadata,
    DROP COLUMN IF EXISTS source_id,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS source_platform,
    DROP COLUMN IF EXISTS source_type;

ALTER TABLE clips
    DROP COLUMN IF EXISTS file_size_bytes,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS original_filename,
    DROP COLUMN IF EXISTS storage_key,
    DROP COLUMN IF EXISTS storage_bucket,
    DROP COLUMN IF EXISTS storage_provider,
    DROP COLUMN IF EXISTS duration_verified,
    DROP COLUMN IF EXISTS duration_seconds,
    DROP COLUMN IF EXISTS source_metadata,
    DROP COLUMN IF EXISTS source_id,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS source_platform,
    DROP COLUMN IF EXISTS source_type;

ALTER TABLE clips ADD CONSTRAINT clips_stream_source_check
    CHECK (stream_source IN ('twitch', 'stream'));
