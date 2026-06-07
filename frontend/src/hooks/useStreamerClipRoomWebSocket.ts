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
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>();
    const reconnectAttemptsRef = useRef(0);
    const maxReconnectAttempts = 10;
    const connectRef = useRef<(() => void) | null>(null);
    const onEventRef = useRef(onEvent);
    const isActiveRef = useRef(true);

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

    const connect = useCallback(async () => {
        if (!enabled || !roomId) return;

        try {
            const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsHost = import.meta.env.VITE_WS_HOST || window.location.host;
            const token = await getSecureItem('token');

            if (!token) {
                setError('Authentication required');
                return;
            }

            const wsUrl = `${wsProtocol}//${wsHost}/api/v1/streamer-clip-rooms/${encodeURIComponent(roomId)}/ws`;
            const authProtocol = `auth.bearer.${btoa(token)}`;
            const ws = new WebSocket(wsUrl, [authProtocol]);
            wsRef.current = ws;

            ws.onopen = () => {
                setIsConnected(true);
                setError(null);
                reconnectAttemptsRef.current = 0;
            };

            ws.onmessage = event => {
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
                console.error('Streamer clip room websocket error:', event);
                setError('Connection error');
            };

            ws.onclose = () => {
                setIsConnected(false);

                if (!isActiveRef.current) {
                    return;
                }

                if (reconnectAttemptsRef.current < maxReconnectAttempts && enabled) {
                    const backoffDelay = Math.min(
                        1000 * Math.pow(2, reconnectAttemptsRef.current),
                        30000,
                    );
                    reconnectAttemptsRef.current += 1;

                    reconnectTimeoutRef.current = window.setTimeout(() => {
                        connectRef.current?.();
                    }, backoffDelay);
                } else if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
                    setError('Connection lost. Unable to reconnect after multiple attempts.');
                }
            };
        } catch (err) {
            setError('Failed to connect to streamer clip room');
            console.error('Streamer clip room websocket connection error:', err);
        }
    }, [enabled, invalidateRoomQueries, roomId]);

    useEffect(() => {
        isActiveRef.current = true;
        connectRef.current = connect;

        queueMicrotask(() => {
            void connect();
        });

        return () => {
            isActiveRef.current = false;
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }
        };
    }, [connect]);

    return {
        isConnected,
        error,
        lastEvent,
    };
}
