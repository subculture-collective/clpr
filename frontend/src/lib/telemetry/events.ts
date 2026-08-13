/**
 * Comprehensive Event Tracking Schema
 *
 * Defines all trackable events across the application for analytics.
 * Events are organized by category and include properties for rich data collection.
 */

// ============================================================================
// Event Categories
// ============================================================================

/**
 * Authentication Events
 */
export const AuthEvents = {
  // User Registration
  SIGNUP_STARTED: 'signup_started',
  SIGNUP_COMPLETED: 'signup_completed',
  SIGNUP_FAILED: 'signup_failed',

  // User Login
  LOGIN_STARTED: 'login_started',
  LOGIN_COMPLETED: 'login_completed',
  LOGIN_FAILED: 'login_failed',

  // User Logout
  LOGOUT: 'logout',

  // OAuth
  OAUTH_REDIRECT: 'oauth_redirect',
  OAUTH_CALLBACK: 'oauth_callback',
} as const;

/**
 * Submission Events (Clips)
 */
export const SubmissionEvents = {
  // Viewing
  SUBMISSION_VIEWED: 'submission_viewed',
  SUBMISSION_LIST_VIEWED: 'submission_list_viewed',

  // Creating
  SUBMISSION_CREATE_STARTED: 'submission_create_started',
  SUBMISSION_CREATE_COMPLETED: 'submission_create_completed',
  SUBMISSION_CREATE_FAILED: 'submission_create_failed',

  // Editing
  SUBMISSION_EDIT_STARTED: 'submission_edit_started',
  SUBMISSION_EDIT_COMPLETED: 'submission_edit_completed',
  SUBMISSION_EDIT_FAILED: 'submission_edit_failed',

  // Deleting
  SUBMISSION_DELETE_STARTED: 'submission_delete_started',
  SUBMISSION_DELETE_COMPLETED: 'submission_delete_completed',
  SUBMISSION_DELETE_FAILED: 'submission_delete_failed',

  // Sharing
  SUBMISSION_SHARE_CLICKED: 'submission_share_clicked',
  SUBMISSION_SHARED: 'submission_shared',
  SUBMISSION_SHARE_COPIED: 'submission_share_copied',

  // Playback
  SUBMISSION_PLAY_STARTED: 'submission_play_started',
  SUBMISSION_PLAY_COMPLETED: 'submission_play_completed',
  SUBMISSION_PLAY_PAUSED: 'submission_play_paused',

  // Rate Limiting
  SUBMISSION_RATE_LIMIT_HIT: 'submission_rate_limit_hit',
  SUBMISSION_RATE_LIMIT_EXPIRED: 'submission_rate_limit_expired',
} as const;

/**
 * Engagement Events
 */
export const EngagementEvents = {
  // Voting
  UPVOTE_CLICKED: 'upvote_clicked',
  DOWNVOTE_CLICKED: 'downvote_clicked',
  VOTE_REMOVED: 'vote_removed',

  // Comments
  COMMENT_CREATE_STARTED: 'comment_create_started',
  COMMENT_CREATE_COMPLETED: 'comment_create_completed',
  COMMENT_CREATE_FAILED: 'comment_create_failed',
  COMMENT_EDIT_STARTED: 'comment_edit_started',
  COMMENT_EDIT_COMPLETED: 'comment_edit_completed',
  COMMENT_DELETE_CLICKED: 'comment_delete_clicked',
  COMMENT_DELETE_COMPLETED: 'comment_delete_completed',
  COMMENT_REPLY_CLICKED: 'comment_reply_clicked',

  // Following
  FOLLOW_CLICKED: 'follow_clicked',
  UNFOLLOW_CLICKED: 'unfollow_clicked',

  // Favorites
  FAVORITE_ADDED: 'favorite_added',
  FAVORITE_REMOVED: 'favorite_removed',

  // Search
  SEARCH_PERFORMED: 'search_performed',
  SEARCH_FILTER_APPLIED: 'search_filter_applied',
  SEARCH_RESULT_CLICKED: 'search_result_clicked',

  // Community
  COMMUNITY_JOINED: 'community_joined',
  COMMUNITY_LEFT: 'community_left',
  COMMUNITY_CREATED: 'community_created',

  // Feed
  FEED_FOLLOWED: 'feed_followed',
  FEED_UNFOLLOWED: 'feed_unfollowed',
} as const;

