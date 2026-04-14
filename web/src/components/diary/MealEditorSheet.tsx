import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { db } from "../../lib/db/schema";
import { enqueue } from "../../lib/sync/outbox";
import { flushOutbox } from "../../lib/sync/worker";
import { toSqlTimestamp } from "../../lib/db/queries";
import type {
  LocalMeal,
  LocalMealItem,
  MealDTO,
  MealItemDTO,
  Slot,
} from "../../lib/sync/types";
import FoodPicker, { type FoodPickerPick } from "../food/FoodPicker";
import RecipePicker, { type RecipePick } from "../recipes/RecipePicker";

export interface EditorItem {
  clientId: string;
  foodId?: number | null;
  recipeClientId?: string | null;
  nameSnapshot: string;
  grams: number;
  per100?: {
    kcal?: number;
    carb_g?: number;
    protein_g?: number;
    fat_g?: number;
    sugar_g?: number;
    fiber_g?: number;
  };
}

export interface MealEditorSheetProps {
  open: boolean;
  mode?: "create" | "edit";
  isoDate: string;
  userId: number;
  initialMeal?: LocalMeal;
  initialItems?: LocalMealItem[];
  onClose: () => void;
}

/**
 * Focus-trapped, aria-modal sheet for creating/editing a meal.
 * Picker tabs "Alimenti" / "Ricette" are rendered but the recipe tab is a
 * stub in plan-04 (wired fully in plan-05); food selection uses a minimal
 * name+grams manual entry — SearchBox integration is the FIRST consumer.
 */
