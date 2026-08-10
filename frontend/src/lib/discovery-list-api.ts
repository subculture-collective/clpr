/**
 * DEPRECATED: Discovery List API module.
 *
 * Discovery lists have been merged into the curated playlist system.
 * Use the playlist API (e.g. GET /playlists with ?is_curated=true) instead.
 *
 * This module is kept for backward compatibility with existing consumers.
 * It will be removed in a future release.
 *
 * @deprecated Use {@link ./playlist-api.ts} or the playlists endpoints.
 */
import { apiClient } from './api';
import type { DiscoveryListWithStats } from '../types/discoveryList';
import type { ClipFeedResponse } from '../types/clip';

export const discoveryListApi = {
  // List all discovery lists
  listDiscoveryLists: async (params?: {
    featured?: boolean;
    limit?: number;
    offset?: number;
  }) => {
    const response = await apiClient.get<DiscoveryListWithStats[]>(
      '/discovery-lists',
      { params }
    );
    return response.data;
  },

  // Get a single discovery list by ID or slug
  getDiscoveryList: async (idOrSlug: string) => {
    const response = await apiClient.get<DiscoveryListWithStats>(
      `/discovery-lists/${idOrSlug}`
    );
    return response.data;
  },

  // Get clips in a discovery list
  getDiscoveryListClips: async (
    listId: string,
    params?: { limit?: number; offset?: number }
  ) => {
    const response = await apiClient.get<ClipFeedResponse>(
      `/discovery-lists/${listId}/clips`,
      { params }
    );
    return response.data;
  },

  // Follow a discovery list
  followDiscoveryList: async (listId: string) => {
    const response = await apiClient.post<{ message: string }>(
      `/discovery-lists/${listId}/follow`
    );
    return response.data;
  },

  // Unfollow a discovery list
  unfollowDiscoveryList: async (listId: string) => {
    const response = await apiClient.delete<{ message: string }>(
      `/discovery-lists/${listId}/follow`
    );
    return response.data;
  },

  // Bookmark a discovery list
  bookmarkDiscoveryList: async (listId: string) => {
    const response = await apiClient.post<{ message: string }>(
      `/discovery-lists/${listId}/bookmark`
    );
    return response.data;
  },

  // Unbookmark a discovery list
  unbookmarkDiscoveryList: async (listId: string) => {
    const response = await apiClient.delete<{ message: string }>(
      `/discovery-lists/${listId}/bookmark`
    );
    return response.data;
  },

  // Get user's followed discovery lists
  getUserFollowedLists: async (params?: { limit?: number; offset?: number }) => {
    const response = await apiClient.get<DiscoveryListWithStats[]>(
      '/users/me/discovery-list-follows',
      { params }
    );
    return response.data;
  },

  // Admin API methods
  admin: {
    // List all discovery lists (including inactive)
    listAllDiscoveryLists: async (params?: {
      limit?: number;
      offset?: number;
    }) => {
      const response = await apiClient.get<DiscoveryListWithStats[]>(
        '/admin/discovery-lists',
        { params }
      );
      return response.data;
    },

    // Create a discovery list
    createDiscoveryList: async (data: {
      name: string;
      description?: string;
      is_featured?: boolean;
    }) => {
      const response = await apiClient.post('/admin/discovery-lists', data);
      return response.data;
    },

    // Update a discovery list
    updateDiscoveryList: async (
      listId: string,
      data: {
        name?: string;
        description?: string;
        is_featured?: boolean;
        is_active?: boolean;
      }
    ) => {
      const response = await apiClient.put(
        `/admin/discovery-lists/${listId}`,
        data
      );
      return response.data;
    },

    // Delete a discovery list
    deleteDiscoveryList: async (listId: string) => {
      const response = await apiClient.delete<{ message: string }>(
        `/admin/discovery-lists/${listId}`
      );
      return response.data;
    },

    // Add a clip to a list
    addClipToList: async (listId: string, clipId: string) => {
      const response = await apiClient.post<{ message: string }>(
        `/admin/discovery-lists/${listId}/clips`,
        { clip_id: clipId }
      );
      return response.data;
    },

    // Remove a clip from a list
    removeClipFromList: async (listId: string, clipId: string) => {
      const response = await apiClient.delete<{ message: string }>(
        `/admin/discovery-lists/${listId}/clips/${clipId}`
      );
      return response.data;
    },

    // Reorder clips in a list
    reorderListClips: async (listId: string, clipIds: string[]) => {
      const response = await apiClient.put<{ message: string }>(
        `/admin/discovery-lists/${listId}/clips/reorder`,
        { clip_ids: clipIds }
      );
      return response.data;
    },
  },
};
