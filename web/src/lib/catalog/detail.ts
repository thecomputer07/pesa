import type { FoodDetail } from "../api";
import { db, type CachedFood } from "../db/schema";
import { catalogSources } from "./sync";

/**
 * Food detail built from the offline catalog, so opening a public food needs
 * no request. Returns null when the food is not in the catalog (a custom food,
 * a recipe food, or no catalog yet): the caller then asks the server.
 */
export async function localFoodDetail(id: number): Promise<FoodDetail | null> {
  const row = await db.foods.get(id);
  if (!row || row.cat !== 1) return null;
  const src = (await catalogSources()).find((s) => s.code === row.source_code);
  if (!src) return null;

  // crudo/cotto: the foods sharing the same parent (or this food as parent).
  const root = row.parent_id ?? row.id;
  const family = await db.foods
    .where("parent_id")
    .equals(root)
    .toArray()
    .then(async (kids) => {
      const parent = row.parent_id ? await db.foods.get(root) : undefined;
      return parent ? [parent, ...kids] : kids;
    });
  const siblings = family
    .filter((f) => f.id !== row.id && f.cat === 1)
    .sort((a, b) => a.state.localeCompare(b.state) || a.id - b.id)
    .map((f) => ({ id: f.id, name_it: f.name_it, state: f.state }));

  return {
    id: row.id,
    name_it: row.name_it,
    name_en: row.name_en,
    category: row.category,
    state: row.state,
    edible_pct: row.edible_pct ?? 100,
    kcal: row.kcal,
    protein_g: row.protein_g,
    fat_g: row.fat_g,
    carb_g: row.carb_g,
    sugar_g: row.sugar_g,
    fiber_g: row.fiber_g,
    water_g: row.water_g,
    ash_g: row.ash_g,
    alcohol_g: row.alcohol_g,
    salt_g: row.salt_g,
    attribution: src.attribution,
    license: src.license,
    license_url: src.license_url,
    source_url: src.source_url,
    last_imported_at: row.last_imported_at,
    images: row.thumb_url ? [{ url: row.thumb_url, is_external: !row.thumb_url.startsWith("/") }] : [],
    thumb_url: row.thumb_url,
    siblings,
  };
}

/** Open Food Facts product with this barcode in the offline catalog, if any. */
export async function findFoodByEan(ean: string): Promise<CachedFood | undefined> {
  const digits = ean.replace(/^0+/, "");
  if (!digits) return undefined;
  // Stored as OFF writes it (usually 13 digits); compare without leading zeros.
  for (const candidate of new Set([ean, digits, digits.padStart(13, "0"), digits.padStart(8, "0")])) {
    const hit = await db.foods.where("ean").equals(candidate).first();
    if (hit) return hit;
  }
  return undefined;
}
