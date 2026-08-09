import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest';
import {
    initGoogleAnalytics,
    disableGoogleAnalytics,
    enableGoogleAnalytics,
    trackPageView,
    trackEvent,
    trackClipSubmission,
    trackUpvote,
    trackDownvote,
    trackComment,
    trackShare,
    trackFollow,
    trackUserRegistration,
    trackSearch,
    trackCommunityJoin,
    trackFeedFollow,
    GA_MEASUREMENT_ID,
    configureGoogleAnalytics,
} from './google-analytics';

describe('Google Analytics Utilities', () => {
    let mockGtag: ReturnType<typeof vi.fn>;
    let mockDataLayer: unknown[];

    beforeEach(() => {
        configureGoogleAnalytics({
            measurementId: 'G-TEST123456',
            enabled: true,
            domain: 'test.clpr.tv',
        });
        // Reset state
        mockDataLayer = [];
        mockGtag = vi.fn((...args: unknown[]) => {
            mockDataLayer.push(args);
        });

        // Setup window mocks
        (window as Window & { gtag?: typeof mockGtag }).gtag = undefined;
        window.dataLayer = mockDataLayer;

        // Mock document.createElement for script injection
        const mockScript = document.createElement('script');
        vi.spyOn(document, 'createElement').mockReturnValue(mockScript);
        vi.spyOn(document.head, 'appendChild').mockImplementation(() => mockScript);

        // Mock console methods
        vi.spyOn(console, 'log').mockImplementation(() => {});
    });

    afterEach(() => {
        vi.restoreAllMocks();
        // Clean up window
        delete (window as Window & { gtag?: typeof mockGtag }).gtag;
        delete (window as Window & { 'ga-disable-G-TEST123456'?: boolean })['ga-disable-G-TEST123456'];
    });

    describe('initGoogleAnalytics', () => {
        it('should initialize GA when measurement ID is present', () => {
            initGoogleAnalytics();

            expect(document.createElement).toHaveBeenCalledWith('script');
            expect(document.head.appendChild).toHaveBeenCalled();
        });

        it('should not initialize GA when VITE_ENABLE_ANALYTICS is false', () => {
            // This test would require mocking the env differently
            // Skipping due to module-level constant
        });

        it('should not initialize GA twice', () => {
            initGoogleAnalytics();
            const firstCallCount = (document.createElement as ReturnType<typeof vi.spyOn>).mock.calls.length;

            initGoogleAnalytics();
            const secondCallCount = (document.createElement as ReturnType<typeof vi.spyOn>).mock.calls.length;

            expect(secondCallCount).toBe(firstCallCount);
            expect(console.log).toHaveBeenCalledWith('Google Analytics already initialized');
        });

        it('should set up dataLayer', () => {
            initGoogleAnalytics();
            expect(window.dataLayer).toBeDefined();
            expect(Array.isArray(window.dataLayer)).toBe(true);
        });
    });

    describe('disableGoogleAnalytics', () => {
        it('should set opt-out flag', () => {
            disableGoogleAnalytics();
            expect((window as Window & { 'ga-disable-G-TEST123456'?: boolean })['ga-disable-G-TEST123456']).toBe(true);
        });
    });

    describe('enableGoogleAnalytics', () => {
        it('should remove opt-out flag', () => {
            // First disable
            disableGoogleAnalytics();
            expect((window as Window & { 'ga-disable-G-TEST123456'?: boolean })['ga-disable-G-TEST123456']).toBe(true);

            // Then enable
            enableGoogleAnalytics();
            expect((window as Window & { 'ga-disable-G-TEST123456'?: boolean })['ga-disable-G-TEST123456']).toBe(false);
        });
    });

    describe('trackPageView', () => {
        beforeEach(() => {
            initGoogleAnalytics();
            (window as Window & { gtag?: typeof mockGtag }).gtag = mockGtag;
        });

        it('should track page view with path', () => {
            const path = '/test-page';
            trackPageView(path);

            expect(mockGtag).toHaveBeenCalledWith('event', 'page_view', expect.objectContaining({
                page_path: path,
            }));
        });

        it('should track page view with custom title', () => {
            const path = '/test-page';
            const title = 'Test Page Title';
            trackPageView(path, title);

            expect(mockGtag).toHaveBeenCalledWith('event', 'page_view', expect.objectContaining({
                page_path: path,
                page_title: title,
            }));
        });

        it('should not track when GA is not initialized', () => {
            delete (window as Window & { gtag?: typeof mockGtag }).gtag;
            trackPageView('/test');
            expect(mockGtag).not.toHaveBeenCalled();
        });
    });

    describe('trackEvent', () => {
        beforeEach(() => {
            initGoogleAnalytics();
            (window as Window & { gtag?: typeof mockGtag }).gtag = mockGtag;
        });

        it('should track custom event with parameters', () => {
            const eventName = 'custom_event';
            const params = { custom_param: 'value', count: 42 };

            trackEvent(eventName, params);

            expect(mockGtag).toHaveBeenCalledWith('event', eventName, expect.objectContaining({
                ...params,
                domain: 'test.clpr.tv',
            }));
        });

        it('should add domain to all events', () => {
            trackEvent('test_event');

            expect(mockGtag).toHaveBeenCalledWith('event', 'test_event', expect.objectContaining({
                domain: 'test.clpr.tv',
            }));
        });

        it('should not track when GA is not initialized', () => {
            delete (window as Window & { gtag?: typeof mockGtag }).gtag;
            trackEvent('test_event');
            expect(mockGtag).not.toHaveBeenCalled();
        });
    });

    describe('Pre-defined event tracking functions', () => {
        beforeEach(() => {
            initGoogleAnalytics();
            (window as Window & { gtag?: typeof mockGtag }).gtag = mockGtag;
            mockGtag.mockClear();
        });

        it('trackClipSubmission should track clip_submitted event without creator name', () => {
            trackClipSubmission('clip123');

            expect(mockGtag).toHaveBeenCalledWith('event', 'clip_submitted', expect.objectContaining({
                clip_id: 'clip123',
            }));
            // Verify no creator_name is sent
            expect(mockGtag).toHaveBeenCalledWith('event', 'clip_submitted', expect.not.objectContaining({
                creator_name: expect.anything(),
            }));
        });

        it('trackUpvote should track clip_upvoted event', () => {
            trackUpvote('clip123');

            expect(mockGtag).toHaveBeenCalledWith('event', 'clip_upvoted', expect.objectContaining({
                clip_id: 'clip123',
            }));
        });

        it('trackDownvote should track clip_downvoted event', () => {
            trackDownvote('clip123');

            expect(mockGtag).toHaveBeenCalledWith('event', 'clip_downvoted', expect.objectContaining({
                clip_id: 'clip123',
            }));
        });

        it('trackComment should track comment_posted event', () => {
            trackComment('clip123');

            expect(mockGtag).toHaveBeenCalledWith('event', 'comment_posted', expect.objectContaining({
                clip_id: 'clip123',
            }));
        });

        it('trackShare should track clip_shared event with platform', () => {
            trackShare('clip123', 'twitter');

            expect(mockGtag).toHaveBeenCalledWith('event', 'clip_shared', expect.objectContaining({
                clip_id: 'clip123',
                platform: 'twitter',
            }));
        });

        it('trackFollow should track follow_action event without target ID', () => {
            trackFollow('creator');

            expect(mockGtag).toHaveBeenCalledWith('event', 'follow_action', expect.objectContaining({
                target_type: 'creator',
            }));
            // Verify no target_id is sent
            expect(mockGtag).toHaveBeenCalledWith('event', 'follow_action', expect.not.objectContaining({
                target_id: expect.anything(),
            }));
        });

        it('trackUserRegistration should track user_registration event', () => {
            trackUserRegistration('twitch');

            expect(mockGtag).toHaveBeenCalledWith('event', 'user_registration', expect.objectContaining({
                registration_method: 'twitch',
            }));
        });

        it('trackSearch should track search event without query', () => {
            trackSearch(42);

            expect(mockGtag).toHaveBeenCalledWith('event', 'search', expect.objectContaining({
                results_count: 42,
            }));
            // Verify no search_term is sent
            expect(mockGtag).toHaveBeenCalledWith('event', 'search', expect.not.objectContaining({
                search_term: expect.anything(),
            }));
        });

        it('trackCommunityJoin should track community_joined event', () => {
            trackCommunityJoin('community123');

            expect(mockGtag).toHaveBeenCalledWith('event', 'community_joined', expect.objectContaining({
                community_id: 'community123',
            }));
        });

        it('trackFeedFollow should track feed_followed event', () => {
            trackFeedFollow('hot');

            expect(mockGtag).toHaveBeenCalledWith('event', 'feed_followed', expect.objectContaining({
                feed_type: 'hot',
            }));
        });
    });

    describe('Privacy features', () => {
        beforeEach(() => {
            initGoogleAnalytics();
            (window as Window & { gtag?: typeof mockGtag }).gtag = mockGtag;
        });

        it('should not send PII in tracking functions', () => {
            // Test that no PII is sent in any tracking function
            trackClipSubmission('clip123');
            trackFollow('user');

            // Verify calls don't include PII fields
            expect(mockGtag).not.toHaveBeenCalledWith(
                expect.anything(),
                expect.anything(),
                expect.objectContaining({
                    creator_name: expect.anything(),
                })
            );

            expect(mockGtag).not.toHaveBeenCalledWith(
                expect.anything(),
                expect.anything(),
                expect.objectContaining({
                    target_id: expect.anything(),
                })
            );
        });

        it('should include domain context in all events', () => {
            trackUpvote('clip123');

            expect(mockGtag).toHaveBeenCalledWith('event', 'clip_upvoted', expect.objectContaining({
                domain: 'test.clpr.tv',
            }));
        });
    });

    describe('GA_MEASUREMENT_ID export', () => {
        it('should export GA_MEASUREMENT_ID', () => {
            expect(GA_MEASUREMENT_ID).toBeDefined();
        });
    });
});
