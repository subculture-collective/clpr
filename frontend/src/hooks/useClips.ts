import * as clipApi from '@/lib/clip-api';
import type {
    Clip,
    ClipFeedFilters,
    ClipFeedResponse,
    FavoritePayload,
    VotePayload,
} from '@/types/clip';
import {
    useInfiniteQuery,
    useMutation,
    useQuery,
    useQueryClient,
} from '@tanstack/react-query';

export { useClipFeed } from './useClipFeed';

// Hook for infinite scrolling scraped clips feed
export const useScrapedClipsFeed = (filters?: ClipFeedFilters) => {
    return useInfiniteQuery({
        queryKey: ['scraped-clips', filters],
        queryFn: ({ pageParam = 1 }) =>
            clipApi.fetchScrapedClips({ pageParam, filters }),
        getNextPageParam: lastPage => {
            return lastPage.has_more ? lastPage.page + 1 : undefined;
        },
        initialPageParam: 1,
        refetchInterval: 30_000,
        refetchIntervalInBackground: false,
    });
};

// Hook for infinite scrolling favorites feed
export const useFavoritesFeed = (
    sort: 'newest' | 'top' | 'discussed' = 'newest',
) => {
    return useInfiniteQuery({
        queryKey: ['favorites', sort],
        queryFn: ({ pageParam = 1 }) =>
            clipApi.fetchFavorites({ pageParam, sort }),
        getNextPageParam: lastPage => {
            return lastPage.has_more ? lastPage.page + 1 : undefined;
        },
        initialPageParam: 1,
    });
};

// Hook to fetch a single clip by ID
export const useClipById = (clipId: string) => {
    return useQuery({
        queryKey: ['clip', clipId],
        queryFn: () => clipApi.fetchClipById(clipId),
        enabled: !!clipId,
    });
};

// Hook for voting on clips
export const useClipVote = () => {
    const queryClient = useQueryClient();

    type VoteContext = {
        previousFeedQueries: ReturnType<typeof queryClient.getQueriesData>;
        previousClip?: Clip;
    };

    const updateClipCaches = (
        clipId: string,
        updater: (clip: Clip) => Clip,
    ) => {
        queryClient.setQueriesData({ queryKey: ['clips'] }, (old: unknown) => {
            const oldData = old as { pages?: ClipFeedResponse[] } | null | undefined;
            if (!oldData?.pages) return oldData;

            return {
                ...oldData,
                pages: oldData.pages.map(page => ({
                    ...page,
                    clips: page.clips.map((clip: Clip) =>
                        clip.id === clipId ? updater(clip) : clip,
                    ),
                })),
            };
        });

        const singleClip = queryClient.getQueryData<Clip>(['clip', clipId]);
        if (singleClip) {
            queryClient.setQueryData<Clip>(
                ['clip', clipId],
                updater(singleClip),
            );
        }
    };

    return useMutation<
        Awaited<ReturnType<typeof clipApi.voteOnClip>>,
        Error,
        VotePayload,
        VoteContext
    >({
        mutationFn: async (payload: VotePayload) => {
            return await clipApi.voteOnClip(payload);
        },
        onMutate: async payload => {
            // Optimistic update for feed and single clip caches
            await queryClient.cancelQueries({ queryKey: ['clips'] });
            await queryClient.cancelQueries({
                queryKey: ['clip', payload.clip_id],
            });

            const previousFeedQueries = queryClient.getQueriesData({
                queryKey: ['clips'],
            });
            const previousClip = queryClient.getQueryData<Clip>([
                'clip',
                payload.clip_id,
            ]);

            updateClipCaches(payload.clip_id, clip => {
                const previousVote = clip.user_vote ?? 0;
                const scoreDelta = payload.vote_type - previousVote;

                let nextUpvoteCount = clip.upvote_count;
                let nextDownvoteCount = clip.downvote_count;

                if (nextUpvoteCount !== undefined) {
                    if (payload.vote_type === 1 && previousVote !== 1) {
                        nextUpvoteCount = nextUpvoteCount + 1;
                    }
                    if (payload.vote_type === -1 && previousVote === 1) {
                        nextUpvoteCount = Math.max(0, nextUpvoteCount - 1);
                    }
                }

                if (nextDownvoteCount !== undefined) {
                    if (payload.vote_type === -1 && previousVote !== -1) {
                        nextDownvoteCount = nextDownvoteCount + 1;
                    }
                    if (payload.vote_type === 1 && previousVote === -1) {
                        nextDownvoteCount = Math.max(0, nextDownvoteCount - 1);
                    }
                }

                return {
                    ...clip,
                    user_vote: payload.vote_type,
                    vote_score: clip.vote_score + scoreDelta,
                    upvote_count: nextUpvoteCount,
                    downvote_count: nextDownvoteCount,
                };
            });

            return { previousFeedQueries, previousClip };
        },
        onError: (_error, payload, context) => {
            // Rollback on error
            context?.previousFeedQueries?.forEach(([queryKey, data]) => {
                queryClient.setQueryData(queryKey, data);
            });
            if (context?.previousClip) {
                queryClient.setQueryData(
                    ['clip', payload.clip_id],
                    context.previousClip,
                );
            }
        },
        onSuccess: (data, payload) => {
            const normalizedUserVote =
                data.user_vote === 0 ? null : data.user_vote;

            updateClipCaches(payload.clip_id, clip => ({
                ...clip,
                user_vote: normalizedUserVote as Clip['user_vote'],
                vote_score: data.vote_score,
                upvote_count: data.upvote_count,
                downvote_count: data.downvote_count,
            }));
        },
    });
};

