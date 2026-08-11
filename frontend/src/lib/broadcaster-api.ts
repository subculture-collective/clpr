import apiClient from './api';
import type { Clip } from '@/types/clip';

export interface PopularBroadcaster {
    broadcaster_id: string;
    broadcaster_name: string;
    clip_count: number;
}

export interface CreatorDiscoveryProfile {
    broadcaster_id: string;
    broadcaster_name: string;
    total_clips: number;
    recent_clips: number;
    total_views: number;
    recent_views: number;
    view_velocity: number;
    follower_count: number;
    first_discovered_at: string;
    latest_clip_at: string;
    latest_clip_thumbnail?: string;
    latest_clip_title?: string;
    twitch_category_name?: string;
    score: number;
}

export interface CreatorDiscoveryRails {
    trending: CreatorDiscoveryProfile[];
    rising: CreatorDiscoveryProfile[];
    new: CreatorDiscoveryProfile[];
}

export interface BroadcasterProfile {
    broadcaster_id: string;
    broadcaster_name: string;
    display_name: string;
    avatar_url?: string;
    bio?: string;
    twitch_url: string;
    total_clips: number;
    follower_count: number;
    total_views: number;
    avg_vote_score: number;
    is_following: boolean;
    updated_at: string;
}

export interface BroadcasterClipsResponse {
    success: boolean;
    data: Clip[];
    meta: {
        page: number;
        limit: number;
        total_items: number;
        total_pages: number;
    };
}

export interface BroadcasterClipsFilters {
    page?: number;
    limit?: number;
    sort?: 'recent' | 'popular' | 'trending';
}

/**
 * Fetch popular broadcasters by clip count
 */
export async function fetchPopularBroadcasters(
    limit: number = 15,
): Promise<PopularBroadcaster[]> {
    const response = await apiClient.get<{
        broadcasters: PopularBroadcaster[];
    }>(`/broadcasters/popular`, { params: { limit } });
    return response.data.broadcasters || [];
}

/** Fetch creator-first discovery rails ranked by momentum and freshness. */
export async function fetchCreatorDiscovery(
    limit: number = 12,
): Promise<CreatorDiscoveryRails> {
    const response = await apiClient.get<{
        success: boolean;
        data: CreatorDiscoveryRails;
    }>('/broadcasters/discover', { params: { limit } });
    return response.data.data;
}

/**
 * Fetch broadcaster profile by broadcaster ID
 */
export async function fetchBroadcasterProfile(
    broadcasterId: string,
): Promise<BroadcasterProfile> {
    const response = await apiClient.get<BroadcasterProfile>(
        `/broadcasters/${broadcasterId}`,
    );
    return response.data;
}

/**
 * Fetch clips for a broadcaster with pagination and sorting
 */
export async function fetchBroadcasterClips(
    broadcasterId: string,
    filters?: BroadcasterClipsFilters,
): Promise<BroadcasterClipsResponse> {
    const params: Record<string, string | number> = {
        page: filters?.page || 1,
        limit: filters?.limit || 20,
    };

    if (filters?.sort) {
        params.sort = filters.sort;
    }

    const response = await apiClient.get<BroadcasterClipsResponse>(
        `/broadcasters/${broadcasterId}/clips`,
        { params },
    );
    return response.data;
}

/**
 * Follow a broadcaster
 */
export async function followBroadcaster(
    broadcasterId: string,
): Promise<{ message: string }> {
    const response = await apiClient.post<{ message: string }>(
        `/broadcasters/${broadcasterId}/follow`,
    );
    return response.data;
}

/**
 * Unfollow a broadcaster
 */
export async function unfollowBroadcaster(
    broadcasterId: string,
): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(
        `/broadcasters/${broadcasterId}/follow`,
    );
    return response.data;
}

/**
 * Live status response interface
 */
export interface BroadcasterLiveStatus {
    broadcaster_id: string;
    user_login?: string;
    user_name?: string;
    is_live: boolean;
    stream_title?: string;
    game_name?: string;
    viewer_count: number;
    started_at?: string;
    last_checked: string;
    created_at: string;
    updated_at: string;
}

/**
 * Live broadcasters response interface
 */
export interface LiveBroadcastersResponse {
    success: boolean;
    data: BroadcasterLiveStatus[];
    meta: {
        page: number;
        limit: number;
        total_items: number;
        total_pages: number;
    };
}

/**
 * Followed live broadcasters response interface
 */
export interface FollowedLiveBroadcastersResponse {
    success: boolean;
    data: BroadcasterLiveStatus[];
}

/**
 * Fetch live status for a specific broadcaster
 */
export async function fetchBroadcasterLiveStatus(
    broadcasterId: string,
): Promise<BroadcasterLiveStatus> {
    const response = await apiClient.get<BroadcasterLiveStatus>(
        `/broadcasters/${broadcasterId}/live-status`,
    );
    return response.data;
}

/**
 * Fetch all currently live broadcasters
 */
export async function fetchLiveBroadcasters(
    page: number = 1,
    limit: number = 50,
): Promise<LiveBroadcastersResponse> {
    const response = await apiClient.get<LiveBroadcastersResponse>(
        `/broadcasters/live`,
        { params: { page, limit } },
    );
    return response.data;
}

/**
 * Fetch live broadcasters that the authenticated user follows
 */
export async function fetchFollowedLiveBroadcasters(): Promise<FollowedLiveBroadcastersResponse> {
    const response =
        await apiClient.get<FollowedLiveBroadcastersResponse>(`/feed/live`);
    return response.data;
}
