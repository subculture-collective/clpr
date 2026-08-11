import apiClient from './api';
import type {
    Clip,
    ClipFeedResponse,
    ClipFeedFilters,
    VotePayload,
    FavoritePayload,
} from '@/types/clip';

/**
 * Helper function to build clip query parameters from filters
 */
function buildClipParams(
    cursor: string | undefined,
    filters?: ClipFeedFilters,
): Record<string, string | number | boolean> {
    const params: Record<string, string | number | boolean> = {
        limit: 20,
    };

    // Add cursor for cursor-based pagination
    if (cursor) {
        params.cursor = cursor;
    }

    // Add filters to params, only if they are defined
    if (filters) {
        if (filters.sort) params.sort = filters.sort;
        if (filters.timeframe) params.timeframe = filters.timeframe;
        if (filters.twitch_category_id || filters.game_id) {
            params.twitch_category_id = filters.twitch_category_id || filters.game_id || '';
        }
        if (filters.creator_id) params.creator_id = filters.creator_id;
        if (filters.language) params.language = filters.language;
        if (filters.nsfw !== undefined) params.nsfw = filters.nsfw;
        if (filters.top10k_streamers !== undefined)
            params.top10k_streamers = filters.top10k_streamers;
        if (filters.show_all_clips !== undefined)
            params.show_all_clips = filters.show_all_clips;
        if (filters.tags && filters.tags.length > 0) {
            params.tag = filters.tags.join(',');
        }
        // New filter parameters
        // Note: Backend currently supports only one Twitch category/creator selection
        // Multi-select support requires backend query modifications for OR logic
        const twitchCategories = filters.twitch_categories || filters.games;
        if (twitchCategories && twitchCategories.length > 0) {
            params.twitch_category_id = twitchCategories[0];
        }
        if (filters.streamers && filters.streamers.length > 0) {
            params.broadcaster_id = filters.streamers[0];
        }
        if (filters.exclude_tags && filters.exclude_tags.length > 0) {
            params.exclude_tags = filters.exclude_tags.join(',');
        }
        if (filters.date_from) params.date_from = filters.date_from;
        if (filters.date_to) params.date_to = filters.date_to;
    }

    return params;
}

/**
 * Fetch clips with pagination and filters
 */
export async function fetchClips({
    cursor,
    filters,
}: {
    cursor?: string;
    filters?: ClipFeedFilters;
}): Promise<ClipFeedResponse> {
    const params = buildClipParams(cursor, filters);

    const response = await apiClient.get<{
        success: boolean;
        clips: Clip[];
        pagination: {
            limit: number;
            offset: number;
            total: number;
            total_pages?: number;
            has_more: boolean;
            cursor?: string;
        };
    }>('/feeds/clips', { params });

    return {
        clips: response.data.clips,
        total: response.data.pagination.total,
        // Only calculate page for offset-based pagination
        page:
            response.data.pagination.cursor ?
                1
            :   Math.floor(
                    response.data.pagination.offset /
                        response.data.pagination.limit,
                ) + 1,
        limit: response.data.pagination.limit,
        has_more: response.data.pagination.has_more,
        cursor: response.data.pagination.cursor,
    };
}

/**
 * Fetch scraped clips (not yet submitted by users) with pagination and filters
 */
export async function fetchScrapedClips({
    pageParam = 1,
    filters,
}: {
    pageParam?: number;
    filters?: ClipFeedFilters;
}): Promise<ClipFeedResponse> {
    const params: Record<string, string | number | boolean> = {
        page: pageParam,
        limit: 10,
    };

    // Add filters manually for scraped clips (still uses offset pagination)
    if (filters) {
        if (filters.sort) params.sort = filters.sort;
        if (filters.timeframe) params.timeframe = filters.timeframe;
        if (filters.twitch_category_id || filters.game_id) {
            params.twitch_category_id = filters.twitch_category_id || filters.game_id || '';
        }
        if (filters.creator_id) params.creator_id = filters.creator_id;
        if (filters.language) params.language = filters.language;
    }

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
    }>('/scraped-clips', { params });

    return {
        clips: response.data.data,
        total: response.data.meta.total,
        page: response.data.meta.page,
        limit: response.data.meta.limit,
        has_more: response.data.meta.has_next,
    };
}

