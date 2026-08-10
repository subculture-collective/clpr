import * as Sentry from '@sentry/react';
import { useEffect } from 'react';
import {
    createRoutesFromChildren,
    matchRoutes,
    useLocation,
    useNavigationType,
} from 'react-router-dom';

interface SentryConfig {
    dsn: string;
    environment?: string;
    release?: string;
    tracesSampleRate?: number;
    enabled?: boolean;
}

/**
 * Initialize Sentry SDK with React integration
 */
export function initSentry(config: SentryConfig): void {
    // Only initialize if enabled and DSN is provided
    if (!config.enabled || !config.dsn) {
        console.log('Sentry is disabled');
        return;
    }

    Sentry.init({
        dsn: config.dsn,
        environment: config.environment || 'development',
        release: config.release,

        // Performance monitoring
        tracesSampleRate: config.tracesSampleRate ?? 1.0,

        // Session replay
        replaysSessionSampleRate: 0.1, // 10% of sessions
        replaysOnErrorSampleRate: 1.0, // 100% of sessions with errors

        integrations: [
            // React Router integration for tracking navigation
            Sentry.reactRouterV7BrowserTracingIntegration({
                useEffect,
                useLocation,
                useNavigationType,
                createRoutesFromChildren,
                matchRoutes,
            }),
            // Replay integration for session recording
            Sentry.replayIntegration({
                maskAllText: true,
                blockAllMedia: true,
            }),
        ],

        // Before sending events, scrub sensitive data
        beforeSend(event) {
            return scrubSensitiveData(event);
        },

        // Don't capture specific errors
        ignoreErrors: [
            // Browser extensions
            'top.GLOBALS',
            // Random network errors
            'Network request failed',
            'NetworkError',
            'Failed to fetch',
            // React hydration errors (non-critical)
            'Hydration failed',
        ],
    });

    console.log(
        `Sentry initialized: environment=${config.environment}, release=${config.release}`
    );
}

/**
 * Scrub sensitive data from Sentry events
 */
function scrubSensitiveData(
    event: Sentry.ErrorEvent
): Sentry.ErrorEvent | null {
    // Scrub sensitive headers
    if (event.request?.headers) {
        delete event.request.headers['Authorization'];
        delete event.request.headers['Cookie'];
        delete event.request.headers['X-CSRF-Token'];
    }

    // Scrub sensitive cookies
    if (event.request?.cookies) {
        delete event.request.cookies['access_token'];
        delete event.request.cookies['refresh_token'];
    }

    // Scrub user PII - keep only hashed ID
    if (event.user?.id) {
        event.user.id = hashUserId(event.user.id);
        delete event.user.email;
        delete event.user.username;
        delete event.user.ip_address;
    }

    // Scrub breadcrumbs
    if (event.breadcrumbs) {
        event.breadcrumbs = event.breadcrumbs.map((breadcrumb) => {
            if (breadcrumb.data) {
                delete breadcrumb.data.password;
                delete breadcrumb.data.token;
                delete breadcrumb.data.secret;
                delete breadcrumb.data.apiKey;
            }
            return breadcrumb;
        });
    }

    return event;
}

/**
 * Hash user ID for privacy (matches backend implementation)
 */
function hashUserId(userId: string | number): string {
    const idStr = String(userId);
    // Use browser's SubtleCrypto API for hashing
    // For simplicity in sync function, use a simple hash
    let hash = 0;
    for (let i = 0; i < idStr.length; i++) {
        const char = idStr.charCodeAt(i);
        hash = (hash << 5) - hash + char;
        hash = hash & hash; // Convert to 32bit integer
    }
    return Math.abs(hash).toString(16).slice(0, 16);
}

/**
 * Set user context in Sentry
 */
export function setUser(userId: string | null, username?: string): void {
    if (!userId) {
        Sentry.setUser(null);
        return;
    }

    Sentry.setUser({
        id: hashUserId(userId),
        username: username,
    });
}

/**
 * Clear user context
 */
export function clearUser(): void {
    Sentry.setUser(null);
}

/**
 * Add breadcrumb for tracking user actions
 */
export function addBreadcrumb(
    message: string,
    category?: string,
    data?: Record<string, unknown>
): void {
    Sentry.addBreadcrumb({
        message,
        category: category || 'user-action',
        data,
        level: 'info',
    });
}

/**
 * Capture exception manually
 */
export function captureException(
    error: Error,
    context?: Record<string, Record<string, unknown>>
): void {
    if (context) {
        Sentry.withScope((scope) => {
            Object.entries(context).forEach(([key, value]) => {
                scope.setContext(key, value);
            });
            Sentry.captureException(error);
        });
    } else {
        Sentry.captureException(error);
    }
}

/**
 * Capture message manually
 */
export function captureMessage(
    message: string,
    level: Sentry.SeverityLevel = 'info'
): void {
    Sentry.captureMessage(message, level);
}

/**
 * Export Sentry for direct access if needed
 */
export { Sentry };

/**
 * Capture an error with component stack context for error boundaries
 */
export function captureBoundaryError(
  error: Error,
  componentStack: string | null | undefined,
): void {
  Sentry.withScope((scope) => {
    scope.setContext('errorBoundary', { componentStack });
    Sentry.captureException(error);
  });
}
