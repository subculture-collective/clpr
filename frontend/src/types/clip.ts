import type { ClipSourceFields, ClipSourceType } from './submission';

export interface ClipSubmitter {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    is_verified?: boolean;
}

export interface Clip extends ClipSourceFields {
    id: string;
    twitch_clip_id: string;
    twitch_clip_url: string;
    embed_url: string;
    title: string;
    creator_name: string;
    creator_id?: string;
    broadcaster_name: string;
    broadcaster_id?: string;
    game_id?: string;
    game_name?: string;
    twitch_category_id?: string;
    twitch_category_name?: string;
    language?: string;
    thumbnail_url?: string;
    duration?: number;
    view_count: number;
    created_at: string;
    imported_at: string;
    video_url?: string; // HLS video URL for clips with adaptive streaming support
    stream_source?: ClipSourceType | 'stream'; // 'twitch' = imported clip, 'stream' = created from live stream
    vote_score: number;
    comment_count: number;
    favorite_count: number;
    is_featured: boolean;
    is_nsfw: boolean;
    is_removed: boolean;
    removed_reason?: string;
    is_hidden?: boolean;
    // User interaction state (from API or local)
    user_vote?: 1 | -1 | null;
    is_favorited?: boolean; // Backend returns is_favorited, not user_favorited
    upvote_count?: number;
    downvote_count?: number;
    // Submitter attribution
    submitted_by?: ClipSubmitter;
    submitted_at?: string;
    // Trending and popularity metrics
    trending_score?: number;
    hot_score?: number;
    popularity_index?: number;
    engagement_count?: number;
    // Watch progress (when available from watch history)
    watch_progress?: {
        progress_seconds: number;
        duration_seconds: number;
        progress_percent: number;
        completed: boolean;
        watched_at: string;
    };
}

export interface ClipFeedResponse {
    clips: Clip[];
    total: number;
    page: number;
    limit?: number;
    total_pages?: number;
    has_more: boolean;
    has_next?: boolean;
    has_prev?: boolean;
    cursor?: string; // Cursor for cursor-based pagination
}

export type SortOption =
    | 'hot'
    | 'new'
    | 'top'
    | 'rising'
    | 'discussed'
    | 'views'
    | 'trending'
    | 'popular';
export type TimeFrame = 'hour' | 'day' | 'week' | 'month' | 'year' | 'all';

export interface ClipFeedFilters {
    sort?: SortOption;
    timeframe?: TimeFrame;
    game_id?: string;
    twitch_category_id?: string;
    games?: string[]; // Multi-select game filter
    twitch_categories?: string[];
    creator_id?: string;
    streamers?: string[]; // Multi-select streamer filter
    tags?: string[];
    exclude_tags?: string[]; // Tags to exclude
    language?: string;
    nsfw?: boolean;
    top10k_streamers?: boolean;
    show_all_clips?: boolean; // If true, show both user-submitted and scraped clips (for discovery)
    date_from?: string; // ISO 8601 date string
    date_to?: string; // ISO 8601 date string
}

export interface VotePayload {
    clip_id: string;
    vote_type: 1 | -1;
}

export interface FavoritePayload {
    clip_id: string;
}

// Filter preset types
export interface FilterPreset {
    id: string;
    user_id: string;
    name: string;
    filters_json: string;
    created_at: string;
    updated_at: string;
}

export interface FilterPresetFilters {
    games?: string[];
    streamers?: string[];
    tags?: string[];
    exclude_tags?: string[];
    date_from?: string;
    date_to?: string;
    sort?: string;
    language?: string;
    nsfw?: boolean;
}

export interface CreateFilterPresetRequest {
    name: string;
    filters: FilterPresetFilters;
}

export interface UpdateFilterPresetRequest {
    name?: string;
    filters?: FilterPresetFilters;
}