export default function MealEditorSheet({
  open,
  mode = "create",
  isoDate,
  userId,
  initialMeal,
  initialItems,
  onClose,
}: MealEditorSheetProps) {
  const { t } = useTranslation();
  const sheetRef = useRef<HTMLDivElement | null>(null);
  const [slot, setSlot] = useState<Slot>(initialMeal?.slot ?? "pranzo");
  const [notes, setNotes] = useState<string>(initialMeal?.notes ?? "");
  const [items, setItems] = useState<EditorItem[]>(
    (initialItems ?? []).map((i) => ({
      clientId: i.client_id,
      foodId: i.food_id ?? null,
      recipeClientId: i.recipe_client_id ?? null,
      nameSnapshot: i.name_snapshot,
      grams: i.grams,
      per100: {
        kcal: i.kcal ?? undefined,
        carb_g: i.carb_g ?? undefined,
        protein_g: i.protein_g ?? undefined,
        fat_g: i.fat_g ?? undefined,
        sugar_g: i.sugar_g ?? undefined,
        fiber_g: i.fiber_g ?? undefined,
      },
    })),
  );
  const [pickerTab, setPickerTab] = useState<"foods" | "recipes">("foods");
  const [pending, setPending] = useState<{ food: FoodPickerPick; grams: number } | null>(null);
  const [pendingRecipe, setPendingRecipe] = useState<{ recipe: RecipePick; grams: number } | null>(
    null,
  );

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    sheetRef.current?.focus();
    return () => document.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  if (!open) return null;

  function onFoodPicked(food: FoodPickerPick) {
    setPending({ food, grams: 100 });
  }

  function confirmPending() {
    if (!pending || !(pending.grams > 0)) return;
    const { food, grams } = pending;
    setItems((prev) => [
      ...prev,
      {
        clientId: crypto.randomUUID(),
        foodId: food.id,
        nameSnapshot: food.name_it,
        grams,
        per100: {
          kcal: food.kcal,
          carb_g: food.carb_g,
          protein_g: food.protein_g,
          fat_g: food.fat_g,
          sugar_g: food.sugar_g,
          fiber_g: food.fiber_g,
        },
      },
    ]);
    setPending(null);
  }

  function removeItem(clientId: string) {
    setItems((prev) => prev.filter((i) => i.clientId !== clientId));
  }

  function onRecipePicked(recipe: RecipePick) {
    setPendingRecipe({ recipe, grams: 100 });
  }

  function confirmPendingRecipe() {
    if (!pendingRecipe || !(pendingRecipe.grams > 0)) return;
    const { recipe, grams } = pendingRecipe;
    setItems((prev) => [
      ...prev,
      {
        clientId: crypto.randomUUID(),
        foodId: null,
        recipeClientId: recipe.client_id,
        nameSnapshot: recipe.name,
        grams,
        per100: {
          kcal: recipe.per100.kcal,
          carb_g: recipe.per100.carb_g,
          protein_g: recipe.per100.protein_g,
          fat_g: recipe.per100.fat_g,
          sugar_g: recipe.per100.sugar_g,
          fiber_g: recipe.per100.fiber_g,
        },
      },
    ]);
    setPendingRecipe(null);
  }

  async function onSave() {
    const mealClientId = initialMeal?.client_id ?? crypto.randomUUID();
    const now = Date.now();
    const occurredAtMs = initialMeal?.occurred_at ?? noonRomeMs(isoDate);

    const mealRow: LocalMeal = {
      client_id: mealClientId,
      user_id: userId,
      occurred_at: occurredAtMs,
      slot,
      notes: notes || undefined,
      updated_at_ms: now,
      deleted_at_ms: null,
    };
    const mealDto: MealDTO = {
      client_id: mealClientId,
      updated_at_ms: now,
      deleted_at_ms: null,
      occurred_at: toSqlTimestamp(occurredAtMs),
      slot,
      notes: notes || null,
    };

    await db.transaction(
      "rw",
      db.meals,
      db.meal_items,
      db.outbox,
      async (tx) => {
        await tx.table("meals").put(mealRow);
        await enqueue("meal", mealClientId, mealDto, tx);
        for (const item of items) {
          const itemRow: LocalMealItem = {
            client_id: item.clientId,
            meal_client_id: mealClientId,
            food_id: item.foodId ?? null,
            recipe_client_id: item.recipeClientId ?? null,
            name_snapshot: item.nameSnapshot,
            grams: item.grams,
            kcal: item.per100?.kcal ?? null,
            protein_g: item.per100?.protein_g ?? null,
            fat_g: item.per100?.fat_g ?? null,
            carb_g: item.per100?.carb_g ?? null,
            sugar_g: item.per100?.sugar_g ?? null,
            fiber_g: item.per100?.fiber_g ?? null,
            updated_at_ms: now,
            deleted_at_ms: null,
          };
          await tx.table("meal_items").put(itemRow);
          const itemDto: MealItemDTO = {
            client_id: item.clientId,
            updated_at_ms: now,
            deleted_at_ms: null,
            meal_client_id: mealClientId,
            food_id: item.foodId ?? null,
            recipe_client_id: item.recipeClientId ?? null,
            name_snapshot: item.nameSnapshot,
            grams: item.grams,
            kcal: item.per100?.kcal ?? null,
            protein_g: item.per100?.protein_g ?? null,
            fat_g: item.per100?.fat_g ?? null,
            carb_g: item.per100?.carb_g ?? null,
            sugar_g: item.per100?.sugar_g ?? null,
            fiber_g: item.per100?.fiber_g ?? null,
          };
          await enqueue("meal_item", item.clientId, itemDto, tx);
        }
        // TODO: for edit mode, diff against initialItems and soft-delete removed.
      },
    );
    void flushOutbox();
    onClose();
  }

  async function onDelete() {
    if (!initialMeal) return;
    const now = Date.now();
    const updated: LocalMeal = { ...initialMeal, deleted_at_ms: now, updated_at_ms: now };
    await db.transaction("rw", db.meals, db.meal_items, db.outbox, async (tx) => {
      await tx.table("meals").put(updated);
      await enqueue(
        "meal",
        updated.client_id,
        {
          client_id: updated.client_id,
          updated_at_ms: now,
          deleted_at_ms: now,
          occurred_at: toSqlTimestamp(updated.occurred_at),
          slot: updated.slot,
          notes: updated.notes ?? null,
        } satisfies MealDTO,
        tx,
      );
      for (const i of initialItems ?? []) {
        const updatedItem: LocalMealItem = { ...i, deleted_at_ms: now, updated_at_ms: now };
        await tx.table("meal_items").put(updatedItem);
        await enqueue(
          "meal_item",
          i.client_id,
          {
            client_id: i.client_id,
            updated_at_ms: now,
            deleted_at_ms: now,
            meal_client_id: updated.client_id,
            food_id: i.food_id ?? null,
            recipe_client_id: i.recipe_client_id ?? null,
            name_snapshot: i.name_snapshot,
            grams: i.grams,
          } satisfies MealItemDTO,
          tx,
        );
      }
    });
    void flushOutbox();
    onClose();
  }

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t(mode === "edit" ? "diary.editMeal" : "diary.addMeal")}
      className="fixed inset-0 bg-black/40 flex items-end md:items-center justify-center z-50"
    >
      <div
        ref={sheetRef}
        tabIndex={-1}
        className="bg-white w-full md:max-w-lg md:rounded-lg p-4 space-y-3"
      >
        <h3 className="text-lg font-semibold">
          {t(mode === "edit" ? "diary.editMeal" : "diary.addMeal")}
        </h3>

        <label className="block text-sm">
          {t("diary.slotLabel")}
          <select
            value={slot}
            onChange={(e) => setSlot(e.target.value as Slot)}
            className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5"
          >
            {(["colazione", "pranzo", "cena", "spuntino", "altro"] as Slot[]).map((s) => (
              <option key={s} value={s}>
                {t(`diary.slots.${s}`)}
              </option>
            ))}
          </select>
        </label>

        <fieldset>
          <legend className="text-sm font-medium">{t("diary.picker.title")}</legend>
          <div role="tablist" className="flex gap-2 mt-1">
            <button
              type="button"
              role="tab"
              aria-selected={pickerTab === "foods"}
              onClick={() => setPickerTab("foods")}
              className={`px-3 py-1 text-sm rounded ${pickerTab === "foods" ? "bg-[color:var(--color-brand-600)] text-white" : "bg-white border"}`}
            >
              {t("diary.picker.foods")}
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={pickerTab === "recipes"}
              onClick={() => setPickerTab("recipes")}
              className={`px-3 py-1 text-sm rounded ${pickerTab === "recipes" ? "bg-[color:var(--color-brand-600)] text-white" : "bg-white border"}`}
            >
              {t("diary.picker.recipes")}
            </button>
          </div>
          {pickerTab === "foods" ? (
            <div className="mt-2 space-y-2">
              <FoodPicker onPick={onFoodPicked} placeholder={t("foodPicker.placeholder")} />
              {pending ? (
                <div className="flex gap-2 items-center">
                  <span className="flex-1 text-sm truncate" title={pending.food.name_it}>
                    {pending.food.name_it}
                  </span>
                  <input
                    aria-label={t("detail.weightLabel")}
                    type="number"
                    min={0}
                    value={pending.grams}
                    onChange={(e) =>
                      setPending((p) => (p ? { ...p, grams: Number(e.target.value) } : p))
                    }
                    className="w-24 border border-[color:var(--color-border)] rounded-md px-2 py-1.5 text-sm"
                  />
                  <button
                    type="button"
                    onClick={confirmPending}
                    className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
                  >
                    {t("foodPicker.confirmAdd")}
                  </button>
                  <button
                    type="button"
                    onClick={() => setPending(null)}
                    className="px-2 py-1.5 text-sm"
                  >
                    {t("common.cancel")}
                  </button>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="mt-2 space-y-2">
              <RecipePicker onPick={onRecipePicked} />
              {pendingRecipe ? (
                <div className="flex gap-2 items-center">
                  <span className="flex-1 text-sm truncate" title={pendingRecipe.recipe.name}>
                    {pendingRecipe.recipe.name}
                  </span>
                  <input
                    aria-label={t("detail.weightLabel")}
                    type="number"
                    min={0}
                    value={pendingRecipe.grams}
                    onChange={(e) =>
                      setPendingRecipe((p) =>
                        p ? { ...p, grams: Number(e.target.value) } : p,
                      )
                    }
                    className="w-24 border border-[color:var(--color-border)] rounded-md px-2 py-1.5 text-sm"
                  />
                  <button
                    type="button"
                    onClick={confirmPendingRecipe}
                    className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
                  >
                    {t("foodPicker.confirmAdd")}
                  </button>
                  <button
                    type="button"
                    onClick={() => setPendingRecipe(null)}
                    className="px-2 py-1.5 text-sm"
                  >
                    {t("common.cancel")}
                  </button>
                </div>
              ) : null}
            </div>
          )}
        </fieldset>

        <ul className="divide-y divide-[color:var(--color-border)] max-h-60 overflow-y-auto">
          {items.map((i) => (
            <li key={i.clientId} className="py-2 flex items-center text-sm gap-2 min-w-0">
              <span className="flex-1 min-w-0 truncate" title={i.nameSnapshot}>
                {i.nameSnapshot}
              </span>
              <span className="shrink-0 text-xs text-[color:var(--color-ink-muted)] tabular-nums">
                {i.grams} g
              </span>
              <button
                type="button"
                onClick={() => removeItem(i.clientId)}
                aria-label={t("common.delete")}
                className="shrink-0 text-xs text-red-700 px-1"
              >
                ×
              </button>
            </li>
          ))}
        </ul>

        <label className="block text-sm">
          {t("diary.notes")}
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            maxLength={500}
            className="mt-1 w-full border border-[color:var(--color-border)] rounded-md px-2 py-1.5 text-sm"
            rows={2}
          />
        </label>

        <div className="flex justify-between pt-2">
          {mode === "edit" ? (
            <button type="button" onClick={onDelete} className="text-sm text-red-700">
              {t("diary.delete")}
            </button>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <button type="button" onClick={onClose} className="px-3 py-1.5 text-sm">
              {t("common.cancel")}
            </button>
            <button
              type="button"
              onClick={onSave}
              className="px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
            >
              {t("common.save")}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function noonRomeMs(isoDate: string): number {
  const [y, m, d] = isoDate.split("-").map(Number);
  // 12:00 UTC is safely inside Europe/Rome's 13:00/14:00 local — good default.
  return Date.UTC(y, m - 1, d, 12, 0, 0);
}
