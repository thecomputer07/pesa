import { db, type CachedFood } from "./schema";
import type { FoodSearchResult, SearchSource } from "../api";

export interface LocalSearchOptions {
  limit?: number;
  state?: string;
  source?: SearchSource;
}

/*
 * Local search over the foods stored on the device. Same rules as the server
 * (internal/repo/foods.go, SearchFoodsFiltered):
 *  - every word must match the start of a word of name/English name/synonyms,
 *    in any order, accents ignored, Italian stopwords dropped;
 *  - "te" / "the" / "tea" are the same word;
 *  - ranking: trigram similarity of the product name (brand stripped) + 5 when
 *    it starts with the first word + 2 when it is that word or contains it as
 *    a whole word + source boost (CREA 3, Ciqual/USDA 2);
 *  - one result per source and product name, preferring one with a photo;
 *    the user's own foods are never merged.
 */

const STOPWORDS = new Set([
  "di", "del", "della", "dello", "dei", "degli", "delle", "al", "alla", "allo",
  "agli", "alle", "con", "e", "o", "la", "il", "lo", "gli", "le", "un", "una",
  "uno", "in", "da", "a", "per",
]);

const SYNONYMS: Record<string, string[]> = {
  the: ["te", "the", "tea"],
  te: ["te", "the", "tea"],
  tea: ["te", "the", "tea"],
};

const SOURCE_BOOST: Record<string, number> = { crea: 3, ciqual: 2, usda: 2 };

function stripAccents(s: string): string {
  return s.normalize("NFD").replace(/\p{Mn}/gu, "");
}

/** Lowercase, letters/digits only, words separated by one space. */
function words(s: string): string {
  return s
    .toLowerCase()
    .normalize("NFC")
    .replace(/[^\p{L}\p{N}]+/gu, " ")
    .trim();
}

interface ParsedQuery {
  normQ: string;
  first: string;
  /** One list of accepted spellings per word, unaccented. */
  patterns: string[][];
}

export function parseQuery(raw: string): ParsedQuery | null {
  const kept = words(raw)
    .split(" ")
    .filter((t) => t !== "" && !STOPWORDS.has(t));
  if (kept.length === 0) return null;
  const patterns = kept.slice(0, 6).map((t) => {
    const u = stripAccents(t);
    return SYNONYMS[u] ?? [u];
  });
  return { normQ: kept.join(" "), first: kept[0], patterns };
}

/** pg_trgm trigrams: per word, padded with two spaces before and one after. */
function trigrams(s: string): Set<string> {
  const out = new Set<string>();
  for (const w of words(s).split(" ")) {
    if (!w) continue;
    const p = `  ${w} `;
    for (let i = 0; i + 3 <= p.length; i++) out.add(p.slice(i, i + 3));
  }
  return out;
}

function similarity(a: Set<string>, b: Set<string>): number {
  if (a.size === 0 || b.size === 0) return 0;
  let shared = 0;
  for (const t of a) if (b.has(t)) shared++;
  return shared / (a.size + b.size - shared);
}

interface Entry {
  row: CachedFood;
  /** " word word " haystack, unaccented: a word start is " " + prefix. */
  hay: string;
  /** Product name without the "Brand — " prefix, lowercased. */
  pname: string;
  dkey: string;
}

let index: Entry[] | null = null;
let building: Promise<Entry[]> | null = null;

/** Drop the in-memory index after the stored foods change. */
export function invalidateSearchIndex(): void {
  index = null;
  building = null;
}

function productName(name: string): string {
  const i = name.indexOf(" — ");
  return i >= 0 ? name.slice(i + 3) : name;
}

async function getIndex(): Promise<Entry[]> {
  if (index) return index;
  if (!building) {
    building = (async () => {
      // With the full catalog on the device, search the catalog and the user's
      // own foods; before that, whatever was cached by earlier online searches.
      const haveCatalog = (await db.meta.get("catalog.cursor")) !== undefined;
      const all = await db.foods.toArray();
      const rows = haveCatalog ? all.filter((f) => f.cat === 1 || f.mine === 1) : all;
      const built = rows.map((row): Entry => {
        const pname = productName(row.name_it).toLowerCase();
        const own = row.mine === 1 || row.source_code === "manual";
        return {
          row,
          hay: ` ${stripAccents(words([row.name_it, row.name_en ?? "", row.synonyms ?? ""].join(" ")))} `,
          pname,
          dkey: `${own ? `u${row.id}` : row.source_code}|${pname.replace(/\s+/g, " ").trim()}`,
        };
      });
      index = built;
      return built;
    })();
  }
  return building;
}

function matches(e: Entry, patterns: string[][]): boolean {
  for (const alts of patterns) {
    if (!alts.some((a) => e.hay.includes(` ${a}`))) return false;
  }
  return true;
}

function sourceMatches(row: CachedFood, source?: SearchSource): boolean {
  if (!source) return true;
  if (source === "mine") return row.mine === 1 || row.source_code === "manual";
  return row.source_code === source && row.mine !== 1;
}

export async function localSearch(query: string, opts: LocalSearchOptions = {}): Promise<CachedFood[]> {
  const q = parseQuery(query);
  if (!q) return [];
  const limit = Math.min(opts.limit ?? 20, 100);
  const entries = await getIndex();

  // Collapse duplicates: per dkey keep the row with a photo, then the lowest id.
  const best = new Map<string, Entry>();
  for (const e of entries) {
    if (opts.state && e.row.state !== opts.state) continue;
    if (!sourceMatches(e.row, opts.source)) continue;
    if (!matches(e, q.patterns)) continue;
    const prev = best.get(e.dkey);
    if (
      !prev ||
      (!prev.row.thumb_url && !!e.row.thumb_url) ||
      (!prev.row.thumb_url === !e.row.thumb_url && e.row.id < prev.row.id)
    ) {
      best.set(e.dkey, e);
    }
  }

  const qTri = trigrams(q.normQ);
  const scored = [...best.values()].map((e) => {
    let score = similarity(trigrams(e.pname), qTri);
    if (e.pname.startsWith(q.first)) score += 5;
    if (e.pname === q.first || e.pname.includes(` ${q.first} `)) score += 2;
    if (e.row.mine !== 1) score += SOURCE_BOOST[e.row.source_code] ?? 0;
    return { e, score };
  });
  scored.sort(
    (a, b) =>
      b.score - a.score ||
      a.e.pname.length - b.e.pname.length ||
      a.e.row.name_it.localeCompare(b.e.row.name_it, "it"),
  );
  return scored.slice(0, limit).map((s) => s.e.row);
}

/** Shape of an API search result, for components shared with online search. */
export function toSearchResult(r: CachedFood): FoodSearchResult {
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
    thumb_url: r.thumb_url,
    source: {
      code: r.source_code,
      license: r.source_license,
      license_url: r.source_license_url,
    },
  };
}

/** Count cached foods — used by the offline indicator / settings page. */
export async function cachedCount(): Promise<number> {
  return db.foods.count();
}
