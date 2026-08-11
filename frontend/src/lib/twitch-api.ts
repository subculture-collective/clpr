// Twitch OAuth and chat integration API functions

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export interface TwitchAuthStatus {
  authenticated: boolean;
  bot_authorized: boolean;
  clip_download_authorized: boolean;
  twitch_username?: string;
}

export function getTwitchBotAuthorizationUrl(channel: string): string {
  const returnTo = `/streamer-tools/${encodeURIComponent(channel)}/clips`;
  return `${API_BASE_URL}/twitch/oauth/authorize?return_to=${encodeURIComponent(returnTo)}`;
}

/**
 * Check if the current user has authenticated with Twitch for chat
 */
export async function checkTwitchAuthStatus(): Promise<TwitchAuthStatus> {
  const response = await fetch(`${API_BASE_URL}/twitch/auth/status`, {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to check Twitch auth status');
  }

  return response.json();
}

/**
 * Revoke Twitch OAuth authentication
 */
export async function revokeTwitchAuth(): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/twitch/auth`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to revoke Twitch auth');
  }
}
