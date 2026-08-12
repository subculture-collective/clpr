import apiClient from './api';
import type { ChannelMember, ChannelRole, Channel } from '@/types/chat';

export interface ChatChannel {
    id: string;
    name: string;
    description?: string;
    creator_id: string;
    channel_type: 'public' | 'private' | 'direct';
    is_active: boolean;
    max_participants?: number;
    created_at: string;
    updated_at: string;
}

export interface ChatMessage {
    id: string;
    channel_id: string;
    user_id: string;
    content: string;
    is_deleted: boolean;
    deleted_at?: string;
    deleted_by?: string;
    created_at: string;
    updated_at: string;
    username?: string;
    display_name?: string;
    avatar_url?: string;
}

export interface ChatBan {
    id: string;
    channel_id: string;
    user_id: string;
    banned_by: string;
    reason?: string;
    expires_at?: string;
    created_at: string;
    banned_by_username?: string;
    target_username?: string;
}

export interface ChatModerationLog {
    id: string;
    channel_id: string;
    moderator_id: string;
    target_user_id?: string;
    action: 'ban' | 'unban' | 'mute' | 'unmute' | 'timeout' | 'delete_message';
    reason?: string;
    metadata?: Record<string, unknown>;
    created_at: string;
    moderator_username?: string;
    target_username?: string;
}

export interface BanUserRequest {
    user_id: string;
    reason?: string;
    duration_minutes?: number;
}

export interface MuteUserRequest {
    user_id: string;
    reason?: string;
    duration_minutes?: number;
}

export interface TimeoutUserRequest {
    user_id: string;
    reason: string;
    duration_minutes: number;
}

export interface DeleteMessageRequest {
    reason?: string;
}

export interface BanUserResponse {
    status: 'banned';
    ban_id: string;
    expires_at?: string;
}

export interface UnbanUserResponse {
    status: 'unbanned';
}

export interface MuteUserResponse {
    status: 'muted';
    mute_id: string;
    expires_at?: string;
}

export interface TimeoutUserResponse {
    status: 'timed_out';
    timeout_id: string;
    expires_at: string;
}

export interface DeleteMessageResponse {
    status: 'deleted';
}

export interface CheckBanResponse {
    is_banned: boolean;
    ban_id?: string;
    expires_at?: string;
    reason?: string;
}

export interface ModerationLogResponse {
    logs: ChatModerationLog[];
    total: number;
    page: number;
    limit: number;
}

/** List chat channels visible to the current user. */
export async function listChannels(): Promise<Channel[]> {
    const response = await apiClient.get<
        Channel[] | { channels?: Channel[] | null }
    >('/chat/channels');
    const payload = response.data;
    if (Array.isArray(payload)) return payload;
    return Array.isArray(payload?.channels) ? payload.channels : [];
}

// Ban a user from a channel
export async function banUser(
    channelId: string,
    request: BanUserRequest
): Promise<BanUserResponse> {
    const response = await apiClient.post<BanUserResponse>(
        `/chat/channels/${channelId}/ban`,
        request
    );
    return response.data;
}

// Unban a user by ban ID (uses moderation API)
export async function unbanUser(banId: string): Promise<UnbanUserResponse> {
    const response = await apiClient.delete<UnbanUserResponse>(
        `/moderation/ban/${banId}`
    );
    return response.data;
}

// Unmute a user in a channel (legacy support)
export async function unmuteUser(
    channelId: string,
    userId: string
): Promise<UnbanUserResponse> {
    const response = await apiClient.delete<UnbanUserResponse>(
        `/chat/channels/${channelId}/ban/${userId}`
    );
    return response.data;
}

// Mute a user in a channel
export async function muteUser(
    channelId: string,
    request: MuteUserRequest
): Promise<MuteUserResponse> {
    const response = await apiClient.post<MuteUserResponse>(
        `/chat/channels/${channelId}/mute`,
        request
    );
    return response.data;
}

// Timeout a user (temporary ban) in a channel
export async function timeoutUser(
    channelId: string,
    request: TimeoutUserRequest
): Promise<TimeoutUserResponse> {
    const response = await apiClient.post<TimeoutUserResponse>(
        `/chat/channels/${channelId}/timeout`,
        request
    );
    return response.data;
}

// Delete a chat message
export async function deleteMessage(
    messageId: string,
    request?: DeleteMessageRequest
): Promise<DeleteMessageResponse> {
    const response = await apiClient.delete<DeleteMessageResponse>(
        `/chat/messages/${messageId}`,
        { data: request }
    );
    return response.data;
}

// Get moderation log for a channel
export async function getModerationLog(
    channelId: string,
    page: number = 1,
    limit: number = 50
): Promise<ModerationLogResponse> {
    const response = await apiClient.get<ModerationLogResponse>(
        `/chat/channels/${channelId}/moderation-log`,
        {
            params: { page, limit },
        }
    );
    return response.data;
}

// Check if a user is banned in a channel
export async function checkUserBan(
    channelId: string,
    userId: string
): Promise<CheckBanResponse> {
    const response = await apiClient.get<CheckBanResponse>(
        `/chat/channels/${channelId}/check-ban`,
        {
            params: { user_id: userId },
        }
    );
    return response.data;
}

