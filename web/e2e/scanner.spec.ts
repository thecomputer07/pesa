import { test, expect } from "@playwright/test";

/**
 * Scanner e2e. The real camera is unavailable in CI, so we install a mock for
 * `navigator.mediaDevices.getUserMedia` and drive the decoder through the
 * `window.__pesaScanResult` test hook the BarcodeScanner component attaches.
 */

test.describe("scanner", () => {
  test("happy path: scan navigates to /alimento/off-<ean>", async ({ page }) => {
    await page.addInitScript(() => {
      // Fake a MediaStream by returning a dummy object with getVideoTracks.
      const fakeTrack = {
        stop() {},
        getCapabilities() {
          return {};
        },
        applyConstraints: async () => {},
      } as unknown as MediaStreamTrack;
      const fakeStream = {
        getTracks: () => [fakeTrack],
        getVideoTracks: () => [fakeTrack],
      } as unknown as MediaStream;
      (navigator as unknown as { mediaDevices: MediaDevices }).mediaDevices = {
        getUserMedia: async () => fakeStream,
        enumerateDevices: async () => [],
      } as unknown as MediaDevices;
    });

    await page.route("**/api/v1/foods/barcode/8001505005707", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          id: 0,
          name_it: "Baci Perugina",
          source: { code: "off", license: "ODbL-1.0" },
          license: "ODbL-1.0",
          state: "packaged",
          edible_pct: 100,
          attribution: "Open Food Facts contributors — ODbL",
          images: [],
        }),
      });
    });

    await page.goto("/scan");
    await page.getByText(/scanner\.promptCta|Attiva fotocamera/i).click();
    // Wait until the BarcodeScanner has mounted and exposed the hook.
    await page.waitForFunction(
      () => typeof (window as unknown as { __pesaScanResult?: unknown }).__pesaScanResult === "function",
    );
    await page.evaluate(() => {
      (
        window as unknown as { __pesaScanResult: (ean: string) => void }
      ).__pesaScanResult("8001505005707");
    });
    await expect(page).toHaveURL(/\/alimento\/off-8001505005707$/);
  });

  test("denied path: shows manual entry and toasts on unknown EAN", async ({ page }) => {
    await page.addInitScript(() => {
      (navigator as unknown as { mediaDevices: MediaDevices }).mediaDevices = {
        getUserMedia: async () => {
          throw Object.assign(new Error("denied"), { name: "NotAllowedError" });
        },
        enumerateDevices: async () => [],
      } as unknown as MediaDevices;
    });
    await page.route("**/api/v1/foods/barcode/**", async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/problem+json",
        body: JSON.stringify({ title: "barcode-not-found", status: 404 }),
      });
    });

    await page.goto("/scan");
    await page.getByText(/scanner\.promptCta|Attiva fotocamera/i).click();
    await page.getByLabel(/scanner\.manualLabel|inserisci codice/i).fill("8001505005707");
    await page.getByText(/scanner\.manualSubmit|invia/i).click();
    await expect(page.getByRole("status")).toBeVisible();
  });
});
