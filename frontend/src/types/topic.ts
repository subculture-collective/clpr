import type { Category } from './category';

export interface ClipTopic {
    clip_id: string;
    topic_id: string;
    topic_name: string;
    topic_slug: string;
    source: 'twitch_category' | 'metadata' | 'tag' | 'transcript' | 'manual' | 'backfill';
    confidence: number;
    evidence: Record<string, unknown>;
    assigned_by_user_id?: string;
    created_at: string;
    updated_at: string;
}

export interface ClipTopicsResponse {
    topics: ClipTopic[];
}

export interface TopicMutationResponse {
    topic: Category;
}
