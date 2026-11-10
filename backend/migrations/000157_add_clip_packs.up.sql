-- Weekly clip packs for broadcasters who feature their clips (#86). Each
-- Monday the countdown worker picks a broadcaster's top clips of the past
-- week, renders them as captioned 9:16 files and a 16:9 highlight reel, and
-- zips them on the render host. The zip is served by expiring link and
-- deleted 7 days after it is ready.
CREATE TABLE clip_packs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    broadcaster_id TEXT NOT NULL CHECK (char_length(broadcaster_id) BETWEEN 1 AND 50),
    broadcaster_name TEXT NOT NULL CHECK (char_length(broadcaster_name) BETWEEN 1 AND 100),
    week_start DATE NOT NULL CHECK (EXTRACT(ISODOW FROM week_start) = 1),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'building', 'ready', 'failed', 'expired')),
    file_name TEXT CHECK (file_name IS NULL OR char_length(file_name) BETWEEN 1 AND 200),
    sha256 TEXT CHECK (sha256 IS NULL OR sha256 ~ '^[0-9a-f]{64}$'),
    bytes BIGINT CHECK (bytes IS NULL OR bytes > 0),
    clip_count INT CHECK (clip_count IS NULL OR clip_count >= 0),
    failure_reason TEXT CHECK (failure_reason IS NULL OR char_length(failure_reason) <= 500),
    ready_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (broadcaster_id, week_start),
    CHECK (status <> 'ready' OR (file_name IS NOT NULL AND sha256 IS NOT NULL AND expires_at IS NOT NULL))
);

CREATE INDEX clip_packs_status_idx ON clip_packs (status, created_at);
CREATE INDEX clip_packs_files_idx ON clip_packs (expires_at) WHERE file_name IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE clip_pack_clips (
    pack_id UUID NOT NULL REFERENCES clip_packs(id) ON DELETE CASCADE,
    position INT NOT NULL CHECK (position >= 1),
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    PRIMARY KEY (pack_id, position),
    UNIQUE (pack_id, clip_id)
);
