import { test, expect, type Page } from "@playwright/test";

/**
 * Airplane mode after a first online visit: search, food detail, calculation
 * and a known barcode must work with no network (catalog in IndexedDB, app
 * shell from the service worker). Needs a production build with real data
 * (e.g. PLAYWRIGHT_BASE_URL=https://pesadia.com).
 *
 *   OFFLINE_EAN  barcode of a product in the catalog (default: Nutella 8000500357781)
 */
const EAN = process.env.OFFLINE_EAN ?? "8000500357781";

async function catalogCount(page: Page): Promise<number> {
  return page.evaluate(
    () =>
      new Promise<number>((resolve) => {
        const open = indexedDB.open("pesa");
        open.onerror = () => resolve(0);
        open.onsuccess = () => {
          const db = open.result;
          if (!db.objectStoreNames.contains("foods")) return resolve(0);
          const req = db.transaction("foods").objectStore("foods").count();
          req.onsuccess = () => resolve(req.result);
          req.onerror = () => resolve(0);
        };
      }),
  );
}

test("works in airplane mode after the first visit", async ({ page, context }) => {
  test.setTimeout(120_000);
  await page.goto("/");
  // Service worker installed and controlling the page (needs one reload).
  await page.evaluate(() => navigator.serviceWorker.ready);
  await page.reload();
  await expect.poll(() => page.evaluate(() => !!navigator.serviceWorker.controller)).toBe(true);
  // Catalog downloaded into IndexedDB.
  await expect.poll(() => catalogCount(page), { timeout: 60_000 }).toBeGreaterThan(10_000);

  await context.setOffline(true);

  // Search: deep link opened offline, results from the device.
  await page.goto("/cerca?q=pasta%20semola");
  const first = page.locator("ul li a").first();
  await expect(first).toBeVisible();
  await expect(first).toContainText(/pasta/i);

  // Detail + calculation.
  await first.click();
  await expect(page).toHaveURL(/\/alimento\/\d+/);
  await expect(page.getByRole("heading", { level: 1 })).toContainText(/pasta/i);
  await expect(page.getByText(/carboidrati/i).first()).toBeVisible();

  // Known barcode, typed by hand (no camera in the test browser).
  await page.goto("/scan");
  await page.getByRole("textbox", { name: /inserisci codice a barre/i }).fill(EAN);
  await page.getByRole("textbox", { name: /inserisci codice a barre/i }).press("Enter");
  await expect(page).toHaveURL(/\/alimento\/\d+/);
  await expect(page.getByRole("heading", { level: 1 })).toContainText(/nutella/i);

  await context.setOffline(false);
});
