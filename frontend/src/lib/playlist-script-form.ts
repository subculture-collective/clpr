import type { PlaylistScriptFormValues } from '@/components/admin/PlaylistScriptForm';
import type { PlaylistScript } from '@/types/playlistScript';

export function scriptToFormValues(
    script: PlaylistScript,
): PlaylistScriptFormValues {
    return {
        name: script.name,
        description: script.description || '',
        visibility: script.visibility,
        is_active: script.is_active,
        strategy: script.strategy || 'standard',
        sort: script.sort,
        timeframe: script.timeframe || 'day',
        clip_limit: script.clip_limit,
        game_id: script.game_id || '',
        game_ids: script.game_ids || [],
        broadcaster_id: script.broadcaster_id || '',
        tag: script.tag || '',
        exclude_tags: script.exclude_tags || [],
        language: script.language || '',
        min_vote_score:
            script.min_vote_score != null ? String(script.min_vote_score) : '',
        min_view_count:
            script.min_view_count != null ? String(script.min_view_count) : '',
        exclude_nsfw: script.exclude_nsfw,
        top_10k_streamers: script.top_10k_streamers,
        seed_clip_id: script.seed_clip_id || '',
        schedule: script.schedule || 'manual',
        retention_days: script.retention_days || 30,
        title_template: script.title_template || '',
    };
}
