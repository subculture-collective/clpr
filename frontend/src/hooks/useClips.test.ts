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
        const createClientAndWrapper = () => {
            const queryClient = new QueryClient({
                defaultOptions: {
                    queries: { retry: false },
                    mutations: { retry: false },
                },
            });
            const wrapper = ({ children }: { children: ReactNode }) =>
                createElement(
                    QueryClientProvider,
                    { client: queryClient },
                    children,
                );
            return { queryClient, wrapper };
        };

        const feedPage = (clips: Clip[]): ClipFeedResponse => ({
            clips,
            total: clips.length,
            page: 1,
            limit: 10,
            has_more: false,
        });

        it('should add favorite successfully', async () => {
            const favoriteResponse = {
                message: 'Clip added to favorites',
                is_favorited: true,
            };
            vi.mocked(clipApi.addFavorite).mockResolvedValue(favoriteResponse);

            const { result } = renderHook(() => useClipFavorite(), {
                wrapper: createWrapper(),
            });

            result.current.mutate({ clip_id: 'clip-1', is_favorited: false });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(result.current.data).toEqual(favoriteResponse);
            const expectedPayload: FavoritePayload = { clip_id: 'clip-1' };
            expect(clipApi.addFavorite).toHaveBeenCalledWith(expectedPayload);
            expect(clipApi.removeFavorite).not.toHaveBeenCalled();
        });

        it('adds a favorite for an unfavorited clip that is only in a feed cache', async () => {
            // Regression: the optimistic update used to flip the feed cache
            // before mutationFn read it, so this sent DELETE instead of POST.
            vi.mocked(clipApi.addFavorite).mockResolvedValue({
                message: 'Clip added to favorites',
                is_favorited: true,
            });
            const { queryClient, wrapper } = createClientAndWrapper();
            const clip = { ...mockClip, is_favorited: false, favorite_count: 5 };
            queryClient.setQueryData(['clips', 'feed'], {
                pages: [feedPage([clip])],
                pageParams: [1],
            });

            const { result } = renderHook(() => useClipFavorite(), { wrapper });
            result.current.mutate({ clip_id: clip.id, is_favorited: false });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(clipApi.addFavorite).toHaveBeenCalledWith({ clip_id: clip.id });
            expect(clipApi.removeFavorite).not.toHaveBeenCalled();
            const feed = queryClient.getQueryData<{ pages: ClipFeedResponse[] }>(
                ['clips', 'feed'],
            );
            expect(feed?.pages[0].clips[0]).toMatchObject({
                is_favorited: true,
                favorite_count: 6,
            });
        });

        it('removes a favorite for a favorited clip that is only in a feed cache', async () => {
            vi.mocked(clipApi.removeFavorite).mockResolvedValue({
                message: 'Clip removed from favorites',
                is_favorited: false,
            });
            const { queryClient, wrapper } = createClientAndWrapper();
            const clip = { ...mockClip, is_favorited: true, favorite_count: 5 };
            queryClient.setQueryData(['clips', 'feed'], {
                pages: [feedPage([clip])],
                pageParams: [1],
            });

            const { result } = renderHook(() => useClipFavorite(), { wrapper });
            result.current.mutate({ clip_id: clip.id, is_favorited: true });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            expect(clipApi.removeFavorite).toHaveBeenCalledWith({ clip_id: clip.id });
            expect(clipApi.addFavorite).not.toHaveBeenCalled();
        });

        it('updates the clip detail cache optimistically', async () => {
            let resolveAdd: (value: { message: string; is_favorited: boolean }) => void = () => {};
            vi.mocked(clipApi.addFavorite).mockImplementation(
                () => new Promise(resolve => { resolveAdd = resolve; }),
            );
            const { queryClient, wrapper } = createClientAndWrapper();
            const clip = { ...mockClip, is_favorited: false, favorite_count: 5 };
            queryClient.setQueryData(['clip', clip.id], clip);

            const { result } = renderHook(() => useClipFavorite(), { wrapper });
            result.current.mutate({ clip_id: clip.id, is_favorited: false });

            await waitFor(() =>
                expect(
                    queryClient.getQueryData<Clip>(['clip', clip.id]),
                ).toMatchObject({ is_favorited: true, favorite_count: 6 }),
            );

            resolveAdd({ message: 'ok', is_favorited: true });
            await waitFor(() => expect(result.current.isSuccess).toBe(true));
            expect(
                queryClient.getQueryState(['clip', clip.id])?.isInvalidated,
            ).toBe(true);
        });

        it('rolls back the clip detail cache on error', async () => {
            vi.mocked(clipApi.addFavorite).mockRejectedValue(new Error('nope'));
            const { queryClient, wrapper } = createClientAndWrapper();
            const clip = { ...mockClip, is_favorited: false, favorite_count: 5 };
            queryClient.setQueryData(['clip', clip.id], clip);

            const { result } = renderHook(() => useClipFavorite(), { wrapper });
            result.current.mutate({ clip_id: clip.id, is_favorited: false });

            await waitFor(() => expect(result.current.isError).toBe(true));
            expect(
                queryClient.getQueryData<Clip>(['clip', clip.id]),
            ).toMatchObject({ is_favorited: false, favorite_count: 5 });
        });

        it('keeps favorites feed entries when adding a favorite', async () => {
            vi.mocked(clipApi.addFavorite).mockResolvedValue({
                message: 'ok',
                is_favorited: true,
            });
            const { queryClient, wrapper } = createClientAndWrapper();
            const other = { ...mockClip, id: 'clip-2', is_favorited: true };
            queryClient.setQueryData(['favorites', 'newest'], {
                pages: [feedPage([other])],
                pageParams: [1],
            });

            const { result } = renderHook(() => useClipFavorite(), { wrapper });
            result.current.mutate({ clip_id: other.id, is_favorited: false });

            await waitFor(() => expect(result.current.isSuccess).toBe(true));
            const favorites = queryClient.getQueryData<{ pages: ClipFeedResponse[] }>(
                ['favorites', 'newest'],
            );
            expect(favorites?.pages[0].clips).toHaveLength(1);
        });

        it('should handle favorite error', async () => {
            const error = new Error('Failed to favorite');
            vi.mocked(clipApi.addFavorite).mockRejectedValue(error);

            const { result } = renderHook(() => useClipFavorite(), {
                wrapper: createWrapper(),
            });

            result.current.mutate({ clip_id: 'clip-1', is_favorited: false });

            await waitFor(() => expect(result.current.isError).toBe(true));

            expect(result.current.error).toEqual(error);
        });
    });
    describe('clip caches outside the main feed', () => {
        const createClient = () => {
            const queryClient = new QueryClient({
                defaultOptions: {
                    queries: { retry: false },
                    mutations: { retry: false },
                },
            });
            const wrapper = ({ children }: { children: ReactNode }) =>
                createElement(QueryClientProvider, { client: queryClient }, children);
            return { queryClient, wrapper };
        };

        const seedOtherCaches = (queryClient: QueryClient, clip: Clip) => {
            queryClient.setQueryData(['search', 'q', 'clips'], { results: { clips: [clip] }, counts: { clips: 1 } });
            queryClient.setQueryData(['user-clips', 'u1', 1], { clips: [clip], total: 1 });
            queryClient.setQueryData(['broadcasterClips', 'b1', 1, 'recent'], { data: [clip] });
            queryClient.setQueryData(['playlist', 'p1'], { items: [{ id: 'item-1', clip }] });
        };

        const readOtherCaches = (queryClient: QueryClient) => [
            queryClient.getQueryData<{ results: { clips: Clip[] } }>(['search', 'q', 'clips'])?.results.clips[0],
            queryClient.getQueryData<{ clips: Clip[] }>(['user-clips', 'u1', 1])?.clips[0],
            queryClient.getQueryData<{ data: Clip[] }>(['broadcasterClips', 'b1', 1, 'recent'])?.data[0],
            queryClient.getQueryData<{ items: { clip: Clip }[] }>(['playlist', 'p1'])?.items[0].clip,
        ];

        it('applies a vote to every cached query holding the clip', async () => {
            vi.mocked(clipApi.voteOnClip).mockResolvedValue({
                message: 'ok', vote_score: 101, upvote_count: 51, downvote_count: 0, user_vote: 1,
            });
            const { queryClient, wrapper } = createClient();
            seedOtherCaches(queryClient, { ...mockClip });

            const { result } = renderHook(() => useClipVote(), { wrapper });
            result.current.mutate({ clip_id: mockClip.id, vote_type: 1 });
            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            for (const clip of readOtherCaches(queryClient)) {
                expect(clip).toMatchObject({ user_vote: 1, vote_score: 101 });
            }
        });

        it('rolls back the vote in every cache on error', async () => {
            vi.mocked(clipApi.voteOnClip).mockRejectedValue(new Error('nope'));
            const { queryClient, wrapper } = createClient();
            seedOtherCaches(queryClient, { ...mockClip });

            const { result } = renderHook(() => useClipVote(), { wrapper });
            result.current.mutate({ clip_id: mockClip.id, vote_type: 1 });
            await waitFor(() => expect(result.current.isError).toBe(true));

            for (const clip of readOtherCaches(queryClient)) {
                expect(clip).toMatchObject({ user_vote: null, vote_score: 100 });
            }
        });

        it('marks the clip favorited everywhere so a second click removes it', async () => {
            vi.mocked(clipApi.addFavorite).mockResolvedValue({ message: 'ok', is_favorited: true });
            const { queryClient, wrapper } = createClient();
            seedOtherCaches(queryClient, { ...mockClip, is_favorited: false, favorite_count: 5 });

            const { result } = renderHook(() => useClipFavorite(), { wrapper });
            result.current.mutate({ clip_id: mockClip.id, is_favorited: false });
            await waitFor(() => expect(result.current.isSuccess).toBe(true));

            for (const clip of readOtherCaches(queryClient)) {
                expect(clip).toMatchObject({ is_favorited: true, favorite_count: 6 });
            }
        });
    });
});
