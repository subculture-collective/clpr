import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { FakeWebSocket, installFakeWebSocket } from '@/test/fake-websocket';
import { useChatWebSocket } from './useChatWebSocket';

const CHANNEL_A = '11111111-1111-4111-8111-111111111111';
const CHANNEL_B = '22222222-2222-4222-8222-222222222222';

async function flushMicrotasks() {
    await act(async () => {
        await Promise.resolve();
    });
}

function messageFrame(id: string, content: string) {
    return {
        type: 'message',
        channel_id: CHANNEL_A,
        user_id: 'u1',
        username: 'alice',
        display_name: 'Alice',
        content,
        message_id: id,
        timestamp: '2026-11-02T00:00:00Z',
    };
}

describe('useChatWebSocket', () => {
    beforeEach(() => {
        vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
        installFakeWebSocket();
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.unstubAllGlobals();
    });

    it('connects to the backend chat route without a subprotocol', async () => {
        renderHook(() => useChatWebSocket({ channelId: CHANNEL_A }));
        await flushMicrotasks();

        expect(FakeWebSocket.instances).toHaveLength(1);
        const ws = FakeWebSocket.instances[0];
        expect(ws.url).toBe(`ws://${window.location.host}/api/v1/chat/channels/${CHANNEL_A}/ws`);
        expect(ws.protocols).toBeUndefined();
    });

    it('encodes the channel id in the URL', async () => {
        renderHook(() => useChatWebSocket({ channelId: 'a/b?c' }));
        await flushMicrotasks();
        expect(FakeWebSocket.instances[0].url).toContain('/api/v1/chat/channels/a%2Fb%3Fc/ws');
    });

    it('parses flat server frames, dedupes replayed history and sends channel_id', async () => {
        const onMessage = vi.fn();
        const onTyping = vi.fn();
        const { result } = renderHook(() =>
            useChatWebSocket({ channelId: CHANNEL_A, onMessage, onTyping }),
        );
        await flushMicrotasks();
        const ws = FakeWebSocket.instances[0];

        act(() => {
            ws.serverOpen();
            ws.serverSend(messageFrame('m1', 'hello'));
            ws.serverSend(messageFrame('m1', 'hello'));
            ws.serverSend({ type: 'typing', channel_id: CHANNEL_A, user_id: 'u2', username: 'bob' });
            ws.serverSend({ type: 'presence', channel_id: CHANNEL_A, presence_type: 'joined' });
        });

        expect(result.current.isConnected).toBe(true);
        expect(result.current.messages).toHaveLength(1);
        expect(result.current.messages[0]).toMatchObject({
            id: 'm1',
            user_id: 'u1',
            username: 'alice',
            content: 'hello',
            created_at: '2026-11-02T00:00:00Z',
        });
        expect(onMessage).toHaveBeenCalledTimes(1);
        expect(onTyping).toHaveBeenCalledWith('bob');

        act(() => {
            result.current.sendMessage('hi');
        });
        const sent = JSON.parse(ws.sent[0]);
        expect(sent).toMatchObject({ type: 'message', channel_id: CHANNEL_A, content: 'hi' });
    });

    it('reconnects after a server close while mounted', async () => {
        renderHook(() => useChatWebSocket({ channelId: CHANNEL_A }));
        await flushMicrotasks();

        act(() => {
            FakeWebSocket.instances[0].serverClose();
        });
        act(() => {
            vi.advanceTimersByTime(1000);
        });
        expect(FakeWebSocket.instances).toHaveLength(2);
    });

    it('does not reconnect after unmount', async () => {
        const { unmount } = renderHook(() => useChatWebSocket({ channelId: CHANNEL_A }));
        await flushMicrotasks();
        const ws = FakeWebSocket.instances[0];

        unmount();
        expect(ws.closeCalls).toBe(1);
        await flushMicrotasks();

        act(() => {
            vi.advanceTimersByTime(60_000);
        });
        expect(FakeWebSocket.instances).toHaveLength(1);
    });

    it('does not reconnect when unmounted during a pending backoff', async () => {
        const { unmount } = renderHook(() => useChatWebSocket({ channelId: CHANNEL_A }));
        await flushMicrotasks();
        act(() => {
            FakeWebSocket.instances[0].serverClose();
        });

        unmount();
        act(() => {
            vi.advanceTimersByTime(60_000);
        });
        expect(FakeWebSocket.instances).toHaveLength(1);
    });

    it('keeps exactly one live socket across a channel change', async () => {
        const { rerender } = renderHook(({ channelId }) => useChatWebSocket({ channelId }), {
            initialProps: { channelId: CHANNEL_A },
        });
        await flushMicrotasks();

        rerender({ channelId: CHANNEL_B });
        await flushMicrotasks();

        expect(FakeWebSocket.instances).toHaveLength(2);
        const [first, second] = FakeWebSocket.instances;
        expect(first.closeCalls).toBe(1);
        expect(second.url).toContain(`/api/v1/chat/channels/${CHANNEL_B}/ws`);

        act(() => {
            vi.advanceTimersByTime(60_000);
        });
        expect(FakeWebSocket.instances).toHaveLength(2);
        expect(second.closeCalls).toBe(0);
    });
});
