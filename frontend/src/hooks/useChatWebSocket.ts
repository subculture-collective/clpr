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
 * Flat frame emitted by the backend chat hub (backend/internal/websocket/types.go
 * ServerMessage). History is replayed as individual `message` frames on connect.
 */
interface ChatServerFrame {
    type: 'message' | 'typing' | 'presence' | 'error' | string;
    channel_id?: string;
    user_id?: string;
    username?: string;
    display_name?: string;
    avatar_url?: string;
    content?: string;
    message_id?: string;
    timestamp?: string;
    presence_type?: string;
    error?: string;
}

const MAX_RECONNECT_ATTEMPTS = 10;

export function buildChatWebSocketUrl(channelId: string): string {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = import.meta.env.VITE_WS_HOST || window.location.host;
    return `${wsProtocol}//${wsHost}/api/v1/chat/channels/${encodeURIComponent(channelId)}/ws`;
}

function frameToChatMessage(frame: ChatServerFrame): ChatMessage | null {
    if (!frame.message_id || !frame.user_id || typeof frame.content !== 'string') {
        return null;
    }
    const timestamp = frame.timestamp ?? new Date().toISOString();
    return {
        id: frame.message_id,
        channel_id: frame.channel_id ?? '',
        user_id: frame.user_id,
        username: frame.username ?? '',
        display_name: frame.display_name,
        avatar_url: frame.avatar_url,
        content: frame.content,
        is_deleted: false,
        created_at: timestamp,
        updated_at: timestamp,
    };
}

function newMessageId(): string | undefined {
    return typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
        ? crypto.randomUUID()
        : undefined;
}

/**
 * Hook for managing the WebSocket connection to a chat channel.
 *
 * Authentication uses the HttpOnly `access_token` cookie: the chat upgrader
 * does not echo a WebSocket subprotocol, so a bearer subprotocol would make the
 * browser fail the handshake.
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
    const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

    // Store callbacks in refs to avoid reconnecting when they change
    const onMessageRef = useRef(onMessage);
    const onTypingRef = useRef(onTyping);
    useEffect(() => {
        onMessageRef.current = onMessage;
    }, [onMessage]);
    useEffect(() => {
        onTypingRef.current = onTyping;
    }, [onTyping]);

    useEffect(() => {
        let cancelled = false;
        let socket: WebSocket | null = null;
        let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
        let reconnectAttempts = 0;
        // History is replayed on every (re)connect; dedupe by message id.
        const seenIds = new Set<string>();

        const detach = (ws: WebSocket) => {
            ws.onopen = null;
            ws.onmessage = null;
            ws.onerror = null;
            ws.onclose = null;
        };

        const connect = () => {
            if (cancelled) return;
            let ws: WebSocket;
            try {
                ws = new WebSocket(buildChatWebSocketUrl(channelId));
            } catch (err) {
                setError('Failed to connect to chat');
                console.error('WebSocket connection error:', err);
                return;
            }
            socket = ws;
            wsRef.current = ws;

            ws.onopen = () => {
                if (cancelled) return;
                setIsConnected(true);
                setError(null);
                reconnectAttempts = 0;
            };

            ws.onmessage = event => {
                if (cancelled) return;
                let frame: ChatServerFrame;
                try {
                    frame = JSON.parse(event.data) as ChatServerFrame;
                } catch (err) {
                    console.error('Error parsing WebSocket message:', err);
                    return;
                }

                if (frame.type === 'message') {
                    const message = frameToChatMessage(frame);
                    if (!message || seenIds.has(message.id)) return;
                    seenIds.add(message.id);
                    setMessages(prev => [...prev, message]);
                    onMessageRef.current?.(message);
                } else if (frame.type === 'typing') {
                    if (frame.username) onTypingRef.current?.(frame.username);
                } else if (frame.type === 'error') {
                    setError(frame.error || 'Chat error');
                }
            };

            ws.onerror = event => {
                if (cancelled) return;
                console.error('WebSocket error:', event);
                setError('Connection error');
            };

            ws.onclose = () => {
                if (cancelled || socket !== ws) return;
                socket = null;
                wsRef.current = null;
                setIsConnected(false);

                if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                    const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
                    reconnectAttempts += 1;
                    reconnectTimer = setTimeout(connect, backoffDelay);
                } else {
                    setError('Connection lost. Unable to reconnect after multiple attempts.');
                }
            };
        };

        // Deferred so a StrictMode mount/unmount/mount cycle never opens the
        // first socket at all.
        queueMicrotask(connect);

        return () => {
            cancelled = true;
            if (reconnectTimer) clearTimeout(reconnectTimer);
            if (socket) {
                detach(socket);
                socket.close();
                if (wsRef.current === socket) wsRef.current = null;
                socket = null;
            }
            // Reset per-channel state so the next channel starts clean.
            setMessages([]);
            setIsConnected(false);
            setError(null);
        };
    }, [channelId]);

    const sendMessage = useCallback(
        (content: string) => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
                wsRef.current.send(
                    JSON.stringify({
                        type: 'message',
                        channel_id: channelId,
                        content,
                        message_id: newMessageId(),
                    }),
                );
            }
        },
        [channelId],
    );

    const sendTyping = useCallback(() => {
        if (wsRef.current?.readyState !== WebSocket.OPEN) {
            return;
        }
        // Throttle: at most one typing frame per 3 seconds
        if (typingTimeoutRef.current) {
            return;
        }

        wsRef.current.send(
            JSON.stringify({
                type: 'typing',
                channel_id: channelId,
            }),
        );

        typingTimeoutRef.current = setTimeout(() => {
            typingTimeoutRef.current = undefined;
        }, 3000);
    }, [channelId]);

    useEffect(
        () => () => {
            if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
        },
        [],
    );

    return {
        messages,
        sendMessage,
        sendTyping,
        isConnected,
        error,
    };
}
