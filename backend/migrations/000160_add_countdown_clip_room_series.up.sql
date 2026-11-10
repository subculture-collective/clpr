-- Clip room series (#88): a countdown series can take a streamer's clip
-- room as its source. Its editions are the room's approved clips, in the
-- room's order, instead of the fastest-rising clips. If the room is deleted
-- the series keeps its source and its editions fail with a reason.
ALTER TABLE countdown_series
    ADD COLUMN source TEXT NOT NULL DEFAULT 'velocity' CHECK (source IN ('velocity', 'clip_room')),
    ADD COLUMN clip_room_id UUID REFERENCES streamer_clip_rooms(id) ON DELETE SET NULL;

COMMENT ON COLUMN countdown_series.source IS
    'velocity: the fastest-rising clips of the window; clip_room: the approved clips of clip_room_id, in room order';

CREATE INDEX streamer_clip_room_items_room_approved_idx
    ON streamer_clip_room_items (room_id, approved_at)
    WHERE status = 'approved' AND clip_id IS NOT NULL;
