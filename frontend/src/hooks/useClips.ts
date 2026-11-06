import * as clipApi from '@/lib/clip-api';
import type {
    Clip,
    ClipFeedFilters,
    FavoritePayload,
    VotePayload,
} from '@/types/clip';
import {
    useInfiniteQuery,
    useMutation,
    useQuery,
    useQueryClient,
    type Query,
    type QueryClient,
    type QueryKey,
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

// Clips appear in many query shapes: feeds ({ pages: [{ clips }] }),
// paginated lists ({ clips } or { data }), search results, playlists, queues
// and single-clip queries. Vote and favorite updates walk every cached query
// and patch each clip object with the matching id, so the change shows
// wherever the clip is rendered.
type ClipCacheSnapshot = Array<[QueryKey, unknown]>;

const isClipRecord = (value: Record<string, unknown>, clipId: string) =>
    value.id === clipId && typeof value.vote_score === 'number';

const patchClipInData = (
    data: unknown,
    clipId: string,
    updater: (clip: Clip) => Clip,
): unknown => {
    if (Array.isArray(data)) {
        let changed = false;
        const next = data.map(item => {
            const patched = patchClipInData(item, clipId, updater);
            if (patched !== item) changed = true;
            return patched;
        });
        return changed ? next : data;
    }
    if (typeof data !== 'object' || data === null) return data;

    const record = data as Record<string, unknown>;
    if (isClipRecord(record, clipId)) return updater(record as unknown as Clip);

    let changed = false;
    const next: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(record)) {
        const patched = patchClipInData(value, clipId, updater);
        if (patched !== value) changed = true;
        next[key] = patched;
    }
    return changed ? next : data;
};

const containsClip = (data: unknown, clipId: string): boolean => {
    if (Array.isArray(data)) return data.some(item => containsClip(item, clipId));
    if (typeof data !== 'object' || data === null) return false;
    const record = data as Record<string, unknown>;
    return (
        isClipRecord(record, clipId) ||
        Object.values(record).some(value => containsClip(value, clipId))
    );
};

const queriesHoldingClip = (clipId: string) => ({
    predicate: (query: Query) => containsClip(query.state.data, clipId),
});

const snapshotClipQueries = async (
    queryClient: QueryClient,
    clipId: string,
): Promise<ClipCacheSnapshot> => {
    await queryClient.cancelQueries(queriesHoldingClip(clipId));
    return queryClient.getQueriesData(queriesHoldingClip(clipId));
};

const updateCachedClip = (
    queryClient: QueryClient,
    clipId: string,
    updater: (clip: Clip) => Clip,
) => {
    queryClient.setQueriesData(queriesHoldingClip(clipId), (old: unknown) =>
        patchClipInData(old, clipId, updater),
    );
};

const restoreClipQueries = (
    queryClient: QueryClient,
    snapshot: ClipCacheSnapshot | undefined,
) => {
    snapshot?.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data);
    });
};

// Hook for voting on clips
export const useClipVote = () => {
    const queryClient = useQueryClient();

    type VoteContext = { previousQueries: ClipCacheSnapshot };

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
            // Optimistic update for every cached query holding the clip
            const previousQueries = await snapshotClipQueries(
                queryClient,
                payload.clip_id,
            );

            updateCachedClip(queryClient, payload.clip_id, clip => {
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

            return { previousQueries };
        },
        onError: (_error, _payload, context) => {
            restoreClipQueries(queryClient, context?.previousQueries);
        },
        onSuccess: (data, payload) => {
            const normalizedUserVote =
                data.user_vote === 0 ? null : data.user_vote;

            updateCachedClip(queryClient, payload.clip_id, clip => ({
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

    type FavoriteContext = { previousQueries: ClipCacheSnapshot };

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

            await queryClient.cancelQueries({ queryKey: ['favorites'] });
            const previousQueries = await snapshotClipQueries(
                queryClient,
                clip_id,
            );

            updateCachedClip(queryClient, clip_id, clip =>
                applyFavoriteState(clip, nextFavorited),
            );

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

            return { previousQueries };
        },
        onError: (_error, _variables, context) => {
            restoreClipQueries(queryClient, context?.previousQueries);
        },
        onSettled: (_data, _error, { clip_id }) => {
            // Refetch to ensure consistency
            queryClient.invalidateQueries({ queryKey: ['favorites'] });
            queryClient.invalidateQueries({ queryKey: ['clip', clip_id] });
        },
    });
};
