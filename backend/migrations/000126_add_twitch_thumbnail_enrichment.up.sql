ALTER TABLE clips
    ADD COLUMN IF NOT EXISTS title_source TEXT NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS structural_tagged_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS vision_processed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS vision_attempted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS vision_attempt_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS vision_error TEXT;

ALTER TABLE clips
    DROP CONSTRAINT IF EXISTS clips_title_source_check;
ALTER TABLE clips
    ADD CONSTRAINT clips_title_source_check
    CHECK (title_source IN ('twitch', 'ai', 'user'));

UPDATE clips
SET structural_tagged_at = auto_tagged_at
WHERE structural_tagged_at IS NULL
  AND auto_tagged_at IS NOT NULL;

UPDATE clips
SET title_source = 'user'
WHERE submitted_by_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_clips_structural_tag_queue
    ON clips (created_at DESC)
    WHERE structural_tagged_at IS NULL AND is_removed = FALSE;

CREATE INDEX IF NOT EXISTS idx_clips_vision_queue
    ON clips (created_at DESC)
    WHERE vision_processed_at IS NULL
      AND submitted_by_user_id IS NULL
      AND thumbnail_url IS NOT NULL
      AND is_removed = FALSE;

CREATE TABLE IF NOT EXISTS clip_enrichments (
    clip_id UUID PRIMARY KEY REFERENCES clips(id) ON DELETE CASCADE,
    source_title TEXT NOT NULL,
    suggested_title TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    basis TEXT NOT NULL CHECK (basis IN ('source_title', 'visible', 'metadata', 'insufficient')),
    evidence JSONB NOT NULL DEFAULT '[]'::jsonb,
    tags TEXT[] NOT NULL DEFAULT '{}',
    title_accepted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP INDEX IF EXISTS idx_clips_auto_tag_queue;
