/// <reference lib="webworker" />
// Custom service worker for Pesa. Uses `injectManifest` strategy so we can
// handle Background Sync events in addition to workbox precaching.

import { precacheAndRoute } from "workbox-precaching";

declare const self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<string | { url: string; revision: string | null }>;
};

precacheAndRoute(self.__WB_MANIFEST);

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
