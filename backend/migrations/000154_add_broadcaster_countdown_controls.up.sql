-- Broadcaster controls for countdowns (#75).

-- A self-service exclusion records which consent text the broadcaster saw.
ALTER TABLE countdown_exclusions
    ADD COLUMN consent_text_version TEXT CHECK (consent_text_version IS NULL OR char_length(consent_text_version) BETWEEN 1 AND 40);

-- Uploaded videos that show a clip excluded after upload must be removed or
-- edited within 7 days. The task list is computed from countdown_uploads and
-- the exclusions; this table records which tasks an admin resolved.
CREATE TABLE countdown_removals (
    render_id UUID NOT NULL REFERENCES countdown_renders(id) ON DELETE CASCADE,
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    resolution TEXT NOT NULL CHECK (resolution IN ('removed', 'edited', 'kept')),
    note TEXT CHECK (note IS NULL OR char_length(note) <= 500),
    resolved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (render_id, clip_id)
);

-- "Feature my clips": the broadcaster granted channel:manage:clips so clpr
-- can fetch their clips through Helix and feature them.
CREATE TABLE broadcaster_feature_optins (
    broadcaster_id TEXT PRIMARY KEY CHECK (char_length(broadcaster_id) BETWEEN 1 AND 50),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT NOT NULL CHECK (char_length(scope) BETWEEN 1 AND 200),
    consent_text_version TEXT NOT NULL CHECK (char_length(consent_text_version) BETWEEN 1 AND 40),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX broadcaster_feature_optins_user_idx ON broadcaster_feature_optins (user_id);
