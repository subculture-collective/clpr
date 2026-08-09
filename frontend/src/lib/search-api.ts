import apiClient from './api';
import type {
    SearchRequest,
    SearchResponse,
    SearchSuggestion,
    TrendingSearch,
    SearchHistoryItem,
    FailedSearch,
    SearchAnalyticsSummary,
    SavedSearch,
    SearchFilters,
} from '../types/search';

type LegacySearchResponse = Omit<SearchResponse, 'results'> & {
    results?: {
        [Key in keyof SearchResponse['results']]?:
            | SearchResponse['results'][Key]
            | null;
    } | null;
};

export function normalizeSearchResponse(
    response: LegacySearchResponse,
): SearchResponse {
    return {
        ...response,
        results: {
            clips: response.results?.clips ?? [],
            creators: response.results?.creators ?? [],
            games: response.results?.games ?? [],
            tags: response.results?.tags ?? [],
        },
    };
}

export const searchApi = {
    // Universal search
    async search(params: SearchRequest): Promise<SearchResponse> {
        const response = await apiClient.get<LegacySearchResponse>('/search', {
            params: {
                q: params.query,
                type: params.type,
                sort: params.sort,
                game_id: params.gameId,
                creator_id: params.creatorId,
                language: params.language,
                tags: params.tags,
                min_votes: params.minVotes,
                date_from: params.dateFrom,
                date_to: params.dateTo,
                page: params.page || 1,
                limit: params.limit || 20,
            },
        });
        return normalizeSearchResponse(response.data);
    },

    // Get autocomplete suggestions
    async getSuggestions(query: string): Promise<SearchSuggestion[]> {
        const response = await apiClient.get<{
            query: string;
            suggestions: SearchSuggestion[];
        }>('/search/suggestions', {
            params: { q: query },
        });
        return response.data.suggestions;
    },

    // Get trending searches
    async getTrendingSearches(
        days: number = 7,
        limit: number = 20,
    ): Promise<TrendingSearch[]> {
        const response = await apiClient.get<{
            trending_searches: TrendingSearch[];
            days: number;
            limit: number;
        }>('/search/trending', {
            params: { days, limit },
        });
        return response.data.trending_searches;
    },

    // Get user's search history (requires authentication)
    async getSearchHistory(limit: number = 20): Promise<SearchHistoryItem[]> {
        const response = await apiClient.get<{
            search_history: SearchHistoryItem[];
            limit: number;
        }>('/search/history', {
            params: { limit },
        });
        return response.data.search_history || [];
    },

    // Get failed searches (admin only)
    async getFailedSearches(
        days: number = 7,
        limit: number = 20,
    ): Promise<FailedSearch[]> {
        const response = await apiClient.get<{
            failed_searches: FailedSearch[];
            days: number;
            limit: number;
        }>('/search/failed', {
            params: { days, limit },
        });
        return response.data.failed_searches;
    },

    // Get search analytics summary (admin only)
    async getSearchAnalytics(
        days: number = 7,
    ): Promise<SearchAnalyticsSummary> {
        const response = await apiClient.get<{
            summary: SearchAnalyticsSummary;
            days: number;
        }>('/search/analytics', {
            params: { days },
        });
        return response.data.summary;
    },

    // Saved searches - local storage based for now
    // TODO: Move to backend API when user preferences endpoint is available
    getSavedSearches(): SavedSearch[] {
        const saved = localStorage.getItem('savedSearches');
        if (!saved) return [];
        try {
            return JSON.parse(saved);
        } catch {
            return [];
        }
    },

    saveSearch(
        query: string,
        filters?: SearchFilters,
        name?: string,
    ): SavedSearch {
        const searches = this.getSavedSearches();
        const newSearch: SavedSearch = {
            id: crypto.randomUUID(),
            query,
            filters,
            name,
            created_at: new Date().toISOString(),
        };
        searches.unshift(newSearch);
        // Keep only last 50
        const limited = searches.slice(0, 50);
        localStorage.setItem('savedSearches', JSON.stringify(limited));
        return newSearch;
    },

    deleteSavedSearch(id: string): void {
        const searches = this.getSavedSearches();
        const filtered = searches.filter(s => s.id !== id);
        localStorage.setItem('savedSearches', JSON.stringify(filtered));
    },

    clearSavedSearches(): void {
        localStorage.removeItem('savedSearches');
    },
};
