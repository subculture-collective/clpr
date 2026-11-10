import { apiClient } from './api';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

/** A broadcaster's countdown controls (#75). */
export interface CountdownControls {
  broadcaster_id: string;
  excluded: boolean;
  exclusion_source?: 'admin' | 'self_service';
  excluded_at?: string;
  featured: boolean;
  featured_at?: string;
  clip_download_authorized: boolean;
  consent_text_version: string;
}

export interface CountdownExclusion {
  id: string;
  broadcaster_id?: string;
  clip_id?: string;
  reason: 'request' | 'dmca' | 'takedown';
  source: 'admin' | 'self_service';
  note?: string;
  created_at: string;
}

export interface CountdownRemovalTask {
  render_id: string;
  kind: 'countdown' | 'short';
  youtube_video_id: string;
  edition_date: string;
  series_slug: string;
  clip_id: string;
  clip_title: string;
  broadcaster_name: string;
  broadcaster_id?: string;
  exclusion_reason: CountdownExclusion['reason'];
  exclusion_source: CountdownExclusion['source'];
  excluded_at: string;
  due_at: string;
}

/** A countdown video or Short viewers can open on YouTube. */
export interface CountdownVideo {
  youtube_video_id: string;
  view_count?: number;
  views_updated_at?: string;
}

/** One of the broadcaster's clips in a public countdown (#84). */
export interface CountdownPlacement {
  edition_date: string;
  series_slug: string;
  series_name: string;
  position: number;
  clip_id: string;
  title: string;
  thumbnail_url?: string;
  countdown?: CountdownVideo;
  short?: CountdownVideo;
}

export const countdownControlsApi = {
  get: async (): Promise<CountdownControls> =>
    (await apiClient.get<{ data: CountdownControls }>('/users/me/countdowns')).data.data,

  placements: async (): Promise<CountdownPlacement[]> =>
    (await apiClient.get<{ data: CountdownPlacement[] }>('/users/me/countdowns/placements')).data.data ?? [],

  setExcluded: async (excluded: boolean): Promise<CountdownControls> =>
    (await apiClient.put<{ data: CountdownControls }>('/users/me/countdowns/exclusion', { excluded })).data.data,

  stopFeaturing: async (): Promise<CountdownControls> =>
    (await apiClient.delete<{ data: CountdownControls }>('/users/me/countdowns/feature')).data.data,

  listExclusions: async (): Promise<CountdownExclusion[]> =>
    (await apiClient.get<{ data: CountdownExclusion[] }>('/admin/countdowns/exclusions')).data.data ?? [],

  createExclusion: async (input: { broadcaster_id?: string; clip_id?: string; reason: CountdownExclusion['reason']; note?: string }) =>
    (await apiClient.post<{ data: CountdownExclusion }>('/admin/countdowns/exclusions', input)).data.data,

  deleteExclusion: async (id: string) => {
    await apiClient.delete(`/admin/countdowns/exclusions/${encodeURIComponent(id)}`);
  },

  listRemovals: async (): Promise<CountdownRemovalTask[]> =>
    (await apiClient.get<{ data: CountdownRemovalTask[] }>('/admin/countdowns/removals')).data.data ?? [],

  resolveRemoval: async (input: { render_id: string; clip_id: string; resolution: 'removed' | 'edited' | 'kept'; note?: string }) => {
    await apiClient.post('/admin/countdowns/removals', input);
  },
};

export function youTubeVideoUrl(video: CountdownVideo, kind: 'countdown' | 'short'): string {
  const id = encodeURIComponent(video.youtube_video_id);
  return kind === 'short' ? `https://www.youtube.com/shorts/${id}` : `https://www.youtube.com/watch?v=${id}`;
}

/** Where "Feature my clips" sends the broadcaster to grant channel:manage:clips. */
export function featureMyClipsUrl(returnTo = '/settings'): string {
  return `${API_BASE_URL}/twitch/oauth/authorize?purpose=feature&return_to=${encodeURIComponent(returnTo)}`;
}
