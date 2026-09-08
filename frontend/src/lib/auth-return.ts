/** Preserve a requested in-app destination without accepting external or auth-loop URLs. */
export function getAuthReturnTo(destination: unknown): string {
    let path = destination;
    if (destination && typeof destination === 'object' && 'pathname' in destination) {
        const location = destination as { pathname?: unknown; search?: unknown; hash?: unknown };
        path = typeof location.pathname === 'string'
            ? location.pathname + (typeof location.search === 'string' ? location.search : '') +
              (typeof location.hash === 'string' ? location.hash : '')
            : '';
    }
    if (typeof path !== 'string' || !path.startsWith('/') || path.startsWith('//') || path.includes('\\') || [...path].some(char => char.charCodeAt(0) < 32 || char.charCodeAt(0) === 127)) {
        return '/';
    }
    const url = new URL(path, 'https://clpr.invalid');
    if (url.origin !== 'https://clpr.invalid' || url.pathname === '/login' || url.pathname.startsWith('/auth/')) {
        return '/';
    }
    return url.pathname + url.search + url.hash;
}
