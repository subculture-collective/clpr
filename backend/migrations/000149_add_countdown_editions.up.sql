-- Daily countdown editions (#76, ADR-002). A series describes one ranked list
-- (overall, one Twitch category, or one language). Each day the scheduler
-- selects one edition per enabled series from the clips created in the
-- series window, ranked by view_velocity.

CREATE TABLE countdown_series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) <= 64),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    game_id TEXT CHECK (game_id IS NULL OR char_length(game_id) BETWEEN 1 AND 50),
    language TEXT CHECK (language IS NULL OR char_length(language) BETWEEN 2 AND 10),
    size INT NOT NULL DEFAULT 10 CHECK (size BETWEEN 1 AND 50),
    window_hours INT NOT NULL DEFAULT 24 CHECK (window_hours BETWEEN 1 AND 168),
    per_broadcaster_cap INT NOT NULL DEFAULT 1 CHECK (per_broadcaster_cap BETWEEN 1 AND 10),
    selection_hour_utc INT NOT NULL DEFAULT 12 CHECK (selection_hour_utc BETWEEN 0 AND 23),
    priority INT NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN countdown_series.priority IS
    'Series with a lower priority select first each day; a clip goes to at most one edition per day';

CREATE TABLE countdown_editions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID NOT NULL REFERENCES countdown_series(id) ON DELETE CASCADE,
    edition_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'selected'
        CHECK (status IN ('selected', 'fetching', 'rendering', 'rendered', 'published', 'failed')),
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL CHECK (window_end > window_start),
    failure_reason TEXT CHECK (failure_reason IS NULL OR char_length(failure_reason) <= 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (series_id, edition_date),
    UNIQUE (id, edition_date)
);

CREATE INDEX countdown_editions_date_idx ON countdown_editions (edition_date DESC);

-- edition_date is repeated here so the database itself guarantees that one
-- clip appears in at most one edition per day.
CREATE TABLE countdown_edition_clips (
    edition_id UUID NOT NULL,
    edition_date DATE NOT NULL,
    position INT NOT NULL CHECK (position >= 1),
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    view_velocity DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (edition_id, position),
    UNIQUE (edition_date, clip_id),
    FOREIGN KEY (edition_id, edition_date) REFERENCES countdown_editions (id, edition_date) ON DELETE CASCADE
);

CREATE INDEX countdown_edition_clips_clip_idx ON countdown_edition_clips (clip_id);

-- A broadcaster or a single clip left out of every future edition: on
-- request, after a DMCA notice, or after a takedown or platform claim.
CREATE TABLE countdown_exclusions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    broadcaster_id TEXT CHECK (broadcaster_id IS NULL OR char_length(broadcaster_id) BETWEEN 1 AND 50),
    clip_id UUID REFERENCES clips(id) ON DELETE CASCADE,
    reason TEXT NOT NULL CHECK (reason IN ('request', 'dmca', 'takedown')),
    source TEXT NOT NULL CHECK (source IN ('admin', 'self_service')),
    note TEXT CHECK (note IS NULL OR char_length(note) <= 500),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((broadcaster_id IS NULL) <> (clip_id IS NULL))
);

CREATE UNIQUE INDEX countdown_exclusions_broadcaster_idx ON countdown_exclusions (broadcaster_id) WHERE broadcaster_id IS NOT NULL;
CREATE UNIQUE INDEX countdown_exclusions_clip_idx ON countdown_exclusions (clip_id) WHERE clip_id IS NOT NULL;
