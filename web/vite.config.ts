import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: "prompt",
      // Phase 4: switched from generateSW → injectManifest so sw.ts can
      // handle Background Sync events in addition to workbox precaching.
      strategies: "injectManifest",
      srcDir: "src/sw",
      filename: "sw.ts",
      injectRegister: null,
      manifest: false,
      includeAssets: [
        "icon.svg",
        "favicon-32.png",
        "icons/icon-192.png",
        "icons/icon-512.png",
        "icons/icon-maskable-512.png",
        "icons/apple-touch-icon.png",
      ],
      injectManifest: {
        globPatterns: ["**/*.{js,css,html,svg,woff2,png,ico}"],
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,svg,woff2,png,ico}"],
        navigateFallback: "/index.html",
        navigateFallbackDenylist: [/^\/api\//, /^\/healthz/, /^\/readyz/],
        cleanupOutdatedCaches: true,
        clientsClaim: false,
        skipWaiting: false,
        runtimeCaching: [
          {
            urlPattern: ({ url }) => url.pathname.startsWith("/api/v1/foods/search"),
            handler: "NetworkFirst",
            options: {
              cacheName: "pesa-foods-search",
              networkTimeoutSeconds: 3,
              expiration: { maxEntries: 200, maxAgeSeconds: 60 * 60 * 24 * 7 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          {
            urlPattern: ({ url }) => /^\/api\/v1\/foods\/\d+$/.test(url.pathname),
            handler: "StaleWhileRevalidate",
            options: {
              cacheName: "pesa-foods-detail",
              expiration: { maxEntries: 500, maxAgeSeconds: 60 * 60 * 24 * 30 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          {
            urlPattern: ({ request }) => request.destination === "font",
            handler: "CacheFirst",
            options: {
              cacheName: "pesa-fonts",
              expiration: { maxEntries: 30, maxAgeSeconds: 60 * 60 * 24 * 365 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
        ],
      },
      devOptions: { enabled: false },
    }),
  ],
  build: {
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    port: 5173,
    host: "0.0.0.0",
    // VITE_API_PROXY_TARGET = "http://backend:8080" dentro Docker compose,
    // fallback localhost per dev nativo (senza container).
    proxy: (() => {
      const target = process.env.VITE_API_PROXY_TARGET || "http://localhost:8080";
      // changeOrigin: lets the target be a real vhost (e.g. https://pesadia.com).
      const p = { target, changeOrigin: true };
      return {
        "/api": p,
        "/healthz": p,
        "/readyz": p,
        "/verify-email": p,
        "/static": p,
      };
    })(),
  },
});
