import { describe, expect, it } from "vitest";
import { aggregateRecipe } from "../aggregate";

describe("aggregateRecipe", () => {
  it("computes total, per-serving, and per-100g", () => {
    const r = aggregateRecipe(
      [
        { grams: 200, snapshot: { carb_g: 50, kcal: 150 } },
        { grams: 100, snapshot: { carb_g: 4, kcal: 20 } },
      ],
      2,
    );
    expect(r.total.carb_g).toBeCloseTo(104, 3); // 100 + 4
    expect(r.perServing.carb_g).toBeCloseTo(52, 3);
    expect(r.per100g.carb_g).toBeCloseTo((104 / 300) * 100, 3);
    expect(r.totalGrams).toBe(300);
  });

  it("defaults servings=0 to 1 to avoid divide-by-zero", () => {
    const r = aggregateRecipe([{ grams: 100, snapshot: { kcal: 100 } }], 0);
    expect(r.perServing.kcal).toBeCloseTo(100, 5);
  });

  it("empty items returns all zeros with no NaN", () => {
    const r = aggregateRecipe([], 4);
    expect(Number.isFinite(r.per100g.kcal)).toBe(true);
    expect(r.total.kcal).toBe(0);
    expect(r.per100g.kcal).toBe(0);
  });

  it("honors totalGramsOverride when provided", () => {
    const r = aggregateRecipe([{ grams: 100, snapshot: { carb_g: 10 } }], 1, 200);
    expect(r.totalGrams).toBe(200);
    expect(r.per100g.carb_g).toBeCloseTo((10 / 200) * 100, 3); // 5
  });

  // TODO: property-based test for floating-point tolerance across 1000 items.
});
