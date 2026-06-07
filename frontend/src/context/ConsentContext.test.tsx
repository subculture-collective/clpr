import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { ConsentProvider, useConsent } from './ConsentContext';

// Mock useAuth hook from AuthContext
vi.mock('./AuthContext', async () => {
    const actual = await vi.importActual('./AuthContext');
    return {
        ...actual,
        useAuth: () => ({
            user: null,
            isAuthenticated: false,
        }),
    };
});

// Test component that uses the consent context
function TestConsumer() {
    const {
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
    } = useConsent();

    return (
        <div>
            <div data-testid='has-consented'>{String(hasConsented)}</div>
            <div data-testid='do-not-track'>{String(doNotTrack)}</div>
            <div data-testid='show-banner'>{String(showConsentBanner)}</div>
            <div data-testid='analytics'>{String(consent.analytics)}</div>
            <div data-testid='advertising'>{String(consent.advertising)}</div>
            <div data-testid='functional'>{String(consent.functional)}</div>
            <div data-testid='can-personalized'>
                {String(canShowPersonalizedAds)}
            </div>
            <div data-testid='can-analytics'>{String(canTrackAnalytics)}</div>
            <button onClick={acceptAll} data-testid='accept-all'>
                Accept All
            </button>
            <button onClick={rejectAll} data-testid='reject-all'>
                Reject All
            </button>
            <button onClick={resetConsent} data-testid='reset'>
                Reset
            </button>
            <button
                onClick={() => updateConsent({ analytics: true })}
                data-testid='enable-analytics'
            >
                Enable Analytics
            </button>
        </div>
    );
}

