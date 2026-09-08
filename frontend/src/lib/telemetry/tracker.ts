/**
 * Unified Analytics Tracker
 * 
 * Provides a single interface for tracking events across multiple analytics providers.
 * Respects user consent and GDPR compliance through ConsentContext.
 * Automatically enriches events with user properties and context.
 */

import { trackEvent as trackGoogleAnalyticsEvent } from '../google-analytics';
import { trackPostHogEvent, identifyPostHogUser, trackPostHogPageView, resetPostHogUser } from '../posthog-analytics';
import type { EventName, EventProperties, BaseEventProperties } from './events';

/**
 * User properties for identification
 */
export interface UserProperties {
  user_id: string;
  username?: string;
  email?: string;
  is_premium?: boolean;
  premium_tier?: string;
  signup_date?: string;
  is_verified?: boolean;
  [key: string]: string | number | boolean | undefined;
}

/**
 * Analytics configuration
 */
interface AnalyticsConfig {
  enabled: boolean;
  debug: boolean;
  userId?: string;
  userProperties?: UserProperties;
}

// Constants for environment variable values
const ENV_TRUE = 'true';

let config: AnalyticsConfig = {
  enabled: false,
  debug: import.meta.env.VITE_ENABLE_DEBUG === ENV_TRUE || (import.meta.env.DEV && import.meta.env.MODE !== 'test'),
};

/**
 * Configure the analytics tracker
 */
export function configureAnalytics(options: Partial<AnalyticsConfig>): void {
  config = {
    ...config,
    ...options,
  };
  
  if (config.debug) {
    console.log('[Analytics] Configured:', config);
  }
}

/**
 * Enable analytics tracking
 * Should be called when user grants consent
 */
export function enableAnalytics(): void {
  config.enabled = true;
  
  if (config.debug) {
    console.log('[Analytics] Enabled');
  }
}

/**
 * Disable analytics tracking
 * Should be called when user revokes consent
 */
export function disableAnalytics(): void {
  config.enabled = false;
  
  if (config.debug) {
    console.log('[Analytics] Disabled');
  }
}

/**
 * Check if analytics is enabled
 */
export function isAnalyticsEnabled(): boolean {
  return config.enabled;
}

/**
 * Identify a user for analytics tracking
 * Call this after user logs in or when user properties change
 */
export function identifyUser(userId: string, properties?: UserProperties): void {
  if (!config.enabled) {
    if (config.debug) {
      console.log('[Analytics] Identify skipped (disabled):', { userId, properties });
    }
    return;
  }
  
  // Update config with user info
  config.userId = userId;
  config.userProperties = properties;
  
  // Identify in PostHog
  identifyPostHogUser(userId, properties);
  
  if (config.debug) {
    console.log('[Analytics] User identified:', { userId, properties });
  }
}

/**
 * Reset user identity (e.g., on logout)
 */
export function resetUser(): void {
  config.userId = undefined;
  config.userProperties = undefined;
  
  // Reset PostHog user identity
  resetPostHogUser();
  
  if (config.debug) {
    console.log('[Analytics] User reset');
  }
}

/**
 * Get common event properties
 * Automatically includes user context, device info, and page context
 */
