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
const MAX_RECONNECT_ATTEMPTS = 10;

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

                if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                    const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
                    reconnectAttempts += 1;
                    reconnectTimer = setTimeout(() => {
                        void connect();
                    }, backoffDelay);
                } else {
                    setError('Connection lost. Unable to reconnect after multiple attempts.');
                }
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
    }, [enabled, invalidateRoomQueries, roomId]);

    return {
        isConnected,
        error,
        lastEvent,
    };
}
