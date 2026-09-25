import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { PropsWithChildren } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { getClipTags } = vi.hoisted(() => ({ getClipTags: vi.fn() }));

vi.mock('../lib/tag-api', () => ({ tagApi: { getClipTags } }));

import { clipTagsQueryKey, useClipTags } from './useTags';

function setup() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: PropsWithChildren) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    return { client, wrapper };
}

describe('useClipTags', () => {
    beforeEach(() => {
        getClipTags.mockReset();
        getClipTags.mockResolvedValue({ tags: [] });
    });

    it('shares one request between every card that shows the same clip', async () => {
        const { wrapper } = setup();
        const first = renderHook(() => useClipTags('clip-1'), { wrapper });
        await waitFor(() => expect(first.result.current.isSuccess).toBe(true));

        const second = renderHook(() => useClipTags('clip-1'), { wrapper });
        const third = renderHook(() => useClipTags('clip-1'), { wrapper });
        await waitFor(() => expect(third.result.current.isSuccess).toBe(true));

        expect(second.result.current.isStale).toBe(false);
        expect(getClipTags).toHaveBeenCalledTimes(1);
    });

    it('is not refetched by feed-wide operations on the clips namespace', async () => {
        const { client, wrapper } = setup();
        const { result } = renderHook(() => useClipTags('clip-1'), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        await client.invalidateQueries({ queryKey: ['clips'] });
        await client.resetQueries({ queryKey: ['clips'] });

        expect(clipTagsQueryKey('clip-1')[0]).not.toBe('clips');
        expect(getClipTags).toHaveBeenCalledTimes(1);
    });

    it('does not request tags without a clip id', () => {
        const { wrapper } = setup();
        renderHook(() => useClipTags(''), { wrapper });
        expect(getClipTags).not.toHaveBeenCalled();
    });
});
