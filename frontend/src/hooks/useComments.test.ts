import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, InfiniteData } from '@tanstack/react-query';
import { createElement, type ReactNode } from 'react';
import {
  useComments,
  useCreateComment,
  useUpdateComment,
  useDeleteComment,
  useCommentVote,
  useReportComment,
} from './useComments';
import * as commentApi from '@/lib/comment-api';
import type {
  Comment,
  CommentFeedResponse,
  CreateCommentPayload,
  UpdateCommentPayload,
  CommentVotePayload,
  ReportCommentPayload,
} from '@/types/comment';

// Mock the comment API
vi.mock('@/lib/comment-api');

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
};

const mockComment: Comment = {
  id: 'comment-1',
  clip_id: 'clip-1',
  user_id: 'user-1',
  username: 'testuser',
  user_avatar: 'https://example.com/avatar.png',
  user_karma: 1234,
  user_role: 'user',
  parent_comment_id: null,
  content: 'Test comment',
  vote_score: 5,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  is_deleted: false,
  is_removed: false,
  reply_count: 0,
  depth: 0,
  child_count: 0,
  user_vote: null,
  replies: [],
};

const mockCommentFeedResponse: CommentFeedResponse = {
  comments: [mockComment],
  total: 1,
  page: 1,
  limit: 10,
  has_more: false,
};

