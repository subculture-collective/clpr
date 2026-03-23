/**
 * Forum API client
 * Handles all forum-related API requests
 */

import { apiClient } from './api';
import type {
  ForumThread,
  ForumReply,
  CreateThreadRequest,
  CreateReplyRequest,
  UpdateReplyRequest,
  ForumThreadsResponse,
  ForumThreadDetailResponse,
  ForumSearchResponse,
  ForumSort,
  ForumAnalyticsResponse,
  PopularDiscussionsResponse,
  HelpfulRepliesResponse,
  FlagContentRequest,
} from '@/types/forum';

interface ListThreadsParams {
  page?: number;
  limit?: number;
  sort?: ForumSort;
  game_id?: string;
  tags?: string[];
}

interface SearchThreadsParams {
  q: string;
  author?: string;
  sort?: 'relevance' | 'date' | 'votes';
  page?: number;
  limit?: number;
}

export const forumApi = {
  /**
   * List forum threads with optional filters
   */
  async listThreads(params: ListThreadsParams = {}): Promise<ForumThreadsResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page) queryParams.append('page', params.page.toString());
    if (params.limit) queryParams.append('limit', params.limit.toString());
    if (params.sort) queryParams.append('sort', params.sort);
    if (params.game_id) queryParams.append('game_id', params.game_id);
    if (params.tags && params.tags.length > 0) {
      params.tags.forEach(tag => queryParams.append('tags', tag));
    }

    const response = await apiClient.get(
      `/forum/threads?${queryParams.toString()}`
    );
    const body = response.data;
    // Backend returns { data: Thread[], meta: {...}, success } — map to our type
    return {
      threads: body.data ?? body.threads ?? [],
      total: body.meta?.count ?? body.total ?? 0,
      page: body.meta?.page ?? body.page ?? 1,
      limit: body.meta?.limit ?? body.limit ?? 20,
    };
  },

  /**
   * Get a single thread with its replies
   */
  async getThread(threadId: string): Promise<ForumThreadDetailResponse> {
    const response = await apiClient.get(
      `/forum/threads/${threadId}`
    );
    const body = response.data;
    // Backend may return { data: { thread, replies }, success } or { thread, replies }
    if (body.data?.thread) {
      return body.data;
    }
    if (body.thread) {
      return body;
    }
    // Fallback: body IS the thread detail
    return { thread: body.data ?? body, replies: body.replies ?? [] };
  },

  /**
   * Create a new forum thread
   */
  async createThread(data: CreateThreadRequest): Promise<ForumThread> {
    const response = await apiClient.post<{ success: boolean; data: ForumThread }>('/forum/threads', data);
    return response.data.data ?? response.data as unknown as ForumThread;
  },

  /**
   * Create a reply to a thread
   */
  async createReply(threadId: string, data: CreateReplyRequest): Promise<ForumReply> {
    const response = await apiClient.post<ForumReply>(
      `/forum/threads/${threadId}/replies`,
      data
    );
    return response.data;
  },

  /**
   * Update an existing reply
   */
  async updateReply(replyId: string, data: UpdateReplyRequest): Promise<ForumReply> {
    const response = await apiClient.patch<ForumReply>(
      `/forum/replies/${replyId}`,
      data
    );
    return response.data;
  },

  /**
   * Delete a reply (soft delete)
   */
  async deleteReply(replyId: string): Promise<void> {
    await apiClient.delete(`/forum/replies/${replyId}`);
  },

  /**
   * Search forum threads and replies
   */
  async search(params: SearchThreadsParams): Promise<ForumSearchResponse> {
    const queryParams = new URLSearchParams();
    queryParams.append('q', params.q);
    if (params.author) queryParams.append('author', params.author);
    if (params.sort) queryParams.append('sort', params.sort);
    if (params.page) queryParams.append('page', params.page.toString());
    if (params.limit) queryParams.append('limit', params.limit.toString());

    const response = await apiClient.get<ForumSearchResponse>(
      `/forum/search?${queryParams.toString()}`
    );
    return response.data;
  },

  /**
   * Get forum analytics data
   */
  async getAnalytics(): Promise<ForumAnalyticsResponse> {
    const response = await apiClient.get<ForumAnalyticsResponse>('/forum/analytics');
    return response.data;
  },

  /**
   * Get popular discussions
   */
  async getPopularDiscussions(timeframe: 'day' | 'week' | 'month' | 'all' = 'week', limit = 20): Promise<PopularDiscussionsResponse> {
    const params = new URLSearchParams();
    params.append('timeframe', timeframe);
    params.append('limit', limit.toString());
    
    const response = await apiClient.get<PopularDiscussionsResponse>(`/forum/popular?${params.toString()}`);
    return response.data;
  },

  /**
   * Get most helpful replies
   */
  async getMostHelpfulReplies(timeframe: 'week' | 'month' | 'all' = 'month', limit = 20): Promise<HelpfulRepliesResponse> {
    const params = new URLSearchParams();
    params.append('timeframe', timeframe);
    params.append('limit', limit.toString());

    const response = await apiClient.get<HelpfulRepliesResponse>(`/forum/helpful-replies?${params.toString()}`);
    return response.data;
  },

  /**
   * Flag content (thread or reply) for moderation review
   */
  async flagContent(data: FlagContentRequest) {
    const response = await apiClient.post('/forum/flag', data);
    return response.data;
  },
};
