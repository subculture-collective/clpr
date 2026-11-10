-- Weekly streamer digest email (#85). Broadcasters who feature their clips
-- get one email a week with their top clips, countdown placements and their
-- clip pack. notify_streamer_digest is its own preference and unsubscribe.
ALTER TABLE notification_preferences
    ADD COLUMN notify_streamer_digest BOOLEAN NOT NULL DEFAULT TRUE;

-- One row per user and week, written before the email is sent, so no digest
-- goes out twice even with several API replicas.
CREATE TABLE streamer_digest_sends (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_start DATE NOT NULL CHECK (EXTRACT(ISODOW FROM week_start) = 1),
    status TEXT NOT NULL DEFAULT 'sending' CHECK (status IN ('sending', 'sent', 'skipped', 'failed')),
    reason TEXT CHECK (reason IS NULL OR char_length(reason) <= 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, week_start)
);