/**
 * Fetch favorite clips with pagination and filters
 */
export async function fetchFavorites({
    pageParam = 1,
    sort = 'newest',
}: {
    pageParam?: number;
    sort?: 'newest' | 'top' | 'discussed';
}): Promise<ClipFeedResponse> {
    const params: Record<string, string | number> = {
        page: pageParam,
        limit: 10,
        sort,
    };

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
    }>('/favorites', { params });

    return {
        clips: response.data.data,
        total: response.data.meta.total,
        page: response.data.meta.page,
        limit: response.data.meta.limit,
        has_more: response.data.meta.has_next,
    };
}

/**
 * Fetch a single clip by ID
 */
export async function fetchClipById(clipId: string): Promise<Clip> {
    const response = await apiClient.get<{
        success: boolean;
        data: Clip;
    }>(`/clips/${clipId}`);

    return response.data.data;
}

/**
 * Vote on a clip
 */
export async function voteOnClip(payload: VotePayload): Promise<{
    message: string;
    vote_score: number;
    upvote_count: number;
    downvote_count: number;
    user_vote: number;
}> {
    const response = await apiClient.post<{
        success: boolean;
        data: {
            message: string;
            vote_score: number;
            upvote_count: number;
            downvote_count: number;
            user_vote: number;
        };
    }>(`/clips/${payload.clip_id}/vote`, {
        vote: payload.vote_type,
    });

    return response.data.data;
}

/**
 * Add a clip to favorites
 */
export async function addFavorite(payload: FavoritePayload): Promise<{
    message: string;
    is_favorited: boolean;
}> {
    const response = await apiClient.post<{
        success: boolean;
        data: {
            message: string;
            is_favorited: boolean;
        };
    }>(`/clips/${payload.clip_id}/favorite`);

    return response.data.data;
}

/**
 * Remove a clip from favorites
 */
export async function removeFavorite(payload: FavoritePayload): Promise<{
    message: string;
    is_favorited: boolean;
}> {
    const response = await apiClient.delete<{
        success: boolean;
        data: {
            message: string;
            is_favorited: boolean;
        };
    }>(`/clips/${payload.clip_id}/favorite`);

    return response.data.data;
}

/**
 * Update clip metadata (title) - creator only
 */
export async function updateClipMetadata(
    clipId: string,
    payload: {
        title?: string;
        tags?: string[];
    },
): Promise<{ message: string }> {
    const response = await apiClient.put<{
        success: boolean;
        data: { message: string };
    }>(`/clips/${clipId}/metadata`, payload);

    return response.data.data;
}

/**
 * Update clip visibility (hidden status) - creator only
 */
export async function updateClipVisibility(
    clipId: string,
    isHidden: boolean,
): Promise<{
    message: string;
    is_hidden: boolean;
}> {
    const response = await apiClient.put<{
        success: boolean;
        data: { message: string; is_hidden: boolean };
    }>(`/clips/${clipId}/visibility`, {
        is_hidden: isHidden,
    });

    return response.data.data;
}

/**
 * Fetch clips for a specific creator
 */
export async function fetchCreatorClips({
    creatorId,
    pageParam = 1,
    limit = 25,
}: {
    creatorId: string;
    pageParam?: number;
    limit?: number;
}): Promise<ClipFeedResponse> {
    const params: Record<string, string | number> = {
        page: pageParam,
        limit,
    };

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
    }>(`/creators/${creatorId}/clips`, { params });

    return {
        clips: response.data.data,
        page: response.data.meta.page,
        total: response.data.meta.total,
        limit: response.data.meta.limit,
        total_pages: response.data.meta.total_pages,
        has_more: response.data.meta.has_next,
        has_next: response.data.meta.has_next,
        has_prev: response.data.meta.has_prev,
    };
}
