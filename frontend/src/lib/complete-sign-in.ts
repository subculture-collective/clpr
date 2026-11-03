import type { NavigateFunction } from 'react-router-dom';
import { getAuthReturnTo } from './auth-return';
import { fetchRecommendationPreferences } from './recommendation-api';
import { AuthEvents, trackEvent } from './telemetry';

/**
 * Finishes a sign-in once session cookies are set: loads the user, records
 * the login, and sends the user to where they were going (or onboarding).
 */
export async function completeSignIn(
  navigate: NavigateFunction,
  refreshUser: () => Promise<void>,
): Promise<void> {
  await refreshUser();

  trackEvent(AuthEvents.LOGIN_COMPLETED, {
    method: 'twitch',
  });

  // Get the intended destination from session storage or default to home
  const returnTo = getAuthReturnTo(sessionStorage.getItem('auth_return_to'));
  sessionStorage.removeItem('auth_return_to');

  if (returnTo === '/') {
    try {
      const preferences = await fetchRecommendationPreferences();
      if (!preferences.onboarding_completed) {
        navigate('/onboarding', { replace: true });
        return;
      }
    } catch {
      // Authentication succeeded; preference setup can remain optional.
    }
  }

  navigate(returnTo, { replace: true });
}
