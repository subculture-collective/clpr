import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));
vi.mock('@/lib/api', () => ({ default: { get } }));

const { useInfinitePlaylist } = await import('./usePlaylist');

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return createElement(QueryClientProvider, { client }, children);
}

describe('useInfinitePlaylist', () => {
  beforeEach(() => get.mockReset());

  it('never exceeds the backend page-size contract and exposes continuation', async () => {
    get.mockResolvedValue({
      data: {
        success: true,
        data: { id: 'playlist-token', clips: [] },
        meta: { page: 1, limit: 100, total: 101, total_pages: 2, has_next: true, has_prev: false },
      },
    });

    const { result } = renderHook(() => useInfinitePlaylist('playlist-token'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(get).toHaveBeenCalledWith('/playlists/share/playlist-token', {
      params: { page: 1, limit: 100 },
    });
    expect(result.current.hasNextPage).toBe(true);
  });
});
