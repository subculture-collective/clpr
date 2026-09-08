/**
 * Analytics Module
 *
 * Unified event tracking system for web and mobile applications.
 *
 * @example
 * ```typescript
 * import { trackEvent, AuthEvents, identifyUser } from '@/lib/telemetry';
 *
 * // Track a login event
 * trackEvent(AuthEvents.LOGIN_COMPLETED, {
 *   method: 'twitch'
 * });
 *
 * // Identify user
 * identifyUser('user123', {
 *   user_id: 'user123',
 *   signup_date: '2024-01-01'
 * });
 * ```
 */

// Export event schema
export {
  AuthEvents,
  SubmissionEvents,
  EngagementEvents,
  NavigationEvents,
  SettingsEvents,
  ErrorEvents,
  PerformanceEvents,
} from './events';

export type {
  EventName,
  EventProperties,
  BaseEventProperties,
  AuthEventProperties,
  SubmissionEventProperties,
  EngagementEventProperties,
  PremiumEventProperties,
  NavigationEventProperties,
  SettingsEventProperties,
  ErrorEventProperties,
  PerformanceEventProperties,
} from './events';

// Export tracker functions
export {
  trackEvent,
  trackPageView,
  trackError,
  trackPerformance,
  identifyUser,
  resetUser,
  enableAnalytics,
  disableAnalytics,
  isAnalyticsEnabled,
  configureAnalytics,
  getAnalyticsConfig,
} from './tracker';

export type {
  UserProperties,
} from './tracker';
