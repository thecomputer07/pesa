import "fake-indexeddb/auto";
import { beforeEach, describe, expect, it } from "vitest";
import { db } from "../schema";
import { upsertFoods } from "../sync";
import { cachedCount, localSearch } from "../search";
import type { FoodSearchResult } from "../../api";

function mk(
  id: number,
  name_it: string,
  extra: Partial<FoodSearchResult> = {},
): FoodSearchResult {
  return {
    id,
    name_it,
    state: "crudo",
    source: { code: "CREA", license: "CC-BY-NC-SA" },
    ...extra,
  };
}

describe("local FTS search", () => {
  beforeEach(async () => {
    await db.foods.clear();
    await db.meta.clear();
    await upsertFoods([
      mk(1, "Pane comune"),
      mk(2, "Pane integrale", { category: "Cereali" }),
      mk(3, "Mela Golden", { name_en: "apple" }),
      mk(4, "Riso parboiled", { state: "cotto" }),
      mk(5, "Caffè espresso"),
    ]);
  });

  it("upserts and counts", async () => {
    expect(await cachedCount()).toBe(5);
  });

  it("prefix-matches by Italian name (accent-insensitive)", async () => {
    const rows = await localSearch("caffe");
    expect(rows.map((r) => r.id)).toEqual([5]);
  });

  it("ranks prefix match higher than substring", async () => {
    const rows = await localSearch("pane");
    expect(rows[0].id).toBeOneOf([1, 2]);
    expect(rows.map((r) => r.id)).toContain(1);
    expect(rows.map((r) => r.id)).toContain(2);
  });

  it("matches english name for fallback", async () => {
    const rows = await localSearch("apple");
    expect(rows.map((r) => r.id)).toContain(3);
  });

  it("filters by state", async () => {
    const cooked = await localSearch("riso", { state: "cotto" });
    expect(cooked.map((r) => r.id)).toEqual([4]);
    const raw = await localSearch("riso", { state: "crudo" });
    expect(raw).toHaveLength(0);
  });

  it("returns empty for a too-short / empty query", async () => {
    expect(await localSearch("")).toEqual([]);
    expect(await localSearch("   ")).toEqual([]);
  });
});

// Polyfill for Vitest's `toBeOneOf` that `@testing-library/jest-dom` doesn't ship.
expect.extend({
  toBeOneOf(received: unknown, expected: unknown[]) {
    const pass = expected.includes(received);
    return {
      pass,
      message: () =>
        pass
          ? `expected ${String(received)} not to be one of ${JSON.stringify(expected)}`
          : `expected ${String(received)} to be one of ${JSON.stringify(expected)}`,
    };
  },
});

declare module "vitest" {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface Assertion<T = unknown> {
    toBeOneOf(expected: T[]): void;
  }
}
