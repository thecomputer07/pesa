import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import App from "./App";
import "./i18n/config";
import { queryClient } from "./lib/queryClient";
import { AuthProvider } from "./lib/auth/AuthContext";
import { initSentry } from "./lib/sentry";
import { registerSW } from "./lib/sw-update";
import { installVisibilityFallback, registerBackgroundSync } from "./lib/sync/bg-sync-register";
import { flushOutbox } from "./lib/sync/worker";
import { pullDelta } from "./lib/sync/pull";
import "./styles/tailwind.css";
import { refreshMyFoods, syncCatalog } from "./lib/catalog/sync";

initSentry();

// Phase 4: install visibility/online listeners + best-effort Background Sync.
// These are side-effect-only and never throw.
installVisibilityFallback();
void registerBackgroundSync();
if (typeof navigator !== "undefined" && navigator.onLine) {
  // Recipes become foods on the server: refresh the local copy after the sync.
  void pullDelta()
    .then(() => flushOutbox())
    .then(() => refreshMyFoods());
}

// Offline catalog: first run downloads it, later only the changes (throttled).
void syncCatalog();
if (typeof window !== "undefined") {
  window.addEventListener("online", () => void syncCatalog());
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <App />
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>,
);

// Register the service worker after mount (best-effort, PWA only in prod).
if (import.meta.env.PROD) {
  void registerSW();
}
