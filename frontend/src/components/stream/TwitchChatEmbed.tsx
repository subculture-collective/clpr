import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuth } from '../../context/AuthContext';
import { checkTwitchAuthStatus } from '../../lib/twitch-api';
import { useRegisterTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';

export interface TwitchChatEmbedProps {
  channel: string;
  position?: 'side' | 'bottom';
}

export function TwitchChatEmbed({ channel, position = 'side' }: TwitchChatEmbedProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);
  const { isAuthenticated: isUserLoggedIn } = useAuth();
  // The chat embed is a Twitch embed too, so floating controls stay clear of it.
  const chatRef = useRef<HTMLDivElement>(null);
  useRegisterTwitchPlayer(chatRef, true);

  const checkAuth = useCallback(async () => {
    try {
      const status = await checkTwitchAuthStatus();
      setIsAuthenticated(status.authenticated);
    } catch (error) {
      console.error('Failed to check Twitch auth status:', error);
      setIsAuthenticated(false);
    } finally {
      setIsCheckingAuth(false);
    }
  }, []);

  useEffect(() => {
    if (isUserLoggedIn) {
      checkAuth();
    } else {
      setIsCheckingAuth(false);
      setIsAuthenticated(false);
    }
  }, [isUserLoggedIn, checkAuth]);

  const handleTwitchLogin = () => {
    // Redirect to backend OAuth endpoint
    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
    window.location.href = `${apiBaseUrl}/twitch/oauth/authorize`;
  };

  const containerClasses = position === 'bottom'
    ? 'w-full h-[500px] mt-4'
    : 'w-full h-full min-h-[600px]';

  return (
    <div className={`flex flex-col border rounded-lg overflow-hidden bg-surface-raised ${containerClasses}`}>
      {/* Header */}
      <div className="flex shrink-0 items-center justify-between p-3 border-b border-border bg-surface">
        <h3 className="font-bold text-white">Twitch Chat</h3>

        {isUserLoggedIn && !isCheckingAuth && !isAuthenticated && (
          <button
            onClick={handleTwitchLogin}
            className="text-sm px-3 py-1 bg-primary-400 hover:bg-primary-300 text-background rounded transition-colors"
          >
            Login to Chat
          </button>
        )}

        {isAuthenticated && (
          <span className="text-sm text-green-600 dark:text-green-400 flex items-center gap-1">
            <span className="w-2 h-2 bg-green-500 rounded-full"></span>
            Connected
          </span>
        )}
      </div>

      {/* Chat Embed: fills the rest of the box so the iframe is never clipped */}
      <div ref={chatRef} className="relative w-full min-h-0 flex-1">
        <iframe
          src={`https://www.twitch.tv/embed/${encodeURIComponent(channel)}/chat?parent=${encodeURIComponent(window.location.hostname)}&darkpopout`}
          className="absolute inset-0 w-full h-full border-0"
          title={`${channel} Twitch Chat`}
          sandbox="allow-storage-access-by-user-activation allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox allow-modals"
        />
      </div>
    </div>
  );
}
