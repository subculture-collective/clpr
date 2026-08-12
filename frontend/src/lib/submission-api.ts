import apiClient from './api';
import type {
  SubmitClipRequest,
  SubmissionResponse,
  SubmissionListResponse,
  SubmissionStatsResponse,
  ModerationQueueResponse,
} from '../types/submission';
import type { Clip } from '../types/clip';

/**
 * Response from checking clip status
 */
export interface ClipStatusResponse {
  success: boolean;
  exists: boolean;
  can_be_claimed: boolean;
  clip?: Clip;
}

/**
 * Response from fetching clip metadata from Twitch
 */
export interface ClipMetadata {
  clip_id: string;
  title: string;
  streamer_name: string;
  game_name: string;
  view_count: number;
  created_at: string;
  thumbnail_url: string;
  duration: number;
  url: string;
}

/**
 * Submit a clip for moderation
 */
export async function submitClip(
  request: SubmitClipRequest
): Promise<SubmissionResponse> {
  const response = await apiClient.post<SubmissionResponse>(
    '/submissions',
    request
  );
  return response.data;
}

/**
 * Check if a clip exists and whether it can be claimed
 */
export async function checkClipStatus(clipId: string): Promise<ClipStatusResponse> {
  const response = await apiClient.get<ClipStatusResponse>(
    `/submissions/check/${clipId}`
  );
  return response.data;
}

/**
 * Get clip metadata from Twitch
 */
export async function getClipMetadata(clipUrl: string): Promise<ClipMetadata> {
  const response = await apiClient.get<{ success: boolean; data: ClipMetadata }>(
    '/submissions/metadata',
    {
      params: { url: clipUrl },
    }
  );
  return response.data.data;
}

/**
 * Get user's submissions
 */
export async function getUserSubmissions(
  page = 1,
  limit = 20
): Promise<SubmissionListResponse> {
  const response = await apiClient.get<SubmissionListResponse>(
    '/submissions',
    {
      params: { page, limit },
    }
  );
  const payload = response.data;
  return {
    success: payload?.success ?? true,
    data: Array.isArray(payload?.data) ? payload.data : [],
    meta: {
      page: payload?.meta?.page ?? page,
      limit: payload?.meta?.limit ?? limit,
      total: payload?.meta?.total ?? 0,
      total_pages: payload?.meta?.total_pages ?? 0,
    },
  };
}

/**
 * Get submission statistics for current user
 */
export async function getSubmissionStats(): Promise<SubmissionStatsResponse> {
  const response = await apiClient.get<SubmissionStatsResponse>(
    '/submissions/stats'
  );
  return response.data;
}

/**
 * Get pending submissions for moderation (admin/moderator only)
 */
export async function getPendingSubmissions(
  page = 1,
  limit = 20
): Promise<ModerationQueueResponse> {
  const response = await apiClient.get<ModerationQueueResponse>(
    '/admin/submissions',
    {
      params: { page, limit },
    }
  );
  return response.data;
}

/**
 * Approve a submission (admin/moderator only)
 */
export async function approveSubmission(submissionId: string): Promise<void> {
  await apiClient.post(`/admin/submissions/${submissionId}/approve`);
}

/**
 * Reject a submission (admin/moderator only)
 */
export async function rejectSubmission(
  submissionId: string,
  reason: string
): Promise<void> {
  await apiClient.post(`/admin/submissions/${submissionId}/reject`, {
    reason,
  });
}

/**
 * Response from bulk operations
 */
export interface BulkOperationResponse {
  success: boolean;
  message: string;
  count: number;
}

/**
 * Bulk approve submissions (admin/moderator only)
 */
export async function bulkApproveSubmissions(
  submissionIds: string[]
): Promise<BulkOperationResponse> {
  const response = await apiClient.post<BulkOperationResponse>(
    '/admin/submissions/bulk-approve',
    {
      submission_ids: submissionIds,
    }
  );
  return response.data;
}

/**
 * Bulk reject submissions (admin/moderator only)
 */
export async function bulkRejectSubmissions(
  submissionIds: string[],
  reason: string
): Promise<BulkOperationResponse> {
  const response = await apiClient.post<BulkOperationResponse>(
    '/admin/submissions/bulk-reject',
    {
      submission_ids: submissionIds,
      reason,
    }
  );
  return response.data;
}
