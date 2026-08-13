import apiClient from './api';
import type {
  Comment,
  CommentFeedResponse,
  CommentSortOption,
  CreateCommentPayload,
  UpdateCommentPayload,
  CommentVotePayload,
  ReportCommentPayload,
} from '@/types/comment';

type ApiComment = {
  id: string;
  clip_id: string;
  user_id: string;
  parent_comment_id: string | null;
  content: string;
  rendered_content?: string;
  vote_score: number;
  reply_count?: number;
  is_edited?: boolean;
  is_removed: boolean;
  removed_reason?: string;
  is_hidden_by_creator_moderation?: boolean;
  creator_moderation_message?: string;
  created_at: string;
  updated_at: string;
  author_username: string;
  author_display_name: string;
  author_avatar_url?: string;
  author_karma?: number;
  author_role?: 'admin' | 'moderator' | 'user' | string;
  user_vote?: 1 | -1 | 0 | null;
  replies?: ApiComment[];
};

const normalizeComment = (comment: ApiComment, depth = 0): Comment => {
  const childCount = comment.reply_count ?? comment.replies?.length ?? 0;

  return {
    id: comment.id,
    clip_id: comment.clip_id,
    user_id: comment.user_id,
    username: comment.author_username,
    user_display_name: comment.author_display_name,
    user_avatar: comment.author_avatar_url,
    user_karma: comment.author_karma,
    user_role: comment.author_role as Comment['user_role'],
    parent_comment_id: comment.parent_comment_id,
    content: comment.content,
    rendered_content: comment.rendered_content,
    vote_score: comment.vote_score,
    created_at: comment.created_at,
    updated_at: comment.updated_at,
    edited_at: comment.is_edited ? comment.updated_at : undefined,
    is_deleted: comment.is_removed && comment.content === '[deleted]',
    is_removed: comment.is_removed,
    removed_reason: comment.removed_reason,
    is_hidden_by_creator_moderation: comment.is_hidden_by_creator_moderation,
    creator_moderation_message: comment.creator_moderation_message,
    depth,
    reply_count: childCount,
    child_count: childCount,
    user_vote: comment.user_vote === 0 ? null : (comment.user_vote as Comment['user_vote']),
    replies: comment.replies?.map((reply) => normalizeComment(reply, depth + 1)) || [],
  };
};

/**
 * Fetch comments for a clip
 */
export async function fetchComments({
  clipId,
  sort = 'best',
  pageParam = 1,
  limit = 10,
  includeReplies = false,
}: {
  clipId: string;
  sort?: CommentSortOption;
  pageParam?: number;
  limit?: number;
  includeReplies?: boolean;
}): Promise<CommentFeedResponse> {
  const response = await apiClient.get<{
    comments: ApiComment[];
    total?: number;
    next_cursor?: number;
    has_more: boolean;
  }>(`/clips/${clipId}/comments`, {
    params: {
      sort,
      cursor: (pageParam - 1) * limit,
      limit,
      include_replies: includeReplies,
    },
  });

  const comments = (response.data.comments || []).map((comment) => normalizeComment(comment));

  return {
    comments,
    total: response.data.total || comments.length,
    page: pageParam,
    limit,
    has_more: response.data.has_more,
  };
}

/**
 * Create a new comment
 */
export async function createComment(
  payload: CreateCommentPayload
): Promise<Comment> {
  const response = await apiClient.post<Comment>(
    `/clips/${payload.clip_id}/comments`,
    {
      content: payload.content,
      parent_comment_id: payload.parent_comment_id,
    }
  );
  return normalizeComment(response.data as unknown as ApiComment);
}

/**
 * Update a comment
 */
export async function updateComment(
  commentId: string,
  payload: UpdateCommentPayload
): Promise<{ message: string }> {
  const response = await apiClient.put<{ message: string }>(
    `/comments/${commentId}`,
    payload
  );
  return response.data;
}

/**
 * Delete a comment
 */
export async function deleteComment(commentId: string): Promise<{ message: string }> {
  const response = await apiClient.delete<{ message: string }>(
    `/comments/${commentId}`
  );
  return response.data;
}

/**
 * Vote on a comment
 */
export async function voteOnComment(
  payload: CommentVotePayload
): Promise<{ message: string }> {
  const response = await apiClient.post<{ message: string }>(
    `/comments/${payload.comment_id}/vote`,
    {
      vote: payload.vote_type,
    }
  );
  return response.data;
}

/**
 * Report a comment
 */
export async function reportComment(
  payload: ReportCommentPayload
): Promise<{ message: string }> {
  const response = await apiClient.post<{ message: string }>(
    '/reports',
    {
      target_type: 'comment',
      target_id: payload.comment_id,
      reason: payload.reason,
      description: payload.description,
    }
  );
  return response.data;
}
