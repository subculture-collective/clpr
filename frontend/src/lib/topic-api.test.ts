import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './api';
import { topicApi } from './topic-api';

vi.mock('./api', () => ({
    apiClient: {
        get: vi.fn(),
        put: vi.fn(),
        post: vi.fn(),
    },
}));

describe('topicApi', () => {
    beforeEach(() => vi.clearAllMocks());

    it('loads public clip topics', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { topics: [] } });
        await expect(topicApi.getClipTopics('clip-1')).resolves.toEqual({ topics: [] });
        expect(apiClient.get).toHaveBeenCalledWith('/clips/clip-1/topics');
    });

    it('sends explicit moderator corrections', async () => {
        vi.mocked(apiClient.put).mockResolvedValue({ data: { topics: [] } });
        await topicApi.replaceClipTopics('clip-1', ['topic-1']);
        expect(apiClient.put).toHaveBeenCalledWith('/admin/clips/clip-1/topics', {
            topic_ids: ['topic-1'],
        });
    });

    it('supports merge and split moderation operations', async () => {
        vi.mocked(apiClient.post)
            .mockResolvedValueOnce({ data: { success: true } })
            .mockResolvedValueOnce({ data: { topic: { id: 'new-topic' } } });

        await topicApi.mergeTopics('source', 'target');
        await topicApi.splitTopic('source', {
            name: 'New topic',
            slug: 'new-topic',
            clip_ids: ['clip-1'],
        });

        expect(apiClient.post).toHaveBeenNthCalledWith(1, '/admin/topics/source/merge', {
            target_topic_id: 'target',
        });
        expect(apiClient.post).toHaveBeenNthCalledWith(2, '/admin/topics/source/split', {
            name: 'New topic',
            slug: 'new-topic',
            clip_ids: ['clip-1'],
        });
    });
});
