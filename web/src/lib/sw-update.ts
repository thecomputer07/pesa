/**
 * Service-worker update flow (prompt strategy).
 * Uses workbox-window to observe lifecycle; exposes a hook for InstallPrompt.
 */
import { Workbox, type WorkboxLifecycleWaitingEvent } from "workbox-window";

let wb: Workbox | null = null;
let onUpdateAvailable: ((accept: () => void) => void) | null = null;

export function setUpdateHandler(handler: (accept: () => void) => void): void {
  onUpdateAvailable = handler;
}

export async function registerSW(): Promise<void> {
  if (!("serviceWorker" in navigator)) return;

  wb = new Workbox("/sw.js");

  const showPrompt = (event: WorkboxLifecycleWaitingEvent) => {
    const accept = () => {
      if (!wb) return;
      wb.addEventListener("controlling", () => window.location.reload());
      wb.messageSkipWaiting();
    };
    if (onUpdateAvailable) onUpdateAvailable(accept);
    // Suppress "unused" warning for event parameter in minimal builds.
    void event;
  };

  // Since workbox-window v6 "waiting" also fires for a SW from another tab
  // (event.isExternal); the old "externalwaiting" event no longer exists.
  wb.addEventListener("waiting", showPrompt);

  try {
    await wb.register();
  } catch (err) {
    // eslint-disable-next-line no-console
    console.warn("[pesa] SW registration failed", err);
  }
}
