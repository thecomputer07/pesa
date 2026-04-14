import { describe, expect, it } from "vitest";
import { computeDayTotals, computeItemNutrients } from "../totals";

describe("computeItemNutrients", () => {
  it("scales linearly from per-100g to grams", () => {
    const n = computeItemNutrients(200, { kcal: 100, carb_g: 50 });
    expect(n.kcal).toBeCloseTo(200, 5);
    expect(n.carb_g).toBeCloseTo(100, 5);
  });

  it("returns zeros when grams is zero or negative", () => {
    const n = computeItemNutrients(0, { kcal: 100 });
    expect(n.kcal).toBe(0);
  });

  it("treats missing per-100g keys as zero (no NaN)", () => {
    const n = computeItemNutrients(100, {});
    expect(Number.isFinite(n.kcal)).toBe(true);
    expect(n.kcal).toBe(0);
  });
});

describe("computeDayTotals", () => {
  it("sums nutrients across items and excludes tombstones", () => {
    const totals = computeDayTotals([
      { grams: 100, food: { kcal: 100, carb_g: 20 } },
      { grams: 50, food: { kcal: 200, carb_g: 40 } },
      { grams: 100, food: { kcal: 999 }, deleted_at_ms: 1 }, // excluded
    ]);
    expect(totals.kcal).toBeCloseTo(100 + 100, 5);
    expect(totals.carb_g).toBeCloseTo(20 + 20, 5);
  });

  it("prefers snapshot over food (recipe snapshot-freeze)", () => {
    const totals = computeDayTotals([
      { grams: 100, snapshot: { carb_g: 50 }, food: { carb_g: 99 } },
    ]);
    expect(totals.carb_g).toBeCloseTo(50, 5);
  });

  it("returns all zeros for empty input", () => {
    const t = computeDayTotals([]);
    expect(t).toEqual({ kcal: 0, carb_g: 0, protein_g: 0, fat_g: 0, sugar_g: 0, fiber_g: 0 });
  });

  // TODO: property-based test for float summation tolerance over 1000 items.
});
