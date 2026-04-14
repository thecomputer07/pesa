import type { FoodDetail } from "./api";

export type Nutrient =
  | "kcal"
  | "protein_g"
  | "fat_g"
  | "carb_g"
  | "sugar_g"
  | "fiber_g"
  | "water_g"
  | "salt_g";

export const NUTRIENTS: Nutrient[] = [
  "kcal",
  "protein_g",
  "fat_g",
  "carb_g",
  "sugar_g",
  "fiber_g",
  "water_g",
  "salt_g",
];

export interface PortionValues {
  grams: number;
  values: Partial<Record<Nutrient, number>>;
}

/**
 * Scale per-100g-edible values to the user's input grams.
 * The Italian CREA dataset expresses values per 100 g of EDIBLE portion;
 * `edible_pct` (0..100) is the yield of the raw/gross weight into edible weight.
 *
 * When `gramsAreEdible=true` (default) the user's grams are already edible,
 * so the scaling factor is `grams / 100`. When `false`, we first multiply
 * by `edible_pct/100` to convert gross → edible.
 */
export function scaleFromPer100g(
  food: Pick<FoodDetail, "kcal" | "protein_g" | "fat_g" | "carb_g" | "sugar_g" | "fiber_g" | "water_g" | "salt_g" | "edible_pct">,
  grams: number,
  opts: { gramsAreEdible?: boolean } = {},
): PortionValues {
  const edibleYield = opts.gramsAreEdible === false ? Math.max(0, food.edible_pct ?? 100) / 100 : 1;
  const effective = grams * edibleYield;
  const factor = effective / 100;

  const out: Partial<Record<Nutrient, number>> = {};
  for (const n of NUTRIENTS) {
    const v = (food as Partial<Record<Nutrient, number | undefined>>)[n];
    if (typeof v === "number" && Number.isFinite(v)) {
      out[n] = roundFor(n, v * factor);
    }
  }
  return { grams, values: out };
}

/** Round to a display-appropriate number of decimals per nutrient. */
function roundFor(n: Nutrient, value: number): number {
  const digits = n === "kcal" ? 0 : 1;
  const p = 10 ** digits;
  return Math.round(value * p) / p;
}

/**
 * Decide if two same-name foods from different sources are "in conflict"
 * (>10% difference in CHO per 100 g).
 */
export function carbConflict(
  a: Pick<FoodDetail, "carb_g">,
  b: Pick<FoodDetail, "carb_g">,
  thresholdPct = 10,
): boolean {
  const ca = a.carb_g;
  const cb = b.carb_g;
  if (typeof ca !== "number" || typeof cb !== "number") return false;
  if (ca === 0 && cb === 0) return false;
  const base = Math.max(Math.abs(ca), Math.abs(cb));
  const delta = Math.abs(ca - cb);
  return (delta / base) * 100 > thresholdPct;
}
