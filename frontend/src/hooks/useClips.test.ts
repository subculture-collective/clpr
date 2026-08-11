import * as clipApi from '@/lib/clip-api';
import type {
    Clip,
    ClipFeedResponse,
    FavoritePayload,
    VotePayload,
} from '@/types/clip';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
    useClipById,
    useClipFavorite,
    useClipFeed,
    useClipVote,
} from './useClips';

// Mock the clip API
vi.mock('@/lib/clip-api');

const createWrapper = () => {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
            mutations: {
                retry: false,
            },
        },
    });

    return ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);
};

const mockClip: Clip = {
    id: 'clip-1',
    twitch_clip_id: 'TwitchClip1',
    twitch_clip_url: 'https://clips.twitch.tv/TwitchClip1',
    embed_url:
        'https://clips.twitch.tv/embed?clip=TwitchClip1&parent=localhost',
    title: 'Amazing Gaming Moment',
    creator_name: 'Creator1',
    creator_id: 'creator-1',
    broadcaster_name: 'Streamer1',
    broadcaster_id: 'streamer-1',
    game_id: 'game-1',
    game_name: 'Fortnite',
    language: 'en',
    thumbnail_url: 'https://example.com/thumbnail.jpg',
    duration: 45,
    view_count: 1000,
    created_at: '2024-01-01T00:00:00Z',
    imported_at: '2024-01-01T00:00:00Z',
    vote_score: 100,
    comment_count: 10,
    favorite_count: 50,
    is_featured: false,
    is_nsfw: false,
    is_removed: false,
    user_vote: null,
    is_favorited: false,
};

const mockClipFeedResponse: ClipFeedResponse = {
    clips: [mockClip],
    total: 1,
    page: 1,
    limit: 10,
    has_more: false,
};

