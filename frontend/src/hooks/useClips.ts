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

// Hook for infinite scrolling clip feed
export const useClipFeed = (filters?: ClipFeedFilters) => {
    return useInfiniteQuery({
        queryKey: ['clips', filters],
        // Use page-based pagination to align with tests and mocks
        queryFn: ({ pageParam = 1 }) =>
            clipApi.fetchClips({ pageParam, filters } as unknown as {
                pageParam: number;
                filters?: ClipFeedFilters;
            }),
        getNextPageParam: lastPage => {
            return lastPage.has_more ? (lastPage.page ?? 1) + 1 : undefined;
        },
        initialPageParam: 1,
    });
};

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
                pages: oldData.pages.map((page: ClipFeedResponse) => ({
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

// Hook for favoriting clips
export const useClipFavorite = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (payload: FavoritePayload) => {
            // Check cache to determine if we should add or remove
            // First, try to get the single-clip cache entry
            let isCurrentlyFavorited: boolean | undefined;
            const singleClip = queryClient.getQueryData<Clip>([
                'clip',
                payload.clip_id,
            ]);
            if (singleClip) {
                isCurrentlyFavorited = singleClip.is_favorited;
            } else {
                // Fallback: iterate cached clips in paginated queries
                const cachedClip = queryClient
                    .getQueriesData({ queryKey: ['clips'] })
                    .flatMap(([, data]) => {
                        const queryData = data as {
                            pages?: ClipFeedResponse[];
                        };
                        return (
                            queryData?.pages?.flatMap(page => page.clips) || []
                        );
                    })
                    .find(clip => clip.id === payload.clip_id);
                isCurrentlyFavorited = cachedClip?.is_favorited;
            }
            if (isCurrentlyFavorited) {
                return await clipApi.removeFavorite(payload);
            } else {
                return await clipApi.addFavorite(payload);
            }
        },
        onMutate: async payload => {
            // Optimistic update for clips queries
            await queryClient.cancelQueries({ queryKey: ['clips'] });
            await queryClient.cancelQueries({ queryKey: ['favorites'] });

            const previousClipsData = queryClient.getQueriesData({
                queryKey: ['clips'],
            });
            const previousFavoritesData = queryClient.getQueriesData({
                queryKey: ['favorites'],
            });

            // Update clips feed
            queryClient.setQueriesData(
                { queryKey: ['clips'] },
                (old: unknown) => {
                    const oldData = old as { pages?: ClipFeedResponse[] } | null | undefined;
                    if (!oldData?.pages) return oldData;

                    return {
                        ...oldData,
                        pages: oldData.pages.map((page: ClipFeedResponse) => ({
                            ...page,
                            clips: page.clips.map((clip: Clip) => {
                                if (clip.id === payload.clip_id) {
                                    const isFavorited = !clip.is_favorited;
                                    return {
                                        ...clip,
                                        is_favorited: isFavorited,
                                        favorite_count:
                                            clip.favorite_count +
                                            (isFavorited ? 1 : -1),
                                    };
                                }
                                return clip;
                            }),
                        })),
                    };
                },
            );

            // Update favorites feed (remove unfavorited clips)
            queryClient.setQueriesData(
                { queryKey: ['favorites'] },
                (old: unknown) => {
                    const oldData = old as { pages?: { clips: Clip[] }[] } | null | undefined;
                    if (!oldData?.pages) return oldData;

                    return {
                        ...oldData,
                        pages: oldData.pages.map((page: ClipFeedResponse) => ({
                            ...page,
                            clips: page.clips.filter(
                                (clip: Clip) => clip.id !== payload.clip_id,
                            ),
                        })),
                    };
                },
            );

            return { previousClipsData, previousFavoritesData };
        },
        onError: (_error, _payload, context) => {
            // Rollback on error
            if (context?.previousClipsData) {
                context.previousClipsData.forEach(([queryKey, data]) => {
                    queryClient.setQueryData(queryKey, data);
                });
            }
            if (context?.previousFavoritesData) {
                context.previousFavoritesData.forEach(([queryKey, data]) => {
                    queryClient.setQueryData(queryKey, data);
                });
            }
        },
        onSettled: () => {
            // Refetch to ensure consistency
            queryClient.invalidateQueries({ queryKey: ['favorites'] });
        },
    });
};