function getCommonProperties(): BaseEventProperties {
  const properties: BaseEventProperties = {
    timestamp: new Date().toISOString(),
    page_path: window.location.pathname,
    page_title: document.title,
    referrer: document.referrer || undefined,
    platform: 'web',
  };
  
  // Add user properties if available
  if (config.userId) {
    properties.user_id = config.userId;
    properties.is_authenticated = true;
  } else {
    properties.is_authenticated = false;
  }
  
  if (config.userProperties) {
    if (config.userProperties.is_premium !== undefined) {
      properties.is_premium = config.userProperties.is_premium;
    }
    if (config.userProperties.premium_tier) {
      properties.premium_tier = config.userProperties.premium_tier;
    }
    if (config.userProperties.signup_date) {
      properties.signup_date = config.userProperties.signup_date;
    }
  }
  
  // Add device context
  if (typeof window !== 'undefined') {
    // Detect device type from user agent
    const ua = navigator.userAgent;
    if (/mobile/i.test(ua)) {
      properties.device_type = 'mobile';
    } else if (/tablet/i.test(ua)) {
      properties.device_type = 'tablet';
    } else {
      properties.device_type = 'desktop';
    }
    
    // Add browser info
    if (/chrome/i.test(ua) && !/edge/i.test(ua)) {
      properties.browser = 'Chrome';
    } else if (/safari/i.test(ua) && !/chrome/i.test(ua)) {
      properties.browser = 'Safari';
    } else if (/firefox/i.test(ua)) {
      properties.browser = 'Firefox';
    } else if (/edge/i.test(ua)) {
      properties.browser = 'Edge';
    }
    
    // Add OS info
    if (/windows/i.test(ua)) {
      properties.os = 'Windows';
    } else if (/mac/i.test(ua)) {
      properties.os = 'macOS';
    } else if (/linux/i.test(ua)) {
      properties.os = 'Linux';
    } else if (/android/i.test(ua)) {
      properties.os = 'Android';
    } else if (/ios|iphone|ipad/i.test(ua)) {
      properties.os = 'iOS';
    }
  }
  
  return properties;
}

/**
 * Track a custom event
 * Sends event to all enabled analytics providers
 */
export function trackEvent(
  eventName: EventName,
  properties?: EventProperties
): void {
  if (!config.enabled) {
    if (config.debug) {
      console.log('[Analytics] Event skipped (disabled):', eventName, properties);
    }
    return;
  }
  
  // Merge common properties with event-specific properties
  const enrichedProperties = {
    ...getCommonProperties(),
    ...properties,
  };
  
  // Track in Google Analytics
  trackGoogleAnalyticsEvent(eventName, enrichedProperties);
  
  // Track in PostHog
  trackPostHogEvent(eventName, enrichedProperties);
  
  if (config.debug) {
    console.log('[Analytics] Event tracked:', eventName, enrichedProperties);
  }
}

/**
 * Track a page view
 */
export function trackPageView(path?: string, title?: string): void {
  if (!config.enabled) {
    if (config.debug) {
      console.log('[Analytics] Page view skipped (disabled):', { path, title });
    }
    return;
  }
  
  const pagePath = path || window.location.pathname;
  const pageTitle = title || document.title;
  
  // Track in PostHog (which will auto-enrich with user properties)
  trackPostHogPageView(pagePath, pageTitle);
  
  // Track as custom event in Google Analytics
  trackGoogleAnalyticsEvent('page_view', {
    page_path: pagePath,
    page_title: pageTitle,
    ...getCommonProperties(),
  });
  
  if (config.debug) {
    console.log('[Analytics] Page view tracked:', { path: pagePath, title: pageTitle });
  }
}

/**
 * Track an error event
 * Convenience method for error tracking with rich context
 */
export function trackError(
  error: Error,
  context?: {
    errorType?: string;
    errorCode?: string;
    apiEndpoint?: string;
    httpStatus?: number;
  }
): void {
  trackEvent('error_occurred', {
    error_type: context?.errorType || error.name,
    error_message: error.message,
    error_stack: error.stack,
    error_code: context?.errorCode,
    api_endpoint: context?.apiEndpoint,
    http_status: context?.httpStatus,
  });
}

/**
 * Track a performance metric
 * Convenience method for performance tracking
 */
export function trackPerformance(
  metricName: string,
  metricValue: number,
  metricUnit: string = 'ms'
): void {
  // Determine the appropriate performance event based on metric name
  let eventName: EventName = 'page_load_time'; // default
  
  if (metricName.includes('api') || metricName.includes('response')) {
    eventName = 'api_response_time';
  } else if (metricName.includes('video')) {
    eventName = 'video_load_time';
  } else if (metricName.includes('search')) {
    eventName = 'search_response_time';
  }
  
  trackEvent(eventName, {
    metric_name: metricName,
    metric_value: metricValue,
    metric_unit: metricUnit,
  });
}

/**
 * Get current configuration (for debugging)
 */
export function getAnalyticsConfig(): Readonly<AnalyticsConfig> {
  return { ...config };
}
