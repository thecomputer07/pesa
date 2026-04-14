import { computeItemNutrients, type Nutrients } from "../diary/totals";

export interface RecipeItemInput {
  grams: number;
  /** Snapshot = per-100g nutrient basis frozen at recipe save time. */
  snapshot: Partial<Nutrients>;
}

export interface RecipeAggregate {
  total: Nutrients;
  perServing: Nutrients;
  per100g: Nutrients;
  totalGrams: number;
}

const Z: Nutrients = {
  kcal: 0,
  carb_g: 0,
  protein_g: 0,
  fat_g: 0,
  sugar_g: 0,
  fiber_g: 0,
};

export function aggregateRecipe(
  items: RecipeItemInput[],
  servings: number,
  totalGramsOverride?: number,
): RecipeAggregate {
  const total = items.reduce<Nutrients>((acc, it) => {
    const n = computeItemNutrients(it.grams, it.snapshot);
    return {
      kcal: acc.kcal + n.kcal,
      carb_g: acc.carb_g + n.carb_g,
      protein_g: acc.protein_g + n.protein_g,
      fat_g: acc.fat_g + n.fat_g,
      sugar_g: acc.sugar_g + n.sugar_g,
      fiber_g: acc.fiber_g + n.fiber_g,
    };
  }, { ...Z });

  const safeServings = servings > 0 ? servings : 1;
  const totalGrams =
    totalGramsOverride != null && totalGramsOverride > 0
      ? totalGramsOverride
      : items.reduce((s, i) => s + i.grams, 0);

  const perServing = scale(total, 1 / safeServings);
  const per100g = totalGrams > 0 ? scale(total, 100 / totalGrams) : { ...Z };
  return { total, perServing, per100g, totalGrams };
}

export function recipePer100g(items: RecipeItemInput[], totalGrams?: number): Nutrients {
  return aggregateRecipe(items, 1, totalGrams).per100g;
}

export function recipePerServing(items: RecipeItemInput[], servings: number): Nutrients {
  return aggregateRecipe(items, servings).perServing;
}

function scale(n: Nutrients, f: number): Nutrients {
  return {
    kcal: n.kcal * f,
    carb_g: n.carb_g * f,
    protein_g: n.protein_g * f,
    fat_g: n.fat_g * f,
    sugar_g: n.sugar_g * f,
    fiber_g: n.fiber_g * f,
  };
}
