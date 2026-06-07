import type { Clip } from './clip';

export type StreamerClipRoomApprovalMode = 'manual' | 'auto';
export type StreamerClipRoomItemStatus = 'pending' | 'approved' | 'rejected' | 'skipped';
export type StreamerClipRoomSourceType = 'clpr' | 'twitch';
export type StreamerClipRoomEventType =
    | 'item_detected'
    | 'item_approved'
    | 'item_rejected'
    | 'items_reordered'
    | 'room_status_changed';

export interface StreamerClipRoom {
    id: string;
    owner_user_id: string;
    twitch_channel: string;
    approval_mode: StreamerClipRoomApprovalMode;
    is_active: boolean;
    last_listener_error?: string;
    listener_started_at?: string;
    created_at: string;
    updated_at: string;
}

export interface StreamerClipRoomItem {
    id: string;
    room_id: string;
    clip_id?: string;
    clip?: Clip;
    source_url: string;
    source_type: StreamerClipRoomSourceType;
    status: StreamerClipRoomItemStatus;
    position?: number;
    twitch_message_id?: string;
    twitch_user_id?: string;
    twitch_username?: string;
    message_text?: string;
    skip_reason?: string;
    detected_at: string;
    approved_at?: string;
    approved_by_user_id?: string;
    rejected_at?: string;
    rejected_by_user_id?: string;
    created_at: string;
    updated_at: string;
}

export interface StreamerClipRoomWithItems {
    room: StreamerClipRoom;
    pending_items: StreamerClipRoomItem[];
    approved_items: StreamerClipRoomItem[];
    skipped_items: StreamerClipRoomItem[];
}

export interface StreamerClipRoomEvent {
    type: StreamerClipRoomEventType;
    data: {
        room?: StreamerClipRoom;
        item?: StreamerClipRoomItem;
        item_ids?: string[];
        [key: string]: unknown;
    };
}
