import { db, type CachedFood } from "./schema";
import { normalise, tokenize } from "./sync";

export interface LocalSearchOptions {
  limit?: number;
  state?: "crudo" | "cotto";
}

/**
 * Simple local FTS over the cached foods table.
 * Scoring:
 *   +3 if name_it starts with the query
 *   +2 per token match in name_it
 *   +1 per token match in category/name_en
 * Results with score 0 are excluded.
 */
export async function localSearch(
  query: string,
  opts: LocalSearchOptions = {},
): Promise<CachedFood[]> {
  const q = normalise(query);
  if (!q) return [];
  const tokens = tokenize(query);
  const limit = Math.min(opts.limit ?? 20, 100);

  // We can't do true FTS in Dexie; scan & score in memory. The cache is small
  // (thousands of rows), so this is fast enough for interactive search.
  const all = await db.foods.toArray();

  const scored: { row: CachedFood; score: number }[] = [];
  for (const row of all) {
    if (opts.state && row.state !== opts.state) continue;

    const nameIT = normalise(row.name_it);
    const nameEN = normalise(row.name_en ?? "");
    const cat = normalise(row.category ?? "");

    let score = 0;
    if (nameIT.startsWith(q)) score += 3;
    for (const t of tokens) {
      if (nameIT.includes(t)) score += 2;
      if (nameEN.includes(t)) score += 1;
      if (cat.includes(t)) score += 1;
    }
    if (score > 0) scored.push({ row, score });
  }

  scored.sort((a, b) => {
    if (b.score !== a.score) return b.score - a.score;
    return a.row.name_it.localeCompare(b.row.name_it, "it");
  });
  return scored.slice(0, limit).map((s) => s.row);
}

/** Count cached foods — used by the offline indicator / settings page. */
export async function cachedCount(): Promise<number> {
  return db.foods.count();
}
