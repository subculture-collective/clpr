import type { Clip } from './clip';
import type { User } from '../lib/auth-api';
import type { Tag } from './tag';

export interface SearchRequest {
  query: string;
  type?: 'clips' | 'creators' | 'games' | 'tags' | 'all';
  sort?: 'relevance' | 'recent' | 'popular';
  gameId?: string;
  creatorId?: string;
  language?: string;
  tags?: string[];
  minVotes?: number;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
  limit?: number;
}

export interface SearchResponse {
  query: string;
  results: SearchResultsByType;
  counts: SearchCounts;
  facets?: SearchFacets;
  meta: SearchMeta;
}

export interface SearchResultsByType {
  clips: Clip[];
  creators: User[];
  games: SearchGame[];
  tags: Tag[];
}

export interface SearchCounts {
  clips: number;
  creators: number;
  games: number;
  tags: number;
}

export interface SearchFacets {
  languages?: FacetBucket[];
  games?: FacetBucket[];
  tags?: FacetBucket[];
  date_range?: DateRangeFacet;
}

export interface FacetBucket {
  key: string;
  label?: string;
  count: number;
}

export interface DateRangeFacet {
  last_hour: number;
  last_day: number;
  last_week: number;
  last_month: number;
  older: number;
}

export interface SearchMeta {
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
}

export interface SearchGame {
  id: string;
  name: string;
  clip_count: number;
}

export interface SearchSuggestion {
  text: string;
  type: 'query' | 'game' | 'creator' | 'tag';
}

export interface SearchFilters {
  gameId?: string;
  creatorId?: string;
  language?: string;
  tags?: string[];
  minVotes?: number;
  dateFrom?: string;
  dateTo?: string;
}

// Advanced search discovery types
export interface TrendingSearch {
  query: string;
  search_count: number;
  unique_users: number;
  avg_results: number;
}

export interface SearchHistoryItem {
  query: string;
  result_count: number;
  created_at: string;
}

export interface FailedSearch {
  query: string;
  search_count: number;
  last_searched: string;
}

export interface SearchAnalyticsSummary {
  total_searches: number;
  unique_users: number;
  failed_searches: number;
  avg_results_per_search: number;
  success_rate: number;
}

export interface SavedSearch {
  id: string;
  query: string;
  filters?: SearchFilters;
  created_at: string;
  name?: string;
}
