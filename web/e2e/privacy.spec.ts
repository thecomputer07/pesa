import { expect, test } from "@playwright/test";

// E2E guards for the /io/privacy surface (PRIV-01/02/05).
// The backend routes are stubbed via route interception so the suite is
// self-contained and doesn't depend on a live MariaDB.

test.describe("privacy surface", () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([
      {
        name: "pesa_session",
        value: "fixture-session",
        url: "http://localhost:5173",
        httpOnly: true,
      },
      {
        name: "pesa_csrf",
        value: "fixture-csrf",
        url: "http://localhost:5173",
      },
    ]);
  });

  test("export panel: request → poll → ready → download link", async ({
    page,
  }) => {
    let calls = 0;
    await page.route("**/api/v1/me/export", async (route) => {
      await route.fulfill({
        status: 202,
        contentType: "application/json",
        body: JSON.stringify({ job_id: 1 }),
      });
    });
    await page.route("**/api/v1/me/export/1", async (route) => {
      calls += 1;
      const body =
        calls === 1
          ? { status: "pending" }
          : {
              status: "ready",
              bytes: 1024,
              download_url: "/api/v1/me/export/download?t=AAAA",
              expires_at: "2026-04-20T00:00:00Z",
            };
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(body),
      });
    });

    await page.goto("/io/privacy");
    await page.getByRole("button", { name: /esporta|export/i }).click();
    await expect(page.getByTestId("download-export")).toBeVisible({
      timeout: 10_000,
    });
  });

  test("delete panel: typed confirmation gates the submit button", async ({
    page,
  }) => {
    await page.goto("/io/privacy");
    await page.getByTestId("delete-start").click();
    await page.getByTestId("delete-password").fill("correct-horse-battery");
    await page.getByRole("button", { name: /continua|continue/i }).click();

    const confirmInput = page.getByTestId("delete-confirm-input");
    const submit = page.getByTestId("delete-submit");

    await confirmInput.fill("ELIMIN");
    await expect(submit).toBeDisabled();

    await confirmInput.fill("elimina");
    await expect(submit).toBeDisabled();

    await confirmInput.fill("ELIMINA");
    await expect(submit).toBeEnabled();
  });
});
