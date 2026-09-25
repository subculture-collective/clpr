import { useEffect } from 'react';
import { isAxiosError } from 'axios';
import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import * as clipApi from '@/lib/clip-api';
import { isRetryableError } from '@/lib/error-utils';
import type { ClipFeedFilters } from '@/types/clip';

// Hook for infinite scrolling clip feed
export const useClipFeed = (filters?: ClipFeedFilters) => {
    const queryClient = useQueryClient();
    const query = useInfiniteQuery({
        queryKey: ['clips', filters],
        queryFn: ({ pageParam }) =>
            clipApi.fetchClips({ cursor: pageParam || undefined, filters }),
        getNextPageParam: lastPage => {
            return lastPage.has_more && lastPage.cursor ?
                    lastPage.cursor
                :   undefined;
        },
        initialPageParam: '',
        retry: (attempt, error) =>
            isRetryableError(error) &&
            queryClient.getDefaultOptions().queries?.retry !== false && attempt < 3,
    });
    useEffect(() => {
        if (isAxiosError(query.error) &&
            query.error.response?.data?.code === 'RANKING_REFRESH_REQUIRED') {
            void queryClient.resetQueries({ queryKey: ['clips', filters], exact: true });
        }
    }, [query.error, queryClient, filters]);
    return query;
};
