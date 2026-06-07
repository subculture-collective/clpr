import type { UserRole } from '../lib/roles';

export type ClipSourceType = 'twitch' | 'external' | 'upload';
export type ClipSourcePlatform =
  | 'twitch'
  | 'kick'
  | 'youtube'
  | 'youtube_shorts'
  | 'tiktok'
  | 'upload';
export type ClipUploadStatus = 'none' | 'pending' | 'uploaded' | 'validated' | 'rejected';
export type ClipStorageVisibility = 'private' | 'public';

export interface ClipSourceFields {
  source_type?: ClipSourceType;
  source_platform?: ClipSourcePlatform;
  source_url?: string;
  source_id?: string;
  source_metadata?: Record<string, unknown>;
  duration_seconds?: number;
  duration_verified?: boolean;
  storage_provider?: string;
  storage_bucket?: string;
  storage_key?: string;
  original_filename?: string;
  mime_type?: string;
  file_size_bytes?: number;
  upload_status?: ClipUploadStatus;
  duration_validation_error?: string;
  storage_visibility?: ClipStorageVisibility;
}

export interface ClipSubmission extends ClipSourceFields {
  id: string;
  user_id: string;
  clip_id?: string; // Set when submission is approved
  twitch_clip_id: string;
  twitch_clip_url: string;
  title?: string;
  custom_title?: string;
  tags?: string[];
  is_nsfw: boolean;
  submission_reason?: string;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
  // Metadata from Twitch
  creator_name?: string;
  creator_id?: string;
  broadcaster_name?: string;
  broadcaster_id?: string;
  broadcaster_name_override?: string;
  game_id?: string;
  game_name?: string;
  thumbnail_url?: string;
  duration?: number;
  view_count: number;
}

export interface ClipSubmissionWithUser extends ClipSubmission {
  user?: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    karma_points: number;
    role: UserRole;
  };
}

export interface SubmitClipRequest extends Partial<ClipSourceFields> {
  clip_url: string;
  custom_title?: string;
  tags?: string[];
  is_nsfw: boolean;
  submission_reason?: string;
  broadcaster_name_override?: string;
}

export interface SubmissionStats {
  user_id: string;
  total_submissions: number;
  approved_count: number;
  rejected_count: number;
  pending_count: number;
  approval_rate: number;
}

export interface SubmissionResponse {
  success: boolean;
  message: string;
  submission: ClipSubmission;
}

export interface SubmissionListResponse {
  success: boolean;
  data: ClipSubmission[];
  meta: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface ModerationQueueResponse {
  success: boolean;
  data: ClipSubmissionWithUser[] | null;
  meta: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface SubmissionStatsResponse {
  success: boolean;
  data: SubmissionStats;
}

export interface ApprovalRequest {
  id: string;
}

export interface RejectionRequest {
  id: string;
  reason: string;
}

export interface RateLimitErrorResponse {
  error: 'rate_limit_exceeded';
  limit: number;
  window: number;
  retry_after: number;
}
