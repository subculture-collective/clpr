-- Media fetched for countdown editions (#77, ADR-002). Files live only on the
-- render host and are deleted within 24 hours; these rows keep what was
-- fetched, from where, and its checksum.

ALTER TABLE countdown_editions DROP CONSTRAINT countdown_editions_status_check;
ALTER TABLE countdown_editions ADD CONSTRAINT countdown_editions_status_check
    CHECK (status IN ('selected', 'fetching', 'fetched', 'rendering', 'rendered', 'published', 'failed'));

CREATE INDEX countdown_editions_status_idx ON countdown_editions (status, edition_date);

CREATE TABLE countdown_edition_media (
    edition_id UUID NOT NULL REFERENCES countdown_editions(id) ON DELETE CASCADE,
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    source TEXT NOT NULL CHECK (source IN ('helix', 'player')),
    file_name TEXT NOT NULL CHECK (char_length(file_name) BETWEEN 1 AND 200),
    sha256 TEXT NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    bytes BIGINT NOT NULL CHECK (bytes > 0),
    duration_seconds DOUBLE PRECISION NOT NULL CHECK (duration_seconds > 0),
    width INT NOT NULL CHECK (width > 0),
    height INT NOT NULL CHECK (height > 0),
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (edition_id, clip_id)
);

CREATE INDEX countdown_edition_media_undeleted_idx ON countdown_edition_media (fetched_at) WHERE deleted_at IS NULL;

-- Clips whose media could not be fetched. Selection and backfill leave them
-- out of every edition on the same date.
CREATE TABLE countdown_edition_skips (
    edition_id UUID NOT NULL REFERENCES countdown_editions(id) ON DELETE CASCADE,
    edition_date DATE NOT NULL,
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    reason TEXT NOT NULL CHECK (char_length(reason) BETWEEN 1 AND 500),
    skipped_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (edition_id, clip_id)
);

CREATE INDEX countdown_edition_skips_date_idx ON countdown_edition_skips (edition_date, clip_id);
