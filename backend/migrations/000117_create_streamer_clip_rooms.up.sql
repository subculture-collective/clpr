CREATE TABLE streamer_clip_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    twitch_channel TEXT NOT NULL,
    approval_mode TEXT NOT NULL DEFAULT 'manual' CHECK (approval_mode IN ('manual', 'auto')),
    is_active BOOLEAN NOT NULL DEFAULT false,
    last_listener_error TEXT,
    listener_started_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE streamer_clip_room_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES streamer_clip_rooms(id) ON DELETE CASCADE,
    clip_id UUID REFERENCES clips(id) ON DELETE SET NULL,
    source_url TEXT NOT NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('clpr', 'twitch')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'skipped')),
    position INTEGER,
    twitch_message_id TEXT,
    twitch_user_id TEXT,
    twitch_username TEXT,
    message_text TEXT,
    skip_reason TEXT,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    approved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    rejected_at TIMESTAMPTZ,
    rejected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT streamer_clip_room_items_approved_position CHECK (
        (status = 'approved' AND position IS NOT NULL)
        OR (status <> 'approved')
    )
);

CREATE UNIQUE INDEX streamer_clip_rooms_owner_channel_idx
    ON streamer_clip_rooms (owner_user_id, lower(twitch_channel));

CREATE INDEX streamer_clip_room_items_room_status_position_idx
    ON streamer_clip_room_items (room_id, status, position NULLS LAST, detected_at DESC);

CREATE INDEX streamer_clip_room_items_room_detected_idx
    ON streamer_clip_room_items (room_id, detected_at DESC);

CREATE UNIQUE INDEX streamer_clip_room_items_message_url_idx
    ON streamer_clip_room_items (room_id, twitch_message_id, source_url)
    WHERE twitch_message_id IS NOT NULL;

CREATE UNIQUE INDEX streamer_clip_room_items_approved_position_idx
    ON streamer_clip_room_items (room_id, position)
    WHERE status = 'approved';
