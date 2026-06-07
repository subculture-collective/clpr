/* eslint-disable react-refresh/only-export-components */
import React, {
    createContext,
    useContext,
    useState,
    useEffect,
    useCallback,
} from 'react';
import {
    initGoogleAnalytics,
    disableGoogleAnalytics,
    enableGoogleAnalytics,
} from '../lib/google-analytics';
import {
    initPostHog,
    disablePostHog,
    enablePostHog,
} from '../lib/posthog-analytics';
import {
    enableAnalytics as enableUnifiedAnalytics,
    disableAnalytics as disableUnifiedAnalytics,
    configureAnalytics,
} from '../lib/telemetry';
import { useAuth } from './AuthContext';
import axios from 'axios';

/**
 * Consent categories for different types of tracking/personalization
 */
export interface ConsentPreferences {
    /** Essential cookies/storage - always true, required for site function */
    essential: boolean;
    /** Functional cookies - language, theme, preferences */
    functional: boolean;
    /** Analytics tracking consent */
    analytics: boolean;
    /** Personalized advertising consent */
    advertising: boolean;
    /** Timestamp of last consent update */
    updatedAt: string | null;
    /** Expiration timestamp (12 months from consent) */
    expiresAt: string | null;
}

/**
 * Default consent preferences (privacy-preserving defaults)
 * Following GDPR principles: no tracking until explicit consent
 */
const DEFAULT_CONSENT: ConsentPreferences = {
    essential: true, // Always required
    functional: false,
    analytics: false,
    advertising: false,
    updatedAt: null,
    expiresAt: null,
};

/**
 * Context value interface
 */
export interface ConsentContextType {
    /** Current consent preferences */
    consent: ConsentPreferences;
    /** Whether user has made a consent decision */
    hasConsented: boolean;
    /** Whether Do Not Track is enabled in browser */
    doNotTrack: boolean;
    /** Whether consent banner should be shown */
    showConsentBanner: boolean;
    /** Update consent preferences */
    updateConsent: (preferences: Partial<ConsentPreferences>) => void;
    /** Accept all optional consent categories */
    acceptAll: () => void;
    /** Reject all optional consent categories */
    rejectAll: () => void;
    /** Reset consent (show banner again) */
    resetConsent: () => void;
    /** Check if personalized ads are allowed (considers DNT and consent) */
    canShowPersonalizedAds: boolean;
    /** Check if analytics tracking is allowed */
    canTrackAnalytics: boolean;
}

const CONSENT_STORAGE_KEY = 'clpr_consent_preferences';
const CONSENT_VERSION = '1.0'; // Increment when consent categories change

const ConsentContext = createContext<ConsentContextType | undefined>(undefined);

/**
 * Detects if Do Not Track is enabled in the browser
 */
function detectDoNotTrack(): boolean {
    // Check navigator.doNotTrack (most browsers)
    if (navigator.doNotTrack === '1') {
        return true;
    }
    // Check window.doNotTrack (older browsers)
    if ((window as { doNotTrack?: string }).doNotTrack === '1') {
        return true;
    }
    // Check navigator.globalPrivacyControl (newer privacy standard)
    if (
        (navigator as { globalPrivacyControl?: boolean })
            .globalPrivacyControl === true
    ) {
        return true;
    }
    return false;
}

/**
 * Checks if consent has expired (12 months)
 */
function isConsentExpired(preferences: ConsentPreferences): boolean {
    if (!preferences.expiresAt) return true;
    return new Date() > new Date(preferences.expiresAt);
}

/**
 * Loads consent from localStorage
 */
function loadStoredConsent(): ConsentPreferences | null {
    try {
        const stored = localStorage.getItem(CONSENT_STORAGE_KEY);
        if (!stored) return null;

        const parsed = JSON.parse(stored);
        // Check version compatibility
        if (parsed.version !== CONSENT_VERSION) {
            // Version mismatch - treat as if no consent given
            return null;
        }

        // Check if consent has expired
        if (isConsentExpired(parsed.preferences)) {
            return null;
        }

        return parsed.preferences;
    } catch {
        return null;
    }
}

/**
 * Saves consent to localStorage
 */
function saveConsent(preferences: ConsentPreferences): void {
    try {
        localStorage.setItem(
            CONSENT_STORAGE_KEY,
            JSON.stringify({
                version: CONSENT_VERSION,
                preferences,
            }),
        );
    } catch (error) {
        console.error('Failed to save consent preferences:', error);
    }
}

/**
 * Syncs consent to backend (for logged-in users)
 */
async function syncConsentToBackend(
    preferences: ConsentPreferences,
): Promise<void> {
    try {
        await axios.post('/api/v1/users/me/consent', {
            essential: preferences.essential,
            functional: preferences.functional,
            analytics: preferences.analytics,
            advertising: preferences.advertising,
        });
    } catch (error) {
        // Silent fail - consent is still saved locally
        console.error('Failed to sync consent to backend:', error);
    }
}

/**
 * Loads consent from backend (for logged-in users)
 */
async function loadConsentFromBackend(): Promise<ConsentPreferences | null> {
    try {
        const response = await axios.get('/api/v1/users/me/consent');
        if (response.data?.success && response.data?.data) {
            const data = response.data.data;
            return {
                essential: data.essential ?? true,
                functional: data.functional ?? false,
                analytics: data.analytics ?? false,
                advertising: data.advertising ?? false,
                updatedAt: data.consent_date || data.updated_at || null,
                expiresAt: data.expires_at || null,
            };
        }
        return null;
    } catch {
        // Silent fail - use local consent
        return null;
    }
}

