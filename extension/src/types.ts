/**
 * Shared type definitions for the Clipper browser extension.
 */

/** Represents the current authentication state stored in chrome.storage.local */
export interface AuthState {
  isAuthenticated: boolean;
  username?: string;
  displayName?: string;
  avatarUrl?: string;
}

/** Information extracted from a Twitch clip URL */
export interface TwitchClipInfo {
  clipId: string;
  clipUrl: string;
  /** Channel name (for twitch.tv/channel/clip/id URLs) */
  channel?: string;
}

/** Clip metadata returned by the Clipper API */
export interface ClipMetadata {
  clip_id: string;
  title: string;
  streamer_name: string;
  game_name: string;
  view_count: number;
  created_at: string;
  thumbnail_url: string;
  duration: number;
  url: string;
}

/** Tag as returned by the Clipper API */
export interface Tag {
  id: string;
  name: string;
  slug: string;
  clip_count?: number;
}

/** Request payload for submitting a clip */
export interface SubmitClipRequest {
  clip_url: string;
  custom_title?: string;
  tags?: string[];
  is_nsfw: boolean;
  submission_reason?: string;
}

/** Response from submitting a clip */
export interface SubmissionResponse {
  success: boolean;
  message: string;
}

/** User profile from GET /auth/me */
export interface UserProfile {
  id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  role: 'user' | 'admin' | 'moderator';
}

/** Message types sent between content scripts and background */
export type ExtensionMessage =
  | { type: 'TWITCH_CLIP_DETECTED'; clipInfo: TwitchClipInfo | null }
  | { type: 'GET_CURRENT_CLIP' }
  | { type: 'OPEN_POPUP'; clipInfo?: TwitchClipInfo };

/** Config stored in chrome.storage.local */
export interface ExtensionConfig {
  apiBaseUrl: string;
  frontendUrl: string;
}

export const DEFAULT_CONFIG: ExtensionConfig = {
  apiBaseUrl: 'https://api.clpr.gg/api/v1',
  frontendUrl: 'https://clpr.gg',
};
