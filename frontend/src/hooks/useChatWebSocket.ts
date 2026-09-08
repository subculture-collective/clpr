import { useEffect, useState, useCallback, useRef } from 'react';
import type { ChatMessage } from '@/types/chat';

interface UseChatWebSocketOptions {
    channelId: string;
    onMessage?: (message: ChatMessage) => void;
    onTyping?: (username: string) => void;
}

interface UseChatWebSocketReturn {
    messages: ChatMessage[];
    sendMessage: (content: string) => void;
    sendTyping: () => void;
    isConnected: boolean;
    error: string | null;
}

/**
 * Hook for managing WebSocket connection to chat channel
 */
export function useChatWebSocket({
    channelId,
    onMessage,
    onTyping,
}: UseChatWebSocketOptions): UseChatWebSocketReturn {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const reconnectAttemptsRef = useRef(0);
    const maxReconnectAttempts = 10;
    const connectRef = useRef<(() => void) | null>(null);

    // Store callbacks in refs to avoid reconnecting when they change
    const onMessageRef = useRef(onMessage);
    const onTypingRef = useRef(onTyping);
    useEffect(() => {
        onMessageRef.current = onMessage;
    }, [onMessage]);
    useEffect(() => {
        onTypingRef.current = onTyping;
    }, [onTyping]);

    const connect = useCallback(() => {
        try {
            // Get WebSocket URL from environment or default
            const wsProtocol =
                window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsHost = import.meta.env.VITE_WS_HOST || window.location.host;
            const wsUrl = `${wsProtocol}//${wsHost}/ws/chat/${channelId}`;

            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                setIsConnected(true);
                setError(null);
                reconnectAttemptsRef.current = 0; // Reset on successful connection
                console.log(`Connected to chat channel: ${channelId}`);
            };

            ws.onmessage = event => {
                try {
                    const data = JSON.parse(event.data);

                    if (data.type === 'message') {
                        const message = data.data as ChatMessage;
                        setMessages(prev => [...prev, message]);
                        onMessageRef.current?.(message);
                    } else if (data.type === 'typing') {
                        onTypingRef.current?.(data.data.username);
                    } else if (data.type === 'history') {
                        // Initial message history
                        setMessages(data.data.messages || []);
                    }
                } catch (err) {
                    console.error('Error parsing WebSocket message:', err);
                }
            };

            ws.onerror = event => {
                console.error('WebSocket error:', event);
                setError('Connection error');
            };

            ws.onclose = () => {
                setIsConnected(false);
                console.log(`Disconnected from chat channel: ${channelId}`);

                // Attempt to reconnect with exponential backoff
                if (reconnectAttemptsRef.current < maxReconnectAttempts) {
                    const backoffDelay = Math.min(
                        1000 * Math.pow(2, reconnectAttemptsRef.current),
                        30000,
                    );
                    const attemptNumber = reconnectAttemptsRef.current + 1;
                    reconnectAttemptsRef.current = attemptNumber;

                    console.log(
                        `Attempting to reconnect in ${backoffDelay}ms (attempt ${attemptNumber}/${maxReconnectAttempts})`,
                    );

                    reconnectTimeoutRef.current = window.setTimeout(() => {
                        connectRef.current?.();
                    }, backoffDelay);
                } else {
                    setError(
                        'Connection lost. Unable to reconnect after multiple attempts.',
                    );
                }
            };
        } catch (err) {
            setError('Failed to connect to chat');
            console.error('WebSocket connection error:', err);
        }
    }, [channelId]);

    useEffect(() => {
        // Store connect in ref so it can be called recursively
        connectRef.current = connect;
        queueMicrotask(() => {
            connect();
        });

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [connect]);

    const sendMessage = useCallback((content: string) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(
                JSON.stringify({
                    type: 'message',
                    content,
                }),
            );
        }
    }, []);

    const sendTyping = useCallback(() => {
        // Only send if WebSocket is connected
        if (wsRef.current?.readyState !== WebSocket.OPEN) {
            return;
        }

        // Clear previous typing timeout
        if (typingTimeoutRef.current) {
            clearTimeout(typingTimeoutRef.current);
        }

        wsRef.current.send(
            JSON.stringify({
                type: 'typing',
            }),
        );

        // Prevent sending typing events too frequently (throttle)
        typingTimeoutRef.current = window.setTimeout(() => {
            typingTimeoutRef.current = undefined;
        }, 3000);
    }, []);

    return {
        messages,
        sendMessage,
        sendTyping,
        isConnected,
        error,
    };
}
