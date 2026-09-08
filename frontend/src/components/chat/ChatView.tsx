import { useNavigate } from 'react-router-dom';
import { useState, useRef, useEffect } from 'react';
import { useChatWebSocket } from '@/hooks/useChatWebSocket';
import { useDesktopNotifications } from '@/hooks/useDesktopNotifications';
import { useAuth } from '@/hooks/useAuth';
import { useToast } from '@/hooks';
import { MessageList } from './MessageList';
import { MessageComposer } from './MessageComposer';
import { TypingIndicator } from './TypingIndicator';
import { Alert } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import { Bell, BellOff, Settings } from 'lucide-react';

interface ChatViewProps {
  channelId: string;
  channelName?: string;
}

/**
 * ChatView component that displays messages and composer for a channel
 */
export function ChatView({ channelId, channelName }: ChatViewProps) {
  const navigate = useNavigate();
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const { user } = useAuth();
  const { showToast } = useToast();
  const typingTimeoutsRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());
  const {
    permission,
    requestPermission,
    showNotification,
    isSupported,
  } = useDesktopNotifications();

  const { messages, sendMessage, sendTyping, isConnected, error } = useChatWebSocket({
    channelId,
    onMessage: (message) => {
      // Show desktop notification for mentions
      if (
        user &&
        message.content.includes(`@${user.username}`) &&
        message.user_id !== user.id
      ) {
        showNotification({
          title: `${channelName || channelId}: ${message.username} mentioned you`,
          body: message.content,
          tag: `mention-${message.id}`,
        });
      }
    },
    onTyping: (username: string) => {
      // Clear existing timeout for this user to extend typing duration
      const existingTimeout = typingTimeoutsRef.current.get(username);
      if (existingTimeout) {
        clearTimeout(existingTimeout);
      }

      // Add user to typing list if not already present
      setTypingUsers((prev) => {
        if (!prev.includes(username)) {
          return [...prev, username];
        }
        return prev;
      });

      // Remove typing indicator after 3 seconds
      const timeoutId = setTimeout(() => {
        setTypingUsers((prev) => prev.filter((u) => u !== username));
        typingTimeoutsRef.current.delete(username);
      }, 3000);

      typingTimeoutsRef.current.set(username, timeoutId);
    },
  });

  const handleSend = (content: string) => {
    sendMessage(content);
  };

  // Cleanup typing timeouts on unmount
  useEffect(() => {
    const typingTimeouts = typingTimeoutsRef.current;
    return () => {
      typingTimeouts.forEach((timeout) => clearTimeout(timeout));
      typingTimeouts.clear();
    };
  }, []);

  const handleNotificationToggle = async () => {
    if (permission === 'granted') {
      // Can't revoke permission programmatically, just inform user
      showToast('Please manage notification permissions in your browser settings', 'info');
    } else {
      await requestPermission();
    }
  };

  return (
    <div className="flex-1 flex flex-col bg-background">
      {/* Channel header */}
      <div className="border-b border-border p-4 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">
            # {channelName || channelId}
          </h2>
          <div className="flex items-center gap-2 mt-1">
            <div
              className={`w-2 h-2 rounded-full ${
                isConnected ? 'bg-success-500' : 'bg-error-500'
              }`}
              aria-label={isConnected ? 'Connected' : 'Disconnected'}
            />
            <span className="text-xs text-muted-foreground">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2">
          {/* Settings button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate(`/chat/channels/${channelId}/settings`)}
            aria-label="Channel settings"
          >
            <Settings className="w-4 h-4" />
          </Button>

          {/* Notification toggle */}
          {isSupported && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleNotificationToggle}
              aria-label={
                permission === 'granted'
                  ? 'Notifications enabled'
                  : 'Enable notifications'
              }
            >
              {permission === 'granted' ? (
                <Bell className="w-4 h-4" />
              ) : (
                <BellOff className="w-4 h-4" />
              )}
            </Button>
          )}
        </div>
      </div>

      {/* Error message */}
      {error && (
        <div className="p-4">
          <Alert variant="error" title="Connection Error">
            {error}
          </Alert>
        </div>
      )}

      {/* Message list */}
      <MessageList messages={messages} />

      {/* Typing indicator */}
      <TypingIndicator typingUsers={typingUsers} />

      {/* Message composer */}
      <MessageComposer
        onSend={handleSend}
        onTyping={sendTyping}
        disabled={!isConnected}
      />
    </div>
  );
}
