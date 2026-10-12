import { apiFetch } from "../api";
import { db, type CachedFood } from "../db/schema";
import { getMeta, setMeta, normalise } from "../db/sync";
import { invalidateSearchIndex } from "../db/search";

/**
 * Offline catalog: the whole public food catalog lives in IndexedDB, so search,
 * food detail, calculations and known barcodes work without network.
 *
 *  - first run: GET /api/v1/catalog downloads everything (~19k foods, ~0.8 MB gzip);
 *  - later: at most every CHECK_EVERY_MS, GET /api/v1/catalog?since=<cursor>
 *    returns only what changed (often a 304: the ETag is the catalog version);
 *  - every FULL_EVERY_MS, or when the row count drifts from the server's,
 *    a full download replaces the local copy (catches deleted rows and photos).
 */

const CURSOR_KEY = "catalog.cursor";
const COUNT_KEY = "catalog.count";
const CHECKED_KEY = "catalog.checked_at";
const FULL_KEY = "catalog.full_at";
const SOURCES_KEY = "catalog.sources";
const ETAG_KEY = "catalog.etag";
const MINE_KEY = "catalog.mine_at";

const CHECK_EVERY_MS = 6 * 60 * 60 * 1000;
const FULL_EVERY_MS = 30 * 24 * 60 * 60 * 1000;

export interface CatalogSource {
  code: string;
  name: string;
  license: string;
  license_url?: string;
  attribution: string;
  source_url?: string;
}

interface CatalogResponse {
  cursor: number;
  count: number;
  full: boolean;
  columns: string[];
  rows: unknown[][];
  gone: number[];
  sources: CatalogSource[];
}

/** True once a full catalog has been stored on this device. */
export async function catalogReady(): Promise<boolean> {
  return (await getMeta(CURSOR_KEY)) !== undefined;
}

export async function catalogSources(): Promise<CatalogSource[]> {
  const raw = await getMeta(SOURCES_KEY);
  if (!raw) return [];
  try {
    return JSON.parse(raw) as CatalogSource[];
  } catch {
    return [];
  }
}

function num(v: unknown): number | undefined {
  return typeof v === "number" ? v : undefined;
}
function str(v: unknown): string | undefined {
  return typeof v === "string" && v !== "" ? v : undefined;
}

function toCached(cols: string[], row: unknown[], sources: Map<string, CatalogSource>): CachedFood {
  const v = (name: string) => row[cols.indexOf(name)];
  const code = String(v("source"));
  const src = sources.get(code);
  const nameIT = String(v("name_it"));
  const nameEN = str(v("name_en"));
  const synonyms = str(v("synonyms"));
  const category = str(v("category"));
  return {
    id: Number(v("id")),
    name_it: nameIT,
    name_en: nameEN,
    synonyms,
    category,
    state: String(v("state")),
    parent_id: num(v("parent_id")),
    edible_pct: num(v("edible_pct")),
    kcal: num(v("kcal")),
    protein_g: num(v("protein_g")),
    fat_g: num(v("fat_g")),
    carb_g: num(v("carb_g")),
    sugar_g: num(v("sugar_g")),
    fiber_g: num(v("fiber_g")),
    water_g: num(v("water_g")),
    ash_g: num(v("ash_g")),
    alcohol_g: num(v("alcohol_g")),
    salt_g: num(v("salt_g")),
    // Photo URL from the id: /static/thumbs/{id} serves the food's first image.
    thumb_url: v("has_thumb") === true ? `/static/thumbs/${Number(v("id"))}` : undefined,
    ean: str(v("ean")),
    last_imported_at: str(v("last_imported_date")),
    source_code: code,
    source_license: src?.license ?? "",
    source_license_url: src?.license_url,
    search_tokens: normalise([nameIT, nameEN ?? "", synonyms ?? "", category ?? ""].join(" ")),
    updated_at: new Date().toISOString(),
    cat: 1,
  };
}

let running: Promise<void> | null = null;

/**
 * Brings the local catalog up to date. Cheap to call often (app start, back
 * online): it does nothing until CHECK_EVERY_MS has passed, unless `force`.
 * Never throws — offline or server errors just leave the local copy as it is.
 */
export function syncCatalog(opts: { force?: boolean } = {}): Promise<void> {
  if (!running) {
    running = doSync(opts.force ?? false)
      .catch(() => undefined)
      .finally(() => {
        running = null;
      });
  }
  return running;
}

async function doSync(force: boolean): Promise<void> {
  if (typeof navigator !== "undefined" && !navigator.onLine) return;
  const now = Date.now();
  const cursor = Number((await getMeta(CURSOR_KEY)) ?? 0);
  const checkedAt = Number((await getMeta(CHECKED_KEY)) ?? 0);
  const fullAt = Number((await getMeta(FULL_KEY)) ?? 0);
  if (!force && cursor > 0 && now - checkedAt < CHECK_EVERY_MS) return;

  const full = cursor === 0 || now - fullAt > FULL_EVERY_MS;
  await fetchAndApply(full ? 0 : cursor);
  await setMeta(CHECKED_KEY, String(now));

  // A delta cannot report hard-deleted rows: if the counts disagree, start over.
  if (!full) {
    const expected = Number((await getMeta(COUNT_KEY)) ?? 0);
    const have = await db.foods.where("cat").equals(1).count();
    if (expected > 0 && have !== expected) await fetchAndApply(0);
  }
  requestPersistentStorage();
}

