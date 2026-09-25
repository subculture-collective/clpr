import { useQuery } from '@tanstack/react-query';
import { categoryApi } from '@/lib/category-api';
import { fetchPopularBroadcasters } from '@/lib/broadcaster-api';
import { useTags } from './useTags';

/**
 * Shared queries for navigation and sidebar rails. Every consumer uses the
 * same key and parameters, so one response serves the top nav, the feed
 * sidebar and the topics page instead of each fetching its own variant.
 */

/** Public topic list. Consumers that want featured topics filter locally. */
export const useTopicCategories = () =>
    useQuery({
        queryKey: ['categories', 'topic'],
        queryFn: () => categoryApi.listCategories({ type: 'topic', public: true }),
    });

/** Most popular tags. Consumers that show fewer slice the shared list. */
export const POPULAR_TAG_LIMIT = 20;
export const usePopularTags = () =>
    useTags({ sort: 'popularity', limit: POPULAR_TAG_LIMIT });

export const usePopularBroadcasters = (limit = 20) =>
    useQuery({
        queryKey: ['broadcasters', 'popular', limit],
        queryFn: () => fetchPopularBroadcasters(limit),
    });
