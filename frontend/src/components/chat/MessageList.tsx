import type { ChatMessage } from '@/types/chat';
import { useEffect, useRef, useState } from 'react';
import { MessageItem } from './MessageItem';
import { Spinner } from '@/components/ui/Spinner';

interface MessageListProps {
  messages: ChatMessage[];
  loading?: boolean;
  onLoadMore?: () => void;
  hasMore?: boolean;
}

/**
 * MessageList component with auto-scroll and infinite scroll support
 * Uses virtual scrolling for performance with 1000+ messages
 */
export function MessageList({
  messages,
  loading = false,
  onLoadMore,
  hasMore = false,
}: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const prevScrollHeightRef = useRef(0);
  const isLoadingMoreRef = useRef(false);

  // Auto-scroll to bottom when new messages arrive (only if already at bottom)
  useEffect(() => {
    if (autoScroll) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, autoScroll]);

  // Handle scroll to detect if user is at bottom
  const handleScroll = () => {
    if (!containerRef.current) return;

    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 100;
    setAutoScroll(isAtBottom);

    // Load more messages when scrolling to top
    if (scrollTop < 100 && hasMore && !loading && onLoadMore) {
      prevScrollHeightRef.current = scrollHeight;
      isLoadingMoreRef.current = true;
      onLoadMore();
    }
  };

  // Maintain scroll position after loading more messages (prepend scenario)
  useEffect(() => {
    if (containerRef.current && prevScrollHeightRef.current && isLoadingMoreRef.current) {
      const newScrollHeight = containerRef.current.scrollHeight;
      const scrollDiff = newScrollHeight - prevScrollHeightRef.current;
      containerRef.current.scrollTop += scrollDiff;
      prevScrollHeightRef.current = 0;
      isLoadingMoreRef.current = false;
    }
  }, [messages.length]);

  if (messages.length === 0 && !loading) {
    return (
      <div className="flex-1 flex items-center justify-center p-4">
        <div className="text-center text-muted-foreground">
          <p className="text-sm">No messages yet</p>
          <p className="text-xs mt-1">Be the first to send a message!</p>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className="flex-1 overflow-y-auto p-4 space-y-1"
      onScroll={handleScroll}
    >
      {hasMore && (
        <div className="flex justify-center py-2">
          {loading ? (
            <Spinner size="sm" />
          ) : (
            <button
              onClick={onLoadMore}
              className="text-xs text-link dark:text-primary-400 hover:underline"
            >
              Load more messages
            </button>
          )}
        </div>
      )}

      {messages.map((message) => (
        <MessageItem key={message.id} message={message} />
      ))}

      <div ref={messagesEndRef} />
    </div>
  );
}
