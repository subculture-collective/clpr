import type {
    PlaylistScriptStrategy,
    PlaylistScriptSchedule,
    PlaylistScript,
} from '@/types/playlistScript';

export interface StrategyMeta {
    label: string;
    description: string;
    group: 'database' | 'twitch';
    requiresGameId?: boolean;
    requiresGameIds?: boolean;
    requiresBroadcasterId?: boolean;
    requiresSeedClipId?: boolean;
    hidesSortTimeframe?: boolean;
}

export const STRATEGY_META: Record<PlaylistScriptStrategy, StrategyMeta> = {
    standard: {
        label: 'Standard',
        description: 'Clip query with sort + timeframe filters',
        group: 'database',
    },
    sleeper_hits: {
        label: 'Sleeper Hits',
        description: 'Hidden gems — high retention but low view count',
        group: 'database',
        hidesSortTimeframe: true,
    },
    viral_velocity: {
        label: 'Viral Velocity',
        description: 'Fastest growing engagement in the last 48 hours',
        group: 'database',
        hidesSortTimeframe: true,
    },
    community_favorites: {
        label: 'Community Favorites',
        description: 'Highest save-to-view ratio among clips',
        group: 'database',
        hidesSortTimeframe: true,
    },
    deep_cuts: {
        label: 'Deep Cuts',
        description: 'High watch progress + good votes but under the radar',
        group: 'database',
        hidesSortTimeframe: true,
    },
    fresh_faces: {
        label: 'Fresh Faces',
        description: 'Best clips from new creators with 5 or fewer clips',
        group: 'database',
        hidesSortTimeframe: true,
    },
    one_per_creator: {
        label: 'One per Creator',
        description: 'One standout clip per creator for a broader mix',
        group: 'database',
        hidesSortTimeframe: true,
    },
    diversity_roulette: {
        label: 'Across the Culture',
        description: 'A creator-first mix spread across live-culture topics',
        group: 'database',
        hidesSortTimeframe: true,
    },
    clip_of_the_day: {
        label: 'Clip of the Day',
        description: 'The strongest current clip by momentum, views, and freshness',
        group: 'database',
        hidesSortTimeframe: true,
    },
    weekend_mix: {
        label: 'Weekend Mix',
        description: 'One clip per creator with a soft topic cap',
        group: 'database',
        hidesSortTimeframe: true,
    },
    similar_vibes: {
        label: 'Similar Vibes',
        description: 'Clips similar to a seed clip using AI embeddings',
        group: 'database',
        requiresSeedClipId: true,
        hidesSortTimeframe: true,
    },
    cross_game_hits: {
        label: 'Cross-Category Hits',
        description: 'Top clips across multiple specified Twitch categories',
        group: 'database',
        requiresGameIds: true,
        hidesSortTimeframe: true,
    },
    controversial: {
        label: 'Controversial',
        description: 'High comment density relative to views',
        group: 'database',
        hidesSortTimeframe: true,
    },
    binge_worthy: {
        label: 'Binge-Worthy',
        description: 'Clips from sessions where users watched 3+ in a row',
        group: 'database',
        hidesSortTimeframe: true,
    },
    rising_stars: {
        label: 'Rising Stars',
        description: 'Creators whose recent performance beats their average',
        group: 'database',
        hidesSortTimeframe: true,
    },
    twitch_top_game: {
        label: 'Twitch Top Category',
        description: 'Import top Twitch clips for a specific category',
        group: 'twitch',
        requiresGameId: true,
        hidesSortTimeframe: true,
    },
    twitch_top_broadcaster: {
        label: 'Twitch Top Creator',
        description: 'Import top Twitch clips for a specific creator channel',
        group: 'twitch',
        requiresBroadcasterId: true,
        hidesSortTimeframe: true,
    },
    twitch_trending: {
        label: 'Twitch Trending',
        description: "Trending clips from Twitch's current top categories",
        group: 'twitch',
        hidesSortTimeframe: true,
    },
    twitch_discovery: {
        label: 'Twitch Discovery',
        description: 'Clips from beyond Twitch’s main categories (rank 11–20)',
        group: 'twitch',
        hidesSortTimeframe: true,
    },
};

export const SCHEDULE_LABELS: Record<PlaylistScriptSchedule, string> = {
    manual: 'Manual only',
    hourly: 'Every hour',
    daily: 'Every day',
    weekly: 'Every week',
    monthly: 'Every month',
};

export const TITLE_PLACEHOLDERS = [
    { token: '{name}', label: 'Script name' },
    { token: '{date}', label: 'Current date (Jan 2, 2006)' },
    { token: '{day}', label: 'Day of week (Monday)' },
    { token: '{week_start}', label: 'Week start (Jan 2)' },
    { token: '{month}', label: 'Month (January 2006)' },
] as const;

/**
 * Client-side replica of the backend buildPlaylistTitle function.
 * Uses current date to preview what the generated title would look like.
 */
export function buildPlaylistTitle(
    script: Pick<PlaylistScript, 'name' | 'title_template'>,
): string {
    const now = new Date();

    if (script.title_template) {
        let title = script.title_template;
        title = title.replaceAll('{name}', script.name);
        title = title.replaceAll(
            '{date}',
            now.toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric',
            }),
        );
        title = title.replaceAll(
            '{day}',
            now.toLocaleDateString('en-US', { weekday: 'long' }),
        );

        // week_start = most recent Monday
        const dayOfWeek = now.getDay();
        const daysToMonday = (dayOfWeek + 6) % 7;
        const monday = new Date(now);
        monday.setDate(monday.getDate() - daysToMonday);
        title = title.replaceAll(
            '{week_start}',
            monday.toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
            }),
        );

        title = title.replaceAll(
            '{month}',
            now.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }),
        );
        return title;
    }

    return `${script.name} \u2022 ${now.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}`;
}

/**
 * Generate a human-readable summary of a playlist script configuration.
 */
export function buildScriptSummary(form: {
    name: string;
    strategy: PlaylistScriptStrategy;
    sort: string;
    timeframe: string;
    clip_limit: number;
    visibility: string;
    schedule: PlaylistScriptSchedule;
    retention_days: number;
}): string {
    const meta = STRATEGY_META[form.strategy];
    const strategyLabel = meta.label.toLowerCase();

    const parts: string[] = [];

    if (form.strategy === 'standard') {
        parts.push(
            `a ${form.visibility} playlist of the top ${form.clip_limit} clips`,
            `from the last ${form.timeframe}, sorted by ${form.sort}`,
        );
    } else {
        parts.push(
            `a ${form.visibility} playlist of ${form.clip_limit} clips`,
            `using the "${strategyLabel}" strategy`,
        );
    }

    if (form.schedule !== 'manual') {
        parts.push(`generated ${form.schedule}`);
    } else {
        parts.push('generated on demand');
    }

    parts.push(`with ${form.retention_days}-day retention`);

    return `This script will create ${parts.join(', ')}.`;
}
