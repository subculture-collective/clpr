export interface Badge {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: 'achievement' | 'staff' | 'special' | 'supporter';
  requirement?: string;
}

export interface UserBadge {
  id: string;
  badge_id: string;
  awarded_at: string;
  awarded_by?: string;
  name: string;
  description: string;
  icon: string;
  category: string;
}

export interface KarmaHistory {
  id: string;
  user_id: string;
  amount: number;
  source: string;
  source_id?: string;
  created_at: string;
}

export interface KarmaBreakdown {
  clip_karma: number;
  comment_karma: number;
  total_karma: number;
}

export interface UserStats {
  user_id: string;
  trust_score: number;
  engagement_score: number;
  total_comments: number;
  total_votes_cast: number;
  total_clips_submitted: number;
  correct_reports: number;
  incorrect_reports: number;
  days_active: number;
  last_active_date?: string;
  updated_at: string;
}

export interface UserReputation {
  user_id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  karma_points: number;
  rank: string;
  trust_score: number;
  engagement_score: number;
  badges: UserBadge[] | null;
  stats?: UserStats | null;
  created_at: string;
}

export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  score: number;
  user_rank: string;
  total_comments?: number;
  total_votes_cast?: number;
  total_clips_submitted?: number;
}

export type LeaderboardType = 'karma' | 'engagement' | 'streamers';

export interface LeaderboardResponse {
  type: LeaderboardType;
  page: number;
  limit: number;
  entries: LeaderboardEntry[];
}
