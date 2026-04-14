import { db, type CachedFood } from "./schema";
import { searchFoods, type FoodSearchResult } from "../api";

const ETAG_KEY = "foods.etag";
const LAST_SYNC_KEY = "foods.last_sync_at";

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

/** Upsert a batch of foods into the local cache. */
export async function upsertFoods(rows: FoodSearchResult[]): Promise<number> {
  if (rows.length === 0) return 0;
  const mapped = rows.map(toCached);
  await db.foods.bulkPut(mapped);
  return mapped.length;
}

/**
 * Warm-cache the first N pages of the catalog for offline use.
 * Uses the existing /search endpoint with an empty-ish query until a
 * dedicated /catalog endpoint ships (Phase 4).
 */
export async function warmCache(
  seedQueries: string[] = ["a", "e", "i", "o", "u"],
  perQuery = 50,
  signal?: AbortSignal,
): Promise<{ imported: number }> {
  let imported = 0;
  for (const q of seedQueries) {
    if (signal?.aborted) break;
    try {
      const resp = await searchFoods(q, { limit: perQuery, page: 1, signal });
      imported += await upsertFoods(resp.results);
    } catch {
      // network issue — skip; partial warm-cache is fine
    }
  }
  await setMeta(LAST_SYNC_KEY, new Date().toISOString());
  return { imported };
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