/**
 * Premium/Subscription Events
 */
export const PremiumEvents = {
  // Pricing Page
  PRICING_PAGE_VIEWED: 'pricing_page_viewed',
  PRICING_TIER_CLICKED: 'pricing_tier_clicked',

  // Checkout Flow
  CHECKOUT_STARTED: 'checkout_started',
  CHECKOUT_PAYMENT_INFO_ENTERED: 'checkout_payment_info_entered',
  CHECKOUT_COMPLETED: 'checkout_completed',
  CHECKOUT_FAILED: 'checkout_failed',
  CHECKOUT_CANCELLED: 'checkout_cancelled',

  // Subscription Management
  SUBSCRIPTION_CREATED: 'subscription_created',
  SUBSCRIPTION_UPDATED: 'subscription_updated',
  SUBSCRIPTION_CANCELLED: 'subscription_cancelled',
  SUBSCRIPTION_RESUMED: 'subscription_resumed',
  SUBSCRIPTION_PAYMENT_FAILED: 'subscription_payment_failed',

  // Upgrade/Downgrade
  UPGRADE_MODAL_VIEWED: 'upgrade_modal_viewed',
  UPGRADE_CLICKED: 'upgrade_clicked',
  DOWNGRADE_CLICKED: 'downgrade_clicked',

  // Feature Paywalls
  PAYWALL_VIEWED: 'paywall_viewed',
  PAYWALL_DISMISSED: 'paywall_dismissed',
  PAYWALL_UPGRADE_CLICKED: 'paywall_upgrade_clicked',
} as const;

/**
 * Navigation Events
 */
export const NavigationEvents = {
  // Page Views
  PAGE_VIEWED: 'page_viewed',
  SCREEN_VIEWED: 'screen_viewed', // For mobile apps

  // Feature Navigation
  FEATURE_CLICKED: 'feature_clicked',
  NAV_LINK_CLICKED: 'nav_link_clicked',
  BACK_BUTTON_CLICKED: 'back_button_clicked',

  // Tabs/Sections
  TAB_CLICKED: 'tab_clicked',
  SECTION_VIEWED: 'section_viewed',

  // External Navigation
  EXTERNAL_LINK_CLICKED: 'external_link_clicked',
} as const;

/**
 * User Settings Events
 */
export const SettingsEvents = {
  // Profile
  PROFILE_VIEWED: 'profile_viewed',
  PROFILE_EDITED: 'profile_edited',
  AVATAR_CHANGED: 'avatar_changed',

  // Preferences
  SETTINGS_VIEWED: 'settings_viewed',
  LANGUAGE_CHANGED: 'language_changed',
  NOTIFICATION_PREFERENCES_CHANGED: 'notification_preferences_changed',
  FEED_AUTOPLAY_CHANGED: 'feed_autoplay_changed',

  // Privacy
  PRIVACY_SETTINGS_CHANGED: 'privacy_settings_changed',
  CONSENT_UPDATED: 'consent_updated',

  // Account
  PASSWORD_CHANGE_STARTED: 'password_change_started',
  PASSWORD_CHANGE_COMPLETED: 'password_change_completed',
  EMAIL_CHANGE_STARTED: 'email_change_started',
  EMAIL_CHANGE_COMPLETED: 'email_change_completed',
  ACCOUNT_DELETED: 'account_deleted',
} as const;

/**
 * Error Events
 */
export const ErrorEvents = {
  // Application Errors
  ERROR_OCCURRED: 'error_occurred',
  API_ERROR: 'api_error',
  NETWORK_ERROR: 'network_error',

  // Form Errors
  FORM_VALIDATION_ERROR: 'form_validation_error',
  FORM_SUBMISSION_ERROR: 'form_submission_error',

  // Media Errors
  VIDEO_PLAYBACK_ERROR: 'video_playback_error',
  IMAGE_LOAD_ERROR: 'image_load_error',
} as const;

/**
 * Performance Events
 */
