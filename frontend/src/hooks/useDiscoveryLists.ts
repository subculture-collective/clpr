/**
 * DEPRECATED: Discovery list React Query hooks.
 *
 * These hooks consume the deprecated discovery-list API. Discovery lists have
 * been merged into the curated playlist system — prefer the playlist hooks in
 * `usePlaylist.ts` instead (e.g. `useInfiniteFeaturedPlaylists`).
 *
 * Kept for backward compatibility; will be removed in a future release.
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { discoveryListApi } from '../lib/discovery-list-api';
import { useAuth } from './useAuth';

export function useDiscoveryLists(params?: {
  featured?: boolean;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: ['discovery-lists', params],
    queryFn: () => discoveryListApi.listDiscoveryLists(params),
  });
}

export function useDiscoveryList(idOrSlug: string) {
  return useQuery({
    queryKey: ['discovery-list', idOrSlug],
    queryFn: () => discoveryListApi.getDiscoveryList(idOrSlug),
    enabled: !!idOrSlug,
  });
}

export function useDiscoveryListClips(
  listId: string,
  params?: { limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: ['discovery-list-clips', listId, params],
    queryFn: () => discoveryListApi.getDiscoveryListClips(listId, params),
    enabled: !!listId,
  });
}

export function useFollowDiscoveryList() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: string; slug?: string; idOrSlug?: string }) =>
      discoveryListApi.followDiscoveryList(id),
    onSuccess: (_data, { id, slug, idOrSlug }) => {
      // Invalidate queries to refresh data
      // Invalidate by ID
      queryClient.invalidateQueries({ queryKey: ['discovery-list', id] });
      // Also invalidate by slug if provided (for slug-based cache keys)
      if (slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', slug] });
      }
      // Also invalidate by the original idOrSlug from URL (in case it differs)
      if (idOrSlug && idOrSlug !== id && idOrSlug !== slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', idOrSlug] });
      }
      queryClient.invalidateQueries({ queryKey: ['discovery-lists'] });
      queryClient.invalidateQueries({ queryKey: ['user-followed-lists'] });
    },
    onError: (error) => {
      console.error('Failed to follow discovery list:', error);
    },
  });
}

export function useUnfollowDiscoveryList() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: string; slug?: string; idOrSlug?: string }) =>
      discoveryListApi.unfollowDiscoveryList(id),
    onSuccess: (_data, { id, slug, idOrSlug }) => {
      // Invalidate queries to refresh data
      // Invalidate by ID
      queryClient.invalidateQueries({ queryKey: ['discovery-list', id] });
      // Also invalidate by slug if provided (for slug-based cache keys)
      if (slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', slug] });
      }
      // Also invalidate by the original idOrSlug from URL (in case it differs)
      if (idOrSlug && idOrSlug !== id && idOrSlug !== slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', idOrSlug] });
      }
      queryClient.invalidateQueries({ queryKey: ['discovery-lists'] });
      queryClient.invalidateQueries({ queryKey: ['user-followed-lists'] });
    },
    onError: (error) => {
      console.error('Failed to unfollow discovery list:', error);
    },
  });
}

export function useBookmarkDiscoveryList() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: string; slug?: string; idOrSlug?: string }) =>
      discoveryListApi.bookmarkDiscoveryList(id),
    onSuccess: (_data, { id, slug, idOrSlug }) => {
      // Invalidate queries to refresh data
      // Invalidate by ID
      queryClient.invalidateQueries({ queryKey: ['discovery-list', id] });
      // Also invalidate by slug if provided (for slug-based cache keys)
      if (slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', slug] });
      }
      // Also invalidate by the original idOrSlug from URL (in case it differs)
      if (idOrSlug && idOrSlug !== id && idOrSlug !== slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', idOrSlug] });
      }
      queryClient.invalidateQueries({ queryKey: ['discovery-lists'] });
    },
    onError: (error) => {
      console.error('Failed to bookmark discovery list:', error);
    },
  });
}

export function useUnbookmarkDiscoveryList() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: string; slug?: string; idOrSlug?: string }) =>
      discoveryListApi.unbookmarkDiscoveryList(id),
    onSuccess: (_data, { id, slug, idOrSlug }) => {
      // Invalidate queries to refresh data
      // Invalidate by ID
      queryClient.invalidateQueries({ queryKey: ['discovery-list', id] });
      // Also invalidate by slug if provided (for slug-based cache keys)
      if (slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', slug] });
      }
      // Also invalidate by the original idOrSlug from URL (in case it differs)
      if (idOrSlug && idOrSlug !== id && idOrSlug !== slug) {
        queryClient.invalidateQueries({ queryKey: ['discovery-list', idOrSlug] });
      }
      queryClient.invalidateQueries({ queryKey: ['discovery-lists'] });
    },
    onError: (error) => {
      console.error('Failed to unbookmark discovery list:', error);
    },
  });
}

export function useUserFollowedLists(params?: {
  limit?: number;
  offset?: number;
}) {
  const { isAuthenticated } = useAuth();

  return useQuery({
    queryKey: ['user-followed-lists', params],
    queryFn: () => discoveryListApi.getUserFollowedLists(params),
    enabled: isAuthenticated,
  });
}
