import type { initSentry as InitSentry } from './sentry';

type SentryConfig = Parameters<typeof InitSentry>[0];

const isEnabled = import.meta.env.VITE_SENTRY_ENABLED === 'true';

export async function initSentry(config: SentryConfig): Promise<void> {
  if (!isEnabled || !config.dsn) {
    console.log('Sentry is disabled');
    return;
  }
  const sentry = await import('./sentry');
  sentry.initSentry(config);
}

export function setUser(userId: string | null, username?: string): void {
  if (!isEnabled) return;
  void import('./sentry').then((sentry) => sentry.setUser(userId, username));
}

export function clearUser(): void {
  if (!isEnabled) return;
  void import('./sentry').then((sentry) => sentry.clearUser());
}

export function captureBoundaryError(
  error: Error,
  componentStack: string | null | undefined,
): void {
  if (!isEnabled) return;
  void import('./sentry').then(({ Sentry }) => {
    Sentry.withScope((scope) => {
      scope.setContext('errorBoundary', { componentStack });
      Sentry.captureException(error);
    });
  });
}
