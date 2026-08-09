export interface AnalyticsRuntimeConfig {
    enabled: boolean;
    autoConsent: boolean;
    googleMeasurementId: string;
    postHogApiKey: string;
    postHogHost: string;
    domain: string;
}

declare global {
    interface Window {
        __CLPR_ANALYTICS_CONFIG__?: Partial<AnalyticsRuntimeConfig>;
    }
}

export function getAnalyticsRuntimeConfig(): AnalyticsRuntimeConfig {
    const buildConfig: AnalyticsRuntimeConfig = {
        enabled: import.meta.env.VITE_ENABLE_ANALYTICS === 'true',
        autoConsent: import.meta.env.VITE_AUTO_CONSENT === 'true',
        googleMeasurementId: import.meta.env.VITE_GA_MEASUREMENT_ID || '',
        postHogApiKey: import.meta.env.VITE_POSTHOG_API_KEY || '',
        postHogHost:
            import.meta.env.VITE_POSTHOG_HOST || 'https://app.posthog.com',
        domain: import.meta.env.VITE_DOMAIN || window.location.hostname,
    };

    return { ...buildConfig, ...window.__CLPR_ANALYTICS_CONFIG__ };
}
