import { useLiveQuery } from "dexie-react-hooks";
import { db } from "../db/schema";
import { enqueue } from "../sync/outbox";
import { flushOutbox } from "../sync/worker";
import { toSqlTimestamp } from "../db/queries";
import type {
  LocalMeal,
  LocalMealItem,
  LocalMealTemplate,
  MealDTO,
  MealItemDTO,
  MealTemplateDTO,
  Slot,
} from "../sync/types";
import type { Nutrients } from "../diary/totals";

export interface TemplateItem {
  food_id?: number | null;
  recipe_client_id?: string | null;
  name_snapshot: string;
  grams: number;
  snapshot: Partial<Nutrients> & { source_food_name?: string };
}

export async function saveTemplate(
  name: string,
  items: TemplateItem[],
  userId: number,
): Promise<string> {
  const clientId = crypto.randomUUID();
  const now = Date.now();
  const payloadJson = JSON.stringify(items);
  const row: LocalMealTemplate = {
    client_id: clientId,
    user_id: userId,
    name,
    payload_json: payloadJson,
    updated_at_ms: now,
    deleted_at_ms: null,
  };
  await db.transaction("rw", db.meal_templates, db.outbox, async (tx) => {
    await tx.table("meal_templates").put(row);
    const dto: MealTemplateDTO = {
      client_id: clientId,
      updated_at_ms: now,
      deleted_at_ms: null,
      name,
      payload_json: items,
    };
    await enqueue("meal_template", clientId, dto, tx);
  });
  void flushOutbox();
  return clientId;
}

export async function saveTemplateFromMeal(
  mealClientId: string,
  name: string,
  userId: number,
): Promise<string> {
  const items = await db.meal_items
    .where("meal_client_id")
    .equals(mealClientId)
    .and((i) => !i.deleted_at_ms)
    .toArray();
  const proto: TemplateItem[] = items.map((i) => ({
    food_id: i.food_id ?? null,
    recipe_client_id: i.recipe_client_id ?? null,
    name_snapshot: i.name_snapshot,
    grams: i.grams,
    snapshot: {
      kcal: i.kcal ?? undefined,
      carb_g: i.carb_g ?? undefined,
      protein_g: i.protein_g ?? undefined,
      fat_g: i.fat_g ?? undefined,
      sugar_g: i.sugar_g ?? undefined,
      fiber_g: i.fiber_g ?? undefined,
      source_food_name: i.name_snapshot,
    },
  }));
  return saveTemplate(name, proto, userId);
}

export async function instantiateTemplate(
  templateClientId: string,
  targetIsoDate: string,
  slot: Slot,
  userId: number,
): Promise<string> {
  const tpl = await db.meal_templates.get(templateClientId);
  if (!tpl || tpl.deleted_at_ms) throw new Error("template_not_found");
  const items = JSON.parse(tpl.payload_json) as TemplateItem[];
  const mealClientId = crypto.randomUUID();
  const occurredAt = noonRomeMs(targetIsoDate);
  const now = Date.now();

  const mealRow: LocalMeal = {
    client_id: mealClientId,
    user_id: userId,
    occurred_at: occurredAt,
    slot,
    notes: undefined,
    updated_at_ms: now,
    deleted_at_ms: null,
  };

  await db.transaction(
    "rw",
    db.meals,
    db.meal_items,
    db.outbox,
    async (tx) => {
      await tx.table("meals").put(mealRow);
      const mealDto: MealDTO = {
        client_id: mealClientId,
        updated_at_ms: now,
        deleted_at_ms: null,
        occurred_at: toSqlTimestamp(occurredAt),
        slot,
        notes: null,
      };
      await enqueue("meal", mealClientId, mealDto, tx);

      for (const p of items) {
        const itemClientId = crypto.randomUUID();
        const itemRow: LocalMealItem = {
          client_id: itemClientId,
          meal_client_id: mealClientId,
          food_id: p.food_id ?? null,
          recipe_client_id: p.recipe_client_id ?? null,
          name_snapshot: p.name_snapshot,
          grams: p.grams,
          kcal: p.snapshot.kcal ?? null,
          protein_g: p.snapshot.protein_g ?? null,
          fat_g: p.snapshot.fat_g ?? null,
          carb_g: p.snapshot.carb_g ?? null,
          sugar_g: p.snapshot.sugar_g ?? null,
          fiber_g: p.snapshot.fiber_g ?? null,
          updated_at_ms: now,
          deleted_at_ms: null,
        };
        await tx.table("meal_items").put(itemRow);
        const itemDto: MealItemDTO = {
          client_id: itemClientId,
          updated_at_ms: now,
          deleted_at_ms: null,
          meal_client_id: mealClientId,
          food_id: p.food_id ?? null,
          recipe_client_id: p.recipe_client_id ?? null,
          name_snapshot: p.name_snapshot,
          grams: p.grams,
          kcal: p.snapshot.kcal ?? null,
          protein_g: p.snapshot.protein_g ?? null,
          fat_g: p.snapshot.fat_g ?? null,
          carb_g: p.snapshot.carb_g ?? null,
          sugar_g: p.snapshot.sugar_g ?? null,
          fiber_g: p.snapshot.fiber_g ?? null,
        };
        await enqueue("meal_item", itemClientId, itemDto, tx);
      }
    },
  );
  void flushOutbox();
  return mealClientId;
}

export async function deleteTemplate(clientId: string): Promise<void> {
  const now = Date.now();
  const tpl = await db.meal_templates.get(clientId);
  if (!tpl) return;
  const tombstoned: LocalMealTemplate = { ...tpl, deleted_at_ms: now, updated_at_ms: now };
  await db.transaction("rw", db.meal_templates, db.outbox, async (tx) => {
    await tx.table("meal_templates").put(tombstoned);
    const dto: MealTemplateDTO = {
      client_id: clientId,
      updated_at_ms: now,
      deleted_at_ms: now,
      name: tpl.name,
      payload_json: JSON.parse(tpl.payload_json),
    };
    await enqueue("meal_template", clientId, dto, tx);
  });
  void flushOutbox();
}

export function useTemplates(userId: number) {
  return useLiveQuery(
    () =>
      db.meal_templates
        .where("user_id")
        .equals(userId)
        .and((t) => !t.deleted_at_ms)
        .toArray(),
    [userId],
  );
}

function noonRomeMs(isoDate: string): number {
  const [y, m, d] = isoDate.split("-").map(Number);
  return Date.UTC(y, m - 1, d, 12, 0, 0);
}
