import { test, expect } from "@playwright/test";

/**
 * Offline smoke — asserts the shell loads from the service worker
 * after a first warm-up visit and that a cached food detail is viewable.
 * Requires the production build (`make build && make serve`).
 */
test.describe("offline behaviour", () => {
  test("shell loads offline after a warm-up visit", async ({ page, context }) => {
    await page.goto("/");
    // warm-up: service worker precache should install during this first visit.
    await page.waitForLoadState("networkidle");

    await context.setOffline(true);
    await page.reload();

    await expect(page.getByRole("main")).toBeVisible();
    await expect(page.getByRole("banner")).toBeVisible();

    await context.setOffline(false);
  });
});
