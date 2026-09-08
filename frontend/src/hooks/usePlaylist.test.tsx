import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { allowTestConsole } from '../test/setup';
import * as playlists from './usePlaylist';

const api = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn(), patch: vi.fn(), put: vi.fn(), delete: vi.fn() }));
vi.mock('@/lib/api', () => ({ default: api }));

const id = '00000000-0000-4000-8000-000000000001';
function harness() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    function wrapper({ children }: { children: ReactNode }) {
        return createElement(QueryClientProvider, { client }, children);
    }
    return { client, wrapper };
}
beforeEach(() => {
    for (const method of Object.values(api)) method.mockReset().mockResolvedValue({ data: { data: { id } } });
});

describe('playlist reads', () => {
    it.each([
        { value: id, path: '/playlists/' + id },
        { value: 'shared-link', path: '/playlists/share/shared-link' },
    ])('resolves $value using the right access boundary', async ({ value, path }) => {
        const { wrapper } = harness();
        const { result } = renderHook(() => playlists.usePlaylist(value), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(api.get).toHaveBeenCalledWith(path, { params: { page: 1, limit: 20 } });
    });

    it.each([
        { name: 'private', useRead: playlists.usePlaylists, path: '/playlists' },
        { name: 'public', useRead: playlists.usePublicPlaylists, path: '/playlists/public' },
        { name: 'featured', useRead: playlists.useFeaturedPlaylists, path: '/playlists/featured' },
    ])('keeps $name lists in their own cache and endpoint', async ({ useRead, path }) => {
        const { wrapper } = harness();
        const { result } = renderHook(() => useRead(2, 10), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(api.get).toHaveBeenCalledWith(path, { params: { page: 2, limit: 10 } });
    });

    it('does not fetch an absent playlist or disabled account list', () => {
        const { wrapper } = harness();
        renderHook(() => {
            playlists.usePlaylist('');
            playlists.useInfinitePlaylist('');
            playlists.usePlaylists(1, 20, false);
        }, { wrapper });
        expect(api.get).not.toHaveBeenCalled();
    });

    it.each([
        { value: 'share-token', limit: 500, expectedLimit: 100, path: '/playlists/share/share-token' },
        { value: id, limit: -1, expectedLimit: 1, path: '/playlists/' + id },
    ])('paginates $value within backend bounds and stops at the last page', async ({ value, limit, expectedLimit, path }) => {
        api.get.mockImplementation(async (_path, { params }) => ({
            data: { data: { id, clips: [{ id: 'clip-' + params.page }] }, meta: { page: params.page, has_next: params.page === 1 } },
        }));
        const { wrapper } = harness();
        const { result } = renderHook(() => playlists.useInfinitePlaylist(value, limit), { wrapper });
        await waitFor(() => expect(result.current.hasNextPage).toBe(true));
        await act(async () => { await result.current.fetchNextPage(); });
        await waitFor(() => expect(result.current.hasNextPage).toBe(false));
        expect(api.get.mock.calls).toEqual([
            [path, { params: { page: 1, limit: expectedLimit } }],
            [path, { params: { page: 2, limit: expectedLimit } }],
        ]);
        expect(result.current.data?.pages.flatMap(page => page.data.clips)).toEqual([{ id: 'clip-1' }, { id: 'clip-2' }]);
    });

    it('continues featured lists using server pagination metadata', async () => {
        api.get.mockImplementation(async (_path, { params }) => ({ data: { data: [], meta: { page: params.page, has_next: params.page === 1 } } }));
        const { wrapper } = harness();
        const { result } = renderHook(() => playlists.useInfiniteFeaturedPlaylists(10), { wrapper });
        await waitFor(() => expect(result.current.hasNextPage).toBe(true));
        await act(async () => { await result.current.fetchNextPage(); });
        await waitFor(() => expect(result.current.hasNextPage).toBe(false));
        expect(api.get).toHaveBeenLastCalledWith('/playlists/featured', { params: { page: 2, limit: 10 } });
    });
});

const edits = [
    { name: 'create', detail: false, list: true, useAction: () => {
        const action = playlists.useCreatePlaylist();
        return () => action.mutateAsync({ title: 'Saved clips', visibility: 'private' });
    } },
    { name: 'update', detail: true, list: true, useAction: () => {
        const action = playlists.useUpdatePlaylist();
        return () => action.mutateAsync({ id, data: { title: 'Renamed' } });
    } },
    { name: 'copy', detail: false, list: true, useAction: () => {
        const action = playlists.useCopyPlaylist();
        return () => action.mutateAsync({ id, data: { title: 'Copy' } });
    } },
    { name: 'delete', detail: false, list: true, useAction: () => {
        const action = playlists.useDeletePlaylist();
        return () => action.mutateAsync(id);
    } },
    { name: 'add clips', detail: true, list: false, useAction: () => {
        const action = playlists.useAddClipsToPlaylist();
        return () => action.mutateAsync({ id, data: { clip_ids: ['clip'] } });
    } },
    { name: 'remove clip', detail: true, list: false, useAction: () => {
        const action = playlists.useRemoveClipFromPlaylist();
        return () => action.mutateAsync({ playlistId: id, clipId: 'clip' });
    } },
    { name: 'reorder', detail: true, list: false, useAction: () => {
        const action = playlists.useReorderPlaylistClips();
        return () => action.mutateAsync({ id, data: { clip_ids: ['clip-2', 'clip-1'] } });
    } },
    ...[
        { name: 'like', useMutation: playlists.useLikePlaylist },
        { name: 'unlike', useMutation: playlists.useUnlikePlaylist },
        { name: 'bookmark', useMutation: playlists.useBookmarkPlaylist },
        { name: 'unbookmark', useMutation: playlists.useUnbookmarkPlaylist },
    ].map(({ name, useMutation }) => ({ name, detail: true, list: true, useAction: () => {
        const action = useMutation();
        return () => action.mutateAsync(id);
    } })),
];

describe('playlist mutations', () => {
    it.each(edits)('$name revalidates affected views without touching another playlist', async ({ useAction, detail, list }) => {
        const { client, wrapper } = harness();
        const keys = [['playlist', id], ['playlist', 'unrelated'], ['playlists', 1, 20]];
        keys.forEach(key => client.setQueryData(key, { title: 'Before' }));
        const { result } = renderHook(useAction, { wrapper });
        await act(async () => { await result.current(); });
        expect(client.getQueryState(keys[0])?.isInvalidated).toBe(detail);
        expect(client.getQueryState(keys[1])?.isInvalidated).toBe(false);
        expect(client.getQueryState(keys[2])?.isInvalidated).toBe(list);
    });

    it.each(edits)('$name preserves cached data and surfaces a failed write', async ({ useAction }) => {
        const failure = new Error('Write rejected');
        for (const method of [api.post, api.patch, api.put, api.delete]) method.mockRejectedValue(failure);
        const { client, wrapper } = harness();
        const key = ['playlist', id];
        client.setQueryData(key, { title: 'Before' });
        const { result } = renderHook(useAction, { wrapper });
        allowTestConsole('error', /Failed to/);
        await act(async () => { await expect(result.current()).rejects.toThrow('Write rejected'); });
        expect(client.getQueryData(key)).toEqual({ title: 'Before' });
        expect(client.getQueryState(key)?.isInvalidated).toBe(false);
    });
});
