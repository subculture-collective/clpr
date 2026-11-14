import { useCallback, useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { getSecureItem } from '@/lib/secure-storage';
import type { StreamerClipRoomEvent } from '@/types/streamerClipRoom';

interface UseStreamerClipRoomWebSocketOptions {
    roomId?: string;
    channel?: string;
    enabled?: boolean;
    onEvent?: (event: StreamerClipRoomEvent) => void;
}

interface UseStreamerClipRoomWebSocketReturn {
    isConnected: boolean;
    error: string | null;
    lastEvent: StreamerClipRoomEvent | null;
}

const streamerClipRoomQueryKey = ['streamer-clip-room'] as const;
const MAX_RECONNECT_DELAY_MS = 30000;

export function useStreamerClipRoomWebSocket({
    roomId,
    channel,
    enabled = true,
    onEvent,
}: UseStreamerClipRoomWebSocketOptions): UseStreamerClipRoomWebSocketReturn {
    const queryClient = useQueryClient();
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [lastEvent, setLastEvent] = useState<StreamerClipRoomEvent | null>(null);
    const onEventRef = useRef(onEvent);

    useEffect(() => {
        onEventRef.current = onEvent;
    }, [onEvent]);

    const invalidateRoomQueries = useCallback(() => {
        queryClient.invalidateQueries({ queryKey: streamerClipRoomQueryKey });
        if (channel) {
            queryClient.invalidateQueries({ queryKey: [...streamerClipRoomQueryKey, channel] });
        }
        if (roomId) {
            queryClient.invalidateQueries({ queryKey: [...streamerClipRoomQueryKey, roomId] });
        }
    }, [channel, queryClient, roomId]);

    useEffect(() => {
        if (!enabled || !roomId) return;

        // Every value below is owned by this effect run, so a socket or timer
        // from a previous room/mount can never reconnect into the current one.
        let cancelled = false;
        let socket: WebSocket | null = null;
        let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
        let reconnectAttempts = 0;

        // Refetch the room first: it catches up on events missed while
        // disconnected, and the authenticated request renews an expired
        // session cookie before the socket upgrade needs it.
        const reconnect = async () => {
            try {
                await queryClient.refetchQueries({ queryKey: streamerClipRoomQueryKey });
            } catch (err) {
                console.error('Streamer clip room refresh before reconnect failed:', err);
            }
            await connect();
        };

        const connect = async () => {
            if (cancelled) return;
            let token: string | null = null;
            try {
                token = await getSecureItem('token');
            } catch (err) {
                console.error('Streamer clip room token lookup failed:', err);
            }
            // The effect may have been cleaned up while awaiting storage.
            if (cancelled) return;

            let ws: WebSocket;
            try {
                const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
                const wsHost = import.meta.env.VITE_WS_HOST || window.location.host;
                const wsUrl = `${wsProtocol}//${wsHost}/api/v1/streamer-clip-rooms/${encodeURIComponent(roomId)}/ws`;
                // Web sessions authenticate with the HttpOnly access_token cookie;
                // a stored bearer token (if any) is passed as a subprotocol, which
                // the backend echoes back.
                ws = token
                    ? new WebSocket(wsUrl, [`auth.bearer.${btoa(token)}`])
                    : new WebSocket(wsUrl);
            } catch (err) {
                setError('Failed to connect to streamer clip room');
                console.error('Streamer clip room websocket connection error:', err);
                return;
            }
            socket = ws;

            ws.onopen = () => {
                if (cancelled) return;
                setIsConnected(true);
                setError(null);
                reconnectAttempts = 0;
            };

            ws.onmessage = event => {
                if (cancelled) return;
                try {
                    const parsed = JSON.parse(event.data) as StreamerClipRoomEvent;
                    setLastEvent(parsed);
                    onEventRef.current?.(parsed);
                    invalidateRoomQueries();
                } catch (err) {
                    console.error('Error parsing streamer clip room websocket message:', err);
                }
            };

            ws.onerror = event => {
                if (cancelled) return;
                console.error('Streamer clip room websocket error:', event);
                setError('Connection error');
            };

            ws.onclose = () => {
                if (cancelled || socket !== ws) return;
                socket = null;
                setIsConnected(false);

                // Keep retrying for as long as the page is open; the delay
                // grows to a cap so a long outage still recovers on its own.
                const backoffDelay = Math.min(
                    1000 * Math.pow(2, reconnectAttempts),
                    MAX_RECONNECT_DELAY_MS,
                );
                reconnectAttempts += 1;
                reconnectTimer = setTimeout(() => {
                    void reconnect();
                }, backoffDelay);
            };
        };

        queueMicrotask(() => {
            void connect();
        });

        return () => {
            cancelled = true;
            if (reconnectTimer) clearTimeout(reconnectTimer);
            if (socket) {
                const ws = socket;
                socket = null;
                ws.onopen = null;
                ws.onmessage = null;
                ws.onerror = null;
                ws.onclose = null;
                ws.close();
            }
            setIsConnected(false);
        };
    }, [enabled, invalidateRoomQueries, queryClient, roomId]);

    return {
        isConnected,
        error,
        lastEvent,
    };
}
