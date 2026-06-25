var CACHE = 'kayis-rahman-v2';
var STATIC = [
  '/',
  '/about/',
  '/resume/',
  '/posts/',
  '/assets/main.css',
  '/assets/scripts.js'
];

self.addEventListener('install', function (e) {
  e.waitUntil(
    caches.open(CACHE).then(function (cache) {
      return cache.addAll(STATIC);
    })
  );
  self.skipWaiting();
});

self.addEventListener('activate', function (e) {
  e.waitUntil(
    caches.keys().then(function (keys) {
      return Promise.all(
        keys.filter(function (k) { return k !== CACHE; }).map(function (k) {
          return caches.delete(k);
        })
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', function (e) {
  // Only handle GET requests for same-origin resources
  if (e.request.method !== 'GET' || !e.request.url.startsWith(self.location.origin)) return;

  // Skip caching dynamic content
  if (e.request.url.endsWith('/search.json')) return;

  e.respondWith(
    caches.match(e.request).then(function (cached) {
      var networkFetch = fetch(e.request).then(function (response) {
        // Only cache successful responses (200)
        if (response.ok && response.status === 200) {
          var clone = response.clone();
          caches.open(CACHE).then(function (cache) { cache.put(e.request, clone); });
        }
        return response;
      });
      // Return cache immediately, update in background (stale-while-revalidate)
      return cached || networkFetch;
    }).catch(function () {
      return caches.match('/');
    })
  );
});
