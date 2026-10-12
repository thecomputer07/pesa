/// <reference lib="webworker" />
// Custom service worker for Pesa. Uses `injectManifest` strategy so we can
// handle Background Sync events in addition to workbox precaching.

import { cleanupOutdatedCaches, createHandlerBoundToURL, precacheAndRoute } from "workbox-precaching";
import { NavigationRoute, registerRoute } from "workbox-routing";
import { CacheFirst, StaleWhileRevalidate } from "workbox-strategies";
import { ExpirationPlugin } from "workbox-expiration";
import { CacheableResponsePlugin } from "workbox-cacheable-response";

declare const self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<string | { url: string; revision: string | null }>;
};

precacheAndRoute(self.__WB_MANIFEST);
cleanupOutdatedCaches();

// Any page opens offline: navigations get the precached app shell.
registerRoute(
  new NavigationRoute(createHandlerBoundToURL("/index.html"), {
    denylist: [/^\/api\//, /^\/healthz/, /^\/readyz/, /^\/static\//, /^\/legal\//],
  }),
);

// Food photos: same URL = same file (content hash), keep them once seen.
// Data (catalog, search, detail) is in IndexedDB, not in the HTTP cache.
registerRoute(
  ({ url }) => url.origin === self.location.origin && url.pathname.startsWith("/static/images/"),
  new CacheFirst({
    cacheName: "pesa-photos",
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }),
      new ExpirationPlugin({ maxEntries: 1000, maxAgeSeconds: 60 * 60 * 24 * 90, purgeOnQuotaError: true }),
    ],
  }),
);

// Photo by food id (offline catalog): the photo of a food can change, so
// serve the copy at once and refresh it in the background.
registerRoute(
  ({ url }) => url.origin === self.location.origin && url.pathname.startsWith("/static/thumbs/"),
  new StaleWhileRevalidate({
    cacheName: "pesa-thumbs",
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }),
      new ExpirationPlugin({ maxEntries: 1000, maxAgeSeconds: 60 * 60 * 24 * 90, purgeOnQuotaError: true }),
    ],
  }),
);

// Legal pages (privacy, cookie): readable offline, refreshed in the background.
registerRoute(
  ({ url }) => url.origin === self.location.origin && url.pathname.startsWith("/legal/"),
  new StaleWhileRevalidate({ cacheName: "pesa-legal" }),
);

// Background Sync: when the browser decides it's time to flush, message all
// open clients so they can run flushOutbox() (which holds the IndexedDB
// connection + CSRF token — both live in the window context, not here).
self.addEventListener("sync", (event: Event) => {
  const e = event as ExtendableEvent & { tag?: string };
  if (e.tag === "pesa-outbox-flush") {
    e.waitUntil(
      (async () => {
        const all = await self.clients.matchAll({
          includeUncontrolled: true,
          type: "window",
        });
        for (const client of all) {
          client.postMessage({ type: "pesa:flush" });
        }
      })(),
    );
  }
});

// Allow an explicit skipWaiting message (used by the existing update prompt).
self.addEventListener("message", (event: ExtendableMessageEvent) => {
  const data = event.data as { type?: string } | undefined;
  if (data?.type === "SKIP_WAITING") {
    void self.skipWaiting();
  }
});
