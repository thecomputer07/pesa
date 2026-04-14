import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import SourcesDataPage from "../SourcesDataPage";

vi.mock("react-i18next", () => ({
  useTranslation: () => ({ t: (k: string) => k }),
}));

const fixture = {
  sources: [
    { code: "crea", name: "CREA", license: "CC BY-NC-SA", attribution: "CREA", scope: "data", license_url: "https://crea.example" },
    { code: "bda-ieo", name: "BDA IEO", license: "cit-nc", attribution: "BDA", scope: "data" },
    { code: "off", name: "Open Food Facts", license: "ODbL-1.0", attribution: "OFF", scope: "data+images", license_url: "https://odbl.example" },
    { code: "wikimedia", name: "Wikimedia", license: "per-file", attribution: "WM", scope: "images" },
  ],
};

beforeEach(() => {
  (global as unknown as { fetch: typeof fetch }).fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => fixture,
  }) as unknown as typeof fetch;
});
afterEach(() => {
  vi.restoreAllMocks();
});

describe("SourcesDataPage", () => {
  it("renders CREA, BDA-IEO and OFF but hides Wikimedia", async () => {
    render(<SourcesDataPage />);
    await waitFor(() => expect(screen.getByText("CREA")).toBeTruthy());
    expect(screen.getByText("BDA IEO")).toBeTruthy();
    expect(screen.getByText("Open Food Facts")).toBeTruthy();
    expect(screen.queryByText("Wikimedia")).toBeNull();
  });

  it("external links use rel=noopener noreferrer and target=_blank", async () => {
    render(<SourcesDataPage />);
    await waitFor(() => expect(screen.getByText("CREA")).toBeTruthy());
    const link = screen.getByText("https://crea.example") as HTMLAnchorElement;
    expect(link.getAttribute("rel")).toBe("noopener noreferrer");
    expect(link.getAttribute("target")).toBe("_blank");
  });

  it("contains no banned medical terms (GUARD-01)", async () => {
    render(<SourcesDataPage />);
    await waitFor(() => expect(screen.getByText("CREA")).toBeTruthy());
    const banned = /(glicemia|insulin|diabete|bolo|I:CHO|FSI|CGM|pump)/i;
    expect(document.body.textContent || "").not.toMatch(banned);
  });
});
