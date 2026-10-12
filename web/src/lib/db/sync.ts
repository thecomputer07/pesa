import { db, type CachedFood } from "./schema";
import type { FoodSearchResult } from "../api";
import { invalidateSearchIndex } from "./search";

const ETAG_KEY = "foods.etag";

/** Normalise text → lowercase, NFD-strip accents, trim punctuation. */
export function normalise(s: string): string {
  return s
    .toLocaleLowerCase("it-IT")
    .normalize("NFD")
    .replace(/\p{Diacritic}/gu, "")
    .replace(/[^\p{Letter}\p{Number}\s]/gu, " ")
    .replace(/\s+/g, " ")
    .trim();
}

export function tokenize(s: string): string[] {
  const n = normalise(s);
  if (!n) return [];
  return Array.from(new Set(n.split(" ").filter((t) => t.length >= 2)));
}

function toCached(r: FoodSearchResult): CachedFood {
  const joined = [r.name_it, r.name_en ?? "", r.category ?? ""].join(" ");
  return {
    id: r.id,
    name_it: r.name_it,
    name_en: r.name_en,
    category: r.category,
    state: r.state,
    kcal: r.kcal,
    carb_g: r.carb_g,
    protein_g: r.protein_g,
    fat_g: r.fat_g,
    sugar_g: r.sugar_g,
    fiber_g: r.fiber_g,
    source_code: r.source.code,
    source_license: r.source.license,
    source_license_url: r.source.license_url,
    search_tokens: normalise(joined),
    updated_at: new Date().toISOString(),
  };
}

/**
 * Caches foods seen in online search results, for offline use before the full
 * catalog is on the device. Rows from the catalog or the user's own foods are
 * richer and stay as they are.
 */
export async function upsertFoods(rows: FoodSearchResult[]): Promise<number> {
  if (rows.length === 0) return 0;
  const existing = await db.foods.bulkGet(rows.map((r) => r.id));
  const mapped = rows
    .filter((_, i) => !existing[i]?.cat && !existing[i]?.mine)
    .map(toCached);
  if (mapped.length === 0) return 0;
  await db.foods.bulkPut(mapped);
  invalidateSearchIndex();
  return mapped.length;
}

export async function getMeta(key: string): Promise<string | undefined> {
  const row = await db.meta.get(key);
  return row?.value;
}

export async function setMeta(key: string, value: string): Promise<void> {
  await db.meta.put({ key, value });
}

export async function getEtag(): Promise<string | undefined> {
  return getMeta(ETAG_KEY);
}

export async function setEtag(etag: string): Promise<void> {
  await setMeta(ETAG_KEY, etag);
}
