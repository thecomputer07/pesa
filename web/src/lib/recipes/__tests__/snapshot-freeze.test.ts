import "fake-indexeddb/auto";
import { beforeEach, describe, expect, it } from "vitest";
import { db } from "../../db/schema";
import { computeItemNutrients } from "../../diary/totals";

// Proves the snapshot-freeze invariant end-to-end at the data layer:
// once a recipe_item is written with snapshot_json, later updates to the
// source food row MUST NOT change the stored snapshot.

beforeEach(async () => {
  await db.foods.clear();
  await db.recipe_items.clear();
});

describe("recipe snapshot freeze", () => {
  it("does not change when source food updates", async () => {
    await db.foods.put({
      id: 1,
      name_it: "pasta",
      state: "raw",
      carb_g: 50,
      source_code: "CREA",
      source_license: "CC-BY",
      search_tokens: "pasta",
      updated_at: new Date().toISOString(),
    });

    await db.recipe_items.put({
      client_id: "00000000-0000-7000-8000-000000000001",
      recipe_client_id: "00000000-0000-7000-8000-00000000000a",
      food_id: 1,
      name_snapshot: "pasta",
      grams: 100,
      snapshot_json: { carb_g: 50, source_food_name: "pasta" },
      updated_at_ms: 1,
      deleted_at_ms: null,
    });

    // Simulate a food cache refresh bumping the per-100g carb value.
    await db.foods.update(1, { carb_g: 60 });

    const stored = await db.recipe_items.get("00000000-0000-7000-8000-000000000001");
    expect(stored!.snapshot_json.carb_g).toBe(50); // frozen

    // Historical consumption math uses the snapshot, not live food:
    const n = computeItemNutrients(100, { carb_g: stored!.snapshot_json.carb_g });
    expect(n.carb_g).toBeCloseTo(50, 5);
  });
});
