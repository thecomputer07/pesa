// Pure nutrient math. No Dexie, no React — unit-testable without DOM.

export interface Nutrients {
  kcal: number;
  carb_g: number;
  protein_g: number;
  fat_g: number;
  sugar_g: number;
  fiber_g: number;
}

const EMPTY: Nutrients = {
  kcal: 0,
  carb_g: 0,
  protein_g: 0,
  fat_g: 0,
  sugar_g: 0,
  fiber_g: 0,
};

/** Scale per-100g nutrient values linearly to `grams`. */
export function computeItemNutrients(
  grams: number,
  per100: Partial<Nutrients>,
): Nutrients {
  if (!(grams > 0)) return { ...EMPTY };
  const f = grams / 100;
  return {
    kcal: (per100.kcal ?? 0) * f,
    carb_g: (per100.carb_g ?? 0) * f,
    protein_g: (per100.protein_g ?? 0) * f,
    fat_g: (per100.fat_g ?? 0) * f,
    sugar_g: (per100.sugar_g ?? 0) * f,
    fiber_g: (per100.fiber_g ?? 0) * f,
  };
}

export interface DayTotalsItem {
  grams: number;
  /** Snapshot wins over food when present (recipe_item / meal_template items). */
  snapshot?: Partial<Nutrients>;
  /** Per-100g values from the live foods cache (fallback when no snapshot). */
  food?: Partial<Nutrients>;
  deleted_at_ms?: number | null;
}

/** Sum nutrients across items, excluding tombstones. */
export function computeDayTotals(items: DayTotalsItem[]): Nutrients {
  return items
    .filter((i) => !i.deleted_at_ms)
    .reduce<Nutrients>((acc, i) => {
      const per100 = i.snapshot ?? i.food ?? {};
      const n = computeItemNutrients(i.grams, per100);
      return {
        kcal: acc.kcal + n.kcal,
        carb_g: acc.carb_g + n.carb_g,
        protein_g: acc.protein_g + n.protein_g,
        fat_g: acc.fat_g + n.fat_g,
        sugar_g: acc.sugar_g + n.sugar_g,
        fiber_g: acc.fiber_g + n.fiber_g,
      };
    }, { ...EMPTY });
}

export const _empty = EMPTY;
