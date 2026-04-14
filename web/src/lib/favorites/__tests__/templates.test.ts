import "fake-indexeddb/auto";
import { beforeEach, describe, expect, it } from "vitest";
import { db } from "../../db/schema";
import { instantiateTemplate, saveTemplate } from "../templates";

beforeEach(async () => {
  await db.meal_templates.clear();
  await db.meals.clear();
  await db.meal_items.clear();
  await db.outbox.clear();
});

describe("saveTemplate + instantiateTemplate", () => {
  it("saveTemplate writes a template row and an outbox entry", async () => {
    const id = await saveTemplate(
      "Colazione classica",
      [{ name_snapshot: "pane", grams: 80, snapshot: { carb_g: 50 } }],
      1,
    );
    const row = await db.meal_templates.get(id);
    expect(row?.name).toBe("Colazione classica");
    expect(await db.outbox.where("kind").equals("meal_template").count()).toBe(1);
  });

  it("instantiateTemplate creates a new meal with new item client_ids", async () => {
    const tplId = await saveTemplate(
      "t1",
      [
        { name_snapshot: "pane", grams: 80, snapshot: { carb_g: 50 } },
        { name_snapshot: "marmellata", grams: 20, snapshot: { carb_g: 60 } },
      ],
      1,
    );
    const mealId = await instantiateTemplate(tplId, "2026-04-13", "colazione", 1);

    const meal = await db.meals.get(mealId);
    expect(meal?.slot).toBe("colazione");

    const items = await db.meal_items.where("meal_client_id").equals(mealId).toArray();
    expect(items).toHaveLength(2);
    // New client_ids must not collide with any existing template payload identifier.
    const ids = new Set(items.map((i) => i.client_id));
    expect(ids.size).toBe(2);
    expect(items.map((i) => i.name_snapshot).sort()).toEqual(["marmellata", "pane"]);
  });

  it("instantiateTemplate on a tombstoned template throws", async () => {
    const tplId = await saveTemplate("t2", [], 1);
    const existing = await db.meal_templates.get(tplId);
    await db.meal_templates.put({ ...existing!, deleted_at_ms: Date.now() });
    await expect(instantiateTemplate(tplId, "2026-04-13", "pranzo", 1)).rejects.toThrow(
      /template_not_found/,
    );
  });

  // TODO: round-trip save-from-meal → delete meal → instantiate → items intact
});
