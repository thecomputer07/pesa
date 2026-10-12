import "fake-indexeddb/auto";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { db } from "../../db/schema";
import { localSearch } from "../../db/search";
import { catalogReady, clearMyFoods, syncCatalog } from "../sync";
import { findFoodByEan, localFoodDetail } from "../detail";

const COLUMNS = [
  "id", "name_it", "name_en", "synonyms", "category", "state", "source",
  "parent_id", "edible_pct", "kcal", "protein_g", "fat_g", "carb_g",
  "sugar_g", "fiber_g", "water_g", "ash_g", "alcohol_g", "salt_g",
  "has_thumb", "ean", "last_imported_date",
];

function row(id: number, name: string, source: string, extra: Partial<Record<string, unknown>> = {}) {
  const v: Record<string, unknown> = {
    id, name_it: name, name_en: null, synonyms: null, category: null, state: "raw",
    source, parent_id: null, edible_pct: 100, kcal: 100, protein_g: 1, fat_g: 1,
    carb_g: 20, sugar_g: 1, fiber_g: 1, water_g: null, ash_g: null, alcohol_g: null,
    salt_g: null, has_thumb: false, ean: null, last_imported_date: null, ...extra,
  };
  return COLUMNS.map((c) => v[c]);
}

const SOURCES = [
  { code: "crea", name: "CREA", license: "CC-BY-3.0", attribution: "CREA Alimenti e Nutrizione" },
  { code: "off", name: "Open Food Facts", license: "ODbL-1.0", attribution: "Open Food Facts contributors" },
  { code: "ciqual", name: "Ciqual", license: "Etalab-2.0", attribution: "ANSES Ciqual" },
];

const FULL = {
  cursor: 1000,
  count: 8,
  full: true,
  columns: COLUMNS,
  gone: [],
  sources: SOURCES,
  rows: [
    row(1, "Pasta di semola", "crea"),
    row(2, "Pasta di semola, cotta", "crea", { state: "cooked", parent_id: 1, carb_g: 30 }),
    row(3, "Barilla — Pasta di semola", "off", { ean: "8076800195057" }),
    row(4, "De Cecco — Pasta di semola", "off", { has_thumb: true }),
    row(5, "Tè alla pesca", "crea"),
    row(6, "Estathé — Estathé Pesca", "off"),
    row(7, "Sushi salmone", "ciqual"),
    row(8, "Caffè espresso", "crea"),
  ],
};

function mockFetch(responses: Record<string, unknown>) {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    const key = Object.keys(responses).find((k) => url.includes(k));
    if (!key) return new Response("not found", { status: 404 });
    return new Response(JSON.stringify(responses[key]), {
      status: 200,
      headers: { "Content-Type": "application/json", ETag: '"cat-1"' },
    });
  });
}

describe("offline catalog", () => {
  beforeEach(async () => {
    await db.foods.clear();
    await db.meta.clear();
    vi.stubGlobal("fetch", mockFetch({ "/api/v1/catalog/state": { cursor: 1000, count: 8 }, "/api/v1/catalog": FULL }));
    await syncCatalog({ force: true });
  });
  afterEach(() => vi.unstubAllGlobals());

  it("stores the whole catalog", async () => {
    expect(await catalogReady()).toBe(true);
    expect(await db.foods.where("cat").equals(1).count()).toBe(8);
  });

  it("finds words in any order, accents ignored", async () => {
    const ids = (await localSearch("semola pasta")).map((r) => r.id);
    expect(ids).toContain(1);
    expect((await localSearch("caffe")).map((r) => r.id)).toEqual([8]);
  });

  it("puts CREA first and collapses same-name OFF products", async () => {
    const ids = (await localSearch("pasta")).map((r) => r.id);
    expect(ids[0]).toBe(1);
    // Barilla and De Cecco are the same OFF product name: one row, the one with a photo.
    expect(ids).toContain(4);
    expect(ids).not.toContain(3);
    expect((await db.foods.get(4))?.thumb_url).toBe("/static/thumbs/4");
  });

  it("treats te/the/tea as the same word and drops stopwords", async () => {
    const ids = (await localSearch("the alla pesca")).map((r) => r.id);
    expect(ids).toContain(5);
    expect((await localSearch("tea pesca")).map((r) => r.id)).toContain(5);
  });

  it("filters by source", async () => {
    const ids = (await localSearch("pasta", { source: "off" })).map((r) => r.id);
    expect(ids).toEqual([4]);
  });

  it("builds the food detail with source and crudo/cotto siblings", async () => {
    const d = await localFoodDetail(2);
    expect(d?.attribution).toBe("CREA Alimenti e Nutrizione");
    expect(d?.siblings.map((s) => s.id)).toEqual([1]);
    const raw = await localFoodDetail(1);
    expect(raw?.siblings.map((s) => s.id)).toEqual([2]);
  });

  it("finds a product by barcode, with or without leading zeros", async () => {
    expect((await findFoodByEan("8076800195057"))?.id).toBe(3);
    expect(await findFoodByEan("0000000000000")).toBeUndefined();
  });

  it("applies a delta: changed rows and removed ids", async () => {
    const delta = {
      ...FULL,
      cursor: 2000,
      count: 8,
      full: false,
      gone: [7],
      rows: [row(9, "Riso basmati", "crea"), row(8, "Caffè espresso ristretto", "crea")],
    };
    vi.stubGlobal(
      "fetch",
      mockFetch({ "/api/v1/catalog/state": { cursor: 2000, count: 8 }, "/api/v1/catalog?since=1000&v=2000": delta }),
    );
    await syncCatalog({ force: true });
    expect(await db.foods.get(7)).toBeUndefined();
    expect((await db.foods.get(8))?.name_it).toBe("Caffè espresso ristretto");
    expect((await localSearch("basmati")).map((r) => r.id)).toEqual([9]);
  });

  it("asks only for the version when nothing changed", async () => {
    const f = mockFetch({ "/api/v1/catalog/state": { cursor: 1000, count: 8 }, "/api/v1/catalog": FULL });
    vi.stubGlobal("fetch", f);
    await syncCatalog({ force: true });
    expect(f.mock.calls.map((c) => String(c[0]))).toEqual(["/api/v1/catalog/state"]);
  });

  it("keeps the user's own foods on a full refresh and forgets them at logout", async () => {
    await db.foods.put({
      id: 100, name_it: "Pane di mia nonna", state: "raw", source_code: "manual",
      source_license: "", search_tokens: "pane di mia nonna", updated_at: "", mine: 1,
    });
    await db.meta.delete("catalog.cursor"); // next sync downloads everything again
    await syncCatalog({ force: true });
    expect(await db.foods.get(100)).toBeDefined();
    expect((await localSearch("nonna")).map((r) => r.id)).toEqual([100]);
    await clearMyFoods();
    expect(await db.foods.get(100)).toBeUndefined();
    expect(await localSearch("nonna")).toEqual([]);
  });
});
