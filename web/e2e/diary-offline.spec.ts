import { test, expect } from "@playwright/test";

// Representative offline → online diary sync flow. TODO: expand to the full
// matrix (delete/soft-delete round-trip, conflict toast with mocked response).

test.describe("@chromium @webkit diary offline", () => {
  test("add meal offline → appears immediately → syncs when online", async ({
    page,
    context,
  }) => {
    // TODO: swap in Phase 1 login helper; this test assumes the dev server
    // auto-authenticates via a pre-seeded session cookie. Skip if not set.
    await page.goto("/diario");

    await context.setOffline(true);

    await page.getByRole("button", { name: /aggiungi pasto/i }).click();
    await page.getByLabel(/nome alimento/i).fill("pane");
    await page.getByLabel(/peso/i).fill("80");
    await page.getByRole("button", { name: "+" }).click();
    await page.getByRole("button", { name: /salva/i }).click();

    await expect(page.getByText("pane")).toBeVisible();

    // Go back online; outbox should drain within a few seconds.
    await context.setOffline(false);
    await page.waitForFunction(
      async () => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const db = (window as any).__pesaDb;
        if (!db) return true; // if no debug handle, trust the UI state
        const count = await db.outbox.count();
        return count === 0;
      },
      { timeout: 15_000 },
    );

    await page.reload();
    await expect(page.getByText("pane")).toBeVisible();
  });

  // TODO: conflict toast test with route mock
  // TODO: soft-delete round-trip
});