export const PerformanceEvents = {
  PAGE_LOAD_TIME: 'page_load_time',
  API_RESPONSE_TIME: 'api_response_time',
  VIDEO_LOAD_TIME: 'video_load_time',
  SEARCH_RESPONSE_TIME: 'search_response_time',
} as const;

// ============================================================================
// Event Property Types
// ============================================================================

/**
 * Common properties that can be included with any event
 */
export interface BaseEventProperties {
  // User context
  user_id?: string;
  is_authenticated?: boolean;
  is_premium?: boolean;
  premium_tier?: string;
  signup_date?: string;

  // Session context
  session_id?: string;
  timestamp?: string;

  // Device context
  platform?: 'web' | 'mobile';
  device_type?: 'mobile' | 'tablet' | 'desktop';
  os?: string;
  browser?: string;

  // Page context
  page_path?: string;
  page_title?: string;
  referrer?: string;

  // Additional metadata
  [key: string]: string | number | boolean | undefined;
}

/**
 * Authentication event properties
 */
export interface AuthEventProperties extends BaseEventProperties {
  method?: 'twitch' | 'email' | 'oauth';
  provider?: string;
  error?: string;
}

/**
 * Submission event properties
 */
export interface SubmissionEventProperties extends BaseEventProperties {
  submission_id?: string;
  clip_id?: string;
  title?: string;
  creator_name?: string;
  broadcaster_name?: string;
  game_name?: string;
  duration?: number;
  is_nsfw?: boolean;
  tags?: string[];
  share_platform?: string;
  playback_position?: number;
  playback_duration?: number;
}

/**
 * Engagement event properties
 */
export interface EngagementEventProperties extends BaseEventProperties {
  target_id?: string;
  target_type?: 'clip' | 'comment' | 'user' | 'creator' | 'community';
  action?: 'add' | 'remove' | 'update';
  search_query?: string;
  search_results_count?: number;
  filter_type?: string;
  filter_value?: string;
}

/**
 * Premium event properties
 */
export interface PremiumEventProperties extends BaseEventProperties {
  tier?: 'pro' | 'premium';
  billing_period?: 'monthly' | 'yearly';
  price?: number;
  currency?: string;
  payment_method?: string;
  promo_code?: string;
  feature?: string;
  cancel_reason?: string;
  error?: string;
}

/**
 * Navigation event properties
 */
export interface NavigationEventProperties extends BaseEventProperties {
  from_page?: string;
  to_page?: string;
  link_text?: string;
  link_url?: string;
  feature_name?: string;
  tab_name?: string;
  section_name?: string;
}

/**
 * Settings event properties
 */
export interface SettingsEventProperties extends BaseEventProperties {
  setting_name?: string;
  old_value?: string;
  new_value?: string;
  consent_type?: 'essential' | 'functional' | 'analytics' | 'advertising';
}

/**
 * Error event properties
 */
export interface ErrorEventProperties extends BaseEventProperties {
  error_type?: string;
  error_message?: string;
  error_code?: string;
  error_stack?: string;
  api_endpoint?: string;
  http_status?: number;
}

/**
 * Performance event properties
 */
export interface PerformanceEventProperties extends BaseEventProperties {
  metric_name?: string;
  metric_value?: number;
  metric_unit?: string;
}

// ============================================================================
// Type Exports
// ============================================================================

export type EventName =
  | typeof AuthEvents[keyof typeof AuthEvents]
  | typeof SubmissionEvents[keyof typeof SubmissionEvents]
  | typeof EngagementEvents[keyof typeof EngagementEvents]
  | typeof PremiumEvents[keyof typeof PremiumEvents]
  | typeof NavigationEvents[keyof typeof NavigationEvents]
  | typeof SettingsEvents[keyof typeof SettingsEvents]
  | typeof ErrorEvents[keyof typeof ErrorEvents]
  | typeof PerformanceEvents[keyof typeof PerformanceEvents];

export type EventProperties =
  | AuthEventProperties
  | SubmissionEventProperties
  | EngagementEventProperties
  | PremiumEventProperties
  | NavigationEventProperties
  | SettingsEventProperties
  | ErrorEventProperties
  | PerformanceEventProperties
  | BaseEventProperties;
