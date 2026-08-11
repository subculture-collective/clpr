export type PlaylistScriptSort =
    | 'hot'
    | 'new'
    | 'top'
    | 'rising'
    | 'discussed'
    | 'trending'
    | 'popular';

export type PlaylistScriptTimeframe =
    | 'hour'
    | 'day'
    | 'week'
    | 'month'
    | 'year';

export type PlaylistScriptSchedule =
    | 'manual'
    | 'hourly'
    | 'daily'
    | 'weekly'
    | 'monthly';

export type PlaylistScriptStrategy =
    | 'standard'
    | 'sleeper_hits'
    | 'viral_velocity'
    | 'community_favorites'
    | 'deep_cuts'
    | 'fresh_faces'
    | 'one_per_creator'
    | 'diversity_roulette'
    | 'clip_of_the_day'
    | 'weekend_mix'
    | 'similar_vibes'
    | 'cross_game_hits'
    | 'controversial'
    | 'binge_worthy'
    | 'rising_stars'
    | 'twitch_top_game'
    | 'twitch_top_broadcaster'
    | 'twitch_trending'
    | 'twitch_discovery';

export type PlaylistScriptVisibility = 'private' | 'public' | 'unlisted';

export interface PlaylistScript {
    id: string;
    name: string;
    description?: string;
    sort: PlaylistScriptSort;
    timeframe?: PlaylistScriptTimeframe;
    clip_limit: number;
    visibility: PlaylistScriptVisibility;
    is_active: boolean;
    schedule: PlaylistScriptSchedule;
    strategy: PlaylistScriptStrategy;
    game_id?: string;
    game_ids?: string[];
    broadcaster_id?: string;
    tag?: string;
    tags: string[];
    tags_logic: 'and' | 'or';
    exclude_tags?: string[];
    language?: string;
    min_vote_score?: number;
    min_view_count?: number;
    exclude_nsfw: boolean;
    top_10k_streamers: boolean;
    seed_clip_id?: string;
    retention_days: number;
    title_template?: string;
    created_by?: string;
    created_at: string;
    updated_at: string;
    last_run_at?: string;
    last_generated_playlist_id?: string;
}

export interface CreatePlaylistScriptRequest {
    name: string;
    description?: string;
    sort: PlaylistScriptSort;
    timeframe?: PlaylistScriptTimeframe;
    clip_limit: number;
    visibility?: PlaylistScriptVisibility;
    is_active?: boolean;
    schedule?: PlaylistScriptSchedule;
    strategy?: PlaylistScriptStrategy;
    game_id?: string;
    game_ids?: string[];
    broadcaster_id?: string;
    tag?: string;
    tags?: string[];
    tags_logic?: 'and' | 'or';
    exclude_tags?: string[];
    language?: string;
    min_vote_score?: number;
    min_view_count?: number;
    exclude_nsfw?: boolean;
    top_10k_streamers?: boolean;
    seed_clip_id?: string;
    retention_days?: number;
    title_template?: string;
}

export interface UpdatePlaylistScriptRequest {
    name?: string;
    description?: string;
    sort?: PlaylistScriptSort;
    timeframe?: PlaylistScriptTimeframe;
    clip_limit?: number;
    visibility?: PlaylistScriptVisibility;
    is_active?: boolean;
    schedule?: PlaylistScriptSchedule;
    strategy?: PlaylistScriptStrategy;
    game_id?: string;
    game_ids?: string[];
    broadcaster_id?: string;
    tag?: string;
    tags?: string[];
    tags_logic?: 'and' | 'or';
    exclude_tags?: string[];
    language?: string;
    min_vote_score?: number;
    min_view_count?: number;
    exclude_nsfw?: boolean;
    top_10k_streamers?: boolean;
    seed_clip_id?: string;
    retention_days?: number;
    title_template?: string;
}
