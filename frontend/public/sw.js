// Service Worker for the clpr PWA
// Provides offline shell caching for static assets only
// Does NOT cache authenticated or sensitive API data

// Bump when precached files (offline page, icons) change so clients refresh them.
const CACHE_NAME = 'clpr-v4';
const OFFLINE_URL = '/offline.html';

// Static assets to cache for offline shell
const STATIC_ASSETS = [
  '/',
  '/offline.html',
  '/manifest.json',
  '/icons/icon-192x192.png',
  '/icons/icon-512x512.png',
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      console.log('[SW] Caching static assets');
      return cache.addAll(STATIC_ASSETS.map(url => new Request(url, { credentials: 'same-origin' })));
    })
    .catch(err => {
      console.error('[SW] Failed to cache static assets. Installation aborted:', err);
      // The install event will fail, and the service worker will not activate.
    })
  );
  // Updates wait for the page to request activation via SKIP_WAITING.
  // The first installation activates normally without forcing a page reload.
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            console.log('[SW] Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
  // Take control of all pages immediately
  return self.clients.claim();
});

// Fetch event - serve cached assets, but never cache API data
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Leave cross-origin requests (Twitch thumbnails, avatars, embeds) to the
  // browser. A fetch() from the worker is checked against connect-src, which
  // does not list image CDNs, so re-fetching them here fails every time.
  if (url.origin !== self.location.origin) {
    return;
  }

  // Leave API, auth and non-GET requests to the browser. Proxying them through
  // respondWith() never cached anything, but it made every API call appear
  // twice in DevTools and added a hop to each request.
  if (
    url.pathname.startsWith('/api/') ||
    url.pathname.startsWith('/auth/') ||
    request.method !== 'GET' ||
    request.headers.get('Authorization') ||
    request.headers.get('Cookie')
  ) {
    return;
  }

  // For static assets: Network first, fallback to cache
  // This ensures users get fresh content when online
  event.respondWith(
    fetch(request)
      .then((response) => {
        // Clone the response before caching
        const responseToCache = response.clone();
        
        if (response.status === 200) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(request, responseToCache);
          }).catch((err) => {
            console.error('[SW] Failed to cache response for', request.url, err);
          });
        }
        
        return response;
      })
      .catch(() => {
        // If network fails, try to serve from cache
        return caches.match(request).then((cachedResponse) => {
          if (cachedResponse) {
            return cachedResponse;
          }
          
          // If requesting a page and nothing cached, show offline page
          if (request.mode === 'navigate') {
            return caches.match(OFFLINE_URL);
          }
          
          // For other resources, return a generic offline response
          return new Response('Offline', {
            status: 503,
            statusText: 'Service Unavailable',
          });
        });
      })
  );
});

// Handle messages from the main app
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});
