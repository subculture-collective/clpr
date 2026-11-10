import { apiClient } from './api';

/** A broadcaster's weekly clip pack (#86). */
export interface ClipPack {
  id: string;
  broadcaster_id: string;
  broadcaster_name: string;
  week_start: string;
  status: 'pending' | 'building' | 'ready' | 'failed' | 'expired';
  bytes?: number;
  clip_count?: number;
  ready_at?: string;
  expires_at?: string;
  download_url?: string;
}

export const clipPackApi = {
  mine: async (): Promise<ClipPack[]> =>
    (await apiClient.get<{ data: ClipPack[] }>('/users/me/clip-packs')).data.data ?? [],
};
