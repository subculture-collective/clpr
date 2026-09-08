import { createElement, type ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, expect, it, vi } from 'vitest';
import * as api from '@/lib/clip-api';
import { useClipFeed } from './useClipFeed';
import type { ClipFeedFilters, ClipFeedResponse } from '@/types/clip';

vi.mock('@/lib/clip-api', () => ({ fetchClips: vi.fn() }));
const filters: ClipFeedFilters = { sort: 'trending', timeframe: 'week' };
const page = (generation: string, cursor?: string): ClipFeedResponse => ({
    clips: [], total: 0, page: 1, cursor, has_more: !!cursor,
    engagement: { generation, period: 'week', partial_coverage: true, estimated: true },
});
const wrapper = () => {
    const client = new QueryClient({ defaultOptions: { queries: { gcTime: Infinity } } });
    return ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client }, children);
};
beforeEach(() => vi.resetAllMocks());

it('restarts an expired ranking from page one without mixing generations or losing filters', async () => {
    vi.mocked(api.fetchClips)
        .mockResolvedValueOnce(page('old', 'old-cursor'))
        .mockRejectedValueOnce({ isAxiosError: true, response: { status: 409, data: { code: 'RANKING_REFRESH_REQUIRED' } } })
        .mockResolvedValueOnce(page('new'));
    const { result } = renderHook(() => useClipFeed(filters), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.hasNextPage).toBe(true));
    await act(async () => { await result.current.fetchNextPage(); });
    await waitFor(() => expect(result.current.data?.pages[0]?.engagement?.generation).toBe('new'));
    expect(result.current.data?.pages).toHaveLength(1);
    expect(result.current.hasNextPage).toBe(false);
    expect(api.fetchClips).toHaveBeenNthCalledWith(2, { cursor: 'old-cursor', filters });
    expect(api.fetchClips).toHaveBeenNthCalledWith(3, { cursor: undefined, filters });
});

it('keeps pages of a retained ranking and preserves partial-history metadata', async () => {
    vi.mocked(api.fetchClips).mockResolvedValueOnce(page('same', 'next')).mockResolvedValueOnce(page('same'));
    const { result } = renderHook(() => useClipFeed(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.hasNextPage).toBe(true));
    await act(async () => { await result.current.fetchNextPage(); });
    await waitFor(() => expect(result.current.data?.pages).toHaveLength(2));
    expect(result.current.data?.pages.every(p => p.engagement?.partial_coverage)).toBe(true);
});

it('does not reset or retry unrelated conflicts', async () => {
    vi.mocked(api.fetchClips).mockRejectedValue({ isAxiosError: true, response: { status: 409, data: { code: 'OTHER' } } });
    const { result } = renderHook(() => useClipFeed(filters), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(api.fetchClips).toHaveBeenCalledTimes(1);
});
