import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useLiveQuery } from "dexie-react-hooks";
import { db } from "../../lib/db/schema";
import { enqueue } from "../../lib/sync/outbox";
import { flushOutbox } from "../../lib/sync/worker";
import { aggregateRecipe } from "../../lib/recipes/aggregate";
import type {
  LocalRecipe,
  LocalRecipeItem,
  RecipeDTO,
  RecipeItemDTO,
  RecipeItemSnapshot,
} from "../../lib/sync/types";
import FoodPicker, { type FoodPickerPick } from "../food/FoodPicker";
import WeightInput from "../WeightInput";
import ConfirmDialog from "../ConfirmDialog";

export interface EditorItem {
  clientId: string;
  foodId?: number | null;
  name: string;
  grams: number;
  per100Snapshot: RecipeItemSnapshot;
}

export interface RecipeEditorProps {
  clientId?: string;
  userId: number;
}

function fmt(n: number | undefined | null, digits = 1): string {
  if (n == null || !Number.isFinite(n)) return "—";
  return n.toFixed(digits).replace(".", ",");
}

export default function RecipeEditor({ clientId, userId }: RecipeEditorProps) {
  const navigate = useNavigate();

  const existing = useLiveQuery(
    async () => (clientId ? db.recipes.get(clientId) : undefined),
    [clientId],
  );
  const existingItems = useLiveQuery(
    async () =>
      clientId
        ? db.recipe_items.where("recipe_client_id").equals(clientId).toArray()
        : [],
    [clientId],
  );

  const [name, setName] = useState("");
  const [items, setItems] = useState<EditorItem[]>([]);
  const [pending, setPending] = useState<{ food: FoodPickerPick; grams: number | null } | null>(
    null,
  );

  useEffect(() => {
    if (existing) setName(existing.name);
  }, [existing]);

  useEffect(() => {
    if (existingItems && existingItems.length > 0 && items.length === 0) {
      setItems(
        existingItems
          .filter((i) => !i.deleted_at_ms)
          .map((i) => ({
            clientId: i.client_id,
            foodId: i.food_id ?? null,
            name: i.name_snapshot,
            grams: i.grams,
            per100Snapshot: i.snapshot_json,
          })),
      );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [existingItems]);

  const sumGrams = items.reduce((s, i) => s + i.grams, 0);
  const aggregate = useMemo(
    () =>
      aggregateRecipe(
        items.map((i) => ({ grams: i.grams, snapshot: i.per100Snapshot })),
        1,
      ),
    [items],
  );

  function onFoodPicked(food: FoodPickerPick) {
    setPending({ food, grams: 100 });
  }

  function confirmPending() {
    if (!pending || !(pending.grams && pending.grams > 0)) return;
    const { food, grams } = pending;
    const snap: RecipeItemSnapshot = {
      kcal: food.kcal,
      carb_g: food.carb_g,
      protein_g: food.protein_g,
      fat_g: food.fat_g,
      sugar_g: food.sugar_g,
      fiber_g: food.fiber_g,
      source_food_name: food.name_it,
    };
    setItems((prev) => [
      ...prev,
      {
        clientId: crypto.randomUUID(),
        foodId: food.id,
        name: food.name_it,
        grams: grams!,
        per100Snapshot: snap,
      },
    ]);
    setPending(null);
  }

  function removeItem(id: string) {
    setItems((prev) => prev.filter((i) => i.clientId !== id));
  }

  function updateItemGrams(id: string, grams: number | null) {
    if (grams == null || grams <= 0) return;
    setItems((prev) => prev.map((i) => (i.clientId === id ? { ...i, grams } : i)));
  }

  function scaleAll(newTotal: number) {
    if (!(newTotal > 0) || sumGrams <= 0) return;
    const factor = newTotal / sumGrams;
    setItems((prev) => prev.map((i) => ({ ...i, grams: Math.round(i.grams * factor * 10) / 10 })));
  }

  const [scaleTarget, setScaleTarget] = useState<number | null>(null);

  async function onSave() {
    if (!name.trim()) return;
    const rcid = clientId ?? crypto.randomUUID();
    const now = Date.now();
    const recipeRow: LocalRecipe = {
      client_id: rcid,
      user_id: userId,
      name: name.trim(),
      servings: 1,
      total_grams: sumGrams > 0 ? sumGrams : null,
      notes: null,
      updated_at_ms: now,
      deleted_at_ms: null,
    };
    const recipeDto: RecipeDTO = {
      client_id: rcid,
      updated_at_ms: now,
      deleted_at_ms: null,
      name: recipeRow.name,
      servings: 1,
      total_grams: sumGrams > 0 ? sumGrams : null,
      notes: null,
    };
    await db.transaction("rw", db.recipes, db.recipe_items, db.outbox, async (tx) => {
      await tx.table("recipes").put(recipeRow);
      await enqueue("recipe", rcid, recipeDto, tx);
      for (const item of items) {
        const row: LocalRecipeItem = {
          client_id: item.clientId,
          recipe_client_id: rcid,
          food_id: item.foodId ?? null,
          name_snapshot: item.name,
          grams: item.grams,
          snapshot_json: item.per100Snapshot,
          updated_at_ms: now,
          deleted_at_ms: null,
        };
        await tx.table("recipe_items").put(row);
        const dto: RecipeItemDTO = {
          client_id: item.clientId,
          updated_at_ms: now,
          deleted_at_ms: null,
          recipe_client_id: rcid,
          food_id: item.foodId ?? null,
          name_snapshot: item.name,
          grams: item.grams,
          snapshot_json: item.per100Snapshot,
        };
        await enqueue("recipe_item", item.clientId, dto, tx);
      }
    });
    void flushOutbox();
    navigate("/ricette");
  }

  const [confirmDelete, setConfirmDelete] = useState(false);

  async function onDelete() {
    if (!clientId || !existing) return;
    setConfirmDelete(false);
    const now = Date.now();
    await db.transaction("rw", db.recipes, db.outbox, async (tx) => {
      await tx
        .table("recipes")
        .put({ ...existing, deleted_at_ms: now, updated_at_ms: now });
      await enqueue(
        "recipe",
        clientId,
        {
          client_id: clientId,
          updated_at_ms: now,
          deleted_at_ms: now,
          name: existing.name,
          servings: existing.servings,
        } satisfies RecipeDTO,
        tx,
      );
    });
    void flushOutbox();
    navigate("/ricette");
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void onSave();
      }}
      className="space-y-5 max-w-xl mx-auto"
    >
      <label className="block">
        <span className="block text-sm font-medium mb-1">Nome ricetta</span>
        <input
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="es. Pasta al pomodoro della nonna"
          className="w-full border border-[color:var(--color-border)] rounded-md px-3 py-2 text-base"
        />
      </label>

      <section className="space-y-2">
        <div className="flex items-baseline justify-between">
          <h2 className="text-base font-semibold">Ingredienti</h2>
          <span className="text-xs text-[color:var(--color-ink-muted)]">
            {items.length} · totale {sumGrams} g
          </span>
        </div>

        {items.length === 0 ? (
          <p className="text-sm text-[color:var(--color-ink-muted)] py-2">
            Nessun ingrediente. Cerca un alimento qui sotto per aggiungerlo.
          </p>
        ) : (
          <ul className="divide-y divide-[color:var(--color-border)] border border-[color:var(--color-border)] rounded-md bg-white">
            {items.map((i) => (
              <li key={i.clientId} className="p-2 flex items-center gap-2">
                <span className="flex-1 min-w-0 truncate text-sm" title={i.name}>
                  {i.name}
                </span>
                <input
                  type="text"
                  inputMode="decimal"
                  aria-label={`Grammi di ${i.name}`}
                  value={i.grams}
                  onChange={(e) => {
                    const v = e.target.value.replace(",", ".");
                    const n = Number(v);
                    if (Number.isFinite(n) && n > 0) updateItemGrams(i.clientId, n);
                  }}
                  className="w-20 shrink-0 border border-[color:var(--color-border)] rounded-md px-2 py-1 text-sm text-right tabular-nums"
                />
                <span className="shrink-0 text-xs text-[color:var(--color-ink-muted)]">g</span>
                <button
                  type="button"
                  onClick={() => removeItem(i.clientId)}
                  aria-label={`Rimuovi ${i.name}`}
                  className="shrink-0 text-red-700 text-lg leading-none px-1"
                >
                  ×
                </button>
              </li>
            ))}
          </ul>
        )}

        {pending ? (
          <div className="flex items-center gap-2 p-2 border border-[color:var(--color-brand-500)] rounded-md bg-[color:var(--color-brand-50)]">
            <span className="flex-1 min-w-0 truncate text-sm" title={pending.food.name_it}>
              {pending.food.name_it}
            </span>
            <div className="shrink-0">
              <WeightInput
                value={pending.grams}
                onChange={(n) => setPending((p) => (p ? { ...p, grams: n } : p))}
                max={5000}
                autoFocus
              />
            </div>
            <button
              type="button"
              onClick={confirmPending}
              className="shrink-0 px-3 py-1.5 text-sm bg-[color:var(--color-brand-600)] text-white rounded-md"
            >
              Aggiungi
            </button>
            <button
              type="button"
              onClick={() => setPending(null)}
              className="shrink-0 px-2 py-1.5 text-sm"
            >
              Annulla
            </button>
          </div>
        ) : (
          <FoodPicker onPick={onFoodPicked} placeholder="Cerca un alimento…" />
        )}
      </section>

      {items.length > 0 ? (
        <section className="space-y-2">
          <h2 className="text-base font-semibold">Peso totale</h2>
          <div className="flex items-center gap-2 text-sm">
            <span className="font-semibold tabular-nums">{sumGrams} g</span>
            <span className="text-[color:var(--color-ink-muted)]">— scala a:</span>
            <input
              type="text"
              inputMode="decimal"
              placeholder="es. 1000"
              value={scaleTarget ?? ""}
              onChange={(e) => {
                const v = e.target.value.replace(",", ".");
                const n = Number(v);
                setScaleTarget(v === "" ? null : Number.isFinite(n) ? n : scaleTarget);
              }}
              className="w-24 border border-[color:var(--color-border)] rounded-md px-2 py-1 text-sm text-right tabular-nums"
            />
            <span className="text-[color:var(--color-ink-muted)]">g</span>
            <button
              type="button"
              onClick={() => {
                if (scaleTarget && scaleTarget > 0) {
                  scaleAll(scaleTarget);
                  setScaleTarget(null);
                }
              }}
              disabled={!scaleTarget || scaleTarget <= 0}
              className="px-2 py-1 text-xs border border-[color:var(--color-border)] rounded-md disabled:opacity-50"
            >
              Applica
            </button>
          </div>
          <p className="text-xs text-[color:var(--color-ink-muted)]">
            Gli ingredienti verranno ridimensionati proporzionalmente.
          </p>
        </section>
      ) : null}

      {items.length > 0 ? (
        <section className="space-y-1 rounded-md border border-[color:var(--color-border)] bg-white p-3">
          <h2 className="text-base font-semibold">Valori per 100 g della ricetta</h2>
          <dl className="grid grid-cols-4 gap-2 text-sm mt-2">
            <Kv label="Carb" value={`${fmt(aggregate.per100g.carb_g)} g`} emphasise />
            <Kv label="Kcal" value={fmt(aggregate.per100g.kcal, 0)} />
            <Kv label="Prot" value={`${fmt(aggregate.per100g.protein_g)} g`} />
            <Kv label="Grassi" value={`${fmt(aggregate.per100g.fat_g)} g`} />
          </dl>
        </section>
      ) : null}

      <div className="flex justify-between items-center pt-2">
        {clientId ? (
          <button
            type="button"
            onClick={() => setConfirmDelete(true)}
            className="px-3 py-2 text-sm text-red-700 border border-red-200 rounded-md"
          >
            Elimina
          </button>
        ) : (
          <span />
        )}
        <button
          type="submit"
          disabled={!name.trim() || items.length === 0}
          className="px-4 py-2 bg-[color:var(--color-brand-600)] text-white rounded-md font-medium disabled:opacity-50"
        >
          Salva ricetta
        </button>
      </div>
      <ConfirmDialog
        open={confirmDelete}
        title="Eliminare la ricetta?"
        message="Questa azione è irreversibile."
        confirmLabel="Elimina"
        variant="danger"
        onConfirm={onDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </form>
  );
}

function Kv({
  label,
  value,
  emphasise,
}: {
  label: string;
  value: string;
  emphasise?: boolean;
}) {
  return (
    <div>
      <dt className="text-xs text-[color:var(--color-ink-muted)]">{label}</dt>
      <dd className={emphasise ? "font-bold text-base tabular-nums" : "tabular-nums"}>{value}</dd>
    </div>
  );
}
