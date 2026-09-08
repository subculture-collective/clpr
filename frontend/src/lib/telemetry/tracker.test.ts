import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  trackEvent,
  trackPageView,
  identifyUser,
  resetUser,
  enableAnalytics,
  disableAnalytics,
  isAnalyticsEnabled,
  configureAnalytics,
  AuthEvents,
} from './index';
import * as googleAnalytics from '../google-analytics';
import * as posthogAnalytics from '../posthog-analytics';

// Mock the analytics modules
vi.mock('../google-analytics');
vi.mock('../posthog-analytics');

describe('Analytics Tracker', () => {
  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks();
    
    // Reset analytics state
    disableAnalytics();
    resetUser();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Configuration', () => {
    it('should start disabled by default', () => {
      expect(isAnalyticsEnabled()).toBe(false);
    });

    it('should enable analytics when requested', () => {
      enableAnalytics();
      expect(isAnalyticsEnabled()).toBe(true);
    });

    it('should disable analytics when requested', () => {
      enableAnalytics();
      disableAnalytics();
      expect(isAnalyticsEnabled()).toBe(false);
    });

    it('should configure analytics', () => {
      configureAnalytics({
        enabled: true,
        userId: 'user123',
      });
      expect(isAnalyticsEnabled()).toBe(true);
    });
  });

  describe('User Identification', () => {
    it('should identify user when analytics is enabled', () => {
      enableAnalytics();
      
      identifyUser('user123', {
        user_id: 'user123',
        username: 'testuser',
      });

      expect(posthogAnalytics.identifyPostHogUser).toHaveBeenCalledWith(
        'user123',
        expect.objectContaining({
          user_id: 'user123',
          username: 'testuser',
        })
      );
    });

    it('should not identify user when analytics is disabled', () => {
      disableAnalytics();
      
      identifyUser('user123', {
        user_id: 'user123',
      });

      expect(posthogAnalytics.identifyPostHogUser).not.toHaveBeenCalled();
    });

    it('should reset user identity', () => {
      enableAnalytics();
      identifyUser('user123');
      resetUser();
      
      // Verify PostHog user is reset
      expect(posthogAnalytics.resetPostHogUser).toHaveBeenCalled();
    });
  });

  describe('Event Tracking', () => {
    it('should track events when analytics is enabled', () => {
      enableAnalytics();
      
      trackEvent(AuthEvents.LOGIN_COMPLETED, {
        method: 'twitch',
      });

      expect(googleAnalytics.trackEvent).toHaveBeenCalledWith(
        AuthEvents.LOGIN_COMPLETED,
        expect.objectContaining({
          method: 'twitch',
        })
      );

      expect(posthogAnalytics.trackPostHogEvent).toHaveBeenCalledWith(
        AuthEvents.LOGIN_COMPLETED,
        expect.objectContaining({
          method: 'twitch',
        })
      );
    });

    it('should not track events when analytics is disabled', () => {
      disableAnalytics();
      
      trackEvent(AuthEvents.LOGIN_COMPLETED, {
        method: 'twitch',
      });

      expect(googleAnalytics.trackEvent).not.toHaveBeenCalled();
      expect(posthogAnalytics.trackPostHogEvent).not.toHaveBeenCalled();
    });

    it('should enrich events with common properties', () => {
      enableAnalytics();
      
      trackEvent(AuthEvents.LOGIN_COMPLETED, {
        method: 'twitch',
      });

      expect(googleAnalytics.trackEvent).toHaveBeenCalledWith(
        AuthEvents.LOGIN_COMPLETED,
        expect.objectContaining({
          method: 'twitch',
          timestamp: expect.any(String),
          page_path: expect.any(String),
          platform: 'web',
        })
      );
    });

    it('should include user context when user is identified', () => {
      enableAnalytics();
      
      identifyUser('user123', {
        user_id: 'user123',
      });

      trackEvent(AuthEvents.LOGIN_COMPLETED);

      expect(googleAnalytics.trackEvent).toHaveBeenCalledWith(
        AuthEvents.LOGIN_COMPLETED,
        expect.objectContaining({
          user_id: 'user123',
          is_authenticated: true,
        })
      );
    });
  });

  describe('Page View Tracking', () => {
    it('should track page views when analytics is enabled', () => {
      enableAnalytics();
      
      trackPageView('/home', 'Home Page');

      expect(posthogAnalytics.trackPostHogPageView).toHaveBeenCalledWith(
        '/home',
        'Home Page'
      );

      // Google Analytics gets the page view as a custom event with enriched properties
      expect(googleAnalytics.trackEvent).toHaveBeenCalledWith(
        'page_view',
        expect.objectContaining({
          platform: 'web',
          timestamp: expect.any(String),
        })
      );
    });

    it('should not track page views when analytics is disabled', () => {
      disableAnalytics();
      
      trackPageView('/home', 'Home Page');

      expect(posthogAnalytics.trackPostHogPageView).not.toHaveBeenCalled();
      expect(googleAnalytics.trackEvent).not.toHaveBeenCalled();
    });
  });
});