// Variables for useClipFavorite. `is_favorited` is the clip's state as shown
// to the user when they clicked, which decides between add and remove. The
// optimistic update below changes the cache before mutationFn runs, so the
// cache cannot be used to make that decision.
export type FavoriteToggleVariables = FavoritePayload & {
    is_favorited: boolean;
};

// Hook for favoriting clips
export const useClipFavorite = () => {
    const queryClient = useQueryClient();

    type FavoriteContext = {
        previousClipsData: ReturnType<typeof queryClient.getQueriesData>;
        previousFavoritesData: ReturnType<typeof queryClient.getQueriesData>;
        previousClip?: Clip;
    };

    const applyFavoriteState = (clip: Clip, isFavorited: boolean): Clip => {
        if (Boolean(clip.is_favorited) === isFavorited) return clip;
        return {
            ...clip,
            is_favorited: isFavorited,
            favorite_count: Math.max(
                0,
                (clip.favorite_count ?? 0) + (isFavorited ? 1 : -1),
            ),
        };
    };

    return useMutation<
        Awaited<ReturnType<typeof clipApi.addFavorite>>,
        Error,
        FavoriteToggleVariables,
        FavoriteContext
    >({
        mutationFn: async ({ clip_id, is_favorited }) => {
            const payload: FavoritePayload = { clip_id };
            return is_favorited ?
                    await clipApi.removeFavorite(payload)
                :   await clipApi.addFavorite(payload);
        },
        onMutate: async ({ clip_id, is_favorited }) => {
            const nextFavorited = !is_favorited;

            await queryClient.cancelQueries({ queryKey: ['clips'] });
            await queryClient.cancelQueries({ queryKey: ['favorites'] });
            await queryClient.cancelQueries({ queryKey: ['clip', clip_id] });

            const previousClipsData = queryClient.getQueriesData({
                queryKey: ['clips'],
            });
            const previousFavoritesData = queryClient.getQueriesData({
                queryKey: ['favorites'],
            });
            const previousClip = queryClient.getQueryData<Clip>([
                'clip',
                clip_id,
            ]);

            // Update clips feeds
            queryClient.setQueriesData(
                { queryKey: ['clips'] },
                (old: unknown) => {
                    const oldData = old as { pages?: ClipFeedResponse[] } | null | undefined;
                    if (!oldData?.pages) return oldData;

                    return {
                        ...oldData,
                        pages: oldData.pages.map(page => ({
                            ...page,
                            clips: page.clips.map((clip: Clip) =>
                                clip.id === clip_id ?
                                    applyFavoriteState(clip, nextFavorited)
                                :   clip,
                            ),
                        })),
                    };
                },
            );

            // Update the clip detail cache
            if (previousClip) {
                queryClient.setQueryData<Clip>(
                    ['clip', clip_id],
                    applyFavoriteState(previousClip, nextFavorited),
                );
            }

            // Remove unfavorited clips from the favorites feed
            if (!nextFavorited) {
                queryClient.setQueriesData(
                    { queryKey: ['favorites'] },
                    (old: unknown) => {
                        const oldData = old as { pages?: { clips: Clip[] }[] } | null | undefined;
                        if (!oldData?.pages) return oldData;

                        return {
                            ...oldData,
                            pages: oldData.pages.map(page => ({
                                ...page,
                                clips: page.clips.filter(
                                    (clip: Clip) => clip.id !== clip_id,
                                ),
                            })),
                        };
                    },
                );
            }

            return { previousClipsData, previousFavoritesData, previousClip };
        },
        onError: (_error, { clip_id }, context) => {
            // Rollback on error
            context?.previousClipsData.forEach(([queryKey, data]) => {
                queryClient.setQueryData(queryKey, data);
            });
            context?.previousFavoritesData.forEach(([queryKey, data]) => {
                queryClient.setQueryData(queryKey, data);
            });
            if (context?.previousClip) {
                queryClient.setQueryData(['clip', clip_id], context.previousClip);
            }
        },
        onSettled: (_data, _error, { clip_id }) => {
            // Refetch to ensure consistency
            queryClient.invalidateQueries({ queryKey: ['favorites'] });
            queryClient.invalidateQueries({ queryKey: ['clip', clip_id] });
        },
    });
};
