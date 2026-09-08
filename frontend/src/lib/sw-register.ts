/**
 * Service Worker Registration
 * Registers the service worker for PWA functionality
 * Only registers in production builds to avoid caching issues during development
 */

export async function registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
    // Only register service worker in production
    if (import.meta.env.DEV) {
        console.log(
            '[SW] Skipping service worker registration in development mode'
        );
        return null;
    }

    // Check if service workers are supported
    if (!('serviceWorker' in navigator)) {
        console.log('[SW] Service workers are not supported in this browser');
        return null;
    }

    try {
        // Register the service worker
        const registration = await navigator.serviceWorker.register('/sw.js', {
            scope: '/',
        });

        console.log(
            '[SW] Service worker registered successfully',
            registration
        );

        // Only an update accepted in this page may reload it. Initial activation
        // and updates accepted in another tab must preserve unfinished input.
        let reloadRequested = false;
        let refreshing = false;
        const promptedWorkers = new WeakSet<ServiceWorker>();
        navigator.serviceWorker.addEventListener('controllerchange', () => {
            if (!reloadRequested || refreshing) return;
            refreshing = true;
            window.location.reload();
        });

        const offerUpdate = (worker: ServiceWorker) => {
            if (!navigator.serviceWorker.controller || promptedWorkers.has(worker)) return;
            promptedWorkers.add(worker);
            if (window.confirm('A new version of Clipper is available. Reload to update?')) {
                reloadRequested = true;
                worker.postMessage({ type: 'SKIP_WAITING' });
            }
        };
        const watchInstallingWorker = () => {
            const worker = registration.installing;
            if (!worker) return;
            worker.addEventListener('statechange', () => {
                if (worker.state === 'installed') offerUpdate(worker);
            });
        };

        registration.addEventListener('updatefound', watchInstallingWorker);
        // Registration can resolve after an installation has already started,
        // or with an update that was left waiting by a previous page visit.
        watchInstallingWorker();
        if (registration.waiting) offerUpdate(registration.waiting);

        return registration;
    } catch (error) {
        console.error('[SW] Service worker registration failed:', error);
        return null;
    }
}

/**
 * Unregister all service workers
 * Useful for debugging or if you need to clear the service worker
 */
export async function unregisterServiceWorker(): Promise<boolean> {
    if (!('serviceWorker' in navigator)) {
        return false;
    }

    try {
        const registrations = await navigator.serviceWorker.getRegistrations();
        for (const registration of registrations) {
            await registration.unregister();
            console.log('[SW] Service worker unregistered');
        }
        return true;
    } catch (error) {
        console.error('[SW] Failed to unregister service worker:', error);
        return false;
    }
}

/**
 * Check if the app is currently installed as a PWA
 */
export function isPWAInstalled(): boolean {
    // Check if running in standalone mode (installed PWA)
    return (
        window.matchMedia('(display-mode: standalone)').matches ||
        (window.navigator as { standalone?: boolean }).standalone === true // iOS Safari
    );
}

/**
 * Check if the app can be installed as a PWA
 */
export function canInstallPWA(): boolean {
    return 'BeforeInstallPromptEvent' in window;
}
