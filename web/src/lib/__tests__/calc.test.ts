import { describe, expect, it } from "vitest";
import { carbConflict, scaleFromPer100g } from "../calc";
import type { FoodDetail } from "../api";

const base: Pick<
  FoodDetail,
  "kcal" | "protein_g" | "fat_g" | "carb_g" | "sugar_g" | "fiber_g" | "water_g" | "salt_g" | "edible_pct"
> = {
  kcal: 265,
  protein_g: 9,
  fat_g: 3.2,
  carb_g: 49,
  sugar_g: 5,
  fiber_g: 2.7,
  water_g: 35,
  salt_g: 1.2,
  edible_pct: 100,
};

describe("scaleFromPer100g", () => {
  it("returns 0 for 0 grams", () => {
    const out = scaleFromPer100g(base, 0);
    expect(out.values.carb_g).toBe(0);
    expect(out.values.kcal).toBe(0);
  });

  it("returns per-100 equal to input when grams=100", () => {
    const out = scaleFromPer100g(base, 100);
    expect(out.values.carb_g).toBeCloseTo(49, 1);
    expect(out.values.kcal).toBe(265);
  });

  it("scales linearly", () => {
    const out = scaleFromPer100g(base, 50);
    expect(out.values.carb_g).toBeCloseTo(24.5, 1);
    expect(out.values.kcal).toBe(133); // 132.5 → rounded to 0 decimals
  });

  it("applies edible_pct when gramsAreEdible=false", () => {
    const out = scaleFromPer100g({ ...base, edible_pct: 80 }, 100, { gramsAreEdible: false });
    // 100 gross → 80 edible → 80% of per-100 values
    expect(out.values.carb_g).toBeCloseTo(49 * 0.8, 1);
  });

  it("ignores missing nutrients silently", () => {
    const out = scaleFromPer100g({ ...base, salt_g: undefined }, 200);
    expect(out.values.salt_g).toBeUndefined();
    expect(out.values.carb_g).toBeCloseTo(98, 1);
  });
});

describe("carbConflict", () => {
  it("flags >10% delta", () => {
    expect(carbConflict({ carb_g: 50 }, { carb_g: 56 })).toBe(true);
  });
  it("ignores ≤10% delta", () => {
    expect(carbConflict({ carb_g: 50 }, { carb_g: 54 })).toBe(false);
  });
  it("returns false when either value is missing", () => {
    expect(carbConflict({ carb_g: 50 }, { carb_g: undefined })).toBe(false);
  });
});