describe('ConsentContext', () => {
    beforeEach(() => {
        // Clear localStorage before each test
        localStorage.clear();
        // Reset navigator.doNotTrack
        Object.defineProperty(navigator, 'doNotTrack', {
            value: '0',
            writable: true,
            configurable: true,
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('should show banner when no consent is stored', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        expect(screen.getByTestId('has-consented')).toHaveTextContent('false');
        expect(screen.getByTestId('show-banner')).toHaveTextContent('true');
    });

    it('should default to privacy-preserving settings', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        expect(screen.getByTestId('analytics')).toHaveTextContent('false');
        expect(screen.getByTestId('advertising')).toHaveTextContent('false');
        expect(screen.getByTestId('functional')).toHaveTextContent('false');
    });

    it('should accept all when clicking accept all', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        act(() => {
            fireEvent.click(screen.getByTestId('accept-all'));
        });

        expect(screen.getByTestId('has-consented')).toHaveTextContent('true');
        expect(screen.getByTestId('show-banner')).toHaveTextContent('false');
        expect(screen.getByTestId('analytics')).toHaveTextContent('true');
        expect(screen.getByTestId('advertising')).toHaveTextContent('true');
        expect(screen.getByTestId('functional')).toHaveTextContent('true');
    });

    it('should reject all when clicking reject all', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        act(() => {
            fireEvent.click(screen.getByTestId('reject-all'));
        });

        expect(screen.getByTestId('has-consented')).toHaveTextContent('true');
        expect(screen.getByTestId('show-banner')).toHaveTextContent('false');
        expect(screen.getByTestId('analytics')).toHaveTextContent('false');
        expect(screen.getByTestId('advertising')).toHaveTextContent('false');
        expect(screen.getByTestId('functional')).toHaveTextContent('false');
    });

    it('should update individual consent preference', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        act(() => {
            fireEvent.click(screen.getByTestId('enable-analytics'));
        });

        expect(screen.getByTestId('has-consented')).toHaveTextContent('true');
        expect(screen.getByTestId('analytics')).toHaveTextContent('true');
        expect(screen.getByTestId('advertising')).toHaveTextContent('false');
    });

    it('should persist consent to localStorage', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        act(() => {
            fireEvent.click(screen.getByTestId('accept-all'));
        });

        const stored = localStorage.getItem('clpr_consent_preferences');
        expect(stored).toBeTruthy();

        const parsed = JSON.parse(stored!);
        expect(parsed.version).toBe('1.0');
        expect(parsed.preferences.analytics).toBe(true);
        expect(parsed.preferences.advertising).toBe(true);
    });

    it('should load consent from localStorage', () => {
        // Pre-set consent in localStorage with future expiration
        const expiresAt = new Date();
        expiresAt.setFullYear(expiresAt.getFullYear() + 1); // 1 year from now

        localStorage.setItem(
            'clpr_consent_preferences',
            JSON.stringify({
                version: '1.0',
                preferences: {
                    essential: true,
                    analytics: true,
                    advertising: false,
                    functional: true,
                    updatedAt: new Date().toISOString(),
                    expiresAt: expiresAt.toISOString(),
                },
            }),
        );

        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        expect(screen.getByTestId('has-consented')).toHaveTextContent('true');
        expect(screen.getByTestId('show-banner')).toHaveTextContent('false');
        expect(screen.getByTestId('analytics')).toHaveTextContent('true');
        expect(screen.getByTestId('advertising')).toHaveTextContent('false');
        expect(screen.getByTestId('functional')).toHaveTextContent('true');
    });

    it('should reset consent and show banner again', () => {
        // Pre-set consent
        localStorage.setItem(
            'clpr_consent_preferences',
            JSON.stringify({
                version: '1.0',
                preferences: {
                    essential: true,
                    analytics: true,
                    advertising: true,
                    functional: true,
                    updatedAt: new Date().toISOString(),
                },
            }),
        );

        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        act(() => {
            fireEvent.click(screen.getByTestId('reset'));
        });

        expect(screen.getByTestId('has-consented')).toHaveTextContent('false');
        expect(screen.getByTestId('show-banner')).toHaveTextContent('true');
        expect(localStorage.getItem('clpr_consent_preferences')).toBeNull();
    });

    it('should detect Do Not Track signal', () => {
        // Mock DNT signal
        Object.defineProperty(navigator, 'doNotTrack', {
            value: '1',
            writable: true,
            configurable: true,
        });

        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        expect(screen.getByTestId('do-not-track')).toHaveTextContent('true');
    });

    it('should disallow personalized ads when DNT is enabled even with consent', () => {
        // Mock DNT signal
        Object.defineProperty(navigator, 'doNotTrack', {
            value: '1',
            writable: true,
            configurable: true,
        });

        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        // Accept all consent
        act(() => {
            fireEvent.click(screen.getByTestId('accept-all'));
        });

        // Even though we consented, DNT should block personalized ads
        expect(screen.getByTestId('advertising')).toHaveTextContent('true');
        expect(screen.getByTestId('can-personalized')).toHaveTextContent(
            'false',
        );
        expect(screen.getByTestId('can-analytics')).toHaveTextContent('false');
    });

    it('should allow personalized ads when consented and no DNT', () => {
        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        act(() => {
            fireEvent.click(screen.getByTestId('accept-all'));
        });

        expect(screen.getByTestId('can-personalized')).toHaveTextContent(
            'true',
        );
        expect(screen.getByTestId('can-analytics')).toHaveTextContent('true');
    });

    it('should throw error when useConsent is used outside provider', () => {
        // Suppress console.error for this test
        const consoleSpy = vi
            .spyOn(console, 'error')
            .mockImplementation(() => {});

        expect(() => {
            render(<TestConsumer />);
        }).toThrow('useConsent must be used within a ConsentProvider');

        consoleSpy.mockRestore();
    });

    it('should invalidate old consent versions', () => {
        // Pre-set consent with old version
        localStorage.setItem(
            'clpr_consent_preferences',
            JSON.stringify({
                version: '0.9',
                preferences: {
                    essential: true,
                    analytics: true,
                    advertising: true,
                    functional: true,
                    updatedAt: new Date().toISOString(),
                },
            }),
        );

        render(
            <ConsentProvider>
                <TestConsumer />
            </ConsentProvider>,
        );

        // Should show banner again due to version mismatch
        expect(screen.getByTestId('has-consented')).toHaveTextContent('false');
        expect(screen.getByTestId('show-banner')).toHaveTextContent('true');
    });
});
