import { test, expect } from "@playwright/test";

test("home renders Pesa heading", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Pesa" })).toBeVisible();
});

test("healthz returns ok", async ({ request }) => {
  const r = await request.get("/healthz");
  expect(r.status()).toBe(200);
  const body = await r.json();
  expect(body.status).toBe("ok");
});
