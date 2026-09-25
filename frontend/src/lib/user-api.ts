import apiClient from './api';
import type { Comment } from '../types/comment';
import type { Clip, ClipFeedResponse } from '../types/clip';

export interface UserCommentsResponse {
    comments: Comment[];
    total: number;
    page: number;
    limit: number;
    has_more: boolean;
}

/**
 * Fetch comments by a user
 */
export async function fetchUserComments(
    userId: string,
    page: number = 1,
    limit: number = 10
): Promise<UserCommentsResponse> {
    const response = await apiClient.get<UserCommentsResponse>(
        `/users/${userId}/comments`,
        {
            params: { page, limit },
        }
    );
    return response.data;
}

/**
 * Fetch clips that a user has upvoted
 */
export async function fetchUserUpvotedClips(
    userId: string,
    page: number = 1,
    limit: number = 10
): Promise<ClipFeedResponse> {
    const response = await apiClient.get<{
        success: boolean;
        data: Clip[];
        meta: {
            page: number;
            limit: number;
            total: number;
            total_pages: number;
            has_next: boolean;
            has_prev: boolean;
        };
    }>(`/users/${userId}/upvoted`, {
        params: { page, limit },
    });

    return {
        clips: response.data.data ?? [],
        total: response.data.meta.total,
        page: response.data.meta.page,
        limit: response.data.meta.limit,
        has_more: response.data.meta.has_next,
    };
}

/**
 * Fetch clips that a user has downvoted
 */
export async function fetchUserDownvotedClips(
    userId: string,
    page: number = 1,
    limit: number = 10
): Promise<ClipFeedResponse> {
    const response = await apiClient.get<{
        success: boolean;
        data: Clip[];
        meta: {
            page: number;
            limit: number;
            total: number;
            total_pages: number;
            has_next: boolean;
            has_prev: boolean;
        };
    }>(`/users/${userId}/downvoted`, {
        params: { page, limit },
    });

    return {
        clips: response.data.data ?? [],
        total: response.data.meta.total,
        page: response.data.meta.page,
        limit: response.data.meta.limit,
        has_more: response.data.meta.has_next,
    };
}

/**
 * Initiate Twitch reauthorization flow
 */
export async function reauthorizeTwitch(): Promise<{ auth_url: string }> {
    const response = await apiClient.post<{ auth_url: string }>(
        '/auth/twitch/reauthorize'
    );
    return response.data;
}

export interface UserProfile {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    bio?: string;
    social_links?: string; // JSON string
    karma_points: number;
    trust_score: number;
    role: string;
    is_banned: boolean;
    ban_reason?: string;
    banned_until?: string;
    follower_count: number;
    following_count: number;
    created_at: string;
    updated_at: string;
    stats: {
        clips_submitted: number;
        total_upvotes: number;
        total_comments: number;
        clips_featured: number;
        broadcasters_followed: number;
    };
    is_following: boolean;
    is_followed_by: boolean;
}

export interface UserActivityItem {
    id: string;
    user_id: string;
    activity_type: string;
    target_id?: string;
    target_type?: string;
    metadata?: string;
    created_at: string;
    username: string;
    user_avatar?: string;
    clip_title?: string;
    clip_id?: string;
    comment_text?: string;
    target_user?: string;
}

export interface FollowerUser {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    bio?: string;
    karma_points: number;
    followed_at: string;
    is_following: boolean;
}

export interface UserActivityResponse {
    success: boolean;
    data: UserActivityItem[];
    meta: {
        page: number;
        limit: number;
        total: number;
        total_pages: number;
        has_next: boolean;
        has_prev: boolean;
    };
}

export interface UserFollowersResponse {
    success: boolean;
    data: FollowerUser[];
    meta: {
        page: number;
        limit: number;
        total: number;
        total_pages: number;
        has_next: boolean;
        has_prev: boolean;
    };
}

/**
 * Fetch a user's complete profile with stats
 */
export async function fetchUserProfile(userId: string): Promise<UserProfile> {
    const response = await apiClient.get<{ success: boolean; data: UserProfile }>(
        `/users/${userId}`
    );
    return response.data.data;
}

/**
 * Fetch clips submitted by a user
 */
export async function fetchUserClips(
    userId: string,
    page: number = 1,
    limit: number = 20
): Promise<ClipFeedResponse> {
    const response = await apiClient.get<{
        success: boolean;
        data: Clip[];
        meta: {
            page: number;
            limit: number;
            total: number;
            total_pages: number;
            has_next: boolean;
            has_prev: boolean;
        };
    }>(`/users/${userId}/clips`, {
        params: { page, limit },
    });

    return {
        clips: response.data.data ?? [],
        total: response.data.meta.total,
        page: response.data.meta.page,
        limit: response.data.meta.limit,
        has_more: response.data.meta.has_next,
    };
}

/**
 * Fetch a user's activity feed
 */
export async function fetchUserActivity(
    userId: string,
    page: number = 1,
    limit: number = 20
): Promise<UserActivityResponse> {
    const response = await apiClient.get<UserActivityResponse>(
        `/users/${userId}/activity`,
        {
            params: { page, limit },
        }
    );
    return response.data;
}

/**
 * Fetch a user's followers
 */
export async function fetchUserFollowers(
    userId: string,
    page: number = 1,
    limit: number = 20
): Promise<UserFollowersResponse> {
    const response = await apiClient.get<UserFollowersResponse>(
        `/users/${userId}/followers`,
        {
            params: { page, limit },
        }
    );
    return response.data;
}

/**
 * Fetch users that a user is following
 */
export async function fetchUserFollowing(
    userId: string,
    page: number = 1,
    limit: number = 20
): Promise<UserFollowersResponse> {
    const response = await apiClient.get<UserFollowersResponse>(
        `/users/${userId}/following`,
        {
            params: { page, limit },
        }
    );
    return response.data;
}

/**
 * Follow a user
 */
export async function followUser(userId: string): Promise<void> {
    await apiClient.post(`/users/${userId}/follow`);
}

/**
 * Unfollow a user
 */
export async function unfollowUser(userId: string): Promise<void> {
    await apiClient.delete(`/users/${userId}/follow`);
}

/**
 * Update social links
 */
export async function updateSocialLinks(socialLinks: {
    twitter?: string;
    twitch?: string;
    discord?: string;
    youtube?: string;
    website?: string;
}): Promise<void> {
    await apiClient.put('/users/me/social-links', socialLinks);
}

/**
 * User suggestion for autocomplete/mentions
 */
export interface UserSuggestion {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    is_verified: boolean;
}

/**
 * Search users for autocomplete (e.g., chat mentions)
 * @param query - Username prefix to search for
 * @param limit - Maximum number of results (default: 10, max: 20)
 */
export async function searchUsersAutocomplete(
    query: string,
    limit: number = 10
): Promise<UserSuggestion[]> {
    if (!query || query.trim() === '') {
        return [];
    }

    const response = await apiClient.get<{
        success: boolean;
        data: UserSuggestion[];
    }>('/users/autocomplete', {
        params: { q: query, limit },
    });

    return response.data.data;
}
