CREATE TABLE tag_suppressions (
    tag_id UUID PRIMARY KEY REFERENCES tags(id) ON DELETE CASCADE,
    suppressed_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason TEXT NOT NULL CHECK (char_length(reason) BETWEEN 1 AND 500),
    suppressed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX tag_suppressions_suppressed_at_idx ON tag_suppressions(suppressed_at DESC);
