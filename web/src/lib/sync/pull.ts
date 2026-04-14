import { db } from "../db/schema";
import type {
  LocalFavorite,
  LocalMeal,
  LocalMealItem,
  LocalMealTemplate,
  LocalRecipe,
  LocalRecipeItem,
  PullResponse,
} from "./types";

let inflight: Promise<void> | null = null;

/**
 * Pull delta since last cursor, applying server rows directly to Dexie
 * (bypassing outbox so we don't re-send what we just received). Single-flight.
 * Pages via the cursor until the server stops advancing or we hit safety cap.
 */
export function pullDelta(): Promise<void> {
  if (inflight) return inflight;
  inflight = doPull().finally(() => {
    inflight = null;
  });
  return inflight;
}

async function doPull(): Promise<void> {
  let cursor = Number((await db.meta.get("sync_cursor"))?.value ?? 0);
  for (let i = 0; i < 20; i++) {
    let resp: Response;
    try {
      resp = await fetch(`/api/v1/sync?since=${cursor}`, { credentials: "same-origin" });
    } catch {
      return;
    }
    if (!resp.ok) return;
    const body = (await resp.json()) as PullResponse;

    const rowCount =
      (body.meals?.length ?? 0) +
      (body.meal_items?.length ?? 0) +
      (body.recipes?.length ?? 0) +
      (body.recipe_items?.length ?? 0) +
      (body.meal_templates?.length ?? 0) +
      (body.favorites?.length ?? 0);

    if (rowCount === 0) break;

    await db.transaction(
      "rw",
      [
        db.meals,
        db.meal_items,
        db.recipes,
        db.recipe_items,
        db.meal_templates,
        db.favorites,
        db.meta,
      ],
      async () => {
        if (body.meals) {
          const rows: LocalMeal[] = body.meals.map((m) => ({
            client_id: m.client_id,
            user_id: 0, // resolved at write; server scoped already
            occurred_at: Date.parse(m.occurred_at),
            slot: m.slot,
            notes: m.notes ?? undefined,
            updated_at_ms: m.updated_at_ms,
            deleted_at_ms: m.deleted_at_ms ?? null,
          }));
          await db.meals.bulkPut(rows);
        }
        if (body.meal_items) {
          const rows: LocalMealItem[] = body.meal_items.map((mi) => ({
            client_id: mi.client_id,
            meal_client_id: mi.meal_client_id,
            food_id: mi.food_id ?? null,
            recipe_client_id: mi.recipe_client_id ?? null,
            name_snapshot: mi.name_snapshot,
            grams: mi.grams,
            kcal: mi.kcal ?? null,
            protein_g: mi.protein_g ?? null,
            fat_g: mi.fat_g ?? null,
            carb_g: mi.carb_g ?? null,
            sugar_g: mi.sugar_g ?? null,
            fiber_g: mi.fiber_g ?? null,
            updated_at_ms: mi.updated_at_ms,
            deleted_at_ms: mi.deleted_at_ms ?? null,
          }));
          await db.meal_items.bulkPut(rows);
        }
        if (body.recipes) {
          const rows: LocalRecipe[] = body.recipes.map((r) => ({
            client_id: r.client_id,
            user_id: 0,
            name: r.name,
            servings: r.servings,
            total_grams: r.total_grams ?? null,
            notes: r.notes ?? null,
            updated_at_ms: r.updated_at_ms,
            deleted_at_ms: r.deleted_at_ms ?? null,
          }));
          await db.recipes.bulkPut(rows);
        }
        if (body.recipe_items) {
          const rows: LocalRecipeItem[] = body.recipe_items.map((ri) => ({
            client_id: ri.client_id,
            recipe_client_id: ri.recipe_client_id,
            food_id: ri.food_id ?? null,
            name_snapshot: ri.name_snapshot,
            grams: ri.grams,
            snapshot_json: ri.snapshot_json,
            updated_at_ms: ri.updated_at_ms,
            deleted_at_ms: ri.deleted_at_ms ?? null,
          }));
          await db.recipe_items.bulkPut(rows);
        }
        if (body.meal_templates) {
          const rows: LocalMealTemplate[] = body.meal_templates.map((t) => ({
            client_id: t.client_id,
            user_id: 0,
            name: t.name,
            payload_json: JSON.stringify(t.payload_json),
            updated_at_ms: t.updated_at_ms,
            deleted_at_ms: t.deleted_at_ms ?? null,
          }));
          await db.meal_templates.bulkPut(rows);
        }
        if (body.favorites) {
          const rows: LocalFavorite[] = body.favorites.map((f) => ({
            client_id: f.client_id,
            user_id: 0,
            food_id: f.food_id,
            updated_at_ms: f.updated_at_ms,
            deleted_at_ms: f.deleted_at_ms ?? null,
          }));
          await db.favorites.bulkPut(rows);
        }
        if (body.server_cursor > cursor) {
          await db.meta.put({ key: "sync_cursor", value: String(body.server_cursor) });
        }
      },
    );

    if (body.server_cursor <= cursor) break;
    cursor = body.server_cursor;
  }
}
