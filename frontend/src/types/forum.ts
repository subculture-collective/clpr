/**
 * Forum type definitions
 * Based on backend API types from forum_handler.go
 */

export interface ForumThread {
  id: string;
  user_id: string;
  username: string;
  title: string;
  content: string;
  game_id?: string;
  game_name?: string;
  tags: string[];
  view_count: number;
  reply_count: number;
  locked: boolean;
  locked_at?: string;
  pinned: boolean;
  created_at: string;
  updated_at: string;
}

export interface ForumReply {
  id: string;
  user_id: string;
  username: string;
  thread_id: string;
  parent_reply_id?: string;
  content: string;
  depth: number;
  path: string;
  created_at: string;
  updated_at: string;
  replies?: ForumReply[];
  is_deleted?: boolean;
  // Vote-related fields
  vote_stats?: VoteStats;
  reputation?: ReputationScore;
}

export interface CreateThreadRequest {
  title: string;
  content: string;
  game_id?: string;
  tags?: string[];
}

export interface CreateReplyRequest {
  content: string;
  parent_reply_id?: string;
}

export interface UpdateReplyRequest {
  content: string;
}

export interface ForumThreadsResponse {
  threads: ForumThread[];
  total: number;
  page: number;
  limit: number;
}

export interface ForumThreadDetailResponse {
  thread: ForumThread;
  replies: ForumReply[];
}

export interface ForumSearchResponse {
  data: SearchResult[];
  meta: {
    page: number;
    limit: number;
    query: string;
    author?: string;
    sort: string;
    count: number;
    has_more: boolean;
  };
}

export interface SearchResult {
  type: 'thread' | 'reply';
  id: string;
  title?: string; // Only for threads
  body: string;
  author_id: string;
  author_name: string;
  thread_id?: string; // Only for replies
  vote_count: number;
  created_at: string;
  headline: string; // Highlighted snippet
  rank: number;
}

export type ForumSort = 'newest' | 'most-replied' | 'trending' | 'hot';

export interface ForumFilters {
  game_id?: string;
  tags?: string[];
  sort?: ForumSort;
  search?: string;
}

// Voting system types
export interface VoteStats {
  upvotes: number;
  downvotes: number;
  net_votes: number;
  user_vote: -1 | 0 | 1; // -1=downvote, 0=no vote, 1=upvote
}

export interface ReputationScore {
  user_id: string;
  score: number;
  badge: 'new' | 'contributor' | 'expert' | 'moderator';
  votes: number;
  threads: number;
  replies: number;
  updated_at: string;
}

export interface VoteRequest {
  vote_value: -1 | 0 | 1;
}

// Analytics types
export interface UserContribution {
  user_id: string;
  username: string;
  thread_count: number;
  reply_count: number;
  reputation_score: number;
}

export interface ForumAnalytics {
  total_threads: number;
  total_replies: number;
  total_users: number;
  posts_today: number;
  posts_this_week: number;
  posts_this_month: number;
  active_users_today: number;
  active_users_week: number;
  trending_topics: string[];
  popular_threads: ForumThread[];
  top_contributors: UserContribution[];
  last_updated: string;
}

export interface ForumAnalyticsResponse {
  success: boolean;
  data: ForumAnalytics;
}

export interface PopularDiscussionsResponse {
  success: boolean;
  data: ForumThread[];
  meta: {
    timeframe: string;
    count: number;
    limit: number;
  };
}

export interface HelpfulReply extends ForumReply {
  thread_title: string;
  net_votes: number;
  upvotes: number;
  downvotes: number;
}

export interface HelpfulRepliesResponse {
  success: boolean;
  data: HelpfulReply[];
  meta: {
    timeframe: string;
    count: number;
    limit: number;
  };
}

export interface FlagContentRequest {
  target_type: 'thread' | 'reply';
  target_id: string;
  reason: 'spam' | 'harassment' | 'off-topic' | 'misinformation' | 'other';
  details?: string;
}
