import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import SourcesImagesPage from "../SourcesImagesPage";

vi.mock("react-i18next", () => ({
  useTranslation: () => ({ t: (k: string) => k }),
}));

const fixture = {
  sources: [
    { code: "crea", name: "CREA", license: "CC BY-NC-SA", attribution: "CREA", scope: "data" },
    { code: "off", name: "Open Food Facts", license: "ODbL-1.0", attribution: "OFF", scope: "data+images" },
    { code: "wikimedia", name: "Wikimedia", license: "per-file", attribution: "WM", scope: "images" },
  ],
};

beforeEach(() => {
  (globalThis as unknown as { fetch: typeof fetch }).fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => fixture,
  }) as unknown as typeof fetch;
});
afterEach(() => vi.restoreAllMocks());

describe("SourcesImagesPage", () => {
  it("shows OFF and Wikimedia but not CREA (images page)", async () => {
    render(<SourcesImagesPage />);
    await waitFor(() => expect(screen.getByText("Open Food Facts")).toBeTruthy());
    expect(screen.getByText("Wikimedia")).toBeTruthy();
    expect(screen.queryByText("CREA")).toBeNull();
  });

  it("shows the CREA/BDA-not-hosted note", async () => {
    render(<SourcesImagesPage />);
    await waitFor(() => expect(screen.getByText("Open Food Facts")).toBeTruthy());
    expect(screen.getByText("sources.imagesPage.creaBdaNote")).toBeTruthy();
  });

  it("contains no banned medical terms (GUARD-01)", async () => {
    render(<SourcesImagesPage />);
    await waitFor(() => expect(screen.getByText("Open Food Facts")).toBeTruthy());
    const banned = /(glicemia|insulin|diabete|bolo|I:CHO|FSI|CGM|pump)/i;
    expect(document.body.textContent || "").not.toMatch(banned);
  });
});
