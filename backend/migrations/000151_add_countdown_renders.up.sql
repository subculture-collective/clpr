-- Rendered countdown videos, Shorts and thumbnails (#78). The files live on
-- the render host until they are published (#79) or expire; these rows keep
-- their checksums.

ALTER TABLE countdown_series
    ADD COLUMN shorts_count INT NOT NULL DEFAULT 0 CHECK (shorts_count BETWEEN 0 AND 10);

COMMENT ON COLUMN countdown_series.shorts_count IS
    'How many top clips of each edition also get a 9:16 Short';

CREATE TABLE countdown_renders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    edition_id UUID NOT NULL REFERENCES countdown_editions(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('countdown', 'short', 'thumbnail')),
    clip_id UUID REFERENCES clips(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL CHECK (char_length(file_name) BETWEEN 1 AND 200),
    sha256 TEXT NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    bytes BIGINT NOT NULL CHECK (bytes > 0),
    duration_seconds DOUBLE PRECISION CHECK (duration_seconds IS NULL OR duration_seconds > 0),
    rendered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CHECK ((kind = 'short') = (clip_id IS NOT NULL)),
    UNIQUE NULLS NOT DISTINCT (edition_id, kind, clip_id)
);

CREATE INDEX countdown_renders_undeleted_idx ON countdown_renders (rendered_at) WHERE deleted_at IS NULL;