describe('useClips', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe('useClipFeed hook', () => {
        it('should fetch clips successfully', async () => {
            vi.mocked(clipApi.fetchClips).mockResolvedValue(
                mockClipFeedResponse
            );

            const { result } = renderHook(() => useClipFeed(), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.data?.pages[0].clips).toEqual([mockClip]);
            expect(clipApi.fetchClips).toHaveBeenCalledWith({
                cursor: undefined,
                filters: undefined,
            });
        });

        it('should fetch clips with filters', async () => {
            vi.mocked(clipApi.fetchClips).mockResolvedValue(
                mockClipFeedResponse
            );

            const filters = { sort: 'hot' as const, game_id: 'game-1' };
            const { result } = renderHook(() => useClipFeed(filters), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(clipApi.fetchClips).toHaveBeenCalledWith({
                cursor: undefined,
                filters,
            });
        });

        it('should handle fetch error', async () => {
            const error = new Error('Failed to fetch clips');
            vi.mocked(clipApi.fetchClips).mockRejectedValue(error);

            const { result } = renderHook(() => useClipFeed(), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isError).toBe(true));

            expect(result.current.error).toEqual(error);
        });

        it('should handle loading state', () => {
            vi.mocked(clipApi.fetchClips).mockImplementation(
                () => new Promise(() => {}) // Never resolves
            );

            const { result } = renderHook(() => useClipFeed(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isLoading).toBe(true);
            expect(result.current.data).toBeUndefined();
        });

        it('should handle pagination with has_more true', async () => {
            const page1Response: ClipFeedResponse = {
                ...mockClipFeedResponse,
                has_more: true,
                cursor: 'feed-cursor',
            };

            vi.mocked(clipApi.fetchClips).mockResolvedValue(page1Response);

            const { result } = renderHook(() => useClipFeed(), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.hasNextPage).toBe(true);
        });

        it('should handle pagination with has_more false', async () => {
            vi.mocked(clipApi.fetchClips).mockResolvedValue(
                mockClipFeedResponse
            );

            const { result } = renderHook(() => useClipFeed(), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.hasNextPage).toBe(false);
        });

        it('should fetch next page when has_more is true', async () => {
            const page1Response: ClipFeedResponse = {
                clips: [mockClip],
                total: 20,
                page: 1,
                limit: 10,
                has_more: true,
                cursor: 'feed-cursor',
            };

            const page2Response: ClipFeedResponse = {
                clips: [{ ...mockClip, id: 'clip-2' }],
                total: 20,
                page: 2,
                limit: 10,
                has_more: false,
            };

            vi.mocked(clipApi.fetchClips)
                .mockResolvedValueOnce(page1Response)
                .mockResolvedValueOnce(page2Response);

            const { result } = renderHook(() => useClipFeed(), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.data?.pages).toHaveLength(1);
            expect(result.current.hasNextPage).toBe(true);

            result.current.fetchNextPage();

            await waitFor(() =>
                expect(result.current.data?.pages).toHaveLength(2)
            );

            expect(clipApi.fetchClips).toHaveBeenCalledTimes(2);
            expect(clipApi.fetchClips).toHaveBeenNthCalledWith(2, {
                cursor: 'feed-cursor',
                filters: undefined,
            });
        });
    });

    describe('useClipById hook', () => {
        it('should fetch clip by ID successfully', async () => {
            vi.mocked(clipApi.fetchClipById).mockResolvedValue(mockClip);

            const { result } = renderHook(() => useClipById('clip-1'), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.data).toEqual(mockClip);
            expect(clipApi.fetchClipById).toHaveBeenCalledWith('clip-1');
        });

        it('should handle fetch error', async () => {
            const error = new Error('Clip not found');
            vi.mocked(clipApi.fetchClipById).mockRejectedValue(error);

            const { result } = renderHook(() => useClipById('clip-1'), {
                wrapper: createWrapper(),
            });

            await waitFor(() => expect(result.current.isError).toBe(true));

            expect(result.current.error).toEqual(error);
        });

        it('should handle loading state', () => {
            vi.mocked(clipApi.fetchClipById).mockImplementation(
                () => new Promise(() => {})
            );

            const { result } = renderHook(() => useClipById('clip-1'), {
                wrapper: createWrapper(),
            });

            expect(result.current.isLoading).toBe(true);
            expect(result.current.data).toBeUndefined();
        });

        it('should not fetch when clipId is empty', () => {
            const { result } = renderHook(() => useClipById(''), {
                wrapper: createWrapper(),
            });

            expect(result.current.fetchStatus).toBe('idle');
            expect(clipApi.fetchClipById).not.toHaveBeenCalled();
        });
    });

    describe('useClipVote hook', () => {
        it('should vote on clip successfully', async () => {
            const voteResponse = {
                message: 'Vote recorded',
                vote_score: 101,
                upvote_count: 51,
                downvote_count: 0,
                user_vote: 1,
            };
            vi.mocked(clipApi.voteOnClip).mockResolvedValue(voteResponse);

            const { result } = renderHook(() => useClipVote(), {
                wrapper: createWrapper(),
            });

            const payload: VotePayload = {
                clip_id: 'clip-1',
                vote_type: 1,
            };

            result.current.mutate(payload);

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.data).toEqual(voteResponse);
            expect(clipApi.voteOnClip).toHaveBeenCalledWith(payload);
        });

        it('should handle vote error', async () => {
            const error = new Error('Failed to vote');
            vi.mocked(clipApi.voteOnClip).mockRejectedValue(error);

            const { result } = renderHook(() => useClipVote(), {
                wrapper: createWrapper(),
            });

            result.current.mutate({ clip_id: 'clip-1', vote_type: 1 });

            await waitFor(() => expect(result.current.isError).toBe(true));

            expect(result.current.error).toEqual(error);
        });

        it('should handle upvote (vote_type: 1)', async () => {
            const voteResponse = {
                message: 'Vote recorded',
                vote_score: 101,
                upvote_count: 51,
                downvote_count: 0,
                user_vote: 1,
            };
            vi.mocked(clipApi.voteOnClip).mockResolvedValue(voteResponse);

            const { result } = renderHook(() => useClipVote(), {
                wrapper: createWrapper(),
            });

            result.current.mutate({ clip_id: 'clip-1', vote_type: 1 });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(clipApi.voteOnClip).toHaveBeenCalledWith({
                clip_id: 'clip-1',
                vote_type: 1,
            });
        });

        it('should handle downvote (vote_type: -1)', async () => {
            const voteResponse = {
                message: 'Vote recorded',
                vote_score: 99,
                upvote_count: 50,
                downvote_count: 1,
                user_vote: -1,
            };
            vi.mocked(clipApi.voteOnClip).mockResolvedValue(voteResponse);

            const { result } = renderHook(() => useClipVote(), {
                wrapper: createWrapper(),
            });

            result.current.mutate({ clip_id: 'clip-1', vote_type: -1 });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(clipApi.voteOnClip).toHaveBeenCalledWith({
                clip_id: 'clip-1',
                vote_type: -1,
            });
        });
    });

    describe('useClipFavorite hook', () => {
        it('should add favorite successfully', async () => {
            const favoriteResponse = {
                message: 'Clip added to favorites',
                is_favorited: true,
            };
            vi.mocked(clipApi.addFavorite).mockResolvedValue(favoriteResponse);

            const { result } = renderHook(() => useClipFavorite(), {
                wrapper: createWrapper(),
            });

            const payload: FavoritePayload = {
                clip_id: 'clip-1',
            };

            result.current.mutate(payload);

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.data).toEqual(favoriteResponse);
            expect(clipApi.addFavorite).toHaveBeenCalledWith(payload);
        });

        it('should toggle favorite based on cache state', async () => {
            // Test adding a favorite when not currently favorited
            vi.clearAllMocks();

            const addResponse = {
                message: 'Clip added to favorites',
                is_favorited: true,
            };
            vi.mocked(clipApi.addFavorite).mockResolvedValue(addResponse);

            const { result: addResult } = renderHook(() => useClipFavorite(), {
                wrapper: createWrapper(),
            });

            addResult.current.mutate({ clip_id: 'clip-1' });

            await waitFor(() => expect(addResult.current.isSuccess).toBe(true));

            expect(addResult.current.data).toEqual(addResponse);
            expect(clipApi.addFavorite).toHaveBeenCalledWith({
                clip_id: 'clip-1',
            });
        });

        it('should handle favorite error', async () => {
            const error = new Error('Failed to favorite');
            vi.mocked(clipApi.addFavorite).mockRejectedValue(error);

            const { result } = renderHook(() => useClipFavorite(), {
                wrapper: createWrapper(),
            });

            result.current.mutate({ clip_id: 'clip-1' });

            await waitFor(() => expect(result.current.isError).toBe(true));

            expect(result.current.error).toEqual(error);
        });
    });
});
