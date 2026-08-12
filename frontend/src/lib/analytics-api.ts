import apiClient from './api';

// Analytics API types
export interface CreatorAnalyticsOverview {
  total_clips: number;
  total_views: number;
  total_upvotes: number;
  total_comments: number;
  avg_engagement_rate: number;
  follower_count: number;
}

export interface CreatorTopClip {
  id: string;
  twitch_clip_id: string;
  twitch_clip_url: string;
  embed_url: string;
  title: string;
  creator_name: string;
  creator_id?: string;
  broadcaster_name: string;
  broadcaster_id?: string;
  game_id?: string;
  game_name?: string;
  language?: string;
  thumbnail_url?: string;
  duration?: number;
  view_count: number;
  created_at: string;
  imported_at: string;
  vote_score: number;
  comment_count: number;
  favorite_count: number;
  is_featured: boolean;
  is_nsfw: boolean;
  is_removed: boolean;
  removed_reason?: string;
  views: number;
  engagement_rate: number;
}

export interface TrendDataPoint {
  date: string;
  value: number;
}

export interface ClipAnalytics {
  clip_id: string;
  total_views: number;
  unique_viewers: number;
  avg_view_duration?: number;
  total_shares: number;
  peak_concurrent_viewers: number;
  retention_rate?: number;
  first_viewed_at?: string;
  last_viewed_at?: string;
  updated_at: string;
}

export interface UserAnalytics {
  user_id: string;
  clips_upvoted: number;
  clips_downvoted: number;
  comments_posted: number;
  clips_favorited: number;
  searches_performed: number;
  days_active: number;
  total_karma_earned: number;
  last_active_at?: string;
  updated_at: string;
}

export interface PlatformOverviewMetrics {
  total_users: number;
  active_users_daily: number | null;
  active_users_monthly: number | null;
  total_clips: number;
  clips_added_today: number;
  total_votes: number;
  total_comments: number;
  avg_session_duration: number | null;
  generated_at: string;
  source: 'canonical' | 'snapshot';
}

export interface GameMetric {
  game_id?: string;
  game_name: string;
  clip_count: number;
  view_count: number;
}

export interface CreatorMetric {
  creator_id?: string;
  creator_name: string;
  clip_count: number;
  view_count: number;
  vote_score: number;
}

export interface TagMetric {
  tag_id: string;
  tag_name: string;
  usage_count: number;
}

export interface ContentMetrics {
  most_popular_games: GameMetric[];
  most_popular_creators: CreatorMetric[];
  trending_tags: TagMetric[];
  avg_clip_vote_score: number;
}

export interface GeographyMetric {
  country: string;
  view_count: number;
  percentage: number;
}

export interface DeviceMetric {
  device_type: string;
  view_count: number;
  percentage: number;
}

export interface CreatorAudienceInsights {
  top_countries: GeographyMetric[];
  device_types: DeviceMetric[];
  total_views: number;
}

// Creator Analytics API
export const getCreatorAnalyticsOverview = async (
  creatorName: string
): Promise<CreatorAnalyticsOverview> => {
  const response = await apiClient.get(
    `/creators/${encodeURIComponent(creatorName)}/analytics/overview`
  );
  return response.data;
};

export const getCreatorTopClips = async (
  creatorName: string,
  params?: { sort?: string; limit?: number }
): Promise<{ clips: CreatorTopClip[]; count: number }> => {
  const response = await apiClient.get(
    `/creators/${encodeURIComponent(creatorName)}/analytics/clips`,
    { params }
  );
  return response.data;
};

export const getCreatorTrends = async (
  creatorName: string,
  params?: { metric?: string; days?: number }
): Promise<{ metric: string; days: number; data: TrendDataPoint[] }> => {
  const response = await apiClient.get(
    `/creators/${encodeURIComponent(creatorName)}/analytics/trends`,
    { params }
  );
  return response.data;
};

export const getCreatorAudienceInsights = async (
  creatorName: string,
  params?: { limit?: number }
): Promise<CreatorAudienceInsights> => {
  const response = await apiClient.get(
    `/creators/${encodeURIComponent(creatorName)}/analytics/audience`,
    { params }
  );
  return response.data;
};

// Clip Analytics API
export const getClipAnalytics = async (
  clipId: string
): Promise<ClipAnalytics> => {
  const response = await apiClient.get(`/clips/${clipId}/analytics`);
  return response.data;
};

export const trackClipView = async (clipId: string): Promise<void> => {
  await apiClient.post(`/clips/${clipId}/track-view`);
};

// User Analytics API
export const getUserStats = async (): Promise<UserAnalytics> => {
  const response = await apiClient.get('/users/me/stats');
  return response.data;
};

// Admin Analytics API
export const getPlatformOverview = async (): Promise<PlatformOverviewMetrics> => {
  const response = await apiClient.get('/admin/analytics/overview');
  return response.data;
};

export const getContentMetrics = async (): Promise<ContentMetrics> => {
  const response = await apiClient.get('/admin/analytics/content');
  return response.data;
};

export const getPlatformTrends = async (
  params?: { metric?: string; days?: number }
): Promise<{ metric: string; days: number; data: TrendDataPoint[] }> => {
  const response = await apiClient.get('/admin/analytics/trends', { params });
  return response.data;
};
