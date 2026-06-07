export interface Comment {
  id: string;
  clip_id: string;
  user_id: string;
  username: string;
  user_display_name?: string;
  user_avatar?: string;
  user_karma?: number;
  user_role?: 'admin' | 'moderator' | 'user';
  user_verified?: boolean;
  parent_comment_id: string | null;
  content: string;
  rendered_content?: string;
  vote_score: number;
  created_at: string;
  updated_at: string;
  edited_at?: string;
  is_deleted: boolean;
  is_removed: boolean;
  removed_reason?: string;
  is_hidden_by_creator_moderation?: boolean;
  creator_moderation_message?: string;
  depth: number;
  reply_count: number;
  child_count: number;
  user_vote?: 1 | -1 | null;
  replies?: Comment[];
}

export interface CommentFeedResponse {
  comments: Comment[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
}

export type CommentSortOption = 'best' | 'top' | 'new' | 'old' | 'controversial';

export interface CreateCommentPayload {
  clip_id: string;
  content: string;
  parent_comment_id?: string | null;
}

export interface UpdateCommentPayload {
  content: string;
}

export interface CommentVotePayload {
  comment_id: string;
  vote_type: 1 | -1;
}

export interface ReportCommentPayload {
  comment_id: string;
  reason: 'spam' | 'harassment' | 'off-topic' | 'misinformation' | 'other';
  description?: string;
}