// Get list of banned users in a channel
export async function getChannelBans(
    channelId: string,
    page: number = 1,
    limit: number = 50
): Promise<{ bans: ChatBan[]; total: number; page: number; limit: number }> {
    const response = await apiClient.get<{
        bans: ChatBan[];
        total: number;
        page: number;
        limit: number;
    }>(`/chat/channels/${channelId}/bans`, {
        params: { page, limit },
    });
    return response.data;
}

// Get bans across all channels (uses moderation API, admin only)
export async function getAllBans(
    page: number = 1,
    limit: number = 50,
    channelId?: string
): Promise<{ bans: ChatBan[]; total: number; page: number; limit: number }> {
    // Calculate offset from page for the moderation API
    const offset = (page - 1) * limit;
    const params: Record<string, string | number> = { limit, offset };
    if (channelId) {
        params.channelId = channelId;
    }
    const response = await apiClient.get<{
        bans: Array<{
            id: string;
            community_id: string;
            banned_user_id: string;
            banned_by_user_id?: string;
            reason?: string;
            banned_at: string;
        }>;
        total: number;
        limit: number;
        offset: number;
    }>('/moderation/bans', { params });
    // Map CommunityBan fields to ChatBan fields
    const mappedBans: ChatBan[] = (response.data.bans || []).map(ban => ({
        id: ban.id,
        channel_id: ban.community_id,
        user_id: ban.banned_user_id,
        banned_by: ban.banned_by_user_id || '',
        reason: ban.reason,
        created_at: ban.banned_at,
    }));
    return {
        bans: mappedBans,
        total: response.data.total,
        page,
        limit: response.data.limit,
    };
}

// Create a new channel
export interface CreateChannelRequest {
    name: string;
    description?: string;
    channel_type?: 'public' | 'private' | 'direct';
}

export async function createChannel(
    request: CreateChannelRequest
): Promise<Channel> {
    const response = await apiClient.post<Channel>(
        '/chat/channels',
        request
    );
    return response.data;
}

// Get current user's role in a channel
export async function getCurrentUserRole(
    channelId: string
): Promise<{ role: ChannelRole }> {
    const response = await apiClient.get<{ role: ChannelRole }>(
        `/chat/channels/${channelId}/role`
    );
    return response.data;
}

// List members of a channel
export async function listChannelMembers(
    channelId: string,
    limit: number = 50,
    offset: number = 0
): Promise<{ members: ChannelMember[]; limit: number; offset: number }> {
    const response = await apiClient.get<{
        members: ChannelMember[];
        limit: number;
        offset: number;
    }>(`/chat/channels/${channelId}/members`, {
        params: { limit, offset },
    });
    return response.data;
}

// Add a member to a channel
export async function addChannelMember(
    channelId: string,
    userId: string,
    role: 'member' | 'moderator' | 'admin' = 'member'
): Promise<ChannelMember> {
    const response = await apiClient.post<ChannelMember>(
        `/chat/channels/${channelId}/members`,
        { user_id: userId, role }
    );
    return response.data;
}

// Remove a member from a channel
export async function removeChannelMember(
    channelId: string,
    userId: string
): Promise<{ status: string }> {
    const response = await apiClient.delete<{ status: string }>(
        `/chat/channels/${channelId}/members/${userId}`
    );
    return response.data;
}

// Update a member's role in a channel
export async function updateChannelMemberRole(
    channelId: string,
    userId: string,
    role: 'member' | 'moderator' | 'admin'
): Promise<ChannelMember> {
    const response = await apiClient.patch<ChannelMember>(
        `/chat/channels/${channelId}/members/${userId}`,
        { role }
    );
    return response.data;
}

// Delete a channel (owner only)
export async function deleteChannel(
    channelId: string
): Promise<{ status: string }> {
    const response = await apiClient.delete<{ status: string }>(
        `/chat/channels/${channelId}`
    );
    return response.data;
}

// Sync bans from Twitch channel
export interface SyncBansFromTwitchRequest {
    channel_name: string;
}

export interface SyncBansFromTwitchResponse {
    job_id: string;
    status: string;
    message?: string;
}

export interface SyncBansProgressResponse {
    job_id: string;
    status: 'pending' | 'in_progress' | 'completed' | 'failed';
    bans_added: number;
    bans_existing: number;
    total_processed: number;
    error?: string;
}

export async function syncBansFromTwitch(
    channelId: string,
    request: SyncBansFromTwitchRequest
): Promise<SyncBansFromTwitchResponse> {
    const response = await apiClient.post<SyncBansFromTwitchResponse>(
        `/chat/sync-bans`,
        {
            ...request,
            channel_id: channelId,
        }
    );
    return response.data;
}

export async function checkSyncBansProgress(
    channelId: string,
    jobId: string
): Promise<SyncBansProgressResponse> {
    const response = await apiClient.get<SyncBansProgressResponse>(
        `/chat/sync-bans/${jobId}/progress`,
        {
            params: { channel_id: channelId },
        }
    );
    return response.data;
}