async function fetchAndApply(since: number): Promise<void> {
  const headers: Record<string, string> = {};
  const etag = await getMeta(ETAG_KEY);
  if (etag && since > 0) headers["If-None-Match"] = etag;
  const res = await apiFetch(since > 0 ? `/catalog?since=${since}` : "/catalog", { headers });
  if (res.status === 304) return;
  if (!res.ok) throw new Error(`catalog HTTP ${res.status}`);
  const body = (await res.json()) as CatalogResponse;
  const sources = new Map(body.sources.map((s) => [s.code, s]));
  const rows = body.rows.map((r) => toCached(body.columns, r, sources));

  await db.transaction("rw", db.foods, db.meta, async () => {
    if (body.full) {
      // Replace the catalog; keep the user's own foods. Rows cached by older
      // versions (no flag) are superseded by the catalog copy.
      await db.foods.filter((f) => f.mine !== 1).delete();
    } else if (body.gone.length > 0) {
      await db.foods.bulkDelete(body.gone);
    }
    await db.foods.bulkPut(rows);
    await db.meta.bulkPut([
      { key: CURSOR_KEY, value: String(body.cursor) },
      { key: COUNT_KEY, value: String(body.count) },
      { key: SOURCES_KEY, value: JSON.stringify(body.sources) },
      ...(body.full ? [{ key: FULL_KEY, value: String(Date.now()) }] : []),
    ]);
  });
  const newEtag = res.headers.get("ETag");
  if (newEtag) await setMeta(ETAG_KEY, newEtag);
  invalidateSearchIndex();
}

interface MineResponse {
  results: {
    id: number;
    name_it: string;
    name_en?: string;
    category?: string;
    state: string;
    kcal?: number;
    carb_g?: number;
    protein_g?: number;
    fat_g?: number;
    sugar_g?: number;
    fiber_g?: number;
    thumb_url?: string;
    recipe_client_id?: string;
    source: { code: string; license: string; license_url?: string };
  }[];
}

/**
 * Copies the signed-in user's own foods (custom foods and recipe foods) to the
 * device so they are found offline too. Private: removed by clearMyFoods() at
 * logout. `force` skips the "at most every CHECK_EVERY_MS" rule (after an edit).
 */
export async function syncMyFoods(userId: number, opts: { force?: boolean } = {}): Promise<void> {
  try {
    if (typeof navigator !== "undefined" && !navigator.onLine) return;
    // Stored as "<user id>:<ms>": another account on this device syncs at once.
    const [lastUser, lastAt] = ((await getMeta(MINE_KEY)) ?? "").split(":");
    const fresh = lastUser === String(userId) && Date.now() - Number(lastAt) < CHECK_EVERY_MS;
    if (!opts.force && fresh) return;
    const res = await apiFetch("/foods/mine?recipes=1");
    if (!res.ok) return;
    const body = (await res.json()) as MineResponse;
    const rows: CachedFood[] = body.results.map((r) => ({
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
      recipe_client_id: r.recipe_client_id,
      source_code: r.source.code,
      source_license: r.source.license,
      source_license_url: r.source.license_url,
      search_tokens: normalise([r.name_it, r.name_en ?? "", r.category ?? ""].join(" ")),
      updated_at: new Date().toISOString(),
      mine: 1,
    }));
    await db.transaction("rw", db.foods, db.meta, async () => {
      await db.foods.where("mine").equals(1).delete();
      await db.foods.bulkPut(rows);
      await db.meta.put({ key: MINE_KEY, value: `${userId}:${Date.now()}` });
    });
    invalidateSearchIndex();
  } catch {
    /* offline or signed out: keep what we have */
  }
}

/** Re-reads the user's own foods now (after creating/editing one, or a recipe sync). */
export async function refreshMyFoods(): Promise<void> {
  const [lastUser] = ((await getMeta(MINE_KEY).catch(() => undefined)) ?? "").split(":");
  if (lastUser) await syncMyFoods(Number(lastUser), { force: true });
}

/** Forgets the user's own foods on this device (logout). */
export async function clearMyFoods(): Promise<void> {
  try {
    await db.foods.where("mine").equals(1).delete();
    await db.meta.delete(MINE_KEY);
    invalidateSearchIndex();
  } catch {
    /* IndexedDB unavailable */
  }
}

/** Ask the browser not to evict the catalog under storage pressure. */
function requestPersistentStorage(): void {
  try {
    void navigator.storage?.persist?.();
  } catch {
    /* not supported */
  }
}