describe('useComments', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useComments hook', () => {
    it('should fetch comments successfully', async () => {
      vi.mocked(commentApi.fetchComments).mockResolvedValue(mockCommentFeedResponse);

      const { result } = renderHook(() => useComments('clip-1', 'best'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data?.pages[0].comments).toEqual([mockComment]);
      expect(commentApi.fetchComments).toHaveBeenCalledWith({
        clipId: 'clip-1',
        sort: 'best',
        pageParam: 1,
        limit: 10,
        includeReplies: true,
      });
    });

    it('should handle fetch error', async () => {
      const error = new Error('Failed to fetch comments');
      vi.mocked(commentApi.fetchComments).mockRejectedValue(error);

      const { result } = renderHook(() => useComments('clip-1'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toEqual(error);
    });

    it('should handle loading state', () => {
      vi.mocked(commentApi.fetchComments).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useComments('clip-1'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should not fetch when clipId is empty', () => {
      const { result } = renderHook(() => useComments(''), {
        wrapper: createWrapper(),
      });

      expect(result.current.fetchStatus).toBe('idle');
      expect(commentApi.fetchComments).not.toHaveBeenCalled();
    });

    it('should support different sort options', async () => {
      vi.mocked(commentApi.fetchComments).mockResolvedValue(mockCommentFeedResponse);

      const { result } = renderHook(() => useComments('clip-1', 'new'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(commentApi.fetchComments).toHaveBeenCalledWith({
        clipId: 'clip-1',
        sort: 'new',
        pageParam: 1,
        limit: 10,
        includeReplies: true,
      });
    });

    it('should handle pagination with has_more', async () => {
      const page1Response: CommentFeedResponse = {
        ...mockCommentFeedResponse,
        has_more: true,
      };

      vi.mocked(commentApi.fetchComments).mockResolvedValue(page1Response);

      const { result } = renderHook(() => useComments('clip-1'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.hasNextPage).toBe(true);
    });
  });

  describe('useCreateComment hook', () => {
    it('should create comment successfully', async () => {
      vi.mocked(commentApi.createComment).mockResolvedValue(mockComment);

      const { result } = renderHook(() => useCreateComment(), {
        wrapper: createWrapper(),
      });

      const payload: CreateCommentPayload = {
        clip_id: 'clip-1',
        content: 'New comment',
        parent_comment_id: null,
      };

      result.current.mutate(payload);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockComment);
      expect(commentApi.createComment).toHaveBeenCalledWith(payload);
    });

    it('should handle create error', async () => {
      const error = new Error('Failed to create comment');
      vi.mocked(commentApi.createComment).mockRejectedValue(error);

      const { result } = renderHook(() => useCreateComment(), {
        wrapper: createWrapper(),
      });

      const payload: CreateCommentPayload = {
        clip_id: 'clip-1',
        content: 'New comment',
      };

      result.current.mutate(payload);

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toEqual(error);
    });

    it('should handle loading state during creation', async () => {
      vi.mocked(commentApi.createComment).mockImplementation(
        () => new Promise(() => {})
      );

      const { result } = renderHook(() => useCreateComment(), {
        wrapper: createWrapper(),
      });

      result.current.mutate({
        clip_id: 'clip-1',
        content: 'New comment',
      });

      await waitFor(() => {
        expect(result.current.isPending).toBe(true);
      });
    });

    it('should optimistically update child_count when creating a reply', async () => {
      const parentComment: Comment = {
        ...mockComment,
        id: 'parent-1',
        reply_count: 2,
        child_count: 2,
        replies: [],
      };

      const replyComment: Comment = {
        ...mockComment,
        id: 'reply-1',
        parent_comment_id: 'parent-1',
        content: 'Reply comment',
        reply_count: 0,
      };

      vi.mocked(commentApi.createComment).mockResolvedValue(replyComment);

      const queryClient = new QueryClient({
        defaultOptions: {
          queries: { retry: false },
          mutations: { retry: false },
        },
      });

      // Pre-populate the query cache with parent comment
      queryClient.setQueryData(['comments', 'clip-1'], {
        pages: [{
          comments: [parentComment],
          total: 1,
          page: 1,
          limit: 10,
          has_more: false,
        }],
        pageParams: [1],
      });

      const wrapper = ({ children }: { children: React.ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);

      const { result } = renderHook(() => useCreateComment(), { wrapper });

      const payload: CreateCommentPayload = {
        clip_id: 'clip-1',
        content: 'Reply comment',
        parent_comment_id: 'parent-1',
      };

      result.current.mutate(payload);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      // Verify child_count was updated
      const updatedData = queryClient.getQueryData(['comments', 'clip-1']) as InfiniteData<CommentFeedResponse>;
      expect(updatedData.pages[0].comments[0].child_count).toBe(3);
    });

    it('should rollback child_count on error when creating a reply fails', async () => {
      const parentComment: Comment = {
        ...mockComment,
        id: 'parent-1',
        child_count: 2,
        replies: [],
      };

      const error = new Error('Failed to create reply');
      vi.mocked(commentApi.createComment).mockRejectedValue(error);

      const queryClient = new QueryClient({
        defaultOptions: {
          queries: { retry: false },
          mutations: { retry: false },
        },
      });

      // Pre-populate the query cache with parent comment
      queryClient.setQueryData(['comments', 'clip-1'], {
        pages: [{
          comments: [parentComment],
          total: 1,
          page: 1,
          limit: 10,
          has_more: false,
        }],
        pageParams: [1],
      });

      const wrapper = ({ children }: { children: React.ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);

      const { result } = renderHook(() => useCreateComment(), { wrapper });

      const payload: CreateCommentPayload = {
        clip_id: 'clip-1',
        content: 'Reply comment',
        parent_comment_id: 'parent-1',
      };

      result.current.mutate(payload);

      await waitFor(() => expect(result.current.isError).toBe(true));

      // Verify child_count was rolled back to original value
      const rolledBackData = queryClient.getQueryData(['comments', 'clip-1']) as InfiniteData<CommentFeedResponse>;
      expect(rolledBackData.pages[0].comments[0].child_count).toBe(2);
      expect(result.current.error).toEqual(error);
    });
  });

  describe('useUpdateComment hook', () => {
    it('should update comment successfully', async () => {
      const successResponse = { message: 'Comment updated successfully' };
      vi.mocked(commentApi.updateComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useUpdateComment(), {
        wrapper: createWrapper(),
      });

      const payload: UpdateCommentPayload = {
        content: 'Updated content',
      };

      result.current.mutate({ commentId: 'comment-1', payload });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(successResponse);
      expect(commentApi.updateComment).toHaveBeenCalledWith('comment-1', payload);
    });

    it('should handle update error', async () => {
      const error = new Error('Failed to update comment');
      vi.mocked(commentApi.updateComment).mockRejectedValue(error);

      const { result } = renderHook(() => useUpdateComment(), {
        wrapper: createWrapper(),
      });

      result.current.mutate({
        commentId: 'comment-1',
        payload: { content: 'Updated' },
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useDeleteComment hook', () => {
    it('should delete comment successfully', async () => {
      const successResponse = { message: 'Comment deleted successfully' };
      vi.mocked(commentApi.deleteComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useDeleteComment(), {
        wrapper: createWrapper(),
      });

      result.current.mutate('comment-1');

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(successResponse);
      expect(commentApi.deleteComment).toHaveBeenCalledWith('comment-1');
    });

    it('should handle delete error', async () => {
      const error = new Error('Failed to delete comment');
      vi.mocked(commentApi.deleteComment).mockRejectedValue(error);

      const { result } = renderHook(() => useDeleteComment(), {
        wrapper: createWrapper(),
      });

      result.current.mutate('comment-1');

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toEqual(error);
    });
  });

  describe('useCommentVote hook', () => {
    it('should vote on comment successfully', async () => {
      const successResponse = { message: 'Vote recorded successfully' };
      vi.mocked(commentApi.voteOnComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useCommentVote(), {
        wrapper: createWrapper(),
      });

      const payload: CommentVotePayload = {
        comment_id: 'comment-1',
        vote_type: 1,
      };

      result.current.mutate(payload);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(successResponse);
      expect(commentApi.voteOnComment).toHaveBeenCalledWith(payload);
    });

    it('should handle vote error', async () => {
      const error = new Error('Failed to vote');
      vi.mocked(commentApi.voteOnComment).mockRejectedValue(error);

      const { result } = renderHook(() => useCommentVote(), {
        wrapper: createWrapper(),
      });

      result.current.mutate({ comment_id: 'comment-1', vote_type: -1 });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toEqual(error);
    });

    it('should handle upvote (vote_type: 1)', async () => {
      const successResponse = { message: 'Vote recorded successfully' };
      vi.mocked(commentApi.voteOnComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useCommentVote(), {
        wrapper: createWrapper(),
      });

      result.current.mutate({ comment_id: 'comment-1', vote_type: 1 });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(commentApi.voteOnComment).toHaveBeenCalledWith({
        comment_id: 'comment-1',
        vote_type: 1,
      });
    });

    it('should handle downvote (vote_type: -1)', async () => {
      const successResponse = { message: 'Vote recorded successfully' };
      vi.mocked(commentApi.voteOnComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useCommentVote(), {
        wrapper: createWrapper(),
      });

      result.current.mutate({ comment_id: 'comment-1', vote_type: -1 });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(commentApi.voteOnComment).toHaveBeenCalledWith({
        comment_id: 'comment-1',
        vote_type: -1,
      });
    });
  });

  describe('useReportComment hook', () => {
    it('should report comment successfully', async () => {
      const successResponse = { message: 'Report submitted successfully' };
      vi.mocked(commentApi.reportComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useReportComment(), {
        wrapper: createWrapper(),
      });

      const payload: ReportCommentPayload = {
        comment_id: 'comment-1',
        reason: 'spam',
        description: 'This is spam',
      };

      result.current.mutate(payload);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(successResponse);
      expect(commentApi.reportComment).toHaveBeenCalledWith(payload);
    });

    it('should handle report error', async () => {
      const error = new Error('Failed to report comment');
      vi.mocked(commentApi.reportComment).mockRejectedValue(error);

      const { result } = renderHook(() => useReportComment(), {
        wrapper: createWrapper(),
      });

      result.current.mutate({
        comment_id: 'comment-1',
        reason: 'harassment',
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toEqual(error);
    });

    it('should handle different report reasons', async () => {
      const successResponse = { message: 'Report submitted successfully' };
      vi.mocked(commentApi.reportComment).mockResolvedValue(successResponse);

      const { result } = renderHook(() => useReportComment(), {
        wrapper: createWrapper(),
      });

      const reasons: Array<ReportCommentPayload['reason']> = [
        'spam',
        'harassment',
        'nsfw',
        'violence',
        'copyright',
        'other',
      ];

      for (const reason of reasons) {
        result.current.mutate({
          comment_id: 'comment-1',
          reason,
        });

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(commentApi.reportComment).toHaveBeenCalledWith({
          comment_id: 'comment-1',
          reason,
        });

        result.current.reset();
      }
    });
  });
});
