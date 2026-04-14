import * as Sentry from "@sentry/react";

/**
 * GlitchTip is Sentry-SDK-compatible. DSN comes from Vite env at build time.
 * Opt-out: cookie `pesa_telemetry_opt_out=1`.
 */
function hasOptOut(): boolean {
  if (typeof document === "undefined") return true;
  return /(?:^|;\s*)pesa_telemetry_opt_out=1(?:;|$)/.test(document.cookie);
}

export function initSentry(): void {
  const dsn = import.meta.env.VITE_SENTRY_DSN as string | undefined;
  if (!dsn || hasOptOut()) return;

  Sentry.init({
    dsn,
    environment: import.meta.env.MODE,
    release: (import.meta.env.VITE_APP_VERSION as string | undefined) ?? "0.0.1",
    tracesSampleRate: 0,
    replaysSessionSampleRate: 0,
    replaysOnErrorSampleRate: 0,
    // GlitchTip does not support all Sentry features; keep minimal.
    integrations: [],
    beforeSend(event) {
      // Strip any accidental PII from URLs/query.
      if (event.request?.url) {
        try {
          const u = new URL(event.request.url);
          u.searchParams.forEach((_v, k) => {
            if (["email", "password", "token"].includes(k.toLowerCase())) {
              u.searchParams.set(k, "[redacted]");
            }
          });
          event.request.url = u.toString();
        } catch {
          /* ignore */
        }
      }
      return event;
    },
  });
}

export { Sentry };
