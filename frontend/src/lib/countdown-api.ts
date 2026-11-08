import { apiClient } from './api';

export interface PublicCountdownClip {
  position: number;
  clip_id: string;
  title: string;
  broadcaster_name: string;
  broadcaster_id?: string;
  twitch_clip_id: string;
  twitch_clip_url: string;
  thumbnail_url?: string;
  duration?: number;
}

export interface PublicCountdownEdition {
  edition_date: string;
  series_slug: string;
  series_name: string;
  size: number;
  clips: PublicCountdownClip[];
}

export interface CountdownFeature {
  edition_date: string;
  series_slug: string;
  series_name: string;
  position: number;
  clip_id: string;
  title: string;
}

export interface CountdownPage {
  editions: PublicCountdownEdition[];
  nextBefore: string | null;
}

export const countdownApi = {
  list: async (params: { before?: string; series?: string; days?: number } = {}): Promise<CountdownPage> => {
    const response = await apiClient.get<{
      data: PublicCountdownEdition[];
      meta: { next_before: string | null };
    }>('/countdowns', { params });
    return { editions: response.data.data ?? [], nextBefore: response.data.meta?.next_before ?? null };
  },

  get: async (date: string, series: string): Promise<PublicCountdownEdition> => {
    const response = await apiClient.get<{ data: PublicCountdownEdition }>(
      `/countdowns/${encodeURIComponent(date)}/${encodeURIComponent(series)}`,
    );
    return response.data.data;
  },

  featured: async (params: { clip_id: string } | { broadcaster_id: string }): Promise<CountdownFeature[]> => {
    const response = await apiClient.get<{ data: CountdownFeature[] }>('/countdowns/featured', { params });
    return response.data.data ?? [];
  },
};

export function countdownPath(edition: { edition_date: string; series_slug: string }): string {
  return `/countdowns/${edition.edition_date}/${edition.series_slug}`;
}

/** Formats a YYYY-MM-DD edition date without shifting it into another day. */
export function formatCountdownDate(date: string, month: 'short' | 'long' = 'long'): string {
  const [year, monthIndex, day] = date.split('-').map(Number);
  return new Date(Date.UTC(year, monthIndex - 1, day)).toLocaleDateString('en-US', {
    year: 'numeric',
    month,
    day: 'numeric',
    timeZone: 'UTC',
  });
}
