import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { tagApi } from "../lib/tag-api";
import type { TagLane } from "../types/tag";

// List tags
export const useTags = (params?: {
  sort?: "popularity" | "alphabetical" | "recent" | "trending" | "curated";
  lane?: TagLane;
  limit?: number;
  page?: number;
}) => {
  return useQuery({
    queryKey: ["tags", params],
    queryFn: () => tagApi.listTags(params),
  });
};

// Search tags
export const useTagSearch = (query: string, enabled = true) => {
  return useQuery({
    queryKey: ["tags", "search", query],
    queryFn: () => tagApi.searchTags(query),
    enabled: enabled && query.length >= 2,
  });
};

// Get tag details
export const useTag = (slug: string) => {
  return useQuery({
    queryKey: ["tags", slug],
    queryFn: () => tagApi.getTag(slug),
    enabled: slug.length > 0,
    retry: false,
  });
};

// Get clips by tag
export const useClipsByTag = (
  slug: string,
  params?: { limit?: number; page?: number }
) => {
  return useQuery({
    queryKey: ["tags", slug, "clips", params],
    queryFn: () => tagApi.getClipsByTag(slug, params),
  });
};

// Get clip tags
export const useClipTags = (clipId: string) => {
  return useQuery({
    queryKey: ["clips", clipId, "tags"],
    queryFn: () => tagApi.getClipTags(clipId),
  });
};

// Add tags to clip
export const useAddTagsToClip = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ clipId, tagSlugs }: { clipId: string; tagSlugs: string[] }) =>
      tagApi.addTagsToClip(clipId, tagSlugs),
    onSuccess: (_, variables) => {
      // Invalidate clip tags
      queryClient.invalidateQueries({
        queryKey: ["clips", variables.clipId, "tags"],
      });
      // Invalidate tags list
      queryClient.invalidateQueries({
        queryKey: ["tags"],
      });
    },
    onError: (error) => {
      console.error("Failed to add tags to clip:", error);
    },
  });
};

// Remove tag from clip
export const useRemoveTagFromClip = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ clipId, tagSlug }: { clipId: string; tagSlug: string }) =>
      tagApi.removeTagFromClip(clipId, tagSlug),
    onSuccess: (_, variables) => {
      // Invalidate clip tags
      queryClient.invalidateQueries({
        queryKey: ["clips", variables.clipId, "tags"],
      });
      // Invalidate tags list
      queryClient.invalidateQueries({
        queryKey: ["tags"],
      });
    },
    onError: (error) => {
      console.error("Failed to remove tag from clip:", error);
    },
  });
};

// Admin: Create tag
export const useCreateTag = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: tagApi.createTag,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tags"],
      });
    },
    onError: (error) => {
      console.error("Failed to create tag:", error);
    },
  });
};

// Admin: Update tag
export const useUpdateTag = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof tagApi.updateTag>[1] }) =>
      tagApi.updateTag(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["tags"],
      });
      queryClient.invalidateQueries({
        queryKey: ["tags", variables.data.slug],
      });
    },
    onError: (error) => {
      console.error("Failed to update tag:", error);
    },
  });
};

// Admin: Delete tag
export const useDeleteTag = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: tagApi.deleteTag,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tags"],
      });
    },
    onError: (error) => {
      console.error("Failed to delete tag:", error);
    },
  });
};
