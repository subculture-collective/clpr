import { useState, useEffect } from 'react';
import { useConsent } from '../../context/ConsentContext';
import { Button } from '../ui/Button';
import { Toggle } from '../ui/Toggle';
import { cn } from '../../lib/utils';
import { Link } from 'react-router-dom';

export interface ConsentBannerProps {
  /** Additional CSS classes */
  className?: string;
}

/**
 * Consent banner component for GDPR/privacy compliance
 * Shows a banner at the bottom of the screen for users to manage their consent preferences
 */
export function ConsentBanner({ className }: ConsentBannerProps) {
  const {
    showConsentBanner,
    consent,
    updateConsent,
    acceptAll,
    rejectAll,
    doNotTrack,
  } = useConsent();

  const [showDetails, setShowDetails] = useState(false);
  // Derive local preferences from consent to avoid sync issues
  const [pendingPreferences, setPendingPreferences] = useState({
    functional: consent.functional,
    analytics: consent.analytics,
    advertising: consent.advertising,
  });

  // Update pending preferences when consent changes (e.g., from settings page)
  useEffect(() => {
    queueMicrotask(() => {
      setPendingPreferences({
        functional: consent.functional,
        analytics: consent.analytics,
        advertising: consent.advertising,
      });
    });
  }, [consent.functional, consent.analytics, consent.advertising]);

  if (!showConsentBanner) {
    return null;
  }

  const handleSavePreferences = () => {
    updateConsent(pendingPreferences);
  };

  return (
    <div
      className={cn(
        'fixed bottom-[calc(4rem+env(safe-area-inset-bottom))] left-0 right-0 z-50 max-h-[70dvh] overflow-y-auto bg-background border-t border-border shadow-2xl md:bottom-0 md:max-h-none',
        'animate-in slide-in-from-bottom duration-300',
        className
      )}
      role="dialog"
      aria-modal="true"
      aria-labelledby="consent-banner-title"
      aria-describedby="consent-banner-description"
    >
      <div className="max-w-6xl mx-auto p-3 pb-[calc(.75rem+env(safe-area-inset-bottom))] md:p-6">
        {doNotTrack && (
          <div className="mb-4 p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <p className="text-sm text-blue-800 dark:text-blue-200">
              <strong>Do Not Track detected:</strong> We respect your browser's Do Not Track setting.
              Personalized ads and analytics tracking will be disabled even if you consent.
            </p>
          </div>
        )}

        {!showDetails ? (
          // Simple view
          <div className="flex flex-col md:flex-row items-start md:items-center gap-3 md:gap-4">
            <div className="flex-1">
              <h2 id="consent-banner-title" className="text-base md:text-lg font-semibold mb-1">
                Privacy & Cookie Preferences
              </h2>
              <p id="consent-banner-description" className="text-xs md:text-sm text-muted-foreground">
                <span className='md:hidden'>Choose whether clpr may use optional cookies for analytics and personalization. </span>
                <span className='hidden md:inline'>We use cookies and similar technologies to personalize ads, analyze traffic,
                and improve your experience. You can customize your preferences or accept/reject all. </span>
                <Link to="/privacy" className="font-semibold text-violet-300 underline underline-offset-2">
                  Privacy Policy
                </Link>
              </p>
            </div>
            <div className="grid grid-cols-3 gap-2 w-full md:flex md:w-auto">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowDetails(true)}
                className="flex-1 md:flex-none"
              >
                Customize
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={rejectAll}
                className="flex-1 border-violet-400 text-violet-300 md:flex-none"
              >
                Reject All
              </Button>
              <Button
                variant="primary"
                size="sm"
                onClick={acceptAll}
                className="flex-1 md:flex-none"
              >
                Accept All
              </Button>
            </div>
          </div>
        ) : (
          // Detailed view
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 id="consent-banner-title" className="text-lg font-semibold">
                Customize Privacy Preferences
              </h2>
              <button
                onClick={() => setShowDetails(false)}
                className="text-muted-foreground hover:text-foreground"
                aria-label="Close detailed preferences"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                </svg>
              </button>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              {/* Essential - Always enabled */}
              <div className="p-4 bg-muted/50 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-medium">Essential</h3>
                  <span className="text-xs bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300 px-2 py-1 rounded">
                    Always Active
                  </span>
                </div>
                <p className="text-sm text-muted-foreground">
                  Required for the website to function properly. Includes authentication,
                  security features, and your saved preferences.
                </p>
              </div>

              {/* Functional */}
              <div className="p-4 bg-muted/50 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-medium">Functional</h3>
                  <Toggle
                    checked={pendingPreferences.functional}
                    onChange={(e) => setPendingPreferences(prev => ({
                      ...prev,
                      functional: e.target.checked
                    }))}
                    disabled={doNotTrack}
                    aria-describedby="functional-description"
                  />
                </div>
                <p id="functional-description" className="text-sm text-muted-foreground">
                  Remember your preferences like language, theme, and other settings
                  to enhance your experience.
                </p>
              </div>

              {/* Analytics */}
              <div className="p-4 bg-muted/50 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-medium">Analytics</h3>
                  <Toggle
                    checked={pendingPreferences.analytics}
                    onChange={(e) => setPendingPreferences(prev => ({
                      ...prev,
                      analytics: e.target.checked
                    }))}
                    disabled={doNotTrack}
                    aria-describedby="analytics-description"
                  />
                </div>
                <p id="analytics-description" className="text-sm text-muted-foreground">
                  Help us understand how you use clpr so we can improve the platform.
                  This includes page views, feature usage, and error tracking.
                </p>
              </div>

              {/* Advertising */}
              <div className="p-4 bg-muted/50 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-medium">Advertising</h3>
                  <Toggle
                    checked={pendingPreferences.advertising}
                    onChange={(e) => setPendingPreferences(prev => ({
                      ...prev,
                      advertising: e.target.checked
                    }))}
                    disabled={doNotTrack}
                    aria-describedby="advertising-description"
                  />
                </div>
                <p id="advertising-description" className="text-sm text-muted-foreground">
                  Allow us to show ads tailored to your interests based on your viewing history
                  and preferences. Without this, you'll see contextual ads instead.
                </p>
              </div>
            </div>

            <div className="flex flex-wrap gap-2 justify-end pt-2 border-t border-border">
              <Button
                variant="outline"
                size="sm"
                onClick={rejectAll}
                className='border-violet-400 text-violet-300'
              >
                Reject All
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={acceptAll}
              >
                Accept All
              </Button>
              <Button
                variant="primary"
                size="sm"
                onClick={handleSavePreferences}
              >
                Save Preferences
              </Button>
            </div>

            <p className="text-xs text-muted-foreground text-center">
              You can change these preferences at any time in{' '}
              <Link to="/settings" className="text-primary-500 hover:underline">
                Settings
              </Link>
              {' '}or{' '}
              <Link to="/privacy" className="text-primary-500 hover:underline">
                Privacy Policy
              </Link>
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
