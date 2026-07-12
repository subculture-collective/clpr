import { apiClient } from './api';
import { generatePKCEParams } from './pkce';
import {
  clearSecureStorage,
  getSecureItem,
  removeSecureItem,
  setSecureItem,
} from './secure-storage';

export interface User {
  id: string;
  twitch_id: string;
  username: string;
  display_name: string;
  email?: string;
  avatar_url?: string;
  bio?: string;
  role: 'user' | 'admin' | 'moderator';
  karma_points: number;
  is_banned: boolean;
  ban_reason?: string;
  banned_until?: string;
  is_verified?: boolean;
  is_premium?: boolean;
  premium_tier?: string;
  created_at: string;
  last_login_at?: string;
}

export interface TestLoginParams {
  user_id?: string;
  username?: string;
}

/**
 * Initiates Twitch OAuth flow with PKCE
 * Generates code verifier/challenge and stores verifier securely
 */
export async function initiateOAuth() {
  const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

  // Generate PKCE parameters
  const { codeVerifier, codeChallenge, state } = await generatePKCEParams();

  // Store code verifier and state securely for callback validation
  await setSecureItem('oauth_code_verifier', codeVerifier);
  await setSecureItem('oauth_state', state);

  // Build OAuth URL with PKCE parameters
  const params = new URLSearchParams({
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
    state,
  });

  // In E2E environments, trigger the OAuth endpoint without performing a full navigation
  if (typeof window !== 'undefined' && (window as unknown as Record<string, unknown>).__E2E_MOCK_OAUTH__) {
    try {
      await fetch(`${apiUrl}/auth/twitch?${params.toString()}`, { credentials: 'include', mode: 'no-cors' });
    } catch (err) {
      console.warn('[initiateOAuth] Mock OAuth fetch failed:', err);
    }
    return;
  }

  // Redirect to backend OAuth endpoint with PKCE params
  window.location.href = `${apiUrl}/auth/twitch?${params.toString()}`;
}

/**
 * Handle OAuth callback with PKCE
 * Validates state and sends code verifier to backend
 */
export async function handleOAuthCallback(code: string, state: string): Promise<{ success: boolean; error?: string }> {
  try {
    // Retrieve stored state and code verifier
    const storedState = await getSecureItem('oauth_state');
    const codeVerifier = await getSecureItem('oauth_code_verifier');

    // Validate state parameter (CSRF protection)
    if (!storedState || storedState !== state) {
      return { success: false, error: 'Invalid state parameter' };
    }

    if (!codeVerifier) {
      return { success: false, error: 'Code verifier not found' };
    }

    // Send code and verifier to backend
    // Backend will validate the verifier against the challenge
    await apiClient.post('/auth/twitch/callback', {
      code,
      state,
      code_verifier: codeVerifier,
    });

    // Clean up stored PKCE parameters
    await removeSecureItem('oauth_state');
    await removeSecureItem('oauth_code_verifier');

    return { success: true };
  } catch (error) {
    // Clean up on error (reuse the imported function)
    await removeSecureItem('oauth_state');
    await removeSecureItem('oauth_code_verifier');

    return {
      success: false,
      error: error instanceof Error ? error.message : 'Authentication failed'
    };
  }
}

/**
 * Gets the current authenticated user
 */
export async function getCurrentUser(): Promise<User> {
  const response = await apiClient.get<User>('/auth/me');
  return response.data;
}

/**
 * Logs out the current user and clears secure storage
 */
export async function logout(): Promise<void> {
  try {
    // Revoke tokens on backend
    await apiClient.post('/auth/logout');
  } finally {
    // Clear all secure storage regardless of API call result
    await clearSecureStorage();
  }
}

/**
 * Test-only login helper for E2E/local environments
 */
export async function testLogin(params: TestLoginParams): Promise<User> {
  const response = await apiClient.post<{ user: User }>('/auth/test-login', params);
  return response.data.user;
}
