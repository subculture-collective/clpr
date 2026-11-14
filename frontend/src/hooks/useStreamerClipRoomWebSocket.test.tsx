import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { FakeWebSocket, installFakeWebSocket } from '@/test/fake-websocket';

const getSecureItem = vi.fn<(key: string) => Promise<string | null>>();
vi.mock('@/lib/secure-storage', () => ({
    getSecureItem: (key: string) => getSecureItem(key),
}));

import { useStreamerClipRoomWebSocket } from './useStreamerClipRoomWebSocket';

const ROOM_A = '11111111-1111-4111-8111-111111111111';
const ROOM_B = '22222222-2222-4222-8222-222222222222';

function wrapper({ children }: { children: ReactNode }) {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

async function flushAsync() {
    await act(async () => {
        for (let i = 0; i < 5; i += 1) await Promise.resolve();
    });
}

describe('useStreamerClipRoomWebSocket', () => {
    beforeEach(() => {
        vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
        installFakeWebSocket();
        getSecureItem.mockReset();
        getSecureItem.mockResolvedValue('tok');
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.unstubAllGlobals();
    });

    it('connects to the room route with the bearer subprotocol', async () => {
        renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), { wrapper });
        await flushAsync();

        expect(FakeWebSocket.instances).toHaveLength(1);
        const ws = FakeWebSocket.instances[0];
        expect(ws.url).toBe(`ws://${window.location.host}/api/v1/streamer-clip-rooms/${ROOM_A}/ws`);
        expect(ws.protocols).toEqual([`auth.bearer.${btoa('tok')}`]);
    });

    it('falls back to cookie auth when no stored token exists', async () => {
        getSecureItem.mockResolvedValue(null);
        const { result } = renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), {
            wrapper,
        });
        await flushAsync();

        expect(FakeWebSocket.instances).toHaveLength(1);
        expect(FakeWebSocket.instances[0].protocols).toBeUndefined();
        expect(result.current.error).toBeNull();
    });

    it('does not open a socket when unmounted while the token lookup is pending', async () => {
        let resolveToken: (value: string | null) => void = () => {};
        getSecureItem.mockReturnValue(
            new Promise(resolve => {
                resolveToken = resolve;
            }),
        );
        const { unmount } = renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), {
            wrapper,
        });
        await flushAsync();
        unmount();

        await act(async () => {
            resolveToken('tok');
            await Promise.resolve();
        });
        expect(FakeWebSocket.instances).toHaveLength(0);
    });

    it('does not reconnect after unmount', async () => {
        const { unmount } = renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), {
            wrapper,
        });
        await flushAsync();
        const ws = FakeWebSocket.instances[0];

        unmount();
        expect(ws.closeCalls).toBe(1);
        await act(async () => {
            vi.advanceTimersByTime(60_000);
        });
        await flushAsync();
        expect(FakeWebSocket.instances).toHaveLength(1);
    });

    it('keeps exactly one live socket across a room change', async () => {
        const { rerender } = renderHook(
            ({ roomId }) => useStreamerClipRoomWebSocket({ roomId }),
            { wrapper, initialProps: { roomId: ROOM_A } },
        );
        await flushAsync();

        rerender({ roomId: ROOM_B });
        await flushAsync();

        expect(FakeWebSocket.instances).toHaveLength(2);
        const [first, second] = FakeWebSocket.instances;
        expect(first.closeCalls).toBe(1);
        expect(second.url).toContain(`/streamer-clip-rooms/${ROOM_B}/ws`);

        await act(async () => {
            vi.advanceTimersByTime(60_000);
        });
        await flushAsync();
        expect(FakeWebSocket.instances).toHaveLength(2);
    });

    it('reconnects after a server close while mounted', async () => {
        renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), { wrapper });
        await flushAsync();

        act(() => {
            FakeWebSocket.instances[0].serverClose();
        });
        await act(async () => {
            vi.advanceTimersByTime(1000);
        });
        await flushAsync();
        expect(FakeWebSocket.instances).toHaveLength(2);
    });

    it('keeps reconnecting at the capped delay after many failures', async () => {
        renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), { wrapper });
        await flushAsync();

        // Fail well past the old 10-attempt limit; 30 s always covers the capped delay.
        for (let attempt = 0; attempt < 15; attempt += 1) {
            act(() => {
                FakeWebSocket.instances[FakeWebSocket.instances.length - 1].serverClose();
            });
            await act(async () => {
                vi.advanceTimersByTime(30_000);
            });
            await flushAsync();
        }

        expect(FakeWebSocket.instances).toHaveLength(16);
    });

    it('refreshes the room query before reconnecting', async () => {
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        const refetch = vi.spyOn(client, 'refetchQueries').mockResolvedValue(undefined);
        renderHook(() => useStreamerClipRoomWebSocket({ roomId: ROOM_A }), {
            wrapper: ({ children }: { children: ReactNode }) => (
                <QueryClientProvider client={client}>{children}</QueryClientProvider>
            ),
        });
        await flushAsync();
        expect(refetch).not.toHaveBeenCalled();

        act(() => {
            FakeWebSocket.instances[0].serverClose();
        });
        await act(async () => {
            vi.advanceTimersByTime(1000);
        });
        await flushAsync();

        expect(refetch).toHaveBeenCalledWith({ queryKey: ['streamer-clip-room'] });
        expect(FakeWebSocket.instances).toHaveLength(2);
    });
});
