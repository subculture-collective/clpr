import { apiClient } from './api';
import type {
    CategoryListResponse,
    CategoryDetailResponse,
} from '../types/category';
import type { GameListResponse } from '../types/game';
import type { ClipFeedResponse } from '../types/clip';

export const categoryApi = {
    // List categories with optional filters
    listCategories: async (params?: { featured?: boolean; public?: boolean; type?: string }) => {
        const response = await apiClient.get<CategoryListResponse>(
            '/categories',
            {
                params,
            },
        );
        return response.data;
    },

    // Get category details
    getCategory: async (slug: string) => {
        const response = await apiClient.get<CategoryDetailResponse>(
            `/categories/${slug}`,
        );
        return response.data;
    },

    // Get games in category
    getCategoryGames: async (
        slug: string,
        params?: { limit?: number; page?: number },
    ) => {
        const response = await apiClient.get<GameListResponse>(
            `/categories/${slug}/games`,
            { params },
        );
        return response.data;
    },

    // Get clips in category
    getCategoryClips: async (
        slug: string,
        params?: {
            limit?: number;
            page?: number;
            sort?: 'hot' | 'new' | 'top' | 'rising';
            timeframe?: 'hour' | 'day' | 'week' | 'month' | 'year' | 'all';
        },
    ) => {
        const response = await apiClient.get<ClipFeedResponse>(
            `/categories/${slug}/clips`,
            { params },
        );
        return response.data;
    },
};
