import { apiClient } from './api';
import type { ClipTopicsResponse, TopicMutationResponse } from '../types/topic';

export const topicApi = {
    getClipTopics: async (clipId: string) => {
        const response = await apiClient.get<ClipTopicsResponse>(`/clips/${clipId}/topics`);
        return response.data;
    },

    replaceClipTopics: async (clipId: string, topicIds: string[]) => {
        const response = await apiClient.put<ClipTopicsResponse>(
            `/admin/clips/${clipId}/topics`,
            { topic_ids: topicIds },
        );
        return response.data;
    },

    mergeTopics: async (sourceTopicId: string, targetTopicId: string) => {
        await apiClient.post(`/admin/topics/${sourceTopicId}/merge`, {
            target_topic_id: targetTopicId,
        });
    },

    splitTopic: async (
        sourceTopicId: string,
        payload: { name: string; slug: string; description?: string; clip_ids: string[] },
    ) => {
        const response = await apiClient.post<TopicMutationResponse>(
            `/admin/topics/${sourceTopicId}/split`,
            payload,
        );
        return response.data;
    },
};
