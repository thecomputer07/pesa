import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import App from "../src/App";
import "../src/i18n/config";
import { AuthProvider } from "../src/lib/auth/AuthContext";

// Smoke: the whole app mounts with the same providers as main.tsx and the
// home page renders (heading + search form). The network is down: the auth
// refresh must fall back to "logged out" without breaking the render.
beforeEach(() => {
  globalThis.fetch = vi.fn().mockRejectedValue(new TypeError("offline")) as unknown as typeof fetch;
});

function renderApp() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={["/"]}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("App", () => {
  it("renders the home page heading", async () => {
    renderApp();
    expect(await screen.findByRole("heading", { level: 1 })).toBeTruthy();
  });

  it("renders the home search form", async () => {
    renderApp();
    expect(await screen.findByRole("search")).toBeTruthy();
  });
});
