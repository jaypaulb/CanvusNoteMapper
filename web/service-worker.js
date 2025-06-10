// NO-OP Service Worker - Disables all caching for development
// This service worker intentionally does nothing to prevent caching issues during development

console.log('Service Worker: Caching disabled for development');

// Install event - no caching
self.addEventListener('install', event => {
  console.log('Service Worker: Install - No caching');
  self.skipWaiting(); // Immediately activate
});

// Activate event - clear all existing caches
self.addEventListener('activate', event => {
  console.log('Service Worker: Activate - Clearing all caches');
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          console.log('Service Worker: Deleting cache:', cacheName);
          return caches.delete(cacheName);
        })
      );
    }).then(() => {
      return self.clients.claim();
    })
  );
});

// Fetch event - always fetch from network, never cache
self.addEventListener('fetch', event => {
  console.log('Service Worker: Fetching from network (no cache):', event.request.url);
  event.respondWith(
    fetch(event.request, {
      cache: 'no-store', // Force fresh fetch every time
      headers: {
        'Cache-Control': 'no-cache',
        'Pragma': 'no-cache'
      }
    })
  );
});

// Send message to all clients that caching is disabled
self.addEventListener('message', event => {
  if (event.data && event.data.type === 'GET_VERSION') {
    event.ports[0].postMessage({
      version: 'no-cache-dev-mode',
      caching: false
    });
  }
});



