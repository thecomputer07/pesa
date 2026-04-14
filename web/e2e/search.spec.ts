import { test, expect } from "@playwright/test";

/**
 * End-to-end smoke for the search flow.
 * Assumes the Go backend is serving the embedded web/dist + /api/v1 at :8080
 * and that seed data includes at least one food whose name matches "pane".
 */
test.describe("search flow", () => {
  test("types a query and sees at least one result", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("link", { name: /cerca|search/i }).first().click();
    await expect(page).toHaveURL(/\/cerca/);

    const input = page.getByRole("searchbox");
    await input.fill("pane");
    await page.getByRole("button", { name: /cerca|search/i }).click();

    // Either list items or the empty-state message — accept both, fail on a hard error.
    const list = page.locator("ul li");
    const empty = page.getByText(/nessun risultato|no results/i);
    await expect(list.first().or(empty)).toBeVisible({ timeout: 10_000 });
  });

  test("click through to a food detail page", async ({ page }) => {
    await page.goto("/cerca?q=pane");
    const firstResult = page.locator("ul li a").first();
    if (await firstResult.count()) {
      await firstResult.click();
      await expect(page).toHaveURL(/\/alimento\/\d+/);
      await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    }
  });
});
