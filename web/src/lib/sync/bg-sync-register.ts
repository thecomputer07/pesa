// Registers the Background Sync tag (Chromium) and installs the iOS-standalone
// visibility fallback + 'online' listener. Imported by both outbox.enqueue
// (to request a flush on each write) and main.tsx (to install listeners once).

const BG_SYNC_TAG = "pesa-outbox-flush";
let fallbackInstalled = false;

/**
 * Ask the SW to schedule a Background Sync flush. No-op when the SyncManager
 * API is unavailable (Safari, private browsing, etc.). Safe to call repeatedly.
 */
export async function registerBackgroundSync(): Promise<void> {
  if (typeof navigator === "undefined" || !("serviceWorker" in navigator)) return;
  try {
    const reg = await navigator.serviceWorker.ready;
    const sync = (reg as unknown as { sync?: { register: (t: string) => Promise<void> } }).sync;
    if (sync && typeof sync.register === "function") {
      await sync.register(BG_SYNC_TAG);
    }
  } catch {
    // best-effort
  }
}

/**
 * Install the iOS-standalone fallback: when the tab becomes visible and the
 * device is online, drain the outbox + pull delta. Also listens for the
 * 'online' event and forwards `pesa:flush` messages from the SW.
 *
 * Functions `flushOutbox` and `pullDelta` are dynamically imported to avoid
 * circular refs with worker.ts / pull.ts.
 */
export function installVisibilityFallback(): void {
  if (fallbackInstalled || typeof document === "undefined") return;
  fallbackInstalled = true;

  const trigger = async () => {
    if (typeof navigator !== "undefined" && navigator.onLine === false) return;
    const [{ flushOutbox }, { pullDelta }] = await Promise.all([
      import("./worker"),
      import("./pull"),
    ]);
    void pullDelta().finally(() => void flushOutbox());
  };

  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") void trigger();
  });
  window.addEventListener("online", () => void trigger());
  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.addEventListener("message", (e: MessageEvent) => {
      const data = e.data as { type?: string } | undefined;
      if (data?.type === "pesa:flush") void trigger();
    });
  }
}
