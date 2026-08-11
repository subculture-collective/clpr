/**
 * Deep Linking Utilities for PWA
 * 
 * Handles deep links from universal links (iOS) and app links (Android)
 * for the Clipper PWA.
 */

export interface DeepLinkRoute {
  pattern: RegExp;
  handler: (matches: RegExpMatchArray) => string;
  description: string;
}

/**
 * Supported deep link routes
 */
export const DEEP_LINK_ROUTES: DeepLinkRoute[] = [
  {
    pattern: /^\/clip\/([a-zA-Z0-9_-]+)$/,
    handler: (matches) => `/clip/${matches[1]}`,
    description: 'Clip detail page',
  },
  {
    pattern: /^\/profile$/,
    handler: () => '/profile',
    description: 'User profile page',
  },
  {
    pattern: /^\/profile\/stats$/,
    handler: () => '/profile/stats',
    description: 'User stats page',
  },
  {
    pattern: /^\/search$/,
    handler: () => '/search',
    description: 'Search page',
  },
  {
    pattern: /^\/submit$/,
    handler: () => '/submit',
    description: 'Submit clip page',
  },
  {
    pattern: /^\/game\/([a-zA-Z0-9_-]+)$/,
    handler: (matches) => `/twitch-category/${matches[1]}`,
    description: 'Legacy Twitch category link',
  },
  {
    pattern: /^\/twitch-category\/([a-zA-Z0-9_-]+)$/,
    handler: (matches) => `/twitch-category/${matches[1]}`,
    description: 'Twitch category page',
  },
  {
    pattern: /^\/creator\/([a-zA-Z0-9_-]+)$/,
    handler: (matches) => `/creator/${matches[1]}`,
    description: 'Creator page',
  },
  {
    pattern: /^\/creator\/([a-zA-Z0-9_-]+)\/analytics$/,
    handler: (matches) => `/creator/${matches[1]}/analytics`,
    description: 'Creator analytics page',
  },
  {
    pattern: /^\/tag\/([a-zA-Z0-9_-]+)$/,
    handler: (matches) => `/tag/${matches[1]}`,
    description: 'Tag page',
  },
  {
    pattern: /^\/discover$/,
    handler: () => '/discover',
    description: 'Discovery page',
  },
  {
    pattern: /^\/new$/,
    handler: () => '/new',
    description: 'New clips feed',
  },
  {
    pattern: /^\/top$/,
    handler: () => '/top',
    description: 'Top clips feed',
  },
  {
    pattern: /^\/rising$/,
    handler: () => '/rising',
    description: 'Rising clips feed',
  },
];

/**
 * Check if a URL is a valid deep link
 */
export function isValidDeepLink(url: string): boolean {
  try {
    const parsedUrl = new URL(url);
    const path = parsedUrl.pathname;
    
    return DEEP_LINK_ROUTES.some(route => route.pattern.test(path));
  } catch {
    // Invalid URL
    return false;
  }
}

/**
 * Parse a deep link URL and return the internal route
 */
export function parseDeepLink(url: string): string | null {
  try {
    const parsedUrl = new URL(url);
    const path = parsedUrl.pathname;
    const searchParams = parsedUrl.search;
    
    for (const route of DEEP_LINK_ROUTES) {
      const matches = path.match(route.pattern);
      if (matches) {
        const internalPath = route.handler(matches);
        return searchParams ? `${internalPath}${searchParams}` : internalPath;
      }
    }
    
    return null;
  } catch {
    // Invalid URL
    return null;
  }
}

/**
 * Handle deep link navigation
 * Returns the internal route to navigate to, or null if not a valid deep link
 */
export function handleDeepLink(url: string): string | null {
  // parseDeepLink already validates and returns null for invalid links
  // No need for redundant isValidDeepLink check
  return parseDeepLink(url);
}

/**
 * Generate a deep link URL for a given internal route
 */
export function generateDeepLink(path: string, baseUrl?: string): string {
  const base = baseUrl || window.location.origin;
  return `${base}${path}`;
}

/**
 * Check if the app is opened via a deep link
 */
export function isOpenedViaDeepLink(): boolean {
  // Check if launched via share target
  // Share target should only trigger on specific paths (e.g., /submit)
  const urlParams = new URLSearchParams(window.location.search);
  const hasShareParams = urlParams.has('url') || urlParams.has('title') || urlParams.has('text');
  
  if (hasShareParams) {
    // Only consider it a share target if we're on a path that expects shared content
    const pathname = window.location.pathname;
    // Share target typically routes to /submit or root that redirects to /submit
    if (pathname === '/' || pathname === '/submit') {
      return true;
    }
  }
  
  // Check if referrer suggests external navigation
  const referrer = document.referrer;
  if (referrer && !referrer.startsWith(window.location.origin)) {
    return true;
  }
  
  return false;
}

/**
 * Extract share target data from URL parameters
 */
export function getShareTargetData(): { url?: string; title?: string; text?: string } | null {
  const urlParams = new URLSearchParams(window.location.search);
  
  const url = urlParams.get('url');
  const title = urlParams.get('title');
  const text = urlParams.get('text');
  
  if (!url && !title && !text) {
    return null;
  }
  
  return {
    url: url || undefined,
    title: title || undefined,
    text: text || undefined,
  };
}
