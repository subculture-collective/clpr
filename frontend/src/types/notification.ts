export interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  link?: string;
  is_read: boolean;
  created_at: string;
  expires_at?: string;
  source_user_id?: string;
  source_content_id?: string;
  source_content_type?: string;
  // Included with source user info
  source_username?: string;
  source_display_name?: string;
  source_avatar_url?: string;
}

export type NotificationType =
  | 'reply'
  | 'mention'
  | 'vote_milestone'
  | 'badge_earned'
  | 'rank_up'
  | 'favorited_clip_comment'
  | 'content_removed'
  | 'warning'
  | 'ban'
  | 'appeal_decision'
  | 'submission_approved'
  | 'submission_rejected'
  | 'new_report'
  | 'pending_submissions'
  | 'system_alert'
  | 'clip_comment'
  | 'clip_view_threshold'
  | 'clip_vote_threshold'
  | 'broadcaster_live'
  | 'stream_live';

export interface NotificationPreferences {
  user_id: string;
  in_app_enabled: boolean;
  email_enabled: boolean;
  email_digest: 'immediate' | 'daily' | 'weekly' | 'never';
  
  // Account & Security
  notify_login_new_device: boolean;
  notify_failed_login: boolean;
  notify_password_changed: boolean;
  notify_email_changed: boolean;
  
  // Content notifications
  notify_replies: boolean;
  notify_mentions: boolean;
  notify_content_trending: boolean;
  notify_content_flagged: boolean;
  notify_votes: boolean;
  notify_favorited_clip_comment: boolean;
  
  // Community notifications
  notify_moderator_message: boolean;
  notify_user_followed: boolean;
  notify_comment_on_content: boolean;
  notify_discussion_reply: boolean;
  notify_badges: boolean;
  notify_rank_up: boolean;
  notify_moderation: boolean;
  
  // Creator-specific notification preferences
  notify_clip_approved: boolean;
  notify_clip_rejected: boolean;
  notify_clip_comments: boolean;
  notify_clip_threshold: boolean;
  
  // Broadcaster & Stream notifications
  notify_broadcaster_live: boolean;
  notify_stream_live: boolean;
  
  // Global preferences
  notify_marketing: boolean;
  notify_policy_updates: boolean;
  notify_platform_announcements: boolean;

  // Weekly streamer digest email for broadcasters who feature their clips (#85)
  notify_streamer_digest?: boolean;
  
  updated_at: string;
}

export interface NotificationListResponse {
  notifications: Notification[];
  unread_count: number;
  page: number;
  limit: number;
  has_more: boolean;
}

export interface NotificationCountResponse {
  unread_count: number;
}

export type NotificationFilter = 'all' | 'unread' | 'read';
