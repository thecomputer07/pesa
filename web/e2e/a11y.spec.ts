import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

const ROUTES = ["/", "/cerca", "/scopo", "/accessibilita", "/chi-siamo"];

for (const route of ROUTES) {
  test(`a11y: ${route} has no serious or critical axe violations`, async ({ page }) => {
    await page.goto(route);
    // Wait for the main landmark to be present.
    await page.locator("main#main").waitFor();
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"])
      .analyze();

    const bad = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(bad, JSON.stringify(bad, null, 2)).toEqual([]);
  });
}
