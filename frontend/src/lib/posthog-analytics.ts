/**
 * PostHog Analytics Integration
 *
 * Provides PostHog tracking functionality with privacy-first approach.
 * Respects user consent preferences and DNT signals.
 *
 * Note: This is a lightweight implementation that loads PostHog dynamically
 * when consent is given, without adding it as a dependency.
 */

import { getAnalyticsRuntimeConfig } from './runtime-config';

// PostHog configuration
export const POSTHOG_API_KEY = getAnalyticsRuntimeConfig().postHogApiKey;
export const POSTHOG_HOST = getAnalyticsRuntimeConfig().postHogHost;
const ANALYTICS_ENABLED = getAnalyticsRuntimeConfig().enabled;

// Track if PostHog is initialized
let posthogInitialized = false;
let posthogInstance: PostHogQueue | null = null;

// PostHog array queue type
type PostHogQueue = unknown[] & {
    people?: { set?: (props: Record<string, unknown>) => void };
    init?: (
        key: string,
        options: Record<string, unknown>,
        name?: string,
    ) => void;
    capture?: (event: string, properties?: Record<string, unknown>) => void;
    identify?: (id: string, traits?: Record<string, unknown>) => void;
    reset?: () => void;
    opt_out_capturing?: () => void;
    opt_in_capturing?: () => void;
};

/**
 * Dynamically loads PostHog using their official snippet pattern
 *
 * SECURITY NOTE: This implementation loads PostHog from their official CDN.
 * For production environments with strict security requirements, consider
 * self-hosting or adding SRI verification.
 */
function loadPostHogScript(): Promise<PostHogQueue> {
    return new Promise((resolve, reject) => {
        // Check if already loaded
        const existingPostHog = (window as { posthog?: PostHogQueue }).posthog;
        if (existingPostHog && typeof existingPostHog.capture === 'function') {
            resolve(existingPostHog);
            return;
        }

        // Initialize the posthog queue (official snippet pattern)
        const win = window as { posthog?: PostHogQueue };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        win.posthog = win.posthog || ([] as any);
        const posthog = win.posthog as PostHogQueue;

        // Queue methods to be called after script loads
        const methods = [
            'init',
            'capture',
            'identify',
            'reset',
            'register',
            'unregister',
            'opt_out_capturing',
            'opt_in_capturing',
            'isFeatureEnabled',
            'onFeatureFlags',
            'getFeatureFlag',
            'getFeatureFlagPayload',
            'reloadFeatureFlags',
            'group',
            'setPersonProperties',
            'resetPersonProperties',
        ];

        for (const method of methods) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (posthog as any)[method] = function (...args: unknown[]) {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                (posthog as any).push([method, ...args]);
            };
        }

        // Load the script
        if (document.querySelector('script[src*="posthog"]')) {
            resolve(posthog);
            return;
        }

        const script = document.createElement('script');
        script.async = true;
        script.src = 'https://us-assets.i.posthog.com/static/array.js';
        script.onload = () => resolve(posthog);
        script.onerror = () =>
            reject(new Error('Failed to load PostHog script'));
        document.head.appendChild(script);
    });
}

/**
 * Initialize PostHog Analytics
 * Should be called after user grants analytics consent
 */
export async function initPostHog(): Promise<void> {
    if (!ANALYTICS_ENABLED) {
        console.log('PostHog is disabled via VITE_ENABLE_ANALYTICS');
        return;
    }

    if (!POSTHOG_API_KEY) {
        console.log('PostHog is not configured (no API key)');
        return;
    }

    if (posthogInitialized) {
        console.log('PostHog already initialized');
        return;
    }

    try {
        // Load PostHog script and get the queue
        const posthog = await loadPostHogScript();
        posthogInstance = posthog;

        // Initialize with privacy-preserving settings
        posthog.init?.(POSTHOG_API_KEY, {
            api_host: POSTHOG_HOST,
            loaded: () => {
                posthogInitialized = true;
                console.log('PostHog initialized');
            },
            capture_pageview: false, // We'll manually track page views
            opt_out_capturing_by_default: false,
            respect_dnt: true,
            autocapture: false, // Disable autocapture for better privacy
            disable_session_recording: true, // Disable session recording
            secure_cookie: true,
            persistence: 'localStorage',
        });
    } catch (error) {
        console.error('Failed to initialize PostHog:', error);
    }
}

/**
 * Disable PostHog Analytics
 * Respects user's opt-out preference
 */
export function disablePostHog(): void {
    if (!posthogInstance) return;

    try {
        posthogInstance.opt_out_capturing?.();
        posthogInitialized = false;
        console.log('PostHog disabled');
    } catch (error) {
        console.error('Failed to disable PostHog:', error);
    }
}

/**
 * Enable PostHog Analytics
 * Re-enables tracking if user grants consent
 */
export async function enablePostHog(): Promise<void> {
    if (!posthogInstance) {
        await initPostHog();
        return;
    }

    try {
        posthogInstance.opt_in_capturing?.();
        posthogInitialized = true;
        console.log('PostHog enabled');
    } catch (error) {
        console.error('Failed to enable PostHog:', error);
    }
}

/**
 * Get PostHog instance
 */
function getPostHog(): PostHogQueue | null {
    if (!posthogInitialized || !posthogInstance) return null;
    return posthogInstance;
}

/**
 * Track a page view
 */
export function trackPostHogPageView(path: string, title?: string): void {
    const posthog = getPostHog();
    if (!posthog?.capture) return;

    posthog.capture('$pageview', {
        $current_url: window.location.href,
        path,
        title: title || document.title,
    });
}

/**
 * Track a custom event
 */
export function trackPostHogEvent(
    eventName: string,
    properties?: Record<string, unknown>,
): void {
    const posthog = getPostHog();
    if (!posthog?.capture) return;

    posthog.capture(eventName, properties);
}

/**
 * Identify a user (when logged in)
 */
export function identifyPostHogUser(
    userId: string,
    traits?: Record<string, unknown>,
): void {
    if (!posthogInstance) return;

    try {
        posthogInstance.identify?.(userId, traits);
    } catch (error) {
        console.error('Failed to identify PostHog user:', error);
    }
}

/**
 * Reset PostHog user identity (on logout)
 */
export function resetPostHogUser(): void {
    if (!posthogInstance) return;

    try {
        posthogInstance.reset?.();
    } catch (error) {
        console.error('Failed to reset PostHog user:', error);
    }
}