/**
 * Consent Provider component
 * Manages user consent for tracking, analytics, and personalized ads
 */
export function ConsentProvider({ children }: { children: React.ReactNode }) {
    const { user } = useAuth();
    const initialDoNotTrack = detectDoNotTrack();
    const storedConsent = loadStoredConsent();
    const autoConsent = import.meta.env.VITE_AUTO_CONSENT === 'true';
    const nowISO = new Date().toISOString();
    // eslint-disable-next-line react-hooks/purity
    const currentTime = Date.now();
    const oneYearISO = new Date(
        currentTime + 365 * 24 * 60 * 60 * 1000,
    ).toISOString();

    const bootstrapConsent: ConsentPreferences | null =
        storedConsent ? storedConsent
        : autoConsent ?
            {
                ...DEFAULT_CONSENT,
                functional: true,
                updatedAt: nowISO,
                expiresAt: oneYearISO,
            }
        :   null;

    const [consent, setConsent] = useState<ConsentPreferences>(
        bootstrapConsent ?? DEFAULT_CONSENT,
    );
    const [hasConsented, setHasConsented] =
        useState<boolean>(!!bootstrapConsent);
    const doNotTrack = initialDoNotTrack;
    const [showConsentBanner, setShowConsentBanner] =
        useState<boolean>(!bootstrapConsent);

    // Initialize analytics if consent already stored and DNT disabled
    useEffect(() => {
        if (hasConsented && consent.analytics && !doNotTrack) {
            initGoogleAnalytics();
            initPostHog();
            configureAnalytics({ enabled: true });
            enableUnifiedAnalytics();
        }
    }, [hasConsented, consent.analytics, doNotTrack]);

    // Load consent from backend for logged-in users
    useEffect(() => {
        if (!user) return;

        loadConsentFromBackend().then(backendConsent => {
            if (backendConsent && !isConsentExpired(backendConsent)) {
                // Backend consent exists and is valid - use it
                queueMicrotask(() => {
                    setConsent(backendConsent);
                    setHasConsented(true);
                    setShowConsentBanner(false);
                    saveConsent(backendConsent); // Sync to local storage

                    // Initialize analytics if consented
                    if (backendConsent.analytics && !doNotTrack) {
                        initGoogleAnalytics();
                        initPostHog();
                        configureAnalytics({ enabled: true });
                        enableUnifiedAnalytics();
                    }
                });
            }
        });
    }, [user, doNotTrack]);

    /**
     * Update consent preferences
     */
    const updateConsent = useCallback(
        async (preferences: Partial<ConsentPreferences>) => {
            const now = new Date();
            const expiresAt = new Date(
                now.getTime() + 365 * 24 * 60 * 60 * 1000,
            ); // 12 months

            const updatedPreferences: ConsentPreferences = {
                ...consent,
                ...preferences,
                essential: true, // Always keep essential enabled
                updatedAt: now.toISOString(),
                expiresAt: expiresAt.toISOString(),
            };

            // Update state first
            setConsent(updatedPreferences);
            setHasConsented(true);
            setShowConsentBanner(false);

            // Save to local storage
            saveConsent(updatedPreferences);

            // Sync to backend if user is logged in (non-blocking)
            if (user) {
                syncConsentToBackend(updatedPreferences);
            }

            // Handle analytics based on analytics consent
            if (updatedPreferences.analytics && !doNotTrack) {
                enableGoogleAnalytics();
                enablePostHog();
                configureAnalytics({ enabled: true });
                enableUnifiedAnalytics();
            } else {
                disableGoogleAnalytics();
                disablePostHog();
                disableUnifiedAnalytics();
            }
        },
        [consent, doNotTrack, user],
    );

    /**
     * Accept all optional consent categories
     */
    const acceptAll = useCallback(() => {
        updateConsent({
            functional: true,
            analytics: true,
            advertising: true,
        });
    }, [updateConsent]);

    /**
     * Reject all optional consent categories
     */
    const rejectAll = useCallback(() => {
        updateConsent({
            functional: false,
            analytics: false,
            advertising: false,
        });
    }, [updateConsent]);

    /**
     * Reset consent to trigger banner again
     */
    const resetConsent = useCallback(() => {
        try {
            localStorage.removeItem(CONSENT_STORAGE_KEY);
        } catch {
            // Ignore storage errors
        }

        // Disable analytics when consent is reset
        disableGoogleAnalytics();
        disablePostHog();

        setConsent(DEFAULT_CONSENT);
        setHasConsented(false);
        setShowConsentBanner(true);
    }, []);

    // Compute derived values
    // Personalized ads require explicit consent AND no DNT signal
    const canShowPersonalizedAds =
        hasConsented && consent.advertising && !doNotTrack;

    // Analytics tracking also respects DNT and consent
    const canTrackAnalytics = hasConsented && consent.analytics && !doNotTrack;

    return (
        <ConsentContext.Provider
            value={{
                consent,
                hasConsented,
                doNotTrack,
                showConsentBanner,
                updateConsent,
                acceptAll,
                rejectAll,
                resetConsent,
                canShowPersonalizedAds,
                canTrackAnalytics,
            }}
        >
            {children}
        </ConsentContext.Provider>
    );
}

/**
 * Hook to access consent context
 */
export function useConsent() {
    const context = useContext(ConsentContext);
    if (context === undefined) {
        throw new Error('useConsent must be used within a ConsentProvider');
    }
    return context;
}
